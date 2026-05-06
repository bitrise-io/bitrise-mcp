package registration

import (
	"context"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var Register = bitrise.Tool{
	APIGroups: []string{"registration"},
	Definition: mcp.NewTool("register",
		mcp.WithDescription("Start registration for a new Bitrise user. Sends a one-time password (OTP) to the provided email address."),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
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
			BaseURL:  bitrise.APIRegistrationBaseURL,
			Path:     "/register", // TODO: confirm actual endpoint path
			Body:     map[string]any{"email": email},
			SkipAuth: true,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
