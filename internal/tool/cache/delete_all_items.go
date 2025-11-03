package cache

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var DeleteAllItems = bitrise.Tool{
	APIGroups: []string{"cache-items"},
	Definition: mcp.NewTool("delete_all_cache_items",
		mcp.WithDescription("Delete all key-value cache items belonging to an app."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodDelete,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/cache", appSlug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
