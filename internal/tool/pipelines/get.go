package pipelines

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var Get = bitrise.Tool{
	APIGroups: []string{"pipelines", "read-only"},
	Definition: mcp.NewTool("get_pipeline",
		mcp.WithDescription("Get a pipeline of a given app."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("pipeline_id",
			mcp.Description("Identifier of the pipeline"),
			mcp.Required(),
		),
		mcp.WithBoolean("verbose",
			mcp.Description("Include all pipeline details. Default: false"),
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
		pipelineID, err := request.RequireString("pipeline_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/pipelines/%s", appSlug, pipelineID),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}

		var response map[string]any
		if err := json.Unmarshal([]byte(res), &response); err != nil {
			return mcp.NewToolResultText(res), nil
		}

		// app.slug is the caller's own input argument — redundant in the response.
		delete(response, "app")
		// Internal implementation details with no diagnostic value.
		delete(response, "number_in_app_scope")
		delete(response, "put_on_hold_at")
		// credit_cost is billing metadata, not relevant for pipeline inspection.
		delete(response, "credit_cost")

		verbose := request.GetBool("verbose", false)
		if !verbose {
			// trigger_params overlaps with top-level trigger fields; the
			// environments array inside it could carry large amount of strings.
			delete(response, "trigger_params")
			// attempts tracks retry history; current_attempt_id at the top level
			// is sufficient for the common case.
			delete(response, "attempts")
		}

		// Clean up per-workflow fields that are either noise or billing metadata.
		if workflows, ok := response["workflows"].([]any); ok {
			for _, wf := range workflows {
				if wfMap, ok := wf.(map[string]any); ok {
					delete(wfMap, "credit_cost")
					// startFailureReason appears on every workflow regardless of
					// whether startup failed; only meaningful when non-empty.
					if v, ok := wfMap["startFailureReason"].(string); ok && v == "" {
						delete(wfMap, "startFailureReason")
					}
				}
			}
		}

		return mcp.NewToolResultStructuredOnly(response), nil
	},
}
