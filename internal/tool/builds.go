package tool

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
)

var listBuilds = Tool{
	APIGroups: []string{"builds", "read-only"},
	Definition: mcp.NewTool("list_builds",
		mcp.WithDescription("List all the builds of a specified Bitrise app or all accessible builds."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
		),
		mcp.WithString("sort_by",
			mcp.Description("Order of builds: created_at (default), running_first"),
			mcp.DefaultString("created_at"),
			mcp.Enum("created_at", "running_first"),
		),
		mcp.WithString("branch",
			mcp.Description("Filter builds by branch"),
		),
		mcp.WithString("workflow",
			mcp.Description("Filter builds by workflow"),
		),
		mcp.WithNumber("status",
			mcp.Description("Filter builds by status (0: not finished, 1: successful, 2: failed, 3: aborted, 4: in-progress)"),
			mcp.Enum("0", "1", "2", "3", "4"),
		),
		mcp.WithString("next",
			mcp.Description("Slug of the first build in the response"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Max number of elements per page (default: 50)"),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params := map[string]string{}
		if v := request.GetString("sort_by", ""); v != "" {
			params["sort_by"] = v
		}
		if v := request.GetString("branch", ""); v != "" {
			params["branch"] = v
		}
		if v := request.GetString("workflow", ""); v != "" {
			params["workflow"] = v
		}
		if _, ok := request.GetArguments()["status"]; ok {
			params["status"] = strconv.Itoa(request.GetInt("status", 0))
		}
		if v := request.GetString("next", ""); v != "" {
			params["next"] = v
		}
		if _, ok := request.GetArguments()["limit"]; ok {
			params["limit"] = strconv.Itoa(request.GetInt("limit", 50))
		}

		path := "/builds"
		if appSlug := request.GetString("app_slug", ""); appSlug != "" {
			path = fmt.Sprintf("/apps/%s/builds", appSlug)
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodGet,
			baseURL: apiBaseURL,
			path:    path,
			params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var triggerBitriseBuild = Tool{
	APIGroups: []string{"builds"},
	Definition: mcp.NewTool("trigger_bitrise_build",
		mcp.WithDescription("Trigger a new build/pipeline for a specified Bitrise app"),
		mcp.WithString("app_slug",
			mcp.Description(`Identifier of the Bitrise app (e.g., "d8db74e2675d54c4" or "8eb495d0-f653-4eed-910b-8d6b56cc0ec7")`),
			mcp.Required(),
		),
		mcp.WithString("branch",
			mcp.Description("The branch to build"),
			mcp.DefaultString("main"),
		),
		mcp.WithString("workflow_id",
			mcp.Description("The workflow to build"),
		),
		mcp.WithString("pipeline_id",
			mcp.Description("The pipeline to build"),
		),
		mcp.WithString("commit_message",
			mcp.Description("The commit message for the build"),
		),
		mcp.WithString("commit_hash",
			mcp.Description("The commit hash for the build"),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		buildParams := map[string]string{
			"branch": request.GetString("branch", "main"),
		}
		if v := request.GetString("workflow_id", ""); v != "" {
			buildParams["workflow_id"] = v
		}
		if v := request.GetString("pipeline_id", ""); v != "" {
			buildParams["pipeline_id"] = v
		}
		if v := request.GetString("commit_message", ""); v != "" {
			buildParams["commit_message"] = v
		}
		if v := request.GetString("commit_hash", ""); v != "" {
			buildParams["commit_hash"] = v
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodPost,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/apps/%s/builds", appSlug),
			body: map[string]any{
				"build_params": buildParams,
				"hook_info": map[string]any{
					"type": "bitrise",
				},
			},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var getBuild = Tool{
	APIGroups: []string{"builds", "read-only"},
	Definition: mcp.NewTool("get_build",
		mcp.WithDescription("Get a specific build of a given app."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("build_slug",
			mcp.Description("Identifier of the build"),
			mcp.Required(),
		),
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

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodGet,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/apps/%s/builds/%s", appSlug, buildSlug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var abortBuild = Tool{
	APIGroups: []string{"builds"},
	Definition: mcp.NewTool("abort_build",
		mcp.WithDescription("Abort a specific build."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("build_slug",
			mcp.Description("Identifier of the build"),
			mcp.Required(),
		),
		mcp.WithString("reason",
			mcp.Description("Reason for aborting the build"),
		),
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

		body := map[string]any{}
		if v := request.GetString("reason", ""); v != "" {
			body["abort_reason"] = v
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodPost,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/apps/%s/builds/%s/abort", appSlug, buildSlug),
			body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var getBuildLog = Tool{
	APIGroups: []string{"builds", "read-only"},
	Definition: mcp.NewTool("get_build_log",
		mcp.WithDescription("Get the build log of a specified build of a Bitrise app."),
		mcp.WithString("app_slug",
			mcp.Description(`Identifier of the Bitrise app (e.g., "d8db74e2675d54c4" or "8eb495d0-f653-4eed-910b-8d6b56cc0ec7")`),
			mcp.Required(),
		),
		mcp.WithString("build_slug",
			mcp.Description("Identifier of the Bitrise build"),
			mcp.Required(),
		),
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

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodGet,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/apps/%s/builds/%s/log", appSlug, buildSlug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var getBuildBitriseYML = Tool{
	APIGroups: []string{"builds", "read-only"},
	Definition: mcp.NewTool("get_build_bitrise_yml",
		mcp.WithDescription("Get the bitrise.yml of a build."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("build_slug",
			mcp.Description("Identifier of the build"),
			mcp.Required(),
		),
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

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodGet,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/apps/%s/builds/%s/bitrise.yml", appSlug, buildSlug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var listBuildWorkflows = Tool{
	APIGroups: []string{"builds", "read-only"},
	Definition: mcp.NewTool("list_build_workflows",
		mcp.WithDescription("List the workflows of an app."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodGet,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/apps/%s/build-workflows", appSlug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
