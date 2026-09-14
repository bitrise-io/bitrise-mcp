// Package insights is a typed client for the public Bitrise Insights API
// (https://api.bitrise.io/insights/api-docs/): read-only build and test
// metrics for one project of a workspace.
//
// The request and response types mirror the published OpenAPI spec. Every
// call reports on exactly one project, timestamps are RFC3339, durations are
// seconds and rates are percentages in 0-100. Filters that take several
// values are sent as repeated query parameters; the branch filter accepts '*'
// as a wildcard.
package insights

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
)

// Granularity is the size of one time bucket in a series.
type Granularity string

const (
	GranularityHourly  Granularity = "hourly"
	GranularityDaily   Granularity = "daily"
	GranularityWeekly  Granularity = "weekly"
	GranularityMonthly Granularity = "monthly"
)

// Window is the reported time window shared by every request: start is
// inclusive, end is exclusive, both RFC3339.
type Window struct {
	Start string
	End   string
}

// BuildFilters narrows the builds a build metrics call reports on. Every
// field is optional and repeatable.
type BuildFilters struct {
	Branches  []string
	Workflows []string
	Pipelines []string
}

// TestFilters narrows the test runs a test metrics call reports on.
type TestFilters struct {
	BuildFilters
	TestSuites []string
}

// BuildTotalsRequest is the request of GET /builds/totals.
type BuildTotalsRequest struct {
	WorkspaceSlug string
	Project       string
	Window
	BuildFilters
}

// BuildSeriesRequest is the request of GET /builds/series.
type BuildSeriesRequest struct {
	BuildTotalsRequest
	Granularity Granularity
	// GroupBy adds one dimension: workflow, pipeline, stage or step.
	GroupBy string
	// Limit caps the number of groups, 1-100; ignored without GroupBy.
	Limit int
}

// TestTotalsRequest is the request of GET /tests/totals.
type TestTotalsRequest struct {
	WorkspaceSlug string
	Project       string
	Window
	TestFilters
}

// TestSeriesRequest is the request of GET /tests/series.
type TestSeriesRequest struct {
	TestTotalsRequest
	Granularity Granularity
	// GroupBy adds one dimension: workflow, pipeline, stage, branch or test_suite.
	GroupBy string
	Limit   int
}

// FlakyTestsRequest is the request of GET /tests/flaky. It takes no pipeline
// filter; the endpoint does not accept one.
type FlakyTestsRequest struct {
	WorkspaceSlug string
	Project       string
	Window
	Branches   []string
	Workflows  []string
	TestSuites []string
	TestCases  []string
	Modules    []string
	// OrderBy is flaky_rerun_count (default) or flaky_rate.
	OrderBy string
	// Order is desc (default) or asc.
	Order   string
	Page    int
	PerPage int
}

// Meta echoes the effective query back on every response.
type Meta struct {
	WorkspaceSlug string `json:"workspace_slug"`
	Project       string `json:"project"`
	Start         string `json:"start"`
	End           string `json:"end"`
	Granularity   string `json:"granularity,omitempty"`
	GroupBy       string `json:"group_by,omitempty"`
}

// BuildTotals holds the build aggregates over the whole window.
type BuildTotals struct {
	BuildCount           int64   `json:"build_count"`
	FailureRate          float64 `json:"failure_rate"`
	DurationP50Seconds   float64 `json:"duration_p50_seconds"`
	DurationP90Seconds   float64 `json:"duration_p90_seconds"`
	TotalDurationSeconds float64 `json:"total_duration_seconds"`
}

// BuildTotalsResponse is the response of GET /builds/totals.
type BuildTotalsResponse struct {
	Meta   Meta        `json:"meta"`
	Totals BuildTotals `json:"totals"`
}

// BuildSeriesPoint is one row of a build series: every metric for one time
// bucket, plus the group label when the request was grouped.
type BuildSeriesPoint struct {
	Timestamp    string `json:"timestamp"`
	Workflow     string `json:"workflow,omitempty"`
	Pipeline     string `json:"pipeline,omitempty"`
	Stage        string `json:"stage,omitempty"`
	Step         string `json:"step,omitempty"`
	TargetBranch string `json:"target_branch,omitempty"`
	GroupID      string `json:"group_id,omitempty"`
	BuildTotals
}

// BuildSeriesResponse is the response of GET /builds/series.
type BuildSeriesResponse struct {
	Meta      Meta               `json:"meta"`
	Truncated bool               `json:"truncated"`
	Series    []BuildSeriesPoint `json:"series"`
}

// TestTotals holds the test run aggregates over the whole window. Counts are
// of test case runs, not of distinct test cases.
type TestTotals struct {
	RunCount             int64   `json:"run_count"`
	FailureRate          float64 `json:"failure_rate"`
	DurationP50Seconds   float64 `json:"duration_p50_seconds"`
	DurationP90Seconds   float64 `json:"duration_p90_seconds"`
	TotalDurationSeconds float64 `json:"total_duration_seconds"`
	FlakyRunCount        int64   `json:"flaky_run_count"`
}

// TestTotalsResponse is the response of GET /tests/totals.
type TestTotalsResponse struct {
	Meta   Meta       `json:"meta"`
	Totals TestTotals `json:"totals"`
}

// TestSeriesPoint is one row of a test series.
type TestSeriesPoint struct {
	Timestamp string `json:"timestamp"`
	Workflow  string `json:"workflow,omitempty"`
	Pipeline  string `json:"pipeline,omitempty"`
	Stage     string `json:"stage,omitempty"`
	Branch    string `json:"branch,omitempty"`
	TestSuite string `json:"test_suite,omitempty"`
	GroupID   string `json:"group_id,omitempty"`
	TestTotals
}

// TestSeriesResponse is the response of GET /tests/series.
type TestSeriesResponse struct {
	Meta      Meta              `json:"meta"`
	Truncated bool              `json:"truncated"`
	Series    []TestSeriesPoint `json:"series"`
}

// Pagination describes the page of a list response.
type Pagination struct {
	Page       int64 `json:"page"`
	PerPage    int64 `json:"per_page"`
	TotalCount int64 `json:"total_count"`
	NextPage   int64 `json:"next_page,omitempty"`
}

// FlakyTestCase is one flaky test case: how often it ran and how often a run
// had to be repeated after a flaky result.
type FlakyTestCase struct {
	TestCase        string  `json:"test_case"`
	TestSuite       string  `json:"test_suite"`
	Module          string  `json:"module"`
	RunCount        int64   `json:"run_count"`
	FlakyRerunCount int64   `json:"flaky_rerun_count"`
	FlakyRate       float64 `json:"flaky_rate"`
}

// FlakyTestsResponse is the response of GET /tests/flaky.
type FlakyTestsResponse struct {
	Meta       Meta            `json:"meta"`
	Pagination Pagination      `json:"pagination"`
	TestCases  []FlakyTestCase `json:"test_cases"`
}

// ErrPlanRequired is returned when the workspace has no paid Insights plan
// (HTTP 402).
var ErrPlanRequired = errors.New("the Insights API requires a paid Insights plan for this workspace")

// GetBuildTotals returns build aggregates for one project over the window.
func GetBuildTotals(ctx context.Context, req BuildTotalsRequest) (*BuildTotalsResponse, error) {
	params := commonParams(req.Project, req.Window)
	addBuildFilters(params, req.BuildFilters)

	var out BuildTotalsResponse
	if err := call(ctx, req.WorkspaceSlug, "builds/totals", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetBuildSeries returns a build metrics time series for one project.
func GetBuildSeries(ctx context.Context, req BuildSeriesRequest) (*BuildSeriesResponse, error) {
	params := commonParams(req.Project, req.Window)
	addBuildFilters(params, req.BuildFilters)
	addSeriesParams(params, req.Granularity, req.GroupBy, req.Limit)

	var out BuildSeriesResponse
	if err := call(ctx, req.WorkspaceSlug, "builds/series", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTestTotals returns test run aggregates for one project over the window.
func GetTestTotals(ctx context.Context, req TestTotalsRequest) (*TestTotalsResponse, error) {
	params := commonParams(req.Project, req.Window)
	addTestFilters(params, req.TestFilters)

	var out TestTotalsResponse
	if err := call(ctx, req.WorkspaceSlug, "tests/totals", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTestSeries returns a test metrics time series for one project.
func GetTestSeries(ctx context.Context, req TestSeriesRequest) (*TestSeriesResponse, error) {
	params := commonParams(req.Project, req.Window)
	addTestFilters(params, req.TestFilters)
	addSeriesParams(params, req.Granularity, req.GroupBy, req.Limit)

	var out TestSeriesResponse
	if err := call(ctx, req.WorkspaceSlug, "tests/series", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListFlakyTests returns one page of the flaky test cases of a project.
func ListFlakyTests(ctx context.Context, req FlakyTestsRequest) (*FlakyTestsResponse, error) {
	params := commonParams(req.Project, req.Window)
	addStrings(params, "branch", req.Branches)
	addStrings(params, "workflow", req.Workflows)
	addStrings(params, "test_suite", req.TestSuites)
	addStrings(params, "test_case", req.TestCases)
	addStrings(params, "module", req.Modules)
	if req.OrderBy != "" {
		params["order_by"] = req.OrderBy
	}
	if req.Order != "" {
		params["order"] = req.Order
	}
	if req.Page > 0 {
		params["page"] = strconv.Itoa(req.Page)
	}
	if req.PerPage > 0 {
		params["per_page"] = strconv.Itoa(req.PerPage)
	}

	var out FlakyTestsResponse
	if err := call(ctx, req.WorkspaceSlug, "tests/flaky", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func call(ctx context.Context, workspaceSlug, endpoint string, params map[string]any, out any) error {
	if workspaceSlug == "" {
		return errors.New("workspace_slug is required")
	}

	res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
		Method:  http.MethodGet,
		BaseURL: bitrise.APIInsightsBaseURL,
		Path:    fmt.Sprintf("/workspaces/%s/%s", workspaceSlug, endpoint),
		Params:  params,
	})
	if err != nil {
		var apiErr *bitrise.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusPaymentRequired {
			return ErrPlanRequired
		}
		return err
	}

	if err := json.Unmarshal([]byte(res), out); err != nil {
		return fmt.Errorf("decode insights response: %w", err)
	}
	return nil
}

func commonParams(project string, w Window) map[string]any {
	return map[string]any{
		"project": project,
		"start":   w.Start,
		"end":     w.End,
	}
}

func addBuildFilters(params map[string]any, f BuildFilters) {
	addStrings(params, "branch", f.Branches)
	addStrings(params, "workflow", f.Workflows)
	addStrings(params, "pipeline", f.Pipelines)
}

func addTestFilters(params map[string]any, f TestFilters) {
	addBuildFilters(params, f.BuildFilters)
	addStrings(params, "test_suite", f.TestSuites)
}

func addSeriesParams(params map[string]any, g Granularity, groupBy string, limit int) {
	if g != "" {
		params["granularity"] = string(g)
	}
	if groupBy != "" {
		params["group_by"] = groupBy
	}
	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}
}

func addStrings(params map[string]any, key string, values []string) {
	if len(values) > 0 {
		params[key] = values
	}
}
