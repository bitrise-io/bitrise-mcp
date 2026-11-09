package storereleases

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var DeleteReleasePreset = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases"},
	Definition: mcp.NewTool("delete_release_preset",
		mcp.WithDescription("Deletes a release preset. Deleting the app's default preset needs 'new_default_template_id', the preset that becomes the default instead (400 otherwise). "+
			"App versions already created from the preset keep their settings. This cannot be undone."),
		mcp.WithString("preset_id",
			mcp.Description(presetIDDescription),
			mcp.Required(),
		),
		mcp.WithString("new_default_template_id",
			mcp.Description("The uuidV4 identifier of another preset of the app that becomes the default. Required when the deleted preset is the default."),
		),
		mcp.WithTitleAnnotation("Delete Release Preset"),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		presetID, err := request.RequireString("preset_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		var params map[string]any
		if v := request.GetString("new_default_template_id", ""); v != "" {
			params = map[string]any{"new_default_template_id": v}
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodDelete,
			BaseURL: bitrise.APIRMAppsBaseURL,
			Path:    fmt.Sprintf("/presets/%s", presetID),
			Params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
