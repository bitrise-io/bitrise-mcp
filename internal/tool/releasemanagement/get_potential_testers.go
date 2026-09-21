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
		mcp.WithDescription("Lists the people who can be added to a tester group, with the slug each of them is added by. Pass 'workspace_slug' to list workspace members before a connected app or a tester group exists, or 'connected_app_id' with 'id' for the candidates of an existing group. "+
			"Service accounts are never listed. 'already_member' means 'already in that tester group', never 'already a tester on the app', and is always false on the workspace scope."),
		mcp.WithString("workspace_slug",
			mcp.Description("The slug of the workspace whose members are listed. Mutually exclusive with 'connected_app_id' and 'id'."),
		),
		mcp.WithString("project_id",
			mcp.Description("The uuidV4 identifier of a project in the workspace, used with 'workspace_slug'. When given, the project's outside contributors are listed alongside the workspace members."),
		),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier of the app the tester group is connected to. Required together with 'id'."),
		),
		mcp.WithString("id",
			mcp.Description("The uuidV4 identifier of the tester group. Required together with 'connected_app_id'."),
		),
		mcp.WithNumber("items_per_page",
			mcp.Description("Specifies the maximum number of potential testers to return per page. Between 1 and 1000, default value is 25."),
			mcp.DefaultNumber(25),
			mcp.Min(1),
			mcp.Max(1000),
		),
		mcp.WithNumber("page",
			mcp.Description("Specifies which page should be returned from the whole result set in a paginated scenario. Default value is 1."),
			mcp.DefaultNumber(1),
			mcp.Min(1),
		),
		mcp.WithString("search",
			mcp.Description("Filters the candidates by email or username, case-insensitively."),
		),
		mcp.WithTitleAnnotation("Get Potential Testers"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		workspaceSlug := request.GetString("workspace_slug", "")
		testerGroupID := request.GetString("id", "")
		connectedAppID := request.GetString("connected_app_id", "")
		projectID := request.GetString("project_id", "")

		params := map[string]any{}
		if v := request.GetInt("items_per_page", 25); v != 25 {
			params["items_per_page"] = strconv.Itoa(v)
		}
		if v := request.GetInt("page", 1); v != 1 {
			params["page"] = strconv.Itoa(v)
		}
		if v := request.GetString("search", ""); v != "" {
			params["search"] = v
		}

		var path string
		switch {
		case workspaceSlug != "":
			if connectedAppID != "" || testerGroupID != "" {
				return mcp.NewToolResultError("give either workspace_slug or connected_app_id with id, not both"), nil
			}
			path = "/potential-testers"
			params["workspace_slug"] = workspaceSlug
			if projectID != "" {
				params["project_id"] = projectID
			}
		case connectedAppID != "" && testerGroupID != "":
			if projectID != "" {
				return mcp.NewToolResultError("project_id only applies to workspace_slug"), nil
			}
			path = fmt.Sprintf("/tester-groups/%s/potential-testers", testerGroupID)
		default:
			return mcp.NewToolResultError("either workspace_slug, or connected_app_id with id, is required"), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIRMBuildDistributionsBaseURL,
			Path:    path,
			Params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
