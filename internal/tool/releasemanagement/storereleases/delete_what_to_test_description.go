package storereleases

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var DeleteWhatToTestDescription = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management"},
	Definition: mcp.NewTool("delete_what_to_test_description",
		mcp.WithDescription("Removes a TestFlight \"What to Test\" description (one locale) from the uploaded release candidate of an iOS app version. "+
			"Find the description ids with 'list_what_to_test_descriptions'. This cannot be undone. iOS only."),
		mcp.WithString("app_version_id",
			mcp.Description(appVersionIDDescription),
			mcp.Required(),
		),
		mcp.WithString("what_to_test_id",
			mcp.Description("The identifier of the \"What to Test\" description (the localization 'id' from 'list_what_to_test_descriptions')."),
			mcp.Required(),
		),
		mcp.WithTitleAnnotation("Delete What to Test Description"),
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
		whatToTestID, err := request.RequireString("what_to_test_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodDelete,
			BaseURL: bitrise.APIRMStoreReleasesBaseURL,
			Path:    fmt.Sprintf("/app-versions/%s/beta-distribution/what-to-test/%s", appVersionID, whatToTestID),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
