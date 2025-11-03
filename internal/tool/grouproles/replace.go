package grouproles

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var Replace = bitrise.Tool{
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

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPut,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/roles/%s", appSlug, roleName),
			Body: map[string]any{
				"groups": groupSlugs,
			},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
