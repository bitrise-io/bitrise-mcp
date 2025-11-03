package apps

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var UpdateBitriseYML = bitrise.Tool{
	APIGroups: []string{"apps"},
	Definition: mcp.NewTool("update_bitrise_yml",
		mcp.WithDescription("Update the Bitrise YML config file of a specified Bitrise app."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app (e.g., \"d8db74e2675d54c4\" or \"8eb495d0-f653-4eed-910b-8d6b56cc0ec7\")"),
			mcp.Required(),
		),
		mcp.WithString("bitrise_yml_as_json",
			mcp.Description("The new Bitrise YML config file content to be updated. It must be a string."),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		ymlContent, err := request.RequireString("bitrise_yml_as_json")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
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
