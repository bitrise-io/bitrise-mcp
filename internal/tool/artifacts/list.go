package artifacts

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bitrise-io/bitrise-mcp/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var List = bitrise.Tool{
	APIGroups: []string{"build-artifacts", "read-only"},
	Definition: mcp.NewTool("list_artifacts",
		mcp.WithDescription("Get a list of all build artifacts."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("build_slug",
			mcp.Description("Identifier of the build"),
			mcp.Required(),
		),
		mcp.WithString("next",
			mcp.Description("Slug of the first artifact in the response"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Max number of elements per page (default: 50)"),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		buildSlug, err := request.RequireString("build_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		params := map[string]string{}
		if v := request.GetString("next", ""); v != "" {
			params["next"] = v
		}
		if _, ok := request.GetArguments()["limit"]; ok {
			params["limit"] = strconv.Itoa(request.GetInt("limit", 50))
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/builds/%s/artifacts", appSlug, buildSlug),
			Params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
