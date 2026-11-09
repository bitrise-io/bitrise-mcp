package storereleases

import (
	"github.com/mark3labs/mcp-go/mcp"
)

const presetIDDescription = "The uuidV4 identifier of the release preset, as listed by 'list_release_presets'."

var presetNotificationKeys = []string{"slack_webhook_url", "slack_notification_integration_id", "teams_webhook_url"}

var presetApprovalItemsSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"summary":     map[string]any{"type": "string", "description": "The name of the approval task."},
		"description": map[string]any{"type": "string"},
	},
	"required": []string{"summary"},
}

func presetSettingOptions() []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithBoolean("default",
			mcp.Description("If true, the preset becomes the app's default preset, and the current default stops being one."),
		),
		mcp.WithString("artifact_source",
			mcp.Description("Where release candidate artifacts come from: 'ci' for Bitrise CI builds or 'api' for externally uploaded installable artifacts. "+
				"Required the first time the release candidate settings are set, unless 'release_branch' or 'workflow' is given, which implies 'ci'. 'api' cannot be combined with 'release_branch' or 'workflow' (422)."),
			mcp.Enum("ci", "api"),
		),
		mcp.WithString("release_branch",
			mcp.Description("The release branch part of the build configuration for the release candidate stage."),
		),
		mcp.WithString("workflow",
			mcp.Description("The workflow part of the build configuration for the release candidate stage."),
		),
		mcp.WithBoolean("auto_select_release_candidate",
			mcp.Description("If true, the latest artifact matching the branch and workflow is selected as the release candidate automatically."),
		),
		mcp.WithBoolean("automatic_store_upload",
			mcp.Description("If true, app versions created from the preset upload every successful build of the given branch and workflow to the store."),
		),
		mcp.WithString("slack_webhook_url",
			mcp.Description("Slack incoming webhook URL for release update notifications."),
		),
		mcp.WithString("slack_notification_integration_id",
			mcp.Description("Identifier of a Slack notification integration from the Workspace Settings, for release update notifications."),
		),
		mcp.WithString("teams_webhook_url",
			mcp.Description("Microsoft Teams incoming webhook URL for release update notifications."),
		),
		mcp.WithArray("approvals",
			mcp.Description("Approval tasks that app versions created from the preset start with, each with a 'summary' and optionally a 'description'."),
			mcp.Items(presetApprovalItemsSchema),
		),
		mcp.WithArray("automation",
			mcp.Description("Workflow or pipeline runs triggered on release events, each with an 'event_name' and a 'workflow_name' or 'pipeline_name'."),
			mcp.Items(automationItemsSchema),
		),
	}
}

// presetSettings builds the request body fields of the given shared
// arguments, except the notifications: the API nests the release candidate
// settings under release_candidate and calls the store upload auto_upload.
func presetSettings(request mcp.CallToolRequest) (map[string]any, error) {
	body := map[string]any{}
	if v, ok, err := optionalBool(request, "default"); err != nil {
		return nil, err
	} else if ok {
		body["default"] = v
	}
	if v, ok, err := optionalBool(request, "automatic_store_upload"); err != nil {
		return nil, err
	} else if ok {
		body["auto_upload"] = v
	}

	releaseCandidate := map[string]any{}
	for arg, key := range map[string]string{"artifact_source": "source", "release_branch": "branch", "workflow": "workflow"} {
		if v := request.GetString(arg, ""); v != "" {
			releaseCandidate[key] = v
		}
	}
	if v, ok, err := optionalBool(request, "auto_select_release_candidate"); err != nil {
		return nil, err
	} else if ok {
		releaseCandidate["auto_select"] = v
	}
	if len(releaseCandidate) > 0 {
		body["release_candidate"] = releaseCandidate
	}

	for _, key := range []string{"approvals", "automation"} {
		v, ok, err := optionalArray(request, key)
		if err != nil {
			return nil, err
		}
		if ok {
			body[key] = v
		}
	}
	return body, nil
}
