package insights

import (
	"context"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/insights"
	"github.com/mark3labs/mcp-go/mcp"
)

var GetBuildTotals = bitrise.Tool{
	APIGroups: []string{"insights", "read-only"},
	Definition: mcp.NewTool("get_insights_build_totals",
		toolOptions(
			"Get build aggregates for a project from Bitrise Insights: build count, failure rate (0-100 %), "+
				"p50/p90 duration and total duration in seconds over the whole time window, without grouping. "+
				windowNote+" The window may span at most 2 years."+planNote,
			scopeOptions(),
			buildFilterOptions("builds"),
			readOnlyAnnotations("Get Insights Build Totals"),
		)...,
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		workspaceSlug, project, window, errResult := scope(request)
		if errResult != nil {
			return errResult, nil
		}

		res, err := insights.GetBuildTotals(ctx, insights.BuildTotalsRequest{
			WorkspaceSlug: workspaceSlug,
			Project:       project,
			Window:        window,
			BuildFilters:  buildFilters(request),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call insights api", err), nil
		}
		return mcp.NewToolResultStructuredOnly(res), nil
	},
}
