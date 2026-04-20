package codepush

import (
	"context"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var CreateDeployment = bitrise.Tool{
	APIGroups: []string{"release-management-code-push", "release-management"},
	Definition: mcp.NewTool("codepush_create_deployment",
		mcp.WithDescription("Create a new CodePush deployment for a Bitrise app."),
		mcp.WithString("name",
			mcp.Description("Name for the new deployment"),
			mcp.Required(),
		),
		mcp.WithString("app_id",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("key",
			mcp.Description("Optional deployment key. If not provided, one will be auto-generated."),
		),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		appID, err := request.RequireString("app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{
			"name":   name,
			"app_id": appID,
		}
		if v := request.GetString("key", ""); v != "" {
			body["key"] = v
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APICodePushBaseURL,
			Path:    "/deployments",
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
