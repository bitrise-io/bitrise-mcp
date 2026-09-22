package storereleases

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var StartBetaTesting = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management"},
	Definition: mcp.NewTool("start_beta_testing",
		mcp.WithDescription("Distributes the uploaded release candidate of an app version to a store beta testing target: a TestFlight testing group for iOS apps, or a Google Play testing track for Android apps. "+
			"The release candidate must already be uploaded to the store. Find the targets with 'list_beta_testing_groups'."),
		mcp.WithString("app_version_id",
			mcp.Description(appVersionIDDescription),
			mcp.Required(),
		),
		mcp.WithString("group_id",
			mcp.Description("The TestFlight testing group id (iOS) or the Google Play testing track name (Android), as listed by 'list_beta_testing_groups'."),
			mcp.Required(),
		),
		mcp.WithTitleAnnotation("Start Beta Testing"),
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
		groupID, err := request.RequireString("group_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPatch,
			BaseURL: bitrise.APIRMStoreReleasesBaseURL,
			Path:    fmt.Sprintf("/app-versions/%s/beta-distribution/testing-groups/%s", appVersionID, url.PathEscape(groupID)),
			Body:    map[string]any{},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
