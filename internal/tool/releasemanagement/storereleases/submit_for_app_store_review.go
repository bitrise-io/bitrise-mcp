package storereleases

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var SubmitForAppStoreReview = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management"},
	Definition: mcp.NewTool("submit_for_app_store_review",
		mcp.WithDescription("Submits the uploaded release candidate of an iOS app version to App Store review, optionally setting the App Store metadata of the version per locale first. "+
			"Confirm with the user before calling: if the App Store release type of the version is AFTER_APPROVAL (see 'get_app_store_release_settings'), Apple publishes the version to the App Store as soon as the review passes. "+
			"Follow the review with 'get_app_store_review_status'. iOS only."),
		mcp.WithString("app_version_id",
			mcp.Description(appVersionIDDescription),
			mcp.Required(),
		),
		mcp.WithBoolean("copy_primary_whats_new",
			mcp.Description("If true, the 'whats_new' of the primary localization is copied into every other localization whose 'whats_new' is empty. Defaults to false."),
		),
		mcp.WithArray("localizations",
			mcp.Description("App Store metadata to set per locale before submitting. Each item needs a 'locale' and may carry 'whats_new' (max 4000 characters), 'description' (max 4000), 'keywords' (max 100), 'promotional_text' (max 170), 'marketing_url' and 'support_url'."),
			mcp.Items(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"locale":           map[string]any{"type": "string", "description": "An App Store Connect locale shortcode, e.g. 'en-US'."},
					"whats_new":        map[string]any{"type": "string"},
					"description":      map[string]any{"type": "string"},
					"keywords":         map[string]any{"type": "string"},
					"promotional_text": map[string]any{"type": "string"},
					"marketing_url":    map[string]any{"type": "string"},
					"support_url":      map[string]any{"type": "string"},
				},
				"required": []string{"locale"},
			}),
		),
		mcp.WithTitleAnnotation("Submit for App Store Review"),
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

		body := map[string]any{}
		v, ok, err := optionalBool(request, "copy_primary_whats_new")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if ok {
			body["copy_primary_whats_new"] = v
		}
		localizations, ok, err := optionalArray(request, "localizations")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if ok {
			body["localizations"] = localizations
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APIRMStoreReleasesBaseURL,
			Path:    fmt.Sprintf("/app-versions/%s/apple-app-store/review/submit", appVersionID),
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
