package user

import (
	"context"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var ListConnectedAccounts = bitrise.Tool{
	APIGroups: []string{"user", "read-only"},
	Definition: mcp.NewTool("list_connected_accounts",
		mcp.WithDescription(
			"List the authenticated user's connected Git provider identities. For each provider "+
				"(GitHub, GitHub App, Bitbucket, GitLab) returns whether it is connected, the linked "+
				"account name and URL, and whether the stored OAuth credentials are in an error state. "+
				"Use this after get_provider_connect_url to detect when the user has completed the "+
				"OAuth flow in their browser.",
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIBaseURL,
			Path:    "/me/identities",
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
