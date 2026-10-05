package storereleases

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var ListWhatToTestDescriptions = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management", "read-only"},
	Definition: mcp.NewTool("list_what_to_test_descriptions",
		mcp.WithDescription("Lists the TestFlight \"What to Test\" descriptions of the uploaded release candidate of an iOS app version, one per locale, along with the app's 'primary_locale'. "+
			"Each localization has an 'id' (used by 'update_what_to_test_description'), a 'locale' and the 'whats_new' text. "+
			"iOS only, and the release candidate must already be uploaded to the store."),
		mcp.WithString("app_version_id",
			mcp.Description(appVersionIDDescription),
			mcp.Required(),
		),
		mcp.WithTitleAnnotation("List What to Test Descriptions"),
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
			Path:    fmt.Sprintf("/app-versions/%s/beta-distribution/what-to-test", appVersionID),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
