package storereleases

import (
	"context"
	"net/http"
	"strconv"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var ListAppVersions = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management", "read-only"},
	Definition: mcp.NewTool("list_app_versions",
		mcp.WithDescription("Lists the store release app versions of a Release Management connected app, newest first. "+
			"An app version is what the Release Management UI calls a release: it walks a release candidate through the store upload, beta testing, approval, store review and release stages. "+
			"Each item carries its 'id' (used as 'app_version_id' by the other store release tools), 'name' (the version string), 'status', 'platform', the release candidate selection and the per-stage statuses."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier of the connected app whose app versions are listed."),
			mcp.Required(),
		),
		mcp.WithNumber("items_per_page",
			mcp.Description("Specifies the maximum number of app versions returned per page. Default value is 10."),
			mcp.DefaultNumber(10),
		),
		mcp.WithNumber("page",
			mcp.Description("Specifies which page should be returned from the whole result set in a paginated scenario. Default value is 1."),
			mcp.DefaultNumber(1),
		),
		mcp.WithTitleAnnotation("List App Versions"),
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

		params := map[string]any{
			"app_id": connectedAppID,
		}
		if v := request.GetInt("items_per_page", 10); v != 10 {
			params["items_per_page"] = strconv.Itoa(v)
		}
		if v := request.GetInt("page", 1); v != 1 {
			params["page"] = strconv.Itoa(v)
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIRMStoreReleasesBaseURL,
			Path:    "/app-versions",
			Params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
