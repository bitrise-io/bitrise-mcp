package storereleases

import (
	"context"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var GetAppStoreDraftVersion = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management", "read-only"},
	Definition: mcp.NewTool("get_app_store_draft_version",
		mcp.WithDescription("Gives back the App Store Connect draft version of an iOS connected app: the latest version that is not released yet and so can be submitted for App Store review. "+
			"Returns its 'version_string', 'app_store_state', 'release_type', phased release state and the matching Release Management 'app_version_id' (null when no app version of that name exists). "+
			"Fails with 404 when there is no unreleased version; one can then be created with 'create_app_store_version'. iOS only."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier of the connected app."),
			mcp.Required(),
		),
		mcp.WithTitleAnnotation("Get App Store Draft Version"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connectedAppID, err := request.RequireString("connected_app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIRMAppsBaseURL,
			Path:    "/apple-app-store/app-versions/draft",
			Params: map[string]any{
				"app_id": connectedAppID,
			},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
