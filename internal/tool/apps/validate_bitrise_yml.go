package apps

import (
	"context"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var ValidateBitriseYML = bitrise.Tool{
	APIGroups: []string{"apps", "read-only"},
	Definition: mcp.NewTool("validate_bitrise_yml",
		mcp.WithDescription("Validate a Bitrise YML config file. This endpoint checks if the provided bitrise.yml is valid."),
		mcp.WithString("bitrise_yml",
			mcp.Description("The Bitrise YML config file content to be validated. It must be a string."),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ymlContent, err := request.RequireString("bitrise_yml")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APIBaseURL,
			Path:    "/validate-bitrise-yml",
			Body: map[string]any{
				"bitrise_yml": ymlContent, // CallAPI adds Content-Type: application/json so have to use that format
			},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
