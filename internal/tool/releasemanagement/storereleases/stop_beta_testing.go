package storereleases

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var StopBetaTesting = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management"},
	Definition: mcp.NewTool("stop_beta_testing",
		mcp.WithDescription("Stops the TestFlight beta testing of the release candidate of an iOS app version on a testing group, removing the build from that group so its testers no longer receive it. "+
			"The reverse of 'start_beta_testing'. iOS only: Google Play testing tracks cannot be stopped this way (422 ERR_INVALID_PLATFORM)."),
		mcp.WithString("app_version_id",
			mcp.Description(appVersionIDDescription),
			mcp.Required(),
		),
		mcp.WithString("group_id",
			mcp.Description("The TestFlight testing group id, as listed by 'list_beta_testing_groups'."),
			mcp.Required(),
		),
		mcp.WithTitleAnnotation("Stop Beta Testing"),
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
		groupID, err := request.RequireString("group_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodDelete,
			BaseURL: bitrise.APIRMStoreReleasesBaseURL,
			Path:    fmt.Sprintf("/app-versions/%s/beta-distribution/testing-groups/%s", appVersionID, url.PathEscape(groupID)),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
