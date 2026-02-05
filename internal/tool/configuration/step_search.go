package configuration

import (
	"context"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var StepSearch = bitrise.Tool{
	APIGroups: []string{"configuration", "read-only"},
	Definition: mcp.NewTool("step_search",
		mcp.WithDescription("Find steps for building workflows or step bundles in a Bitrise YML config file. Finds steps based on name, description, tags or maintainers."),
		mcp.WithString("query",
			mcp.Description("The phrase to search steps for like `clone`, `npm`, `deploy` etc."),
			mcp.Required(),
		),
		mcp.WithArray("categories",
			// Available categories are documented here: https://docs.bitrise.io/en/bitrise-ci/workflows-and-pipelines/developing-your-own-bitrise-step/developing-a-new-step.html#category
			mcp.WithStringEnumItems([]string{"build", "code-sign", "test", "deploy", "notification", "access-control", "artifact-info", "installer", "dependency", "utility"}),
			mcp.Description("Categories to filter steps."),
		),
		mcp.WithArray("maintainers",
			// Available values are listed here: https://github.com/bitrise-io/bitrise-workflow-editor/blob/master/source/javascripts/core/models/Step.ts#L6
			mcp.WithStringEnumItems([]string{"bitrise", "verified", "community"}),
			mcp.Description("Filter steps by maintainers. Use `bitrise` to only look for official steps."),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := request.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		categories := request.GetStringSlice("categories", []string{})
		maintainers := request.GetStringSlice("maintainers", []string{})

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIBaseURL,
			Path:    "/search-steps",
			Params: map[string]any{
				"query":       query,
				"categories":  categories,
				"maintainers": maintainers,
			},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
