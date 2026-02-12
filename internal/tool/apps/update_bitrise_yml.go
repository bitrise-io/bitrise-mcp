package apps

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var UpdateBitriseYML = bitrise.Tool{
	APIGroups: []string{"apps"},
	Definition: mcp.NewTool("update_bitrise_yml",
		mcp.WithDescription("Update the Bitrise YML config file of a specified Bitrise app. Prefer using file_path parameter for better performance if applicable."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app (e.g., \"d8db74e2675d54c4\" or \"8eb495d0-f653-4eed-910b-8d6b56cc0ec7\")"),
			mcp.Required(),
		),
		mcp.WithString("bitrise_yml_as_json",
			mcp.Description("The new Bitrise YML config file content to be updated. It must be a string. Either this or file_path must be provided, but not both."),
		),
		mcp.WithString("file_path",
			mcp.Description("Path to the Bitrise YML config file to be updated. Either this or bitrise_yml_as_json must be provided, but not both. Recommended for large files as it's more performant."),
		),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		ymlContent := request.GetString("bitrise_yml_as_json", "")
		filePath := request.GetString("file_path", "")

		if len(ymlContent) == 0 && len(filePath) == 0 {
			return mcp.NewToolResultError("either 'bitrise_yml_as_json' or 'file_path' must be provided"), nil
		}
		if len(ymlContent) > 0 && len(filePath) > 0 {
			return mcp.NewToolResultError("only one of 'bitrise_yml_as_json' or 'file_path' should be provided, not both"), nil
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
			Path:    fmt.Sprintf("/apps/%s/bitrise.yml", appSlug),
			Body: map[string]any{
				"app_config_datastore_yaml": ymlContent,
			},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
