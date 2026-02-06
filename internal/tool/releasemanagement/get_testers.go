package releasemanagement

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var GetTesters = bitrise.Tool{
	APIGroups: []string{"release-management", "read-only"},
	Definition: mcp.NewTool("get_testers",
		mcp.WithDescription("Gives back a list of testers that has been associated with a tester group related to a specific connected app."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier of the app the tester group is connected to. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithString("tester_group_id",
			mcp.Description("The uuidV4 identifier of a tester group. If given, only testers within this specific tester group will be returned."),
		),
		mcp.WithNumber("items_per_page",
			mcp.Description("Specifies the maximum number of testers to be returned that have been added to a tester group related to the specific connected app. Default value is 10."),
			mcp.DefaultNumber(10),
		),
		mcp.WithNumber("page",
			mcp.Description("Specifies which page should be returned from the whole result set in a paginated scenario. Default value is 1."),
			mcp.DefaultNumber(1),
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

		params := map[string]any{}
		if v := request.GetString("tester_group_id", ""); v != "" {
			params["tester_group_id"] = v
		}
		if v := request.GetInt("items_per_page", 10); v != 10 {
			params["items_per_page"] = strconv.Itoa(v)
		}
		if v := request.GetInt("page", 1); v != 1 {
			params["page"] = strconv.Itoa(v)
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIRMBaseURL,
			Path:    fmt.Sprintf("/connected-apps/%s/testers", connectedAppID),
			Params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
