package releasemanagement

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var ListBuildDistributionVersions = bitrise.Tool{
	APIGroups: []string{"release-management", "read-only"},
	Definition: mcp.NewTool("list_build_distribution_versions",
		mcp.WithDescription("Lists Build Distribution versions. Release Management offers a convenient, secure solution to distribute the builds of your mobile apps to testers without having to engage with either TestFlight or Google Play. Once you have installable artifacts, Bitrise can generate both private and public install links that testers or other stakeholders can use to install the app on real devices via over-the-air installation. Build distribution allows you to define tester groups that can receive notifications about installable artifacts. The email takes the notified testers to the test build page, from where they can install the app on their own device. Build distribution versions are the app versions available for testers."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier of the app the build distribution is connected to. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithNumber("items_per_page",
			mcp.Description("Specifies the maximum number of build distribution versions returned per page. Default value is 10."),
			mcp.DefaultNumber(10),
		),
		mcp.WithNumber("page",
			mcp.Description("Specifies which page should be returned from the whole result set in a paginated scenario. Default value is 1."),
			mcp.DefaultNumber(1),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connectedAppID, err := request.RequireString("connected_app_id")
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

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIRMBaseURL,
			Path:    fmt.Sprintf("/connected-apps/%s/build-distributions", connectedAppID),
			Params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
