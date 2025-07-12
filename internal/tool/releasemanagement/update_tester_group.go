package releasemanagement

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var UpdateTesterGroup = bitrise.Tool{
	APIGroups: []string{"release-management"},
	Definition: mcp.NewTool("update_tester_group",
		mcp.WithDescription("Updates the given tester group. The name and the auto notification setting can be updated optionally."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier of the related Release Management connected app."),
			mcp.Required(),
		),
		mcp.WithString("id",
			mcp.Description("The uuidV4 identifier of the tester group to which testers will be added."),
			mcp.Required(),
		),
		mcp.WithString("name",
			mcp.Description("The new name for the tester group. Must be unique in the scope of the related connected app."),
		),
		mcp.WithBoolean("auto_notify",
			mcp.Description("If set to true it indicates the tester group will receive email notifications automatically from now on about new installable builds."),
			mcp.DefaultBool(false),
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

		body := map[string]any{}
		if v := request.GetString("name", ""); v != "" {
			body["name"] = v
		}
		if v := request.GetBool("auto_notify", false); v {
			body["auto_notify"] = v
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPut,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/v1/connected-apps/%s/tester-groups/%s", connectedAppID, id),
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
