package insights

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/stretchr/testify/assert"
)

func withServer(t *testing.T, handler http.HandlerFunc) context.Context {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	prev := bitrise.APIInsightsBaseURL
	bitrise.APIInsightsBaseURL = srv.URL
	t.Cleanup(func() { bitrise.APIInsightsBaseURL = prev })

	return bitrise.ContextWithPAT(context.Background(), "test-token")
}

func TestGetBuildSeries_sendsRepeatedFiltersAndDecodes(t *testing.T) {
	var got *http.Request
	ctx := withServer(t, func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"meta": {"workspace_slug": "ws", "project": "p1", "start": "2026-07-01T00:00:00Z", "end": "2026-07-03T00:00:00Z", "granularity": "daily", "group_by": "workflow"},
			"truncated": false,
			"series": [
				{"timestamp": "2026-07-01T00:00:00Z", "workflow": "ci", "build_count": 42, "failure_rate": 2, "duration_p50_seconds": 301, "duration_p90_seconds": 870, "total_duration_seconds": 12800}
			]
		}`))
	})

	res, err := GetBuildSeries(ctx, BuildSeriesRequest{
		BuildTotalsRequest: BuildTotalsRequest{
			WorkspaceSlug: "ws",
			Project:       "p1",
			Window:        Window{Start: "2026-07-01T00:00:00Z", End: "2026-07-03T00:00:00Z"},
			BuildFilters:  BuildFilters{Branches: []string{"main", "release/*"}},
		},
		Granularity: GranularityDaily,
		GroupBy:     "workflow",
		Limit:       5,
	})
	if !assert.NoError(t, err) {
		return
	}

	if !assert.NotNil(t, got) {
		return
	}
	assert.Equal(t, "/workspaces/ws/builds/series", got.URL.Path)
	assert.Equal(t, "test-token", got.Header.Get("Authorization"))
	q := got.URL.Query()
	assert.Equal(t, "p1", q.Get("project"))
	assert.Equal(t, []string{"main", "release/*"}, q["branch"])
	assert.Equal(t, "daily", q.Get("granularity"))
	assert.Equal(t, "workflow", q.Get("group_by"))
	assert.Equal(t, "5", q.Get("limit"))
	assert.Empty(t, q["workflow"], "unset filters are not sent")

	assert.Equal(t, "workflow", res.Meta.GroupBy)
	if !assert.Len(t, res.Series, 1) {
		return
	}
	assert.Equal(t, "ci", res.Series[0].Workflow)
	assert.Equal(t, int64(42), res.Series[0].BuildCount)
	assert.InDelta(t, 870, res.Series[0].DurationP90Seconds, 0)
}

func TestListFlakyTests_sendsPagingAndDecodes(t *testing.T) {
	var got *http.Request
	ctx := withServer(t, func(w http.ResponseWriter, r *http.Request) {
		got = r
		_, _ = w.Write([]byte(`{
			"meta": {"workspace_slug": "ws", "project": "p1", "start": "2026-07-01T00:00:00Z", "end": "2026-07-31T00:00:00Z"},
			"pagination": {"page": 2, "per_page": 10, "total_count": 25, "next_page": 3},
			"test_cases": [{"test_case": "test_login", "test_suite": "Auth", "module": "app.auth", "run_count": 412, "flaky_rerun_count": 17, "flaky_rate": 4.1}]
		}`))
	})

	res, err := ListFlakyTests(ctx, FlakyTestsRequest{
		WorkspaceSlug: "ws",
		Project:       "p1",
		Window:        Window{Start: "2026-07-01T00:00:00Z", End: "2026-07-31T00:00:00Z"},
		OrderBy:       "flaky_rate",
		Page:          2,
		PerPage:       10,
	})
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "/workspaces/ws/tests/flaky", got.URL.Path)
	q := got.URL.Query()
	assert.Equal(t, "flaky_rate", q.Get("order_by"))
	assert.Equal(t, "2", q.Get("page"))
	assert.Equal(t, "10", q.Get("per_page"))
	assert.Empty(t, q.Get("order"), "defaults are left to the API")

	assert.Equal(t, int64(3), res.Pagination.NextPage)
	if !assert.Len(t, res.TestCases, 1) {
		return
	}
	assert.Equal(t, "test_login", res.TestCases[0].TestCase)
	assert.InDelta(t, 4.1, res.TestCases[0].FlakyRate, 0.001)
}

func TestCall_mapsPaymentRequiredToErrPlanRequired(t *testing.T) {
	ctx := withServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusPaymentRequired)
		_, _ = w.Write([]byte(`{"code":"ERR_PLAN_REQUIRED","message":"the Insights API requires a paid Insights plan"}`))
	})

	_, err := GetTestTotals(ctx, TestTotalsRequest{WorkspaceSlug: "ws", Project: "p1"})
	assert.ErrorIs(t, err, ErrPlanRequired)
}

func TestCall_passesOtherAPIErrorsThrough(t *testing.T) {
	ctx := withServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"no access to project"}`))
	})

	_, err := GetTestSeries(ctx, TestSeriesRequest{TestTotalsRequest: TestTotalsRequest{WorkspaceSlug: "ws", Project: "p1"}})
	var apiErr *bitrise.APIError
	if !assert.ErrorAs(t, err, &apiErr) {
		return
	}
	assert.Equal(t, http.StatusForbidden, apiErr.StatusCode)
	assert.Contains(t, apiErr.Body, "no access to project")
}

func TestCall_requiresWorkspaceSlug(t *testing.T) {
	_, err := GetBuildTotals(bitrise.ContextWithPAT(context.Background(), "t"), BuildTotalsRequest{Project: "p1"})
	assert.EqualError(t, err, "workspace_slug is required")
}
