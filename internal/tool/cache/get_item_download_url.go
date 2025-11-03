package cache

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var GetItemDownloadURL = bitrise.Tool{
	APIGroups: []string{"cache-items", "read-only"},
	Definition: mcp.NewTool("get_cache_item_download_url",
		mcp.WithDescription("Get the download URL for a cache item."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("cache_item_id",
			mcp.Description("Key of the cache item"),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		cacheItemID, err := request.RequireString("cache_item_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/cache-items/%s/download", appSlug, cacheItemID),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
