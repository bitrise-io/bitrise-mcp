package storereleases

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var SubmitReleaseCandidateForBetaReview = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management"},
	Definition: mcp.NewTool("submit_release_candidate_for_beta_review",
		mcp.WithDescription("Submits the uploaded release candidate of an iOS app version for TestFlight beta app review, which external testing groups require before they can receive the build. "+
			"iOS only: Android app versions are rejected with 422 ERR_INVALID_PLATFORM."),
		mcp.WithString("app_version_id",
			mcp.Description(appVersionIDDescription),
			mcp.Required(),
		),
		mcp.WithBoolean("auto_notify_enabled",
			mcp.Description("Whether TestFlight should notify testers automatically once the build is approved. Defaults to false."),
		),
		mcp.WithTitleAnnotation("Submit Release Candidate for Beta Review"),
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

		body := map[string]any{}
		v, ok, err := optionalBool(request, "auto_notify_enabled")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if ok {
			body["auto_notify_enabled"] = v
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APIRMStoreReleasesBaseURL,
			Path:    fmt.Sprintf("/app-versions/%s/release-candidate/submit-for-beta-review", appVersionID),
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
