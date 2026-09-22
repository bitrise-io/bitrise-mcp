package storereleases

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var UpdateAppVersion = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management"},
	Definition: mcp.NewTool("update_app_version",
		mcp.WithDescription("Updates a store release app version (a release in the Release Management UI). Only the given arguments change; omitted ones are left as they are. "+
			"To pin a specific installable artifact as the release candidate, send 'release_candidate_id' together with 'release_candidate_locked' set to true, otherwise the call is rejected with 409. "+
			"Sending 'approvals' or 'automation' replaces the whole existing list. Send an empty string as 'description', 'slack_webhook_url' or 'teams_webhook_url' to clear it. "+
			"Setting 'status' stops Release Management from managing the release for good: the app version moves to a final status and cannot be modified afterwards (409), so confirm with the user before doing so."),
		mcp.WithString("app_version_id",
			mcp.Description(appVersionIDDescription),
			mcp.Required(),
		),
		mcp.WithString("artifact_source",
			mcp.Description("Where release candidate artifacts come from: 'ci' for Bitrise CI builds (required when 'release_branch' or 'workflow' is set) or 'api' for externally uploaded installable artifacts."),
			mcp.Enum("ci", "api"),
		),
		mcp.WithString("description",
			mcp.Description("The new description of the app version. An empty string clears it."),
		),
		mcp.WithString("release_branch",
			mcp.Description("The release branch part of the build configuration for the release candidate stage."),
		),
		mcp.WithString("workflow",
			mcp.Description("The workflow part of the build configuration for the release candidate stage."),
		),
		mcp.WithString("release_candidate_id",
			mcp.Description("The uuidV4 identifier of the installable artifact to select as the release candidate. Must be sent with 'release_candidate_locked' set to true."),
		),
		mcp.WithBoolean("release_candidate_locked",
			mcp.Description("If true, the currently selected release candidate is locked and the latest matching artifact is not selected automatically."),
		),
		mcp.WithBoolean("automatic_store_upload",
			mcp.Description("If true, Release Management uploads every successful build of the given branch and workflow to the store. Turning it on also starts uploading the currently selected release candidate right away, if there is one."),
		),
		mcp.WithString("slack_webhook_url",
			mcp.Description("Slack incoming webhook URL for release update notifications. An empty string clears it."),
		),
		mcp.WithString("slack_notification_integration_id",
			mcp.Description("Identifier of a Slack notification integration from the Workspace Settings, for release update notifications."),
		),
		mcp.WithString("teams_webhook_url",
			mcp.Description("Microsoft Teams incoming webhook URL for release update notifications. An empty string clears it."),
		),
		mcp.WithString("status",
			mcp.Description("Closes the release: 'abandoned' when it will not ship, 'completed_externally' when it shipped outside Release Management. Both are final and irreversible; any running staged rollout schedule is dropped."),
			mcp.Enum("abandoned", "completed_externally"),
		),
		mcp.WithArray("approvals",
			mcp.Description("Tasks for the approval stage, each with a 'summary' and optionally a 'description' and a 'due_date'. Replaces the existing approval task list."),
			mcp.Items(approvalItemsSchema),
		),
		mcp.WithArray("automation",
			mcp.Description("Workflow or pipeline runs triggered on release events, each with an 'event_name' and a 'workflow_name' or 'pipeline_name'. Replaces the existing automation list."),
			mcp.Items(automationItemsSchema),
		),
		mcp.WithTitleAnnotation("Update App Version"),
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

		body := map[string]any{}
		for _, key := range []string{
			"artifact_source", "release_branch", "workflow", "release_candidate_id",
			"slack_notification_integration_id", "status",
		} {
			if v := request.GetString(key, ""); v != "" {
				body[key] = v
			}
		}
		for _, key := range []string{"description", "slack_webhook_url", "teams_webhook_url"} {
			v, ok, err := optionalString(request, key)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if ok {
				body[key] = v
			}
		}
		for _, key := range []string{"release_candidate_locked", "automatic_store_upload"} {
			v, ok, err := optionalBool(request, key)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if ok {
				body[key] = v
			}
		}
		for _, key := range []string{"approvals", "automation"} {
			v, ok, err := optionalArray(request, key)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if ok {
				body[key] = v
			}
		}
		if len(body) == 0 {
			return mcp.NewToolResultError("nothing to update: give at least one argument besides app_version_id"), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPatch,
			BaseURL: bitrise.APIRMStoreReleasesBaseURL,
			Path:    fmt.Sprintf("/app-versions/%s", appVersionID),
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
