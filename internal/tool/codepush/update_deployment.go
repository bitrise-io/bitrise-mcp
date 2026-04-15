package codepush

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var UpdateDeployment = bitrise.Tool{
	APIGroups: []string{"code-push"},
	Definition: mcp.NewTool("codepush_update_deployment",
		mcp.WithDescription("Update the name of an existing CodePush deployment."),
		mcp.WithString("id",
			mcp.Description("Identifier (UUID) of the CodePush deployment"),
			mcp.Required(),
		),
		mcp.WithString("name",
			mcp.Description("New name for the deployment"),
			mcp.Required(),
		),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		name, err := request.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPatch,
			BaseURL: bitrise.APICodePushBaseURL,
			Path:    fmt.Sprintf("/deployments/%s", id),
			Body:    map[string]any{"name": name},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
