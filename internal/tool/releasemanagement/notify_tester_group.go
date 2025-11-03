package releasemanagement

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var NotifyTesterGroup = bitrise.Tool{
	APIGroups: []string{"release-management"},
	Definition: mcp.NewTool("notify_tester_group",
		mcp.WithDescription("Notifies a tester group about a new test build."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier of the related Release Management connected app."),
			mcp.Required(),
		),
		mcp.WithString("id",
			mcp.Description("The uuidV4 identifier of the tester group whose members will be notified about the test build."),
			mcp.Required(),
		),
		mcp.WithString("test_build_id",
			mcp.Description("The unique identifier of the test build what will be sent in the notification of the tester group."),
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
		testBuildID, err := request.RequireString("test_build_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{
			"test_build_id": testBuildID,
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APIRMBaseURL,
			Path:    fmt.Sprintf("/connected-apps/%s/tester-groups/%s/notify", connectedAppID, id),
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
