package releasemanagement

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var GetTesterGroup = bitrise.Tool{
	APIGroups: []string{"release-management", "read-only"},
	Definition: mcp.NewTool("get_tester_group",
		mcp.WithDescription("Gives back the details of the selected tester group."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier of the app the tester group is connected to. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithString("id",
			mcp.Description("The uuidV4 identifier of the tester group. This field is mandatory."),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connectedAppID, err := request.RequireString("connected_app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIRMBaseURL,
			Path:    fmt.Sprintf("/connected-apps/%s/tester-groups/%s", connectedAppID, id),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
