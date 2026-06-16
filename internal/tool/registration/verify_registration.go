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
		mcp.WithDescription("Verify a pending Bitrise registration using the OTP sent to the user's email. Pass the `pending_signup_id` returned by `register`. Returns an `api_token` (a Bitrise personal access token) and, when a workspace was auto-created, a `workspace_slug`. If the code is rejected as invalid, retry with the same `pending_signup_id`; if it has expired or hit the attempt limit, call `register` again for a fresh code. After success, authenticate the MCP connection so the other tools work: set `Authorization: Bearer <api_token>` on the user's Bitrise server entry. Give the user BOTH a CLI command and a copy-pastable JSON snippet — e.g. Claude Code: `claude mcp add --transport http bitrise https://mcp.bitrise.io -H \"Authorization: Bearer <api_token>\"` — and let them use whichever fits (ask which client they use if unsure). Then have them reconnect for it to take effect, and explain how for their client (don't just say \"reconnect\"). The token expires in 24 hours, after which they'll need to register again."),
		mcp.WithTitleAnnotation("Verify Registration"),
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
