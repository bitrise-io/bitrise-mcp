package storereleases

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var UpdateGooglePlayReleaseNotes = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management"},
	Definition: mcp.NewTool("update_google_play_release_notes",
		mcp.WithDescription("Sets the Google Play release notes of the uploaded release candidate of an Android app version, per language. "+
			"Android only, and a release candidate must already be chosen (400 otherwise)."),
		mcp.WithString("app_version_id",
			mcp.Description(appVersionIDDescription),
			mcp.Required(),
		),
		mcp.WithArray("release_notes",
			mcp.Description("The release notes, one item per language, each with a 'language' code (a Google Play supported language, e.g. 'en-US') and the 'text' (max 500 characters)."),
			mcp.Required(),
			mcp.Items(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"language": map[string]any{"type": "string"},
					"text":     map[string]any{"type": "string"},
				},
				"required": []string{"language", "text"},
			}),
		),
		mcp.WithTitleAnnotation("Update Google Play Release Notes"),
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
		releaseNotes, err := requireArray(request, "release_notes")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPut,
			BaseURL: bitrise.APIRMStoreReleasesBaseURL,
			Path:    fmt.Sprintf("/app-versions/%s/google-play-store/localized-release-notes", appVersionID),
			Body: map[string]any{
				"release_notes": releaseNotes,
			},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
