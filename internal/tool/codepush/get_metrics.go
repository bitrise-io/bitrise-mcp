package codepush

import (
	"context"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var GetMetrics = bitrise.Tool{
	APIGroups: []string{"code-push", "read-only"},
	Definition: mcp.NewTool("codepush_get_metrics",
		mcp.WithDescription("Get workspace-level CodePush usage metrics including data transfer, storage, and monthly active users, along with their limits and billing cycle information."),
		mcp.WithString("workspace_slug",
			mcp.Description("Slug of the Bitrise workspace"),
			mcp.Required(),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		workspaceSlug, err := request.RequireString("workspace_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APICodePushBaseURL,
			Path:    "/metrics",
			Params:  map[string]any{"workspace_slug": workspaceSlug},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
