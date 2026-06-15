package registration

import (
	"context"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

type VerifyRegistrationResponse struct {
	UserSlug       string `json:"user_slug"`
	APIToken       string `json:"api_token"`
	TokenExpiresAt string `json:"token_expires_at"`
	WorkspaceSlug  string `json:"workspace_slug,omitempty"`
}

var VerifyRegistration = bitrise.Tool{
	APIGroups: []string{"registration"},
	Definition: mcp.NewTool("verify_registration",
		mcp.WithDescription("Verify a pending Bitrise registration using the OTP sent to the user's email. Pass the `pending_signup_id` returned by `register`. Returns an `api_token` (a Bitrise personal access token) and, when a workspace was auto-created, a `workspace_slug`. After a successful call, save the token into the user's MCP client config so the Bitrise MCP server is authenticated: find the Bitrise server entry and set its `Authorization` header to `Bearer <api_token>` (for clients that connect through `mcp-remote`, such as Claude Desktop, add `--header \"Authorization: Bearer <api_token>\"` to its `args` instead). Common config files: Claude Desktop — `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\\Claude\\claude_desktop_config.json` (Windows); Cursor — `~/.cursor/mcp.json`; VS Code — `.vscode/mcp.json` or user `settings.json`. If you can't edit the file, output the exact JSON snippet for the user to paste; if you don't know which client they use, ask first. Then have them restart or reconnect their MCP client. The token expires in 24 hours, after which they'll need to register again."),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
		mcp.WithOutputSchema[VerifyRegistrationResponse](),
		mcp.WithString("pending_signup_id",
			mcp.Description("The pending_signup_id returned by the `register` tool"),
			mcp.Required(),
		),
		mcp.WithString("otp",
			mcp.Description("One-time password sent to the email address"),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		pendingSignupID, err := request.RequireString("pending_signup_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		otp, err := request.RequireString("otp")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:   http.MethodPost,
			BaseURL:  bitrise.APIBaseURL,
			Path:     "/agent-signup/confirm",
			Body:     map[string]any{"pending_signup_id": pendingSignupID, "otp": otp},
			SkipAuth: true,
		})
		if err != nil {
			return apiErrorResult(err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
