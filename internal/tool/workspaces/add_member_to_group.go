package workspaces

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var AddMemberToGroup = bitrise.Tool{
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

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPut,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/groups/%s/members/%s", groupSlug, userSlug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
