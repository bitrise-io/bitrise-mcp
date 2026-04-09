package builds

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var List = bitrise.Tool{
	APIGroups: []string{"builds", "read-only"},
	Definition: mcp.NewTool("list_builds",
		mcp.WithDescription("List all the builds of a specified Bitrise app or all accessible builds."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
		),
		mcp.WithString("sort_by",
			mcp.Description("Order of builds: created_at (default), running_first"),
			mcp.DefaultString("created_at"),
			mcp.Enum("created_at", "running_first"),
		),
		mcp.WithString("branch",
			mcp.Description("Filter builds by branch"),
		),
		mcp.WithString("workflow",
			mcp.Description("Filter builds by workflow"),
		),
		mcp.WithNumber("status",
			mcp.Description("Filter builds by status (0: not finished, 1: successful, 2: failed, 3: aborted, 4: in-progress)"),
			mcp.Enum("0", "1", "2", "3", "4"),
		),
		mcp.WithString("next",
			mcp.Description("Slug of the first build in the response"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Max number of elements per page (default: 50)"),
		),
		mcp.WithBoolean("verbose",
			mcp.Description("Include all build details. Default: false"),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params := map[string]any{}
		if v := request.GetString("sort_by", ""); v != "" {
			params["sort_by"] = v
		}
		if v := request.GetString("branch", ""); v != "" {
			params["branch"] = v
		}
		if v := request.GetString("workflow", ""); v != "" {
			params["workflow"] = v
		}
		if _, ok := request.GetArguments()["status"]; ok {
			params["status"] = strconv.Itoa(request.GetInt("status", 0))
		}
		if v := request.GetString("next", ""); v != "" {
			params["next"] = v
		}
		if _, ok := request.GetArguments()["limit"]; ok {
			params["limit"] = strconv.Itoa(request.GetInt("limit", 50))
		}

		appSlug := request.GetString("app_slug", "")
		path := "/builds"
		if appSlug != "" {
			path = fmt.Sprintf("/apps/%s/builds", appSlug)
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIBaseURL,
			Path:    path,
			Params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}

		var response map[string]any
		if err := json.Unmarshal([]byte(res), &response); err != nil {
			return mcp.NewToolResultText(res), nil
		}

		verbose := request.GetBool("verbose", false)
		if builds, ok := response["data"].([]any); ok {
			for _, item := range builds {
				build, ok := item.(map[string]any)
				if !ok {
					continue
				}

				// credit_cost is billing metadata — not useful for understanding build
				// results and repeated verbatim on every row.
				delete(build, "credit_cost")
				// Redundant with commit_hash; the URL can be reconstructed from the repo
				// URL and hash when needed.
				delete(build, "commit_view_url")
				// Internal processing/delivery state — not meaningful to the caller.
				delete(build, "environment_prepare_finished_at")
				delete(build, "is_processed")
				delete(build, "is_status_sent")
				delete(build, "log_format")
				// pull_request_id == 0 means this is not a PR build; omit rather than
				// forcing the caller to distinguish 0 from a real PR number.
				if v, ok := build["pull_request_id"].(float64); ok && v == 0 {
					delete(build, "pull_request_id")
				}

				// original_build_params duplicates branch, commit_hash, commit_message
				// and adds only internal PR plumbing fields (head/merge branch refs)
				// that are rarely needed. The pull_request_id at the top level is enough.
				if !verbose {
					delete(build, "original_build_params")
				}

				// When the caller has already scoped by app_slug, the embedded
				// repository object repeats the same app metadata on every build row.
				// Without app_slug it at least identifies which app owns the build, so
				// we keep a minimal subset instead of the full object.
				if repo, ok := build["repository"].(map[string]any); ok {
					if appSlug != "" && !verbose {
						delete(build, "repository")
					} else if !verbose {
						build["repository"] = map[string]any{
							"slug":       repo["slug"],
							"title":      repo["title"],
							"repo_owner": repo["repo_owner"],
							"repo_name":  repo["repo_slug"],
						}
					}
				}
			}
		}

		return mcp.NewToolResultStructuredOnly(response), nil
	},
}
