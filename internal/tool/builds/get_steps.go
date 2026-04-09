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
		mcp.WithDescription("Get step statuses of a specific build of a given app."),
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
			Path:    fmt.Sprintf("/apps/%s/builds/%s/log/summary", appSlug, buildSlug),
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
