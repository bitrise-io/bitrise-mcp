package user

import (
	"context"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

// Me is the single identity tool of the server: one Bitrise account covers
// every product, so there is no Dev Environments variant.
var Me = bitrise.Tool{
	APIGroups: []string{"user", "account", "read-only", "dev-environments", "dev-environments-read-only"},
	Definition: mcp.NewTool("me",
		mcp.WithDescription("Get the currently authenticated Bitrise user account (username, slug, email). One account covers every Bitrise product, including Dev Environments. Needs a user token (OAuth or Personal Access Token); a Workspace API Token has no user."),
		mcp.WithTitleAnnotation("Get Current User"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIBaseURL,
			Path:    "/me",
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
