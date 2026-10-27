package builds

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var GetTestResults = bitrise.Tool{
	APIGroups: []string{"builds", "read-only"},
	Definition: mcp.NewTool("get_build_test_results",
		mcp.WithDescription("Get the test results of a finished build of a given app: test suites and test cases with their status, duration and failure message. Requires a plan that includes test reports."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("build_slug",
			mcp.Description("Identifier of the build"),
			mcp.Required(),
		),
		mcp.WithBoolean("failed_only",
			mcp.Description("Only return failed and errored test cases, and the test suites containing them. Suite counts still cover every test case. Default: false"),
		),
		mcp.WithTitleAnnotation("Get Build Test Results"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		buildSlug, err := request.RequireString("build_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/builds/%s/test-results", appSlug, buildSlug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}

		if !request.GetBool("failed_only", false) {
			return mcp.NewToolResultText(res), nil
		}

		var response map[string]any
		if err := json.Unmarshal([]byte(res), &response); err != nil {
			return mcp.NewToolResultText(res), nil
		}
		if suites, ok := response["test_suites"].([]any); ok {
			response["test_suites"] = failedTestSuites(suites)
		}

		return mcp.NewToolResultStructuredOnly(response), nil
	},
}

func failedTestSuites(suites []any) []any {
	failed := []any{}
	for _, s := range suites {
		suite, ok := s.(map[string]any)
		if !ok {
			continue
		}

		cases := []any{}
		if testCases, ok := suite["test_cases"].([]any); ok {
			for _, c := range testCases {
				if testCase, ok := c.(map[string]any); ok && isFailedTestCase(testCase) {
					cases = append(cases, testCase)
				}
			}
		}

		var nested []any
		if nestedSuites, ok := suite["test_suites"].([]any); ok {
			nested = failedTestSuites(nestedSuites)
		}

		if len(cases) == 0 && len(nested) == 0 {
			continue
		}
		suite["test_cases"] = cases
		suite["test_suites"] = nested
		failed = append(failed, suite)
	}

	return failed
}

func isFailedTestCase(testCase map[string]any) bool {
	status, _ := testCase["status"].(string)
	return status == "failed" || status == "error"
}
