package builds

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var GetTestResults = bitrise.Tool{
	APIGroups: []string{"builds", "read-only"},
	Definition: mcp.NewTool("get_build_test_results",
		mcp.WithDescription("Get the test results of a finished build of a given app: test suites and test cases with their status, duration and failure message. Requires a plan that includes test reports. Filters are applied on the server; with any filter or limit, test suites without matching test cases are left out, while suite counts still cover every test case."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("build_slug",
			mcp.Description("Identifier of the build"),
			mcp.Required(),
		),
		mcp.WithArray("status",
			mcp.WithStringEnumItems([]string{"passed", "failed", "skipped", "error"}),
			mcp.Description("Keep only test cases with these statuses. Use [\"failed\", \"error\"] for every test that needs fixing. A test that failed and then passed on a retry has status passed and flaky true."),
		),
		mcp.WithString("suite",
			mcp.Description("Keep only test cases under the test suite with this exact name, at any nesting level"),
		),
		mcp.WithString("class_name",
			mcp.Description("Keep only test cases with this exact class name"),
		),
		mcp.WithBoolean("flaky",
			mcp.Description("Keep only test cases that are (true) or are not (false) flaky in this build, meaning their retries ended with different statuses"),
		),
		mcp.WithString("sort_by",
			mcp.Description("Sort the test cases within each test suite. Ascending status lists error, failed, skipped, then passed."),
			mcp.Enum("duration", "name", "status"),
		),
		mcp.WithString("sort_order",
			mcp.Description("Direction of sort_by. Default: asc"),
			mcp.Enum("asc", "desc"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of test cases to return, counted in response order across all test suites"),
			mcp.Min(1),
			mcp.Max(1000),
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

		params := map[string]any{}
		if v := request.GetStringSlice("status", nil); len(v) > 0 {
			params["status"] = v
		}
		for _, key := range []string{"suite", "class_name", "sort_by", "sort_order"} {
			if v := request.GetString(key, ""); v != "" {
				params[key] = v
			}
		}
		if _, ok := request.GetArguments()["flaky"]; ok {
			params["flaky"] = strconv.FormatBool(request.GetBool("flaky", false))
		}
		if _, ok := request.GetArguments()["limit"]; ok {
			params["limit"] = strconv.Itoa(request.GetInt("limit", 0))
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/builds/%s/test-results", appSlug, buildSlug),
			Params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
