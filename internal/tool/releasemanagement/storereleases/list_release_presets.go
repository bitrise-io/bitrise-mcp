package storereleases

import (
	"context"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var ListReleasePresets = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases"},
	Definition: mcp.NewTool("list_release_presets",
		mcp.WithDescription("Lists the release presets of a Release Management connected app. A release preset is a template for new app versions: "+
			"the release candidate source, branch and workflow, automatic store upload, Slack and Teams notifications, approval tasks and automation. "+
			"Pass a preset's 'id' as 'presets_id' to 'create_app_version' to fill a new app version from it. 'default' marks the app's default preset."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier of the connected app whose release presets are listed."),
			mcp.Required(),
		),
		mcp.WithTitleAnnotation("List Release Presets"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connectedAppID, err := request.RequireString("connected_app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIRMAppsBaseURL,
			Path:    "/presets",
			Params:  map[string]any{"app_id": connectedAppID},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
