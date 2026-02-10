package apps

import (
	"context"
	"net/http"
	"strconv"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var List = bitrise.Tool{
	APIGroups: []string{"apps", "read-only"},
	Definition: mcp.NewTool("list_apps",
		mcp.WithDescription("List all apps for the currently authenticated user account"),
		mcp.WithString("sort_by",
			mcp.Description("Order of the apps. If set, you should accept the response as sorted"),
			mcp.Enum("last_build_at", "created_at"),
			mcp.DefaultString("last_build_at"),
		),
		mcp.WithString("next",
			mcp.Description("Slug of the first app in the response"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Max number of elements per page (default: 50)"),
			mcp.DefaultNumber(50),
		),
		mcp.WithString("title",
			mcp.Description("Filter apps by title"),
		),
		mcp.WithString("project_type",
			mcp.Description("Filter apps by project type (e.g., 'ios', 'android')"),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params := map[string]any{
			"sort_by": request.GetString("sort_by", "last_build_at"),
			"limit":   strconv.Itoa(request.GetInt("limit", 50)),
		}
		if v := request.GetString("next", ""); v != "" {
			params["next"] = v
		}
		if v := request.GetString("title", ""); v != "" {
			params["title"] = v
		}
		if v := request.GetString("project_type", ""); v != "" {
			params["project_type"] = v
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIBaseURL,
			Path:    "/apps",
			Params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
