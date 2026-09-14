package insights

import (
	"context"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/insights"
	"github.com/mark3labs/mcp-go/mcp"
)

var GetTestSeries = bitrise.Tool{
	APIGroups: []string{"insights", "read-only"},
	Definition: mcp.NewTool("insights_get_test_series",
		toolOptions(
			"Get a test metrics time series for a project from Bitrise Insights: one row per time bucket, oldest first, "+
				"each with test run count, failure rate (0-100 %), p50/p90 duration and total duration in seconds, and flaky run count. "+
				"Use 'group_by' to also split by workflow, pipeline, stage, branch or test_suite. "+
				windowNote+seriesWindowNote+planNote,
			scopeOptions(),
			buildFilterOptions("test runs from builds"),
			[]mcp.ToolOption{testSuiteOption},
			seriesOptions("workflow", "pipeline", "stage", "branch", "test_suite"),
			readOnlyAnnotations("Get Insights Test Series"),
		)...,
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		workspaceSlug, project, window, errResult := scope(request)
		if errResult != nil {
			return errResult, nil
		}
		granularity, groupBy, limit := seriesParams(request)

		res, err := insights.GetTestSeries(ctx, insights.TestSeriesRequest{
			TestTotalsRequest: insights.TestTotalsRequest{
				WorkspaceSlug: workspaceSlug,
				Project:       project,
				Window:        window,
				TestFilters: insights.TestFilters{
					BuildFilters: buildFilters(request),
					TestSuites:   request.GetStringSlice("test_suite", nil),
				},
			},
			Granularity: granularity,
			GroupBy:     groupBy,
			Limit:       limit,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call insights api", err), nil
		}
		return mcp.NewToolResultStructuredOnly(res), nil
	},
}
