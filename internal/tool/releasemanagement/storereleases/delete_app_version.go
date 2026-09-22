package storereleases

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var DeleteAppVersion = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management"},
	Definition: mcp.NewTool("delete_app_version",
		mcp.WithDescription("Deletes a store release app version (a release in the Release Management UI) with its stages, approval tasks, automations and event history. "+
			"This cannot be undone and does not touch anything already uploaded to or published on the store, so always confirm with the user before calling. "+
			"To stop managing a release while keeping its record, set its 'status' with 'update_app_version' instead."),
		mcp.WithString("app_version_id",
			mcp.Description(appVersionIDDescription),
			mcp.Required(),
		),
		mcp.WithTitleAnnotation("Delete App Version"),
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

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodDelete,
			BaseURL: bitrise.APIRMStoreReleasesBaseURL,
			Path:    fmt.Sprintf("/app-versions/%s", appVersionID),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
