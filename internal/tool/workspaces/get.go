package workspaces

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var Get = bitrise.Tool{
	APIGroups: []string{"workspaces", "read-only"},
	Definition: mcp.NewTool("get_workspace",
		mcp.WithDescription("Get details for one workspace"),
		mcp.WithString("workspace_slug",
			mcp.Description("Slug of the Bitrise workspace"),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		workspaceSlug, err := request.RequireString("workspace_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/organizations/%s", workspaceSlug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
