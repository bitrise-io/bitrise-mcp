package apps

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var GetBitriseYML = bitrise.Tool{
	APIGroups: []string{"apps", "read-only"},
	Definition: mcp.NewTool("get_bitrise_yml",
		mcp.WithDescription("Get the current Bitrise YML config file of a specified Bitrise app."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app (e.g., \"d8db74e2675d54c4\" or \"8eb495d0-f653-4eed-910b-8d6b56cc0ec7\")"),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/bitrise.yml", appSlug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
