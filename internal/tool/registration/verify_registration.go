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
		mcp.WithDescription("Verify a pending Bitrise registration using the OTP sent to the user's email. Pass the `pending_signup_id` returned by `register`. Returns an `api_token` and (only when a workspace was auto-created) a `workspace_slug`. After a successful call: locate the user's Bitrise MCP server entry in their MCP client config (common locations: Claude Desktop `claude_desktop_config.json`, Cursor `~/.cursor/mcp.json`, VS Code `settings.json`) and update it — for stdio transport set the `BITRISE_TOKEN` environment variable to the new `api_token`; for HTTP transport set the `Authorization: Bearer <api_token>` header on that entry. If you can't edit the file (e.g. remote MCP context), output the exact JSON snippet for the user to paste. If you don't know which client the user runs, ask. Then tell the user to reconnect or restart their MCP client for the changes to take effect."),
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
