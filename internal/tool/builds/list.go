package builds

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var List = bitrise.Tool{
	APIGroups: []string{"builds", "read-only"},
	Definition: mcp.NewTool("list_builds",
		mcp.WithDescription("List all the builds of a specified Bitrise app or all accessible builds."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
		),
		mcp.WithString("sort_by",
			mcp.Description("Order of builds: created_at (default), running_first"),
			mcp.DefaultString("created_at"),
			mcp.Enum("created_at", "running_first"),
		),
		mcp.WithString("branch",
			mcp.Description("Filter builds by branch"),
		),
		mcp.WithString("workflow",
			mcp.Description("Filter builds by workflow"),
		),
		mcp.WithNumber("status",
			mcp.Description("Filter builds by status (0: not finished, 1: successful, 2: failed, 3: aborted, 4: in-progress)"),
			mcp.Enum("0", "1", "2", "3", "4"),
		),
		mcp.WithString("next",
			mcp.Description("Slug of the first build in the response"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Max number of elements per page (default: 50)"),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params := map[string]any{}
		if v := request.GetString("sort_by", ""); v != "" {
			params["sort_by"] = v
		}
		if v := request.GetString("branch", ""); v != "" {
			params["branch"] = v
		}
		if v := request.GetString("workflow", ""); v != "" {
			params["workflow"] = v
		}
		if _, ok := request.GetArguments()["status"]; ok {
			params["status"] = strconv.Itoa(request.GetInt("status", 0))
		}
		if v := request.GetString("next", ""); v != "" {
			params["next"] = v
		}
		if _, ok := request.GetArguments()["limit"]; ok {
			params["limit"] = strconv.Itoa(request.GetInt("limit", 50))
		}

		path := "/builds"
		if appSlug := request.GetString("app_slug", ""); appSlug != "" {
			path = fmt.Sprintf("/apps/%s/builds", appSlug)
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIBaseURL,
			Path:    path,
			Params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
