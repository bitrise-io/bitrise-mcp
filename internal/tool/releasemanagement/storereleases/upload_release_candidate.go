package storereleases

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var UploadReleaseCandidate = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management"},
	Definition: mcp.NewTool("upload_release_candidate",
		mcp.WithDescription("Starts uploading the release candidate of an app version to the store: TestFlight for iOS apps, the Google Play Console for Android apps. "+
			"The upload runs asynchronously in a Bitrise build; follow it with 'get_release_candidate_upload_status'. "+
			"Unless the app version has 'automatic_store_upload' enabled, the release candidate gets locked by this call. "+
			"Fails with 422 when no suitable artifact is selected and with 409 when an upload cannot be started (for example one is already running)."),
		mcp.WithString("app_version_id",
			mcp.Description(appVersionIDDescription),
			mcp.Required(),
		),
		mcp.WithTitleAnnotation("Upload Release Candidate"),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appVersionID, err := request.RequireString("app_version_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPut,
			BaseURL: bitrise.APIRMStoreReleasesBaseURL,
			Path:    fmt.Sprintf("/app-versions/%s/release-candidate/upload", appVersionID),
			Body:    map[string]any{},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
