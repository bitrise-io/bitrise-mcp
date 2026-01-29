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
		mcp.WithDescription("Validate a Bitrise YML config file. Use this tool to verify any changes made in bitrise.yml."),
		mcp.WithString("bitrise_yml",
			mcp.Description("The Bitrise YML config file content to be validated. It must be a string."),
			mcp.Required(),
		),
		mcp.WithString("app_slug",
			mcp.Description("Slug of a Bitrise app (as returned by the list_apps tool). Specifying this value allows for validating the YML against workspace-specific settings like available stacks, machine types, license pools etc."),
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
