package user

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var GetProviderConnectURL = bitrise.Tool{
	APIGroups: []string{"user", "read-only"},
	Definition: mcp.NewTool("get_provider_connect_url",
		mcp.WithDescription(
			"Returns a URL the user can open in a browser to connect a Git provider (GitHub, "+
				"Bitbucket, or GitLab) to their Bitrise account via OAuth. The URL points at a page "+
				"on app.bitrise.io that auto-submits the OmniAuth form using the user's existing "+
				"Bitrise session — clicking the URL takes the user straight to the provider's "+
				"authorize page. This tool only returns the URL; the OAuth handshake itself runs in "+
				"the user's browser and cannot be performed via PAT. Use list_connected_accounts "+
				"to verify the connection afterwards. Response is JSON "+
				`{"provider": "<provider>", "connect_url": "<url>"}`+
				"; open the connect_url value in a browser.",
		),
		mcp.WithString("provider",
			mcp.Description("Git provider identifier."),
			mcp.Enum("github", "bitbucket", "gitlab"),
			mcp.Required(),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		provider, err := request.RequireString("provider")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/me/identities/connect-url/%s", provider),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
