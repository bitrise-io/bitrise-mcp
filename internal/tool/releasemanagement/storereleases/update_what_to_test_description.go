package storereleases

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var UpdateWhatToTestDescription = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management"},
	Definition: mcp.NewTool("update_what_to_test_description",
		mcp.WithDescription("Replaces the text of an existing TestFlight \"What to Test\" description of the uploaded release candidate of an iOS app version. "+
			"Find the description ids with 'list_what_to_test_descriptions'. iOS only."),
		mcp.WithString("app_version_id",
			mcp.Description(appVersionIDDescription),
			mcp.Required(),
		),
		mcp.WithString("what_to_test_id",
			mcp.Description("The identifier of the \"What to Test\" description (the localization 'id' from 'list_what_to_test_descriptions')."),
			mcp.Required(),
		),
		mcp.WithString("whats_new",
			mcp.Description("The new \"What to Test\" text shown to testers in that locale."),
			mcp.Required(),
		),
		mcp.WithTitleAnnotation("Update What to Test Description"),
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
		whatsNew, err := request.RequireString("whats_new")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPatch,
			BaseURL: bitrise.APIRMStoreReleasesBaseURL,
			Path:    fmt.Sprintf("/app-versions/%s/beta-distribution/what-to-test/%s", appVersionID, whatToTestID),
			Body: map[string]any{
				"whats_new": whatsNew,
			},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
