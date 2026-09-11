package insights

import (
	"context"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/insights"
	"github.com/mark3labs/mcp-go/mcp"
)

var testSuiteOption = mcp.WithArray("test_suite",
	mcp.WithStringItems(),
	mcp.Description("Keep only runs of these test suites"),
)

var GetTestTotals = bitrise.Tool{
	APIGroups: []string{"insights", "read-only"},
	Definition: mcp.NewTool("get_insights_test_totals",
		toolOptions(
			"Get test aggregates for a project from Bitrise Insights: test run count, failure rate (0-100 %), "+
				"p50/p90 duration and total duration in seconds, and flaky run count over the whole time window, without grouping. "+
				"Counts are of test case runs, not of distinct test cases. "+
				windowNote+" The window may span at most 2 years."+planNote,
			scopeOptions(),
			buildFilterOptions("test runs from builds"),
			[]mcp.ToolOption{testSuiteOption},
			readOnlyAnnotations("Get Insights Test Totals"),
		)...,
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		workspaceSlug, project, window, errResult := scope(request)
		if errResult != nil {
			return errResult, nil
		}

		res, err := insights.GetTestTotals(ctx, insights.TestTotalsRequest{
			WorkspaceSlug: workspaceSlug,
			Project:       project,
			Window:        window,
			TestFilters: insights.TestFilters{
				BuildFilters: buildFilters(request),
				TestSuites:   request.GetStringSlice("test_suite", nil),
			},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call insights api", err), nil
		}
		return mcp.NewToolResultStructuredOnly(res), nil
	},
}
