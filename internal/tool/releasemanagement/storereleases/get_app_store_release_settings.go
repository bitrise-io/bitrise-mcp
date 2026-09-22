package storereleases

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var GetAppStoreReleaseSettings = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management", "read-only"},
	Definition: mcp.NewTool("get_app_store_release_settings",
		mcp.WithDescription("Gives back the App Store release settings of an iOS app version as set in App Store Connect: 'release_type' (MANUAL, AFTER_APPROVAL or SCHEDULED), 'phased_release' and, for scheduled releases, the 'earliest_release_date'. "+
			"iOS only; fails with 422 ERR_APP_STORE_VERSION_NOT_FOUND when the version does not exist in App Store Connect."),
		mcp.WithString("app_version_id",
			mcp.Description(appVersionIDDescription),
			mcp.Required(),
		),
		mcp.WithTitleAnnotation("Get App Store Release Settings"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appVersionID, err := request.RequireString("app_version_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIRMStoreReleasesBaseURL,
			Path:    fmt.Sprintf("/app-versions/%s/apple-app-store/release/settings", appVersionID),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
