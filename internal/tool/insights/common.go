// Package insights exposes the public Bitrise Insights API as MCP tools:
// build and test metrics for one project of a workspace. It requires a paid
// Insights plan on the workspace.
package insights

import (
	"github.com/bitrise-io/bitrise-mcp/v2/internal/insights"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	planNote = " Requires a paid Bitrise Insights plan on the workspace. " +
		"The project is identified by its Insights project ID, not by the app slug."

	windowNote = "Time window: 'start' is inclusive and 'end' exclusive, both RFC3339 in UTC (e.g. 2026-07-01T00:00:00Z)."

	seriesWindowNote = " The window may span at most 7 days hourly, 90 days daily, 1 year weekly and 2 years monthly."
)

// scopeOptions are the parameters every Insights tool takes.
func scopeOptions() []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithString("workspace_slug",
			mcp.Description("Slug of the workspace the project belongs to"),
			mcp.Required(),
		),
		mcp.WithString("project",
			mcp.Description("ID of the Insights project to report on"),
			mcp.Required(),
		),
		mcp.WithString("start",
			mcp.Description("Start of the reported time window, inclusive, RFC3339 (e.g. 2026-07-01T00:00:00Z)"),
			mcp.Required(),
		),
		mcp.WithString("end",
			mcp.Description("End of the reported time window, exclusive, RFC3339. Must be after 'start'"),
			mcp.Required(),
		),
	}
}

// buildFilterOptions are the branch/workflow/pipeline filters of the build
// and test metrics tools.
func buildFilterOptions(subject string) []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithArray("branch",
			mcp.WithStringItems(),
			mcp.Description("Keep only "+subject+" on these git branches. '*' is a wildcard: release/* matches every release branch"),
		),
		mcp.WithArray("workflow",
			mcp.WithStringItems(),
			mcp.Description("Keep only "+subject+" of these workflows"),
		),
		mcp.WithArray("pipeline",
			mcp.WithStringItems(),
			mcp.Description("Keep only "+subject+" of these pipelines"),
		),
	}
}

func seriesOptions(groupByValues ...string) []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithString("granularity",
			mcp.Description("Size of one time bucket. Default: daily"),
			mcp.Enum("hourly", "daily", "weekly", "monthly"),
			mcp.DefaultString("daily"),
		),
		mcp.WithString("group_by",
			mcp.Description("Split the series by one more dimension: one row per bucket per group, each row naming its group in the matching field. Omit for a single series over the whole project"),
			mcp.Enum(groupByValues...),
		),
		mcp.WithNumber("limit",
			mcp.Description("Max number of groups to return, 1-100 (default: 20), keeping the largest. Ignored without group_by"),
			mcp.Min(1),
			mcp.Max(100),
		),
	}
}

func readOnlyAnnotations(title string) []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithTitleAnnotation(title),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	}
}

func toolOptions(description string, groups ...[]mcp.ToolOption) []mcp.ToolOption {
	opts := []mcp.ToolOption{mcp.WithDescription(description)}
	for _, g := range groups {
		opts = append(opts, g...)
	}
	return opts
}

// scope reads the parameters every Insights tool takes. The second return
// value is a ready-to-return tool error when one is missing.
func scope(request mcp.CallToolRequest) (workspaceSlug, project string, window insights.Window, errResult *mcp.CallToolResult) {
	var err error
	if workspaceSlug, err = request.RequireString("workspace_slug"); err != nil {
		return "", "", insights.Window{}, mcp.NewToolResultError(err.Error())
	}
	if project, err = request.RequireString("project"); err != nil {
		return "", "", insights.Window{}, mcp.NewToolResultError(err.Error())
	}
	if window.Start, err = request.RequireString("start"); err != nil {
		return "", "", insights.Window{}, mcp.NewToolResultError(err.Error())
	}
	if window.End, err = request.RequireString("end"); err != nil {
		return "", "", insights.Window{}, mcp.NewToolResultError(err.Error())
	}
	return workspaceSlug, project, window, nil
}

func buildFilters(request mcp.CallToolRequest) insights.BuildFilters {
	return insights.BuildFilters{
		Branches:  request.GetStringSlice("branch", nil),
		Workflows: request.GetStringSlice("workflow", nil),
		Pipelines: request.GetStringSlice("pipeline", nil),
	}
}

func seriesParams(request mcp.CallToolRequest) (insights.Granularity, string, int) {
	return insights.Granularity(request.GetString("granularity", "")),
		request.GetString("group_by", ""),
		request.GetInt("limit", 0)
}
