package insights

import (
	"context"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/insights"
	"github.com/mark3labs/mcp-go/mcp"
)

var GetBuildSeries = bitrise.Tool{
	APIGroups: []string{"insights", "read-only"},
	Definition: mcp.NewTool("insights_get_build_series",
		toolOptions(
			"Get a build metrics time series for a project from Bitrise Insights: one row per time bucket, oldest first, "+
				"each with build count, failure rate (0-100 %), p50/p90 duration and total duration in seconds. "+
				"Use 'group_by' to also split by workflow, pipeline, stage or step. "+
				windowNote+seriesWindowNote+planNote,
			scopeOptions(),
			buildFilterOptions("builds"),
			seriesOptions("workflow", "pipeline", "stage", "step"),
			readOnlyAnnotations("Get Insights Build Series"),
		)...,
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		workspaceSlug, project, window, errResult := scope(request)
		if errResult != nil {
			return errResult, nil
		}
		granularity, groupBy, limit := seriesParams(request)

		res, err := insights.GetBuildSeries(ctx, insights.BuildSeriesRequest{
			BuildTotalsRequest: insights.BuildTotalsRequest{
				WorkspaceSlug: workspaceSlug,
				Project:       project,
				Window:        window,
				BuildFilters:  buildFilters(request),
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
