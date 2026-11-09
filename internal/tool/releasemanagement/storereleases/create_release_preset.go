package storereleases

import (
	"context"
	"net/http"
	"slices"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var CreateReleasePreset = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases"},
	Definition: mcp.NewTool("create_release_preset", slices.Concat(
		[]mcp.ToolOption{
			mcp.WithDescription("Creates a release preset for a Release Management connected app: a template that 'create_app_version' fills new app versions from when given its id as 'presets_id'. " +
				"An app can have at most 5 presets, and creating one needs a Standard license on the app (412 otherwise)."),
			mcp.WithString("connected_app_id",
				mcp.Description("The uuidV4 identifier of the connected app the preset belongs to."),
				mcp.Required(),
			),
			mcp.WithString("template_name",
				mcp.Description("The name of the preset."),
				mcp.Required(),
			),
		},
		presetSettingOptions(),
		[]mcp.ToolOption{
			mcp.WithTitleAnnotation("Create Release Preset"),
			mcp.WithReadOnlyHintAnnotation(false),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithOpenWorldHintAnnotation(true),
			mcp.WithIdempotentHintAnnotation(false),
		},
	)...),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connectedAppID, err := request.RequireString("connected_app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		templateName, err := request.RequireString("template_name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body, err := presetSettings(request)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		body["app_id"] = connectedAppID
		body["template_name"] = templateName

		notifications := map[string]any{}
		for _, key := range presetNotificationKeys {
			if v := request.GetString(key, ""); v != "" {
				notifications[key] = v
			}
		}
		if len(notifications) > 0 {
			body["notifications"] = notifications
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APIRMAppsBaseURL,
			Path:    "/presets",
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
