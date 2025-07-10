package tool

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
)

var listCacheItems = Tool{
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

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodGet,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/apps/%s/cache-items", appSlug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var deleteAllCacheItems = Tool{
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

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodDelete,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/apps/%s/cache", appSlug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var deleteCacheItem = Tool{
	APIGroups: []string{"cache-items"},
	Definition: mcp.NewTool("delete_cache_item",
		mcp.WithDescription("Delete a key-value cache item."),
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

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodDelete,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/apps/%s/cache/%s", appSlug, cacheItemID),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var getCacheItemDownloadURL = Tool{
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

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodGet,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/apps/%s/cache-items/%s/download", appSlug, cacheItemID),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
