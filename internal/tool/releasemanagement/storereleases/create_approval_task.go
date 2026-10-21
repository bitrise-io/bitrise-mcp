package storereleases

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var CreateApprovalTask = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management"},
	Definition: mcp.NewTool("create_approval_task",
		mcp.WithDescription("Adds an approval task to the approval stage of an app version. The release cannot move past the approval stage until every task is completed."),
		mcp.WithString("app_version_id",
			mcp.Description(appVersionIDDescription),
			mcp.Required(),
		),
		mcp.WithString("summary",
			mcp.Description("The name of the approval task."),
			mcp.Required(),
		),
		mcp.WithString("description",
			mcp.Description("A detailed explanation of the approval task, providing context beyond the summary."),
		),
		mcp.WithString("assigned_user_slug",
			mcp.Description("The slug of the user to assign the task to. Only the assignee or a project admin can complete an assigned task, and only its creator or a project admin can edit it."),
		),
		mcp.WithString("due_date",
			mcp.Description("A valid date in the future by which the task should be completed."),
		),
		mcp.WithTitleAnnotation("Create Approval Task"),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appVersionID, err := request.RequireString("app_version_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		summary, err := request.RequireString("summary")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{
			"summary": summary,
		}
		for _, key := range []string{"description", "assigned_user_slug", "due_date"} {
			if v := request.GetString(key, ""); v != "" {
				body[key] = v
			}
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APIRMStoreReleasesBaseURL,
			Path:    fmt.Sprintf("/app-versions/%s/approvals", appVersionID),
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
