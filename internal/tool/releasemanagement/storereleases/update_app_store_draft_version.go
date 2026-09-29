package storereleases

import (
	"context"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var UpdateAppStoreDraftVersion = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management"},
	Definition: mcp.NewTool("update_app_store_draft_version",
		mcp.WithDescription("Renames the App Store Connect draft version of an iOS connected app (the latest unreleased version, see 'get_app_store_draft_version') to the given version string, so it matches the Release Management app version to be released. "+
			"A Release Management app version named 'version_string' must already exist under the connected app. iOS only."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier of the connected app."),
			mcp.Required(),
		),
		mcp.WithString("version_string",
			mcp.Description("The new version string of the draft, e.g. '1.2.3'. Must match the name of an existing Release Management app version."),
			mcp.Required(),
		),
		mcp.WithTitleAnnotation("Update App Store Draft Version"),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connectedAppID, err := request.RequireString("connected_app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		versionString, err := request.RequireString("version_string")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPatch,
			BaseURL: bitrise.APIRMAppsBaseURL,
			Path:    "/apple-app-store/app-versions/draft",
			Body: map[string]any{
				"app_id":         connectedAppID,
				"version_string": versionString,
			},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
