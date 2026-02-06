package releasemanagement

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var GetPotentialTesters = bitrise.Tool{
	APIGroups: []string{"release-management", "read-only"},
	Definition: mcp.NewTool("get_potential_testers",
		mcp.WithDescription("Gets a list of potential testers whom can be added as testers to a specific tester group. The list consists of Bitrise users having access to the related Release Management connected app."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier of the app the tester group is connected to. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithString("id",
			mcp.Description("The uuidV4 identifier of the tester group. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithNumber("items_per_page",
			mcp.Description("Specifies the maximum number of potential testers to return having access to a specific connected app. Default value is 10."),
			mcp.DefaultNumber(10),
		),
		mcp.WithNumber("page",
			mcp.Description("Specifies which page should be returned from the whole result set in a paginated scenario. Default value is 1."),
			mcp.DefaultNumber(1),
		),
		mcp.WithString("search",
			mcp.Description("Searches for potential testers based on email or username using a case-insensitive approach."),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connectedAppID, err := request.RequireString("connected_app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		params := map[string]any{}
		if v := request.GetInt("items_per_page", 10); v != 10 {
			params["items_per_page"] = strconv.Itoa(v)
		}
		if v := request.GetInt("page", 1); v != 1 {
			params["page"] = strconv.Itoa(v)
		}
		if v := request.GetString("search", ""); v != "" {
			params["search"] = v
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIRMBaseURL,
			Path:    fmt.Sprintf("/connected-apps/%s/tester-groups/%s/potential-testers", connectedAppID, id),
			Params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
