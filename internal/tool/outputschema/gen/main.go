// Command gen regenerates the tool output schemas under
// internal/tool/outputschema/schemas from the published OpenAPI documents of
// the Bitrise APIs the tools call.
//
// Every tool whose handler passes an API response through gets the JSON
// Schema of that response (derived from the API's OpenAPI document, or, for
// the Release Management APIs which publish examples instead of schemas,
// inferred from those examples). Tools that post-process the response list the
// properties they strip so the schema matches what the tool actually returns.
// Tools that return plain text (a bitrise.yml, a guide, a status message) get
// the text envelope the tool belt wraps such results in. Tools that build a Go
// value declare their schema in code with mcp.WithOutputSchema and are not
// listed here.
//
// Usage:
//
//	go run ./internal/tool/outputschema/gen              # fetch the hosted specs
//	go run ./internal/tool/outputschema/gen -specs DIR   # read <name>.json from DIR
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// specURLs are the hosted OpenAPI documents, one per backend.
var specURLs = map[string]string{
	"ci":      "https://api-docs.bitrise.io/docs/swagger.json",
	"rde":     "https://api.bitrise.io/rde/api-docs/swagger.json",
	"rm-apps": "https://api.bitrise.io/release-management/api-docs/release_management/v2/swagger.json",
	"rm-bd":   "https://api.bitrise.io/release-management/api-docs/release_management/v2/build_distributions/swagger.json",
	"rm-cp":   "https://api.bitrise.io/release-management/api-docs/release_management/v2/code_push/swagger.json",
}

// mapping ties a tool to the API operation whose response it returns.
type mapping struct {
	Tool   string
	Spec   string
	Method string
	Path   string
	// Drop lists properties the tool handler removes before returning, as
	// dotted paths where "[]" descends into array items, e.g. "data[].credit_cost".
	Drop []string
	// Note is appended to the schema description.
	Note string
	// Add lists properties the tool handler adds to the response, keyed by
	// the dotted path of the object they are added to ("" for the root).
	Add map[string]map[string]any
}

var mappings = []mapping{
	// Bitrise API v0.1 — user, apps, builds, artifacts, workspaces, webhooks,
	// cache, pipelines, group roles, configuration.
	{Tool: "me", Spec: "ci", Method: "get", Path: "/me"},
	{Tool: "list_apps", Spec: "ci", Method: "get", Path: "/apps"},
	{Tool: "register_app", Spec: "ci", Method: "post", Path: "/apps/register"},
	{Tool: "finish_bitrise_app", Spec: "ci", Method: "post", Path: "/apps/{app-slug}/finish"},
	{Tool: "get_app", Spec: "ci", Method: "get", Path: "/apps/{app-slug}"},
	{Tool: "delete_app", Spec: "ci", Method: "delete", Path: "/apps/{app-slug}"},
	{Tool: "update_app", Spec: "ci", Method: "patch", Path: "/apps/{app-slug}"},
	{Tool: "get_bitrise_yml", Spec: "ci", Method: "get", Path: "/apps/{app-slug}/bitrise.yml"},
	{Tool: "update_bitrise_yml", Spec: "ci", Method: "post", Path: "/apps/{app-slug}/bitrise.yml"},
	{Tool: "list_branches", Spec: "ci", Method: "get", Path: "/apps/{app-slug}/branches"},
	{Tool: "register_ssh_key", Spec: "ci", Method: "post", Path: "/apps/{app-slug}/register-ssh-key"},
	{Tool: "register_webhook", Spec: "ci", Method: "post", Path: "/apps/{app-slug}/register-webhook"},
	{Tool: "trigger_bitrise_build", Spec: "ci", Method: "post", Path: "/apps/{app-slug}/builds"},
	{Tool: "list_builds", Spec: "ci", Method: "get", Path: "/apps/{app-slug}/builds",
		Drop: []string{"data[].credit_cost", "data[].commit_view_url", "data[].environment_prepare_finished_at", "data[].is_processed", "data[].is_status_sent", "data[].log_format"},
		Note: "pull_request_id is omitted when 0; original_build_params and the full repository object are included only with verbose=true (otherwise repository is reduced to slug, title, repo_owner and repo_name, and dropped entirely when app_slug was given)."},
	{Tool: "get_build", Spec: "ci", Method: "get", Path: "/apps/{app-slug}/builds/{build-slug}",
		Drop: []string{"data.credit_cost", "data.commit_view_url", "data.environment_prepare_finished_at", "data.is_processed", "data.is_status_sent", "data.log_format"},
		Note: "pull_request_id is omitted when 0; original_build_params is included only with verbose=true."},
	{Tool: "abort_build", Spec: "ci", Method: "post", Path: "/apps/{app-slug}/builds/{build-slug}/abort"},
	{Tool: "get_build_bitrise_yml", Spec: "ci", Method: "get", Path: "/apps/{app-slug}/builds/{build-slug}/bitrise.yml"},
	{Tool: "list_build_workflows", Spec: "ci", Method: "get", Path: "/apps/{app-slug}/build-workflows"},
	{Tool: "list_artifacts", Spec: "ci", Method: "get", Path: "/apps/{app-slug}/builds/{build-slug}/artifacts"},
	{Tool: "get_artifact", Spec: "ci", Method: "get", Path: "/apps/{app-slug}/builds/{build-slug}/artifacts/{artifact-slug}"},
	{Tool: "delete_artifact", Spec: "ci", Method: "delete", Path: "/apps/{app-slug}/builds/{build-slug}/artifacts/{artifact-slug}"},
	{Tool: "update_artifact", Spec: "ci", Method: "patch", Path: "/apps/{app-slug}/builds/{build-slug}/artifacts/{artifact-slug}"},
	{Tool: "list_workspaces", Spec: "ci", Method: "get", Path: "/organizations"},
	{Tool: "get_workspace", Spec: "ci", Method: "get", Path: "/organizations/{org-slug}"},
	{Tool: "get_workspace_groups", Spec: "ci", Method: "get", Path: "/organizations/{org-slug}/groups"},
	{Tool: "create_workspace_group", Spec: "ci", Method: "post", Path: "/organizations/{org-slug}/groups"},
	{Tool: "get_workspace_members", Spec: "ci", Method: "get", Path: "/organizations/{org-slug}/members"},
	{Tool: "invite_member_to_workspace", Spec: "ci", Method: "post", Path: "/organizations/{org-slug}/members"},
	{Tool: "add_member_to_group", Spec: "ci", Method: "post", Path: "/groups/{group-slug}/add_member",
		Note: "The tool calls PUT /groups/{group-slug}/members/{user-slug}, which is not in the published API document; this is the schema of the equivalent add_member operation."},
	{Tool: "list_outgoing_webhooks", Spec: "ci", Method: "get", Path: "/apps/{app-slug}/outgoing-webhooks"},
	{Tool: "create_outgoing_webhook", Spec: "ci", Method: "post", Path: "/apps/{app-slug}/outgoing-webhooks"},
	{Tool: "delete_outgoing_webhook", Spec: "ci", Method: "delete", Path: "/apps/{app-slug}/outgoing-webhooks/{app-webhook-slug}"},
	{Tool: "update_outgoing_webhook", Spec: "ci", Method: "put", Path: "/apps/{app-slug}/outgoing-webhooks/{app-webhook-slug}"},
	{Tool: "list_cache_items", Spec: "ci", Method: "get", Path: "/apps/{app-slug}/cache-items"},
	{Tool: "delete_all_cache_items", Spec: "ci", Method: "delete", Path: "/apps/{app-slug}/cache-items"},
	{Tool: "delete_cache_item", Spec: "ci", Method: "delete", Path: "/apps/{app-slug}/cache-items/{cache-item-id}"},
	{Tool: "get_cache_item_download_url", Spec: "ci", Method: "get", Path: "/apps/{app-slug}/cache-items/{cache-item-id}/download"},
	{Tool: "list_pipelines", Spec: "ci", Method: "get", Path: "/apps/{app-slug}/pipelines",
		Drop: []string{"data[].credit_cost", "data[].is_processed"},
		Note: "pull_request_id is omitted when 0; trigger_params is included only with verbose=true."},
	{Tool: "get_pipeline", Spec: "ci", Method: "get", Path: "/apps/{app-slug}/pipelines/{pipeline-id}",
		Drop: []string{"app", "number_in_app_scope", "put_on_hold_at", "credit_cost", "workflows[].credit_cost"},
		Note: "trigger_params and attempts are included only with verbose=true; workflows[].startFailureReason is omitted when empty."},
	{Tool: "abort_pipeline", Spec: "ci", Method: "post", Path: "/apps/{app-slug}/pipelines/{pipeline-id}/abort"},
	{Tool: "rebuild_pipeline", Spec: "ci", Method: "post", Path: "/apps/{app-slug}/pipelines/{pipeline-id}/rebuild"},
	{Tool: "list_group_roles", Spec: "ci", Method: "get", Path: "/apps/{app-slug}/roles/{role-name}"},
	{Tool: "replace_group_roles", Spec: "ci", Method: "put", Path: "/apps/{app-slug}/roles/{role-name}"},
	{Tool: "validate_bitrise_yml", Spec: "ci", Method: "post", Path: "/validate-bitrise-yml"},
	{Tool: "step_search", Spec: "ci", Method: "get", Path: "/search-steps"},
	{Tool: "step_inputs", Spec: "ci", Method: "get", Path: "/step-inputs"},
	{Tool: "list_available_stacks", Spec: "ci", Method: "get", Path: "/available-stacks"},

	// Release Management v2 — apps.
	{Tool: "create_connected_app", Spec: "rm-apps", Method: "post", Path: "/"},
	{Tool: "update_connected_app", Spec: "rm-apps", Method: "patch", Path: "/{id}"},
	{Tool: "list_connected_apps", Spec: "rm-apps", Method: "get", Path: "/"},
	{Tool: "get_connected_app", Spec: "rm-apps", Method: "get", Path: "/{id}"},
	{Tool: "list_installable_artifacts", Spec: "rm-apps", Method: "get", Path: "/installable-artifacts"},
	{Tool: "generate_installable_artifact_upload_url", Spec: "rm-apps", Method: "get", Path: "/installable-artifacts/{id}/upload-url"},
	{Tool: "get_installable_artifact_upload_and_proc_status", Spec: "rm-apps", Method: "get", Path: "/installable-artifacts/{id}/status"},
	{Tool: "set_installable_artifact_public_install_page", Spec: "rm-apps", Method: "patch", Path: "/installable-artifacts/{id}/public-install-page"},

	// Release Management v2 — build distributions.
	{Tool: "list_build_distribution_versions", Spec: "rm-bd", Method: "get", Path: "/"},
	{Tool: "list_build_distribution_version_test_builds", Spec: "rm-bd", Method: "get", Path: "/test-builds"},
	{Tool: "create_tester_group", Spec: "rm-bd", Method: "post", Path: "/tester-groups"},
	{Tool: "notify_tester_group", Spec: "rm-bd", Method: "post", Path: "/tester-groups/{id}/notify"},
	{Tool: "add_testers_to_tester_group", Spec: "rm-bd", Method: "post", Path: "/tester-groups/{id}/add-testers"},
	{Tool: "update_tester_group", Spec: "rm-bd", Method: "put", Path: "/tester-groups/{id}"},
	{Tool: "list_tester_groups", Spec: "rm-bd", Method: "get", Path: "/tester-groups"},
	{Tool: "get_tester_group", Spec: "rm-bd", Method: "get", Path: "/tester-groups/{id}"},
	{Tool: "get_potential_testers", Spec: "rm-bd", Method: "get", Path: "/potential-testers"},
	{Tool: "get_testers", Spec: "rm-bd", Method: "get", Path: "/testers"},

	// Release Management v2 — CodePush.
	{Tool: "codepush_list_deployments", Spec: "rm-cp", Method: "get", Path: "/deployments"},
	{Tool: "codepush_get_deployment", Spec: "rm-cp", Method: "get", Path: "/deployments/{id}"},
	{Tool: "codepush_create_deployment", Spec: "rm-cp", Method: "post", Path: "/deployments"},
	{Tool: "codepush_update_deployment", Spec: "rm-cp", Method: "patch", Path: "/deployments/{id}"},
	{Tool: "codepush_delete_deployment", Spec: "rm-cp", Method: "delete", Path: "/deployments/{id}"},
	{Tool: "codepush_promote_deployment", Spec: "rm-cp", Method: "post", Path: "/deployments/{id}/promote"},
	{Tool: "codepush_rollback_deployment", Spec: "rm-cp", Method: "post", Path: "/deployments/{id}/rollback"},
	{Tool: "codepush_list_updates", Spec: "rm-cp", Method: "get", Path: "/updates"},
	{Tool: "codepush_get_update", Spec: "rm-cp", Method: "get", Path: "/updates/{id}"},
	{Tool: "codepush_patch_update", Spec: "rm-cp", Method: "patch", Path: "/updates/{id}"},
	{Tool: "codepush_delete_update", Spec: "rm-cp", Method: "delete", Path: "/updates/{id}"},
	{Tool: "codepush_get_update_status", Spec: "rm-cp", Method: "get", Path: "/updates/{id}/status"},
	{Tool: "codepush_generate_update_upload_url", Spec: "rm-cp", Method: "get", Path: "/updates/{id}/upload-url"},
	{Tool: "codepush_get_metrics", Spec: "rm-cp", Method: "get", Path: "/metrics"},

	// Dev Environments (RDE).
	{Tool: "bitrise_devenv_list", Spec: "rde", Method: "get", Path: "/v1/workspaces/{workspaceId}/sessions"},
	{Tool: "bitrise_devenv_get", Spec: "rde", Method: "get", Path: "/v1/workspaces/{workspaceId}/sessions/{sessionId}"},
	{Tool: "bitrise_devenv_create", Spec: "rde", Method: "post", Path: "/v1/workspaces/{workspaceId}/sessions"},
	{Tool: "bitrise_devenv_update", Spec: "rde", Method: "patch", Path: "/v1/workspaces/{workspaceId}/sessions/{sessionId}"},
	{Tool: "bitrise_devenv_restore", Spec: "rde", Method: "post", Path: "/v1/workspaces/{workspaceId}/sessions/{sessionId}/restore"},
	{Tool: "bitrise_devenv_terminate", Spec: "rde", Method: "post", Path: "/v1/workspaces/{workspaceId}/sessions/{sessionId}/terminate"},
	{Tool: "bitrise_devenv_delete_terminated", Spec: "rde", Method: "post", Path: "/v1/workspaces/{workspaceId}/sessions:delete-terminated"},
	{Tool: "bitrise_devenv_compare_template", Spec: "rde", Method: "get", Path: "/v1/workspaces/{workspaceId}/sessions/{sessionId}/template-diff"},
	{Tool: "bitrise_devenv_list_session_notifications", Spec: "rde", Method: "get", Path: "/v1/workspaces/{workspaceId}/sessions/{sessionId}/notifications"},
	{Tool: "bitrise_devenv_list_templates", Spec: "rde", Method: "get", Path: "/v1/workspaces/{workspaceId}/templates"},
	{Tool: "bitrise_devenv_get_template", Spec: "rde", Method: "get", Path: "/v1/workspaces/{workspaceId}/templates/{templateId}"},
	{Tool: "bitrise_devenv_create_template", Spec: "rde", Method: "post", Path: "/v1/workspaces/{workspaceId}/templates"},
	{Tool: "bitrise_devenv_update_template", Spec: "rde", Method: "patch", Path: "/v1/workspaces/{workspaceId}/templates/{templateId}"},
	{Tool: "bitrise_devenv_delete_template", Spec: "rde", Method: "delete", Path: "/v1/workspaces/{workspaceId}/templates/{templateId}"},
	{Tool: "bitrise_devenv_list_warm_pools", Spec: "rde", Method: "get", Path: "/v1/workspaces/{workspaceId}/warm-pools"},
	{Tool: "bitrise_devenv_get_warm_pool", Spec: "rde", Method: "get", Path: "/v1/workspaces/{workspaceId}/warm-pools/{warmPoolId}"},
	{Tool: "bitrise_devenv_create_warm_pool", Spec: "rde", Method: "post", Path: "/v1/workspaces/{workspaceId}/warm-pools"},
	{Tool: "bitrise_devenv_update_warm_pool", Spec: "rde", Method: "patch", Path: "/v1/workspaces/{workspaceId}/warm-pools/{warmPoolId}"},
	{Tool: "bitrise_devenv_list_saved_inputs", Spec: "rde", Method: "get", Path: "/v1/saved-inputs"},
	{Tool: "bitrise_devenv_get_saved_input", Spec: "rde", Method: "get", Path: "/v1/saved-inputs/{savedInputId}"},
	{Tool: "bitrise_devenv_create_saved_input", Spec: "rde", Method: "post", Path: "/v1/saved-inputs"},
	{Tool: "bitrise_devenv_update_saved_input", Spec: "rde", Method: "patch", Path: "/v1/saved-inputs/{savedInputId}"},
	{Tool: "bitrise_devenv_delete_saved_input", Spec: "rde", Method: "delete", Path: "/v1/saved-inputs/{savedInputId}"},
	{Tool: "bitrise_devenv_list_stacks", Spec: "rde", Method: "get", Path: "/v1/workspaces/{workspaceId}/stacks"},
	{Tool: "bitrise_devenv_list_machine_types", Spec: "rde", Method: "get", Path: "/v1/workspaces/{workspaceId}/machine-types"},
	{Tool: "bitrise_devenv_get_workspace_usage", Spec: "rde", Method: "get", Path: "/v1/workspaces/{workspaceId}/usage"},
	{Tool: "bitrise_devenv_click", Spec: "rde", Method: "post", Path: "/v1/workspaces/{workspaceId}/sessions/{sessionId}/click"},
	{Tool: "bitrise_devenv_type", Spec: "rde", Method: "post", Path: "/v1/workspaces/{workspaceId}/sessions/{sessionId}/type"},
	{Tool: "bitrise_devenv_scroll", Spec: "rde", Method: "post", Path: "/v1/workspaces/{workspaceId}/sessions/{sessionId}/scroll"},
	{Tool: "bitrise_devenv_mouse_drag", Spec: "rde", Method: "post", Path: "/v1/workspaces/{workspaceId}/sessions/{sessionId}/mouse-drag"},
	{Tool: "bitrise_devenv_open_remote_access", Spec: "rde", Method: "post", Path: "/v1/workspaces/{workspaceId}/sessions/{sessionId}/open-remote-access"},
	{Tool: "bitrise_devenv_create_preview_link", Spec: "rde", Method: "post", Path: "/v1/workspaces/{workspaceId}/preview-links"},
}

// textEnvelope is the schema of a tool whose result is plain text: the tool
// belt wraps such results as {"content": <text>} in structuredContent.
func textEnvelope(description string) map[string]any {
	return map[string]any{
		"type":        "object",
		"description": description,
		"properties": map[string]any{
			"content": map[string]any{"type": "string", "description": "The text result of the tool."},
		},
	}
}

// manuals are schemas that cannot be derived from a published API document.
var manuals = map[string]map[string]any{
	"bitrise_devenv_device_guide":     textEnvelope("The requested device-session guide as markdown."),
	"bitrise_devenv_upload":           textEnvelope("A status message confirming the upload and where the files were extracted on the session."),
	"bitrise_devenv_download":         textEnvelope("A status message confirming the download and where the files were extracted locally."),
	"bitrise_devenv_delete":           textEnvelope("A status message confirming the session was deleted (the API returns an empty body on success)."),
	"bitrise_devenv_delete_warm_pool": textEnvelope("A status message confirming the warm pool was deleted (the API returns an empty body on success)."),
	// GET /apps/{app-slug}/builds/{build-slug}/log/summary is not in the
	// published API document; this mirrors a live response, without the fields
	// the tool strips (app_id, build_id, agent_info; and unless verbose:
	// cli_info, has_build_environment_setup_logs, is_log_archived; per step:
	// collection, support_url, release_notes).
	"get_build_steps": {
		"type":        "object",
		"description": "Step-level summary of a build: every workflow that ran with the status, timing and identity of each step. Not in the published API document; mirrors a live response.",
		"properties": map[string]any{
			"cli_info":                         map[string]any{"type": "object", "description": "Bitrise CLI details of the run (verbose only)."},
			"has_build_environment_setup_logs": map[string]any{"type": "boolean", "description": "verbose only"},
			"is_log_archived":                  map[string]any{"type": "boolean", "description": "verbose only"},
			"execution": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"build_environment_setup": map[string]any{
						"type":        "object",
						"description": "Status and timing of the build environment setup phase.",
						"properties": map[string]any{
							"status":      map[string]any{"type": "string"},
							"started_at":  map[string]any{"type": "string", "format": "date-time"},
							"finished_at": map[string]any{"type": "string", "format": "date-time"},
						},
					},
					"step_bundles": map[string]any{
						"type":        []any{"object", "null"},
						"description": "Step bundles used by the run, keyed by bundle UUID (referenced from steps[].step_bundle_uuid).",
						"additionalProperties": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"id":    map[string]any{"type": "string"},
								"title": map[string]any{"type": "string"},
							},
						},
					},
					"with_groups": map[string]any{"type": []any{"object", "null"}, "description": "\"with\" groups used by the run, keyed by UUID."},
					"workflows": map[string]any{
						"type":        "array",
						"description": "Workflows in execution order, including utility workflows (IDs starting with _).",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"uuid":        map[string]any{"type": "string"},
								"workflow_id": map[string]any{"type": "string", "description": "Workflow ID as defined in bitrise.yml."},
								"steps": map[string]any{
									"type": "array",
									"items": map[string]any{
										"type": "object",
										"properties": map[string]any{
											"uuid":             map[string]any{"type": "string", "description": "Step run UUID; pass it to get_build_log as step_uuid to read that step's log."},
											"step_id":          map[string]any{"type": "string", "description": "Step ID (e.g. git-clone) or step source URL."},
											"title":            map[string]any{"type": "string"},
											"version":          map[string]any{"type": "string"},
											"status":           map[string]any{"type": "string", "description": "success, failed, skipped, skipped_with_run_if, aborted, ..."},
											"status_reason":    map[string]any{"type": "string", "description": "Why the step was skipped, when it was."},
											"started_at":       map[string]any{"type": "string", "format": "date-time"},
											"finished_at":      map[string]any{"type": "string", "format": "date-time"},
											"source_code_url":  map[string]any{"type": "string"},
											"step_bundle_uuid": map[string]any{"type": "string", "description": "UUID of the step bundle this step belongs to, if any."},
											"errors": map[string]any{
												"type": "array",
												"items": map[string]any{
													"type": "object",
													"properties": map[string]any{
														"code":    map[string]any{"type": "integer"},
														"message": map[string]any{"type": "string"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	},
}

func main() {
	specsDir := flag.String("specs", "", "directory with <name>.json spec files instead of fetching the hosted ones")
	out := flag.String("out", "internal/tool/outputschema/schemas", "output directory")
	flag.Parse()

	specs := map[string]map[string]any{}
	for name, url := range specURLs {
		var raw []byte
		var err error
		if *specsDir != "" {
			raw, err = os.ReadFile(filepath.Join(*specsDir, name+".json"))
		} else {
			raw, err = fetch(url)
		}
		if err != nil {
			log.Fatalf("load spec %s: %v", name, err)
		}
		var doc map[string]any
		if err := json.Unmarshal(raw, &doc); err != nil {
			log.Fatalf("parse spec %s: %v", name, err)
		}
		specs[name] = doc
	}

	if err := os.MkdirAll(*out, 0o755); err != nil {
		log.Fatal(err)
	}
	stale, _ := filepath.Glob(filepath.Join(*out, "*.json"))
	for _, f := range stale {
		_ = os.Remove(f)
	}

	written := map[string]bool{}
	for _, m := range mappings {
		if written[m.Tool] {
			log.Fatalf("tool %s mapped twice", m.Tool)
		}
		schema, err := schemaFor(specs, m)
		if err != nil {
			log.Fatalf("%s: %v", m.Tool, err)
		}
		write(*out, m.Tool, schema)
		written[m.Tool] = true
	}
	for tool, schema := range manuals {
		if written[tool] {
			log.Fatalf("tool %s is both mapped and manual", tool)
		}
		write(*out, tool, schema)
		written[tool] = true
	}
	fmt.Printf("wrote %d schemas to %s\n", len(written), *out)
}

func fetch(url string) ([]byte, error) {
	client := http.Client{Timeout: 60 * time.Second}
	res, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d", url, res.StatusCode)
	}
	return io.ReadAll(res.Body)
}

// Property descriptions: the published documents carry paragraph-long field
// comments on every nested field, which multiplied the size of the tools/list
// response every MCP client downloads (the Dev Environments session object
// alone is repeated on seven tools). Keep the first sentence on the
// top-level properties, where it disambiguates the result, and drop it
// deeper down, where the property names carry the meaning.
const (
	maxNestedDescription = 120
	describedDepth       = 1
)

// trimDescriptions shortens or drops nested descriptions in place; depth is
// the property nesting level of schema (0 = root).
func trimDescriptions(schema map[string]any, depth int) {
	for _, k := range []string{"properties", "additionalProperties", "items"} {
		switch v := schema[k].(type) {
		case map[string]any:
			if k == "properties" {
				for _, p := range v {
					if pm, ok := p.(map[string]any); ok {
						if depth+1 > describedDepth {
							delete(pm, "description")
						} else {
							shortenDescription(pm)
						}
						trimDescriptions(pm, depth+1)
					}
				}
			} else {
				delete(v, "description")
				trimDescriptions(v, depth)
			}
		}
	}
}

func shortenDescription(schema map[string]any) {
	d, ok := schema["description"].(string)
	if !ok {
		return
	}
	d = strings.Join(strings.Fields(d), " ")
	if i := strings.Index(d, ". "); i > 0 {
		d = d[:i+1]
	}
	if r := []rune(d); len(r) > maxNestedDescription {
		d = string(r[:maxNestedDescription-1]) + "…"
	}
	if d == "" {
		delete(schema, "description")
		return
	}
	schema["description"] = d
}

func write(dir, tool string, schema map[string]any) {
	trimDescriptions(schema, 0)
	if d, ok := schema["description"].(string); ok {
		schema["description"] = strings.Join(strings.Fields(d), " ")
	}
	b, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Fatalf("marshal %s: %v", tool, err)
	}
	if err := os.WriteFile(filepath.Join(dir, tool+".json"), append(b, '\n'), 0o644); err != nil {
		log.Fatal(err)
	}
}

// schemaFor derives the output schema of one mapped tool.
func schemaFor(specs map[string]map[string]any, m mapping) (map[string]any, error) {
	doc, ok := specs[m.Spec]
	if !ok {
		return nil, fmt.Errorf("unknown spec %q", m.Spec)
	}
	paths, _ := doc["paths"].(map[string]any)
	item, ok := paths[m.Path].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("path %s not in spec %s", m.Path, m.Spec)
	}
	op, ok := item[m.Method].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s %s not in spec %s", strings.ToUpper(m.Method), m.Path, m.Spec)
	}
	responses, _ := op["responses"].(map[string]any)

	var body map[string]any
	var code string
	for _, c := range []string{"200", "201", "202", "204"} {
		if r, ok := responses[c].(map[string]any); ok {
			body = r
			code = c
			break
		}
	}
	if body == nil {
		return nil, fmt.Errorf("no success response for %s %s", m.Method, m.Path)
	}

	title, _ := doc["info"].(map[string]any)["title"].(string)
	desc := fmt.Sprintf("Response of %s %s (%s, HTTP %s).", strings.ToUpper(m.Method), m.Path, title, code)
	if m.Note != "" {
		desc += " " + m.Note
	}

	var schema map[string]any
	if _, isOAS3 := doc["openapi"]; isOAS3 {
		schema = oas3Schema(body)
	} else {
		if s, ok := body["schema"].(map[string]any); ok {
			c := &converter{defs: definitions(doc), nullableOptional: true}
			schema = c.convert(s, map[string]bool{})
		}
	}

	schema = envelope(schema)
	for _, d := range m.Drop {
		if err := drop(schema, strings.Split(d, ".")); err != nil {
			return nil, fmt.Errorf("drop %q: %w", d, err)
		}
	}
	for path, props := range m.Add {
		target := schema
		if path != "" {
			for _, seg := range strings.Split(path, ".") {
				next, ok := target["properties"].(map[string]any)[seg].(map[string]any)
				if !ok {
					return nil, fmt.Errorf("add: property %q not found", seg)
				}
				target = next
			}
		}
		existing, _ := target["properties"].(map[string]any)
		if existing == nil {
			existing = map[string]any{}
			target["properties"] = existing
		}
		for name, prop := range props {
			existing[name] = prop
		}
	}
	if existing, _ := schema["description"].(string); existing != "" {
		desc += " " + existing
	}
	schema["description"] = desc
	return schema, nil
}

// envelope makes sure the schema describes a JSON object, mirroring how the
// tool belt builds structuredContent: an object passes through, an array is
// wrapped as {"items": [...]}, anything else (text, empty body) as
// {"content": "<text>"}.
func envelope(schema map[string]any) map[string]any {
	if schema == nil {
		return textEnvelope("The API returns an empty body on success, so content is empty.")
	}
	switch schema["type"] {
	case "object", nil:
		props, _ := schema["properties"].(map[string]any)
		_, hasAdditional := schema["additionalProperties"]
		if len(props) == 0 && !hasAdditional {
			return textEnvelope("The API returns an empty body on success, so content is empty.")
		}
		if len(props) > 0 || schema["type"] == "object" {
			return schema
		}
	case "array":
		return map[string]any{
			"type":       "object",
			"properties": map[string]any{"items": schema},
			"required":   []any{"items"},
		}
	}
	return textEnvelope("The raw text body of the response.")
}

// drop removes the property at path from schema (descending into array items
// at "[]"-suffixed segments).
func drop(schema map[string]any, path []string) error {
	seg := path[0]
	intoItems := strings.HasSuffix(seg, "[]")
	seg = strings.TrimSuffix(seg, "[]")
	props, _ := schema["properties"].(map[string]any)
	child, ok := props[seg].(map[string]any)
	if !ok {
		return fmt.Errorf("property %q not found", seg)
	}
	if len(path) == 1 {
		delete(props, seg)
		if req, ok := schema["required"].([]any); ok {
			var kept []any
			for _, r := range req {
				if r != seg {
					kept = append(kept, r)
				}
			}
			if len(kept) == 0 {
				delete(schema, "required")
			} else {
				schema["required"] = kept
			}
		}
		return nil
	}
	if intoItems {
		items, ok := child["items"].(map[string]any)
		if !ok {
			return fmt.Errorf("%q is not an array of objects", seg)
		}
		child = items
	}
	return drop(child, path[1:])
}

// --- Swagger 2.0 conversion ---------------------------------------------------

type converter struct {
	defs map[string]any
	// nullableOptional makes properties that are not required accept null
	// too: the Bitrise API v0.1 returns null for unset optional fields (e.g.
	// abort_reason, tag, with_groups) and the Dev Environments gateway emits
	// null for unset message fields (e.g. a template's deviceSpec), although
	// the documents type them as strings/objects.
	nullableOptional bool
}

func definitions(doc map[string]any) map[string]any {
	if d, ok := doc["definitions"].(map[string]any); ok {
		return d
	}
	if c, ok := doc["components"].(map[string]any); ok {
		if s, ok := c["schemas"].(map[string]any); ok {
			return s
		}
	}
	return map[string]any{}
}

var copiedKeywords = []string{"type", "description", "format", "enum", "minimum", "maximum", "minLength", "maxLength", "pattern", "default", "title"}

func (c *converter) convert(s map[string]any, seen map[string]bool) map[string]any {
	if ref, ok := s["$ref"].(string); ok {
		name := ref[strings.LastIndex(ref, "/")+1:]
		if seen[name] {
			return map[string]any{"type": "object", "description": "Recursive reference to " + name + "."}
		}
		target, ok := c.defs[name].(map[string]any)
		if !ok {
			return map[string]any{"type": "object", "description": "Unresolved reference " + ref + "."}
		}
		seen[name] = true
		out := c.convert(target, seen)
		delete(seen, name)
		// A $ref sibling description (swagger 2 tooling emits them) wins.
		if d, ok := s["description"].(string); ok && d != "" {
			out["description"] = d
		}
		return out
	}

	out := map[string]any{}
	for _, k := range copiedKeywords {
		if v, ok := s[k]; ok {
			out[k] = v
		}
	}
	if allOf, ok := s["allOf"].([]any); ok {
		merged := map[string]any{"type": "object", "properties": map[string]any{}}
		for _, part := range allOf {
			p, _ := part.(map[string]any)
			cp := c.convert(p, seen)
			for k, v := range cp {
				switch k {
				case "properties":
					for pk, pv := range v.(map[string]any) {
						merged["properties"].(map[string]any)[pk] = pv
					}
				case "required":
					merged["required"] = append(toAnySlice(merged["required"]), toAnySlice(v)...)
				default:
					if _, exists := merged[k]; !exists {
						merged[k] = v
					}
				}
			}
		}
		for k, v := range out {
			merged[k] = v
		}
		out = merged
	}
	if items, ok := s["items"].(map[string]any); ok {
		out["items"] = c.convert(items, seen)
	}
	if props, ok := s["properties"].(map[string]any); ok {
		required := map[string]bool{}
		for _, r := range toAnySlice(s["required"]) {
			if rs, ok := r.(string); ok {
				required[rs] = true
			}
		}
		converted := map[string]any{}
		for name, p := range props {
			pm, _ := p.(map[string]any)
			cp := c.convert(pm, seen)
			if c.nullableOptional && !required[name] {
				allowNull(cp)
			}
			converted[name] = cp
		}
		out["properties"] = converted
		if out["type"] == nil {
			out["type"] = "object"
		}
		if len(required) > 0 {
			out["required"] = s["required"]
		}
	}
	if ap, ok := s["additionalProperties"].(map[string]any); ok {
		out["additionalProperties"] = c.convert(ap, seen)
	}
	if nullable, _ := s["x-nullable"].(bool); nullable {
		allowNull(out)
	}
	if nullable, _ := s["nullable"].(bool); nullable {
		allowNull(out)
	}
	return out
}

func allowNull(schema map[string]any) {
	switch t := schema["type"].(type) {
	case string:
		if t == "null" {
			return
		}
		schema["type"] = []any{t, "null"}
		if enum, ok := schema["enum"].([]any); ok {
			schema["enum"] = append(enum, nil)
		}
	}
}

func toAnySlice(v any) []any {
	s, _ := v.([]any)
	return s
}

// --- OpenAPI 3 responses ------------------------------------------------------

// oas3Schema returns the response schema of an OpenAPI 3 response: the
// declared application/json schema when there is one, else a schema inferred
// from every published example value.
func oas3Schema(response map[string]any) map[string]any {
	content, _ := response["content"].(map[string]any)
	js, _ := content["application/json"].(map[string]any)
	if js == nil {
		return nil
	}
	if s, ok := js["schema"].(map[string]any); ok {
		c := &converter{defs: map[string]any{}}
		return c.convert(s, map[string]bool{})
	}
	var values []any
	if ex, ok := js["example"]; ok {
		values = append(values, ex)
	}
	if exs, ok := js["examples"].(map[string]any); ok {
		names := make([]string, 0, len(exs))
		for n := range exs {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			if e, ok := exs[n].(map[string]any); ok {
				if v, ok := e["value"]; ok {
					values = append(values, v)
				}
			}
		}
	}
	if len(values) == 0 {
		return nil
	}
	s := infer(values)
	if s["type"] == nil {
		return nil
	}
	if d, _ := s["description"].(string); d == "" {
		s["description"] = "Inferred from the examples published in the API document; field types reflect the example values."
	} else {
		s["description"] = d + " Inferred from the examples published in the API document."
	}
	return s
}

// infer builds a schema describing every value in values.
func infer(values []any) map[string]any {
	types := map[string]bool{}
	var objects []map[string]any
	var arrays [][]any
	for _, v := range values {
		switch t := v.(type) {
		case nil:
			types["null"] = true
		case bool:
			types["boolean"] = true
		case string:
			types["string"] = true
		case float64:
			if t == math.Trunc(t) {
				types["integer"] = true
			} else {
				types["number"] = true
			}
		case map[string]any:
			types["object"] = true
			objects = append(objects, t)
		case []any:
			types["array"] = true
			arrays = append(arrays, t)
		}
	}
	if types["number"] {
		delete(types, "integer")
	}
	out := map[string]any{}
	if len(objects) > 0 {
		keys := map[string]bool{}
		for _, o := range objects {
			for k := range o {
				keys[k] = true
			}
		}
		props := map[string]any{}
		for k := range keys {
			var kv []any
			for _, o := range objects {
				if v, ok := o[k]; ok {
					kv = append(kv, v)
				}
			}
			ps := infer(kv)
			// Examples rarely show every field unset; accept null like the
			// documented APIs do.
			if _, isArr := ps["type"].([]any); !isArr {
				allowNull(ps)
			}
			props[k] = ps
		}
		out["properties"] = props
	}
	if len(arrays) > 0 {
		var elems []any
		for _, a := range arrays {
			elems = append(elems, a...)
		}
		if len(elems) > 0 {
			out["items"] = infer(elems)
		} else {
			out["items"] = map[string]any{}
		}
	}
	names := make([]string, 0, len(types))
	for t := range types {
		names = append(names, t)
	}
	sort.Strings(names)
	switch len(names) {
	case 0:
	case 1:
		out["type"] = names[0]
	default:
		if len(names) == 1 || (len(names) == 2 && types["null"]) {
			// e.g. ["null","string"]: keep the real type first for readability.
			for _, n := range names {
				if n != "null" {
					out["type"] = []any{n, "null"}
				}
			}
		} else {
			arr := make([]any, len(names))
			for i, n := range names {
				arr[i] = n
			}
			out["type"] = arr
		}
	}
	if len(names) == 1 && names[0] == "null" {
		// Only null in the examples: the real type is unknown.
		delete(out, "type")
		out["description"] = "Only null appears in the published examples."
	}
	return out
}
