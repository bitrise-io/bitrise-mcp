package builds

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var GetSteps = bitrise.Tool{
	APIGroups: []string{"builds", "read-only"},
	Definition: mcp.NewTool("get_build_steps",
		mcp.WithDescription("Get the steps of a specific build of a given app with their status and timing. For failed or aborted steps it includes the failure reason (errors), and for skipped steps why they were skipped (status_reason)."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("build_slug",
			mcp.Description("Identifier of the build"),
			mcp.Required(),
		),
		mcp.WithArray("status",
			mcp.WithStringEnumItems([]string{"not_started", "started", "success", "failed", "failed_skippable", "preparation_failed", "skipped", "skipped_with_run_if", "aborted", "aborted_with_custom_timeout", "aborted_with_no_output"}),
			mcp.Description("Keep only steps with these statuses; workflows without any are left out. Use [\"failed\", \"preparation_failed\", \"aborted\", \"aborted_with_custom_timeout\", \"aborted_with_no_output\"] for every step that broke the build."),
		),
		mcp.WithString("step_id",
			mcp.Description("Keep only the runs of this step, e.g. xcode-test. A step used several times is returned once per run, each with its own uuid."),
		),
		mcp.WithBoolean("verbose",
			mcp.Description("Include all build details. Default: false"),
		),
		mcp.WithTitleAnnotation("Get Build Steps"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		buildSlug, err := request.RequireString("build_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		params := map[string]any{}
		if v := request.GetStringSlice("status", nil); len(v) > 0 {
			params["status"] = v
		}
		if v := request.GetString("step_id", ""); v != "" {
			params["step_id"] = v
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/builds/%s/log/summary", appSlug, buildSlug),
			Params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}

		var response map[string]any
		if err := json.Unmarshal([]byte(res), &response); err != nil {
			return mcp.NewToolResultText(res), nil
		}

		// app_id and build_id are redundant — the caller already knows them from
		// the request parameters, and they're available via get_build.
		delete(response, "app_id")
		delete(response, "build_id")
		// agent_info describes the runner infrastructure, not the build steps.
		// Machine type and stack are already on the build object itself.
		delete(response, "agent_info")

		verbose := request.GetBool("verbose", false)
		if !verbose {
			// All internal implementation details.
			delete(response, "cli_info")
			delete(response, "has_build_environment_setup_logs")
			delete(response, "is_log_archived")
		}

		// collection, support_url, and release_notes are almost always the same
		// as source_code_url.
		if execution, ok := response["execution"].(map[string]any); ok {
			if workflows, ok := execution["workflows"].([]any); ok {
				for _, wf := range workflows {
					if wfMap, ok := wf.(map[string]any); ok {
						if steps, ok := wfMap["steps"].([]any); ok {
							for _, step := range steps {
								if stepMap, ok := step.(map[string]any); ok {
									delete(stepMap, "collection")
									delete(stepMap, "support_url")
									delete(stepMap, "release_notes")
								}
							}
						}
					}
				}
			}
		}

		return mcp.NewToolResultStructuredOnly(response), nil
	},
}
