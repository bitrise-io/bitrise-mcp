package storereleases

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var CreateWhatToTestDescription = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management"},
	Definition: mcp.NewTool("create_what_to_test_description",
		mcp.WithDescription("Adds a TestFlight \"What to Test\" description in a new locale to the uploaded release candidate of an iOS app version. "+
			"Use 'update_what_to_test_description' for a locale that already has one. iOS only, and the release candidate must already be uploaded to the store."),
		mcp.WithString("app_version_id",
			mcp.Description(appVersionIDDescription),
			mcp.Required(),
		),
		mcp.WithString("locale",
			mcp.Description("The locale of the description, e.g. 'en-US'. See Apple's supported App Store Connect locale shortcodes."),
			mcp.Required(),
		),
		mcp.WithString("whats_new",
			mcp.Description("The \"What to Test\" text shown to testers in that locale."),
			mcp.Required(),
		),
		mcp.WithTitleAnnotation("Create What to Test Description"),
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
		locale, err := request.RequireString("locale")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		whatsNew, err := request.RequireString("whats_new")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APIRMStoreReleasesBaseURL,
			Path:    fmt.Sprintf("/app-versions/%s/beta-distribution/what-to-test", appVersionID),
			Body: map[string]any{
				"locale":    locale,
				"whats_new": whatsNew,
			},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
