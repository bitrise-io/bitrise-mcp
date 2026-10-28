package insights

import (
	"context"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/insights"
	"github.com/mark3labs/mcp-go/mcp"
)

var ListFlakyTests = bitrise.Tool{
	APIGroups: []string{"insights", "read-only"},
	Definition: mcp.NewTool("insights_list_flaky_tests",
		toolOptions(
			"List the flaky test cases of a project from Bitrise Insights: test cases that were rerun after a flaky result "+
				"in the time window, one page at a time, ranked by rerun count or flaky rate. "+
				windowNote+" The window may span at most 2 years."+planNote,
			scopeOptions(),
			[]mcp.ToolOption{
				mcp.WithArray("branch",
					mcp.WithStringItems(),
					mcp.Description("Keep only test runs from builds on these git branches. '*' is a wildcard: release/* matches every release branch"),
				),
				mcp.WithArray("workflow",
					mcp.WithStringItems(),
					mcp.Description("Keep only test runs from builds of these workflows"),
				),
				testSuiteOption,
				mcp.WithArray("test_case",
					mcp.WithStringItems(),
					mcp.Description("Keep only these test cases"),
				),
				mcp.WithArray("module",
					mcp.WithStringItems(),
					mcp.Description("Keep only test cases of these modules"),
				),
				mcp.WithString("order_by",
					mcp.Description("Ranking: flaky_rerun_count (default, absolute number of flaky reruns) or flaky_rate (percentage of runs that were flaky)"),
					mcp.Enum("flaky_rerun_count", "flaky_rate"),
					mcp.DefaultString("flaky_rerun_count"),
				),
				mcp.WithString("order",
					mcp.Description("Direction of the ranking. Default: desc"),
					mcp.Enum("asc", "desc"),
					mcp.DefaultString("desc"),
				),
				mcp.WithNumber("page",
					mcp.Description("Page of the result set, 1-based (default: 1)"),
					mcp.Min(1),
				),
				mcp.WithNumber("per_page",
					mcp.Description("Test cases per page, 1-100 (default: 20)"),
					mcp.Min(1),
					mcp.Max(100),
				),
			},
			readOnlyAnnotations("List Insights Flaky Tests"),
			[]mcp.ToolOption{mcp.WithOutputSchema[insights.FlakyTestsResponse]()},
		)...,
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		workspaceSlug, project, window, errResult := scope(request)
		if errResult != nil {
			return errResult, nil
		}

		res, err := insights.ListFlakyTests(ctx, insights.FlakyTestsRequest{
			WorkspaceSlug: workspaceSlug,
			Project:       project,
			Window:        window,
			Branches:      request.GetStringSlice("branch", nil),
			Workflows:     request.GetStringSlice("workflow", nil),
			TestSuites:    request.GetStringSlice("test_suite", nil),
			TestCases:     request.GetStringSlice("test_case", nil),
			Modules:       request.GetStringSlice("module", nil),
			OrderBy:       request.GetString("order_by", ""),
			Order:         request.GetString("order", ""),
			Page:          request.GetInt("page", 0),
			PerPage:       request.GetInt("per_page", 0),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call insights api", err), nil
		}
		return mcp.NewToolResultStructuredOnly(res), nil
	},
}
