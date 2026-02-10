package pipelines

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var Rebuild = bitrise.Tool{
	APIGroups: []string{"pipelines"},
	Definition: mcp.NewTool("rebuild_pipeline",
		mcp.WithDescription("Rebuild a pipeline."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("pipeline_id",
			mcp.Description("Identifier of the pipeline"),
			mcp.Required(),
		),
		mcp.WithBoolean("partial",
			mcp.Description("Whether to rebuild only unsuccessful workflows and their dependents"),
			mcp.DefaultBool(false),
		),
		mcp.WithString("triggered_by",
			mcp.Description("Who triggered the rebuild"),
		),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
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

		body := map[string]any{}
		if _, ok := request.GetArguments()["partial"]; ok {
			body["partial"] = request.GetBool("partial", false)
		}
		if v := request.GetString("triggered_by", ""); v != "" {
			body["triggered_by"] = v
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/pipelines/%s/rebuild", appSlug, pipelineID),
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
