package storereleases

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var GetGooglePlayRelease = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management", "read-only"},
	Definition: mcp.NewTool("get_google_play_release",
		mcp.WithDescription("Gives back the current Google Play production rollout of an Android app version as its 'user_fraction': a number between 0 and 1 where 1 is a full release, or null when the release candidate is not on the production track. "+
			"Android only, and a release candidate must already be chosen (400 otherwise)."),
		mcp.WithString("app_version_id",
			mcp.Description(appVersionIDDescription),
			mcp.Required(),
		),
		mcp.WithTitleAnnotation("Get Google Play Release"),
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
			Path:    fmt.Sprintf("/app-versions/%s/google-play-store/release", appVersionID),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
