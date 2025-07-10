package tool

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
)

var listWorkspaces = Tool{
	APIGroups: []string{"workspaces", "read-only"},
	Definition: mcp.NewTool("list_workspaces",
		mcp.WithDescription("List the workspaces the user has access to"),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodGet,
			baseURL: apiBaseURL,
			path:    "/organizations",
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var getWorkspace = Tool{
	APIGroups: []string{"workspaces", "read-only"},
	Definition: mcp.NewTool("get_workspace",
		mcp.WithDescription("Get details for one workspace"),
		mcp.WithString("workspace_slug",
			mcp.Description("Slug of the Bitrise workspace"),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		workspaceSlug, err := request.RequireString("workspace_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodGet,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/organizations/%s", workspaceSlug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var getWorkspaceGroups = Tool{
	APIGroups: []string{"workspaces", "read-only"},
	Definition: mcp.NewTool("get_workspace_groups",
		mcp.WithDescription("Get the groups in a workspace"),
		mcp.WithString("workspace_slug",
			mcp.Description("Slug of the Bitrise workspace"),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		workspaceSlug, err := request.RequireString("workspace_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodGet,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/organizations/%s/groups", workspaceSlug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var createWorkspaceGroup = Tool{
	APIGroups: []string{"workspaces"},
	Definition: mcp.NewTool("create_workspace_group",
		mcp.WithDescription("Create a new group in a workspace."),
		mcp.WithString("workspace_slug",
			mcp.Description("Slug of the Bitrise workspace"),
			mcp.Required(),
		),
		mcp.WithString("group_name",
			mcp.Description("Name of the group"),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		workspaceSlug, err := request.RequireString("workspace_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		groupName, err := request.RequireString("group_name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodPost,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/organizations/%s/groups", workspaceSlug),
			body: map[string]any{
				"name": groupName,
			},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var getWorkspaceMembers = Tool{
	APIGroups: []string{"workspaces", "read-only"},
	Definition: mcp.NewTool("get_workspace_members",
		mcp.WithDescription("Get the members of a workspace"),
		mcp.WithString("workspace_slug",
			mcp.Description("Slug of the Bitrise workspace"),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		workspaceSlug, err := request.RequireString("workspace_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodGet,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/organizations/%s/members", workspaceSlug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var inviteMemberToWorkspace = Tool{
	APIGroups: []string{"workspaces"},
	Definition: mcp.NewTool("invite_member_to_workspace",
		mcp.WithDescription("Invite new Bitrise users to a workspace."),
		mcp.WithString("workspace_slug",
			mcp.Description("Slug of the Bitrise workspace"),
			mcp.Required(),
		),
		mcp.WithString("email",
			mcp.Description("Email address of the user"),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		workspaceSlug, err := request.RequireString("workspace_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		email, err := request.RequireString("email")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodPost,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/organizations/%s/members", workspaceSlug),
			body: map[string]any{
				"email": email,
			},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var addMemberToGroup = Tool{
	APIGroups: []string{"workspaces"},
	Definition: mcp.NewTool("add_member_to_group",
		mcp.WithDescription("Add a member to a group."),
		mcp.WithString("group_slug",
			mcp.Description("Slug of the group"),
			mcp.Required(),
		),
		mcp.WithString("user_slug",
			mcp.Description("Slug of the user"),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		groupSlug, err := request.RequireString("group_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		userSlug, err := request.RequireString("user_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodPut,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/groups/%s/members/%s", groupSlug, userSlug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
