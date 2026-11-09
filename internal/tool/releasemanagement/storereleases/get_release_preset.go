package storereleases

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var GetReleasePreset = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases"},
	Definition: mcp.NewTool("get_release_preset",
		mcp.WithDescription("Gets a release preset: its template name, whether it is the app's default, the release candidate settings, automatic store upload, notification settings, approval tasks and automation."),
		mcp.WithString("preset_id",
			mcp.Description(presetIDDescription),
			mcp.Required(),
		),
		mcp.WithTitleAnnotation("Get Release Preset"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		presetID, err := request.RequireString("preset_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIRMAppsBaseURL,
			Path:    fmt.Sprintf("/presets/%s", presetID),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
