package storereleases

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var UpdateReleasePreset = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases"},
	Definition: mcp.NewTool("update_release_preset", slices.Concat(
		[]mcp.ToolOption{
			mcp.WithDescription("Updates a release preset. Only the given arguments change; omitted ones are left as they are. " +
				"Sending 'approvals' or 'automation' replaces the whole existing list. The notification arguments are merged into the preset's current notification settings, and an empty string clears one. " +
				"To change the app's default preset, set 'default' to true on the new default; setting it to false on the current default fails. " +
				"App versions already created from the preset keep their settings."),
			mcp.WithString("preset_id",
				mcp.Description(presetIDDescription),
				mcp.Required(),
			),
			mcp.WithString("template_name",
				mcp.Description("The new name of the preset."),
			),
		},
		presetSettingOptions(),
		[]mcp.ToolOption{
			mcp.WithTitleAnnotation("Update Release Preset"),
			mcp.WithReadOnlyHintAnnotation(false),
			mcp.WithDestructiveHintAnnotation(true),
			mcp.WithOpenWorldHintAnnotation(true),
			mcp.WithIdempotentHintAnnotation(true),
		},
	)...),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		presetID, err := request.RequireString("preset_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body, err := presetSettings(request)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if v := request.GetString("template_name", ""); v != "" {
			body["template_name"] = v
		}

		changes := map[string]string{}
		for _, key := range presetNotificationKeys {
			v, ok, err := optionalString(request, key)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if ok {
				changes[key] = v
			}
		}
		if len(changes) > 0 {
			// The API stores the notification settings as one value, so
			// sending only some of them would clear the others.
			notifications, err := currentPresetNotifications(ctx, presetID)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("read the current notification settings", err), nil
			}
			for key, v := range changes {
				if v == "" {
					delete(notifications, key)
				} else {
					notifications[key] = v
				}
			}
			body["notifications"] = notifications
		}
		if len(body) == 0 {
			return mcp.NewToolResultError("nothing to update: give at least one argument besides preset_id"), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPut,
			BaseURL: bitrise.APIRMAppsBaseURL,
			Path:    fmt.Sprintf("/presets/%s", presetID),
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

func currentPresetNotifications(ctx context.Context, presetID string) (map[string]any, error) {
	res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
		Method:  http.MethodGet,
		BaseURL: bitrise.APIRMAppsBaseURL,
		Path:    fmt.Sprintf("/presets/%s", presetID),
	})
	if err != nil {
		return nil, err
	}
	var preset struct {
		Notifications map[string]any `json:"notifications"`
	}
	if err := json.Unmarshal([]byte(res), &preset); err != nil {
		return nil, fmt.Errorf("decode preset: %w", err)
	}
	if preset.Notifications == nil {
		preset.Notifications = map[string]any{}
	}
	return preset.Notifications, nil
}
