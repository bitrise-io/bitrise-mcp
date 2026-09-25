package devenvironments

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/devenv"
	"github.com/mark3labs/mcp-go/mcp"
)

// ListSessionNotifications retrieves notifications for a session.
var ListSessionNotifications = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_list_session_notifications",
		mcp.WithTitleAnnotation("List session notifications"),
		mcp.WithDescription(`List notifications for a devenv session. Notifications are events sent by the VM (e.g., AI agent stopped, permission prompt, idle).

Results are ordered by creation time (newest first by default). Supports cursor-based pagination via created_before/created_after timestamps.`),
		mcp.WithString("session_id",
			mcp.Description("The unique identifier (UUID) of the session"),
			mcp.Required(),
		),
		mcp.WithString("created_before",
			mcp.Description("Only return notifications created before this timestamp (RFC3339, exclusive). Used for backward pagination."),
		),
		mcp.WithString("created_after",
			mcp.Description("Only return notifications created after this timestamp (RFC3339, exclusive). Used for polling new notifications."),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of notifications to return (1-100, default 50)"),
		),
		mcp.WithString("order",
			mcp.Description("Sort order by created_at: DESC (default, newest first) or ASC (oldest first)"),
			mcp.Enum("DESC", "ASC"),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sessionID, err := requireUUID(request, "session_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		params := map[string]string{}
		if v := request.GetString("created_before", ""); v != "" {
			params["createdBefore"] = v
		}
		if v := request.GetString("created_after", ""); v != "" {
			params["createdAfter"] = v
		}
		if v, ok := request.GetArguments()["limit"]; ok {
			if num, ok := v.(float64); ok {
				params["limit"] = strconv.Itoa(int(num))
			}
		}
		if v := request.GetString("order", ""); v != "" {
			params["order"] = "SORT_ORDER_" + v
		}

		res, err := devenv.CallAPI(ctx, devenv.CallAPIParams{
			Method: http.MethodGet,
			Path:   devenv.WsPath(ctx, fmt.Sprintf("/sessions/%s/notifications", sessionID)),
			Params: params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("list session notifications", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
