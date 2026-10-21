package storereleases

import (
	"context"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var releaseEventNames = []string{
	"release_candidate_set", "testflight_upload_finished", "google_play_store_upload_finished",
	"beta_review_approved", "beta_review_rejected", "release_for_apple_app_store_testing_group",
	"release_on_google_play_store_testing_track", "approvals_completed", "submitted_for_review",
	"review_cancelled", "review_status_changed", "release_started", "release_percentage_changed",
	"release_completed",
}

var approvalItemsSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"summary":     map[string]any{"type": "string", "description": "The name of the approval task."},
		"description": map[string]any{"type": "string"},
		"due_date":    map[string]any{"type": "string", "description": "A valid date in the future."},
	},
	"required": []string{"summary"},
}

var automationItemsSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"event_name":    map[string]any{"type": "string", "enum": releaseEventNames},
		"workflow_name": map[string]any{"type": "string", "description": "Workflow to run on the event. Give either this or 'pipeline_name'."},
		"pipeline_name": map[string]any{"type": "string", "description": "Pipeline to run on the event. Give either this or 'workflow_name'."},
	},
	"required": []string{"event_name"},
}

var CreateAppVersion = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management"},
	Definition: mcp.NewTool("create_app_version",
		mcp.WithDescription("Creates a store release app version (a release in the Release Management UI) under a connected app. "+
			"Nothing is uploaded or published by this call: it opens the release with its release candidate stage in progress. "+
			"Values given here take precedence over the ones coming from 'presets_id'."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier of the connected app to create the app version under."),
			mcp.Required(),
		),
		mcp.WithString("name",
			mcp.Description("The name of the app version. For iOS apps this is used as the App Store Connect version and must be in X.Y.Z format."),
			mcp.Required(),
		),
		mcp.WithString("artifact_source",
			mcp.Description("Where release candidate artifacts come from: 'ci' for Bitrise CI builds (required when 'release_branch' or 'workflow' is set) or 'api' for externally uploaded installable artifacts."),
			mcp.Enum("ci", "api"),
		),
		mcp.WithString("description",
			mcp.Description("The description of the app version."),
		),
		mcp.WithString("release_branch",
			mcp.Description("The release branch part of the build configuration for the release candidate stage."),
		),
		mcp.WithString("workflow",
			mcp.Description("The workflow part of the build configuration for the release candidate stage."),
		),
		mcp.WithBoolean("automatic_store_upload",
			mcp.Description("If true, Release Management uploads every successful build of the given branch and workflow to the store."),
		),
		mcp.WithBoolean("release_candidate_locked",
			mcp.Description("If true, the release candidate is locked and the latest matching artifact is not selected automatically."),
		),
		mcp.WithString("slack_webhook_url",
			mcp.Description("Slack incoming webhook URL for release update notifications, matching ^https://.*\\.slack\\.com/services/T[a-zA-Z0-9]*/B[a-zA-Z0-9]*/[a-zA-Z0-9]*$."),
		),
		mcp.WithString("slack_notification_integration_id",
			mcp.Description("Identifier of a Slack notification integration from the Workspace Settings, for release update notifications."),
		),
		mcp.WithString("teams_webhook_url",
			mcp.Description("Microsoft Teams incoming webhook URL for release update notifications, matching ^https://(?:.*\\.webhook|outlook)\\.office(?:365)?\\.com/webhookb2/.*"),
		),
		mcp.WithArray("approvals",
			mcp.Description("Tasks for the approval stage, each with a 'summary' and optionally a 'description' and a 'due_date'."),
			mcp.Items(approvalItemsSchema),
		),
		mcp.WithArray("automation",
			mcp.Description("Workflow or pipeline runs triggered on release events, each with an 'event_name' and a 'workflow_name' or 'pipeline_name'."),
			mcp.Items(automationItemsSchema),
		),
		mcp.WithString("presets_id",
			mcp.Description("Identifier of a preset template to fill the app version from. Presets can carry the release branch and workflow, automatic store upload, notification URLs, approvals and automations."),
		),
		mcp.WithTitleAnnotation("Create App Version"),
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
		name, err := request.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{
			"app_id": connectedAppID,
			"name":   name,
		}
		for _, key := range []string{
			"artifact_source", "description", "release_branch", "workflow",
			"slack_webhook_url", "slack_notification_integration_id", "teams_webhook_url", "presets_id",
		} {
			if v := request.GetString(key, ""); v != "" {
				body[key] = v
			}
		}
		for _, key := range []string{"automatic_store_upload", "release_candidate_locked"} {
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

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APIRMStoreReleasesBaseURL,
			Path:    "/app-versions",
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
