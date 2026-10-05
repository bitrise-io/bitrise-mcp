package storereleases

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var ReleaseToGooglePlay = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management"},
	Definition: mcp.NewTool("release_to_google_play",
		mcp.WithDescription("Releases the uploaded release candidate of an Android app version on the Google Play production track to the given fraction of users, or raises the fraction of a rollout already in progress. "+
			"This publishes to production and cannot be undone (a rollout can only be halted in the Play Console), so always confirm with the user before calling. "+
			"Android only, and a release candidate must already be chosen (400 otherwise)."),
		mcp.WithString("app_version_id",
			mcp.Description(appVersionIDDescription),
			mcp.Required(),
		),
		mcp.WithNumber("user_fraction",
			mcp.Description("The fraction of users to release to: greater than 0 and at most 1, where 1 is a full release (e.g. 0.1 for 10%)."),
			mcp.Required(),
			mcp.Max(1),
		),
		mcp.WithTitleAnnotation("Release to Google Play"),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appVersionID, err := request.RequireString("app_version_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		userFraction, err := request.RequireFloat("user_fraction")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if userFraction <= 0 || userFraction > 1 {
			return mcp.NewToolResultError("user_fraction must be greater than 0 and at most 1"), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APIRMStoreReleasesBaseURL,
			Path:    fmt.Sprintf("/app-versions/%s/google-play-store/release", appVersionID),
			Body: map[string]any{
				"user_fraction": userFraction,
			},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
