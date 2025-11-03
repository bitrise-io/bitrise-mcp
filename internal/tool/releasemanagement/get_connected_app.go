package releasemanagement

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var GetConnectedApp = bitrise.Tool{
	APIGroups: []string{"release-management", "read-only"},
	Definition: mcp.NewTool("get_connected_app",
		mcp.WithDescription("Gives back a Release Management connected app for the authenticated account."),
		mcp.WithString("id",
			mcp.Description("Identifier of the Release Management connected app"),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIRMBaseURL,
			Path:    fmt.Sprintf("/connected-apps/%s", id),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
