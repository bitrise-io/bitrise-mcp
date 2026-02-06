package workspaces

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var CreateWorkspaceGroup = bitrise.Tool{
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
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
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

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/organizations/%s/groups", workspaceSlug),
			Body: map[string]any{
				"name": groupName,
			},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
