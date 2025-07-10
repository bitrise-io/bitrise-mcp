package tool

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
)

var listGroupRoles = Tool{
	APIGroups: []string{"group-roles", "read-only"},
	Definition: mcp.NewTool("list_group_roles",
		mcp.WithDescription("List group roles for an app"),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("role_name",
			mcp.Description("Name of the role"),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		roleName, err := request.RequireString("role_name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodGet,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/apps/%s/roles/%s", appSlug, roleName),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var replaceGroupRoles = Tool{
	APIGroups: []string{"group-roles"},
	Definition: mcp.NewTool("replace_group_roles",
		mcp.WithDescription("Replace group roles for an app."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("role_name",
			mcp.Description("Name of the role"),
			mcp.Required(),
		),
		mcp.WithArray("group_slugs",
			mcp.Description("List of group slugs"),
			mcp.Required(),
			mcp.WithStringItems(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		roleName, err := request.RequireString("role_name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		groupSlugs, err := request.RequireStringSlice("group_slugs")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodPut,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/apps/%s/roles/%s", appSlug, roleName),
			body: map[string]any{
				"groups": groupSlugs,
			},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
