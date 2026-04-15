package codepush

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var RollbackDeployment = bitrise.Tool{
	APIGroups: []string{"release-management-code-push", "release-management"},
	Definition: mcp.NewTool("codepush_rollback_deployment",
		mcp.WithDescription("Rollback a CodePush deployment to its previous version, or to a specific package if package_id is provided."),
		mcp.WithString("id",
			mcp.Description("Identifier (UUID) of the CodePush deployment to rollback"),
			mcp.Required(),
		),
		mcp.WithString("package_id",
			mcp.Description("Optional UUID of a specific package to rollback to. Defaults to the previous package."),
		),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{}
		if v := request.GetString("package_id", ""); v != "" {
			body["package_id"] = v
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APICodePushBaseURL,
			Path:    fmt.Sprintf("/deployments/%s/rollback", id),
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
