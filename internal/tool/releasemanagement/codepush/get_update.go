package codepush

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var GetUpdate = bitrise.Tool{
	APIGroups: []string{"release-management-code-push", "release-management", "read-only"},
	Definition: mcp.NewTool("codepush_get_update",
		mcp.WithDescription("Get a specific CodePush update by its ID."),
		mcp.WithString("id",
			mcp.Description("Identifier (UUID) of the CodePush update"),
			mcp.Required(),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APICodePushBaseURL,
			Path:    fmt.Sprintf("/updates/%s", id),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
