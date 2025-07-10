package tool

import (
	"context"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
)

var me = Tool{
	APIGroups: []string{"user", "read-only"},
	Definition: mcp.NewTool("me",
		mcp.WithDescription("Get user info for the currently authenticated user account"),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodGet,
			baseURL: apiBaseURL,
			path:    "/me",
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
