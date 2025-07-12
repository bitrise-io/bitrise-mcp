package workspaces

import (
	"context"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var List = bitrise.Tool{
	APIGroups: []string{"workspaces", "read-only"},
	Definition: mcp.NewTool("list_workspaces",
		mcp.WithDescription("List the workspaces the user has access to"),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIBaseURL,
			Path:    "/organizations",
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
