package releasemanagement

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var CreateTesterGroup = bitrise.Tool{
	APIGroups: []string{"release-management"},
	Definition: mcp.NewTool("create_tester_group",
		mcp.WithDescription("Creates a tester group for a Release Management connected app. Tester groups can be used to distribute installable artifacts to testers automatically. When a new installable artifact is available, the tester groups can either automatically or manually be notified via email. The notification email will contain a link to the installable artifact page for the artifact within Bitrise Release Management. A Release Management connected app can have multiple tester groups. Project team members of the connected app can be selected to be testers and added to the tester group. This endpoint has an elevated access level requirement. Only the owner of the related Bitrise Workspace, a workspace manager or the related project's admin can manage tester groups."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier of the related Release Management connected app."),
			mcp.Required(),
		),
		mcp.WithString("name",
			mcp.Description("The name for the new tester group. Must be unique in the scope of the connected app."),
		),
		mcp.WithBoolean("auto_notify",
			mcp.Description("If set to true it indicates that the tester group will receive notifications automatically."),
			mcp.DefaultBool(false),
		),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connectedAppID, err := request.RequireString("connected_app_id")
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
			Method:  http.MethodPost,
			BaseURL: bitrise.APIRMBaseURL,
			Path:    fmt.Sprintf("/connected-apps/%s/tester-groups", connectedAppID),
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
