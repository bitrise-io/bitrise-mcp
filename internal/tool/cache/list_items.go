package cache

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var ListItems = bitrise.Tool{
	APIGroups: []string{"cache-items", "read-only"},
	Definition: mcp.NewTool("list_cache_items",
		mcp.WithDescription("List the key-value cache items belonging to an app."),
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
			Method:  http.MethodGet,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/cache-items", appSlug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
