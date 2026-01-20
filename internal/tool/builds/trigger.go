package builds

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var Trigger = bitrise.Tool{
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
		mcp.WithArray("environments",
			mcp.Description(`Custom environment variables for the build.`),
			mcp.Items(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"mapped_to": map[string]any{
						"type":        "string",
						"description": "The name of the environment variable",
					},
					"value": map[string]any{
						"type":        "string",
						"description": "The value of the environment variable",
					},
					"is_expand": map[string]any{
						"type":        "boolean",
						"description": "Whether to expand environment variable references in the value",
					},
				},
				"required": []string{"mapped_to", "value"},
			}),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		buildParams := map[string]any{
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
		if v, ok := request.GetArguments()["environments"]; ok {
			buildParams["environments"] = v
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/builds", appSlug),
			Body: map[string]any{
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
