package builds

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var Get = bitrise.Tool{
	APIGroups: []string{"builds", "read-only"},
	Definition: mcp.NewTool("get_build",
		mcp.WithDescription("Get a specific build of a given app."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("build_slug",
			mcp.Description("Identifier of the build"),
			mcp.Required(),
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
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		buildSlug, err := request.RequireString("build_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/builds/%s", appSlug, buildSlug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}

		var response map[string]any
		if err := json.Unmarshal([]byte(res), &response); err != nil {
			return mcp.NewToolResultText(res), nil
		}

		verbose := request.GetBool("verbose", false)
		if !verbose {
			if build, ok := response["data"].(map[string]any); ok {
				// original_build_params duplicates branch, commit_hash, commit_message
				// and adds internal PR branch references (head/merge branch) already
				// captured by pull_request_id.
				delete(build, "original_build_params")
				// credit_cost is billing metadata, not relevant for build inspection.
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
			}
		}

		return mcp.NewToolResultStructuredOnly(response), nil
	},
}
