package builds

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFailedTestSuites(t *testing.T) {
	given := `[
		{
			"name": "UnitTests",
			"test_count": 3,
			"failure_count": 1,
			"test_cases": [
				{"name": "testPasses", "status": "passed"},
				{"name": "testFails", "status": "failed", "message": "expected true"},
				{"name": "testSkipped", "status": "skipped"}
			],
			"test_suites": [
				{
					"name": "NestedPassing",
					"test_cases": [{"name": "testNestedPasses", "status": "passed"}],
					"test_suites": []
				},
				{
					"name": "NestedErroring",
					"test_cases": [{"name": "testNestedErrors", "status": "error", "message": "crashed"}],
					"test_suites": []
				}
			]
		},
		{
			"name": "UITests",
			"test_cases": [{"name": "testUIPasses", "status": "passed"}],
			"test_suites": []
		}
	]`
	want := `[
		{
			"name": "UnitTests",
			"test_count": 3,
			"failure_count": 1,
			"test_cases": [
				{"name": "testFails", "status": "failed", "message": "expected true"}
			],
			"test_suites": [
				{
					"name": "NestedErroring",
					"test_cases": [{"name": "testNestedErrors", "status": "error", "message": "crashed"}],
					"test_suites": []
				}
			]
		}
	]`

	var suites []any
	if err := json.Unmarshal([]byte(given), &suites); err != nil {
		t.Fatal(err)
	}

	got, err := json.Marshal(failedTestSuites(suites))
	if err != nil {
		t.Fatal(err)
	}
	assert.JSONEq(t, want, string(got))
}
