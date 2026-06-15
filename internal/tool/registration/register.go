package registration

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

type RegisterResponse struct {
	PendingSignupID string `json:"pending_signup_id"`
	ExpiresAt       string `json:"expires_at"`
}

var Register = bitrise.Tool{
	APIGroups: []string{"registration"},
	Definition: mcp.NewTool("register",
		mcp.WithDescription("Start registration for a new Bitrise user. Sends a one-time password (OTP) to the provided email address and returns a `pending_signup_id`. After this returns successfully, ask the user for the OTP that was sent to their email, then call `verify_registration` with that OTP and the `pending_signup_id` from this response."),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
		mcp.WithOutputSchema[RegisterResponse](),
		mcp.WithString("email",
			mcp.Description("Email address of the user to register"),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		email, err := request.RequireString("email")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:   http.MethodPost,
			BaseURL:  bitrise.APIBaseURL,
			Path:     "/agent-signup/start",
			Body:     map[string]any{"email": email},
			SkipAuth: true,
		})
		if err != nil {
			return apiErrorResult(err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

// apiErrorResult converts a *bitrise.APIError into a structured tool result so
// the agent can branch on `code` / `status` rather than parsing a flat string.
// Falls back to the raw error message for non-API errors.
func apiErrorResult(err error) *mcp.CallToolResult {
	var apiErr *bitrise.APIError
	if !errors.As(err, &apiErr) {
		return mcp.NewToolResultErrorFromErr("call api", err)
	}

	payload := map[string]any{"status": apiErr.StatusCode}
	var parsed map[string]any
	if jsonErr := json.Unmarshal([]byte(apiErr.Body), &parsed); jsonErr == nil {
		for k, v := range parsed {
			payload[k] = v
		}
	} else {
		payload["body"] = apiErr.Body
	}

	out, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return mcp.NewToolResultError(apiErr.Error())
	}
	return mcp.NewToolResultError(string(out))
}
