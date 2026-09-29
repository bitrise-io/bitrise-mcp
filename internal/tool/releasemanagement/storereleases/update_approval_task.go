package storereleases

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var UpdateApprovalTask = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management"},
	Definition: mcp.NewTool("update_approval_task",
		mcp.WithDescription("Updates an approval task of an app version, including marking it completed or reopening it with 'completed'. Only the given arguments change. "+
			"On an assigned task, only the assignee or a project admin may change 'completed', and only the creator or a project admin may change the other fields (403 otherwise)."),
		mcp.WithString("app_version_id",
			mcp.Description(appVersionIDDescription),
			mcp.Required(),
		),
		mcp.WithString("task_id",
			mcp.Description("The identifier of the approval task, as listed by 'list_approval_tasks'."),
			mcp.Required(),
		),
		mcp.WithString("summary",
			mcp.Description("The new name of the approval task."),
		),
		mcp.WithString("description",
			mcp.Description("The new detailed explanation of the approval task. An empty string clears it."),
		),
		mcp.WithString("assigned_user_slug",
			mcp.Description("The slug of the user to assign the task to."),
		),
		mcp.WithBoolean("completed",
			mcp.Description("Set to true to complete the task (approve), or to false to reopen it. Omit to leave unchanged."),
		),
		mcp.WithTitleAnnotation("Update Approval Task"),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appVersionID, err := request.RequireString("app_version_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		taskID, err := request.RequireString("task_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{}
		for _, key := range []string{"summary", "assigned_user_slug"} {
			if v := request.GetString(key, ""); v != "" {
				body[key] = v
			}
		}
		if v, ok, err := optionalString(request, "description"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		} else if ok {
			body["description"] = v
		}
		v, ok, err := optionalBool(request, "completed")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if ok {
			body["completed"] = v
		}
		if len(body) == 0 {
			return mcp.NewToolResultError("nothing to update: give at least one of summary, description, assigned_user_slug or completed"), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPatch,
			BaseURL: bitrise.APIRMStoreReleasesBaseURL,
			Path:    fmt.Sprintf("/app-versions/%s/approvals/%s", appVersionID, taskID),
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
