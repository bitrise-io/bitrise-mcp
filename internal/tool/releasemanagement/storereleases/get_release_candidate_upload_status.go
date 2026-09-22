package storereleases

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var GetReleaseCandidateUploadStatus = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management", "read-only"},
	Definition: mcp.NewTool("get_release_candidate_upload_status",
		mcp.WithDescription("Gives back the store upload status of the release candidate of an app version: 'upload_state', the 'upload_build_url' of the Bitrise build doing the upload, the 'release_candidate_id' and 'release_candidate_build_slug', and for iOS apps the App Store Connect 'processing_state' of the uploaded build."),
		mcp.WithString("app_version_id",
			mcp.Description(appVersionIDDescription),
			mcp.Required(),
		),
		mcp.WithTitleAnnotation("Get Release Candidate Upload Status"),
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
			Path:    fmt.Sprintf("/app-versions/%s/release-candidate/upload/status", appVersionID),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
