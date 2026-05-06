package registration

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

type VerifyRegistrationResponse struct {
	UserSlug       string `json:"user_slug"`
	APIToken       string `json:"api_token"`
	TokenExpiresAt string `json:"token_expires_at"`
	WorkspaceSlug  string `json:"workspace_slug"`
	NextSteps      string `json:"next_steps"`
}

var VerifyRegistration = bitrise.Tool{
	APIGroups: []string{"registration"},
	Definition: mcp.NewTool("verify_registration",
		mcp.WithDescription("Verify a pending Bitrise registration using the OTP sent to the user's email. Returns an API token and workspace slug for use with authenticated tools."),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
		mcp.WithOutputSchema[VerifyRegistrationResponse](),
		mcp.WithString("email",
			mcp.Description("Email address used during registration"),
			mcp.Required(),
		),
		mcp.WithString("otp",
			mcp.Description("One-time password sent to the email address"),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		email, err := request.RequireString("email")
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
			Body:     map[string]any{"email": email, "otp": otp},
			SkipAuth: true,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}

		var parsed VerifyRegistrationResponse
		if err := json.Unmarshal([]byte(res), &parsed); err != nil {
			return mcp.NewToolResultErrorFromErr("parse response", err), nil
		}

		parsed.NextSteps = "Registration verified successfully. Ask the user if they would like to update their MCP configuration to use the new credentials (set BITRISE_TOKEN to the returned api_token). If yes, guide them through updating the config, then ask them to restart the MCP agent for the changes to take effect."

		out, err := json.Marshal(parsed)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("marshal response", err), nil
		}
		return mcp.NewToolResultText(string(out)), nil
	},
}
