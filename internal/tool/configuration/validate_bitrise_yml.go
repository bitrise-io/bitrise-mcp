package configuration

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var ValidateBitriseYML = bitrise.Tool{
	APIGroups: []string{"configuration", "read-only"},
	Definition: mcp.NewTool("validate_bitrise_yml",
		mcp.WithDescription("Validate a Bitrise YML config file. Use this tool to verify any changes made in bitrise.yml. Prefer using file_path parameter for better performance if applicable."),
		mcp.WithString("bitrise_yml",
			mcp.Description("The Bitrise YML config file content to be validated. It must be a string. Either this or file_path must be provided, but not both."),
		),
		mcp.WithString("file_path",
			mcp.Description("Path to the Bitrise YML config file to be validated. Either this or bitrise_yml must be provided, but not both. Recommended for large files as it's more performant."),
		),
		mcp.WithString("app_slug",
			mcp.Description("Slug of a Bitrise app (as returned by the list_apps tool). Specifying this value allows for validating the YML against workspace-specific settings like available stacks, machine types, license pools etc."),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ymlContent := request.GetString("bitrise_yml", "")
		filePath := request.GetString("file_path", "")

		if len(ymlContent) == 0 && len(filePath) == 0 {
			return mcp.NewToolResultError("either 'bitrise_yml' or 'file_path' must be provided"), nil
		}
		if len(ymlContent) > 0 && len(filePath) > 0 {
			return mcp.NewToolResultError("only one of 'bitrise_yml' or 'file_path' should be provided, not both"), nil
		}

		if len(filePath) > 0 {
			content, err := os.ReadFile(filePath)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to read file: %v", err)), nil
			}
			ymlContent = string(content)
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
