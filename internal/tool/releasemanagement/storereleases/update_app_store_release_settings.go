package storereleases

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var UpdateAppStoreReleaseSettings = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management"},
	Definition: mcp.NewTool("update_app_store_release_settings",
		mcp.WithDescription("Updates the App Store release settings of an iOS app version in App Store Connect. Both 'release_type' and 'phased_release' must be given; the API rejects a partial update. "+
			"Mind that AFTER_APPROVAL makes Apple publish the version to end users as soon as the review passes, and SCHEDULED publishes it at 'earliest_release_date'. iOS only."),
		mcp.WithString("app_version_id",
			mcp.Description(appVersionIDDescription),
			mcp.Required(),
		),
		mcp.WithString("release_type",
			mcp.Description("When Apple publishes the version once the review passes: MANUAL waits for 'release_to_app_store', AFTER_APPROVAL publishes immediately, SCHEDULED publishes at 'earliest_release_date'. Case-insensitive."),
			mcp.Required(),
			mcp.Enum("MANUAL", "AFTER_APPROVAL", "SCHEDULED"),
		),
		mcp.WithBoolean("phased_release",
			mcp.Description("Whether to roll the version out over a 7-day period (true) or to all users at once (false)."),
			mcp.Required(),
		),
		mcp.WithString("earliest_release_date",
			mcp.Description("The earliest date-time to publish the version after approval, e.g. '2026-10-01T09:00:00Z'. Only allowed with 'release_type' SCHEDULED."),
		),
		mcp.WithTitleAnnotation("Update App Store Release Settings"),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appVersionID, err := request.RequireString("app_version_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		releaseType, err := request.RequireString("release_type")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		phasedRelease, err := request.RequireBool("phased_release")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{
			"release_type":   releaseType,
			"phased_release": phasedRelease,
		}
		if v := request.GetString("earliest_release_date", ""); v != "" {
			body["earliest_release_date"] = v
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPatch,
			BaseURL: bitrise.APIRMStoreReleasesBaseURL,
			Path:    fmt.Sprintf("/app-versions/%s/apple-app-store/release/settings", appVersionID),
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
