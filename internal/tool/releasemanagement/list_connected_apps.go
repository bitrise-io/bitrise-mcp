package releasemanagement

import (
	"context"
	"net/http"
	"strconv"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var ListConnectedApps = bitrise.Tool{
	APIGroups: []string{"release-management", "read-only"},
	Definition: mcp.NewTool("list_connected_apps",
		mcp.WithDescription("List Release Management connected apps available for the authenticated account within a workspace."),
		mcp.WithString("workspace_slug",
			mcp.Description("Identifier of the Bitrise workspace for the Release Management connected apps. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithString("project_id",
			mcp.Description("Specifies which Bitrise Project you want to get associated connected apps for"),
		),
		mcp.WithString("platform",
			mcp.Description("Filters for a specific mobile platform for the list of connected apps. Available values are: 'ios' and 'android'."),
			mcp.Enum("ios", "android"),
		),
		mcp.WithString("search",
			mcp.Description("Search by bundle ID (for ios), package name (for android), or app title (for both platforms). The filter is case-sensitive."),
		),
		mcp.WithNumber("items_per_page",
			mcp.Description("Specifies the maximum number of connected apps returned per page. Default value is 10."),
			mcp.DefaultNumber(10),
		),
		mcp.WithNumber("page",
			mcp.Description("Specifies which page should be returned from the whole result set in a paginated scenario. Default value is 1."),
			mcp.DefaultNumber(1),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		workspaceSlug, err := request.RequireString("workspace_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		params := map[string]any{
			"workspace_slug": workspaceSlug,
		}
		if v := request.GetString("project_id", ""); v != "" {
			params["project_id"] = v
		}
		if v := request.GetString("platform", ""); v != "" {
			params["platform"] = v
		}
		if v := request.GetString("search", ""); v != "" {
			params["search"] = v
		}
		if v := request.GetInt("items_per_page", 10); v != 10 {
			params["items_per_page"] = strconv.Itoa(v)
		}
		if page := request.GetInt("page", 1); page != 1 {
			params["page"] = strconv.Itoa(page)
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIRMBaseURL,
			Path:    "/connected-apps",
			Params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
