package codepush

import (
	"context"
	"net/http"
	"strconv"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var ListDeployments = bitrise.Tool{
	APIGroups: []string{"release-management-code-push", "release-management", "read-only"},
	Definition: mcp.NewTool("codepush_list_deployments",
		mcp.WithDescription("List CodePush deployments for a Bitrise app."),
		mcp.WithString("app_id",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("search",
			mcp.Description("Search deployments by name. The filter is case-sensitive."),
		),
		mcp.WithNumber("items_per_page",
			mcp.Description("Maximum number of deployments returned per page. Default value is 10."),
			mcp.DefaultNumber(10),
		),
		mcp.WithNumber("page",
			mcp.Description("Page number to return from the paginated result set. Default value is 1."),
			mcp.DefaultNumber(1),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appID, err := request.RequireString("app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		params := map[string]any{
			"app_id": appID,
		}
		if v := request.GetString("search", ""); v != "" {
			params["search"] = v
		}
		if v := request.GetInt("items_per_page", 10); v != 10 {
			params["items_per_page"] = strconv.Itoa(v)
		}
		if v := request.GetInt("page", 1); v != 1 {
			params["page"] = strconv.Itoa(v)
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APICodePushBaseURL,
			Path:    "/deployments",
			Params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
