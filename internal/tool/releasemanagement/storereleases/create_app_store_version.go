package storereleases

import (
	"context"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var CreateAppStoreVersion = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management"},
	Definition: mcp.NewTool("create_app_store_version",
		mcp.WithDescription("Creates the App Store Connect version record for an iOS app version, which is needed before it can be submitted for App Store review. "+
			"A Release Management app version with the same name as 'version_string' must already exist under the connected app (404 ERR_APP_VERSION_NOT_FOUND otherwise); its release type and, on a Standard license, its default release method and preset release notes are applied to the new record. "+
			"Nothing is published by this call. iOS only."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier of the connected app."),
			mcp.Required(),
		),
		mcp.WithString("version_string",
			mcp.Description("The version string to create in App Store Connect, e.g. '1.2.3'. Must match the name of an existing Release Management app version."),
			mcp.Required(),
		),
		mcp.WithTitleAnnotation("Create App Store Version"),
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
		versionString, err := request.RequireString("version_string")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APIRMAppsBaseURL,
			Path:    "/apple-app-store/app-versions",
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
