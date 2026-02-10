package pipelines

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var List = bitrise.Tool{
	APIGroups: []string{"pipelines", "read-only"},
	Definition: mcp.NewTool("list_pipelines",
		mcp.WithDescription("List all pipelines and standalone builds of an app."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("after",
			mcp.Description("List pipelines/standalone builds run after a given date (RFC3339 time format)"),
		),
		mcp.WithString("before",
			mcp.Description("List pipelines/standalone builds run before a given date (RFC3339 time format)"),
		),
		mcp.WithString("branch",
			mcp.Description("Filter by the branch which was built"),
		),
		mcp.WithNumber("build_number",
			mcp.Description("Filter by the pipeline/standalone build number"),
		),
		mcp.WithString("commit_message",
			mcp.Description("Filter by the commit message of the pipeline/standalone build"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Max number of elements per page (default: 10)"),
		),
		mcp.WithString("pipeline",
			mcp.Description("Filter by the name of the pipeline"),
		),
		mcp.WithString("status",
			mcp.Description("Filter by the status of the pipeline/standalone build"),
			mcp.Enum("on_hold", "running", "succeeded", "failed", "aborted", "succeeded_with_abort"),
		),
		mcp.WithString("trigger_event_type",
			mcp.Description("Filter by the event that triggered the pipeline/standalone build"),
			mcp.Enum("push", "pull-request", "tag"),
		),
		mcp.WithString("workflow",
			mcp.Description("Filter by the name of the workflow used for the pipeline/standalone build"),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		params := map[string]any{}
		if v := request.GetString("after", ""); v != "" {
			params["after"] = v
		}
		if v := request.GetString("before", ""); v != "" {
			params["before"] = v
		}
		if v := request.GetString("branch", ""); v != "" {
			params["branch"] = v
		}
		if _, ok := request.GetArguments()["build_number"]; ok {
			params["build_number"] = strconv.Itoa(request.GetInt("build_number", 0))
		}
		if v := request.GetString("commit_message", ""); v != "" {
			params["commit_message"] = v
		}
		if _, ok := request.GetArguments()["limit"]; ok {
			params["limit"] = strconv.Itoa(request.GetInt("limit", 10))
		}
		if v := request.GetString("pipeline", ""); v != "" {
			params["pipeline"] = v
		}
		if v := request.GetString("status", ""); v != "" {
			params["status"] = v
		}
		if v := request.GetString("trigger_event_type", ""); v != "" {
			params["trigger_event_type"] = v
		}
		if v := request.GetString("workflow", ""); v != "" {
			params["workflow"] = v
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/pipelines", appSlug),
			Params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
