package builds

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var Abort = bitrise.Tool{
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
		mcp.WithString("abort_reason",
			mcp.Description("Reason for aborting the build"),
		),
		mcp.WithBoolean("abort_with_success",
			mcp.Description("If set to true, the aborted build will be marked as successful"),
			mcp.DefaultBool(false),
		),
		mcp.WithBoolean("skip_git_status_report",
			mcp.Description("If set to true, skip sending git status report"),
			mcp.DefaultBool(false),
		),
		mcp.WithBoolean("skip_notifications",
			mcp.Description("If set to true, skip sending notifications"),
			mcp.DefaultBool(false),
		),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(true),
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
		abortReason := request.GetString("abort_reason", "aborted via MCP")
		abortWithSuccess := request.GetBool("abort_with_success", false)
		skipGitStatusReport := request.GetBool("skip_git_status_report", false)
		skipNotifications := request.GetBool("skip_notifications", false)

		body := map[string]any{
			"abort_reason":           abortReason,
			"abort_with_success":     abortWithSuccess,
			"skip_git_status_report": skipGitStatusReport,
			"skip_notifications":     skipNotifications,
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/builds/%s/abort", appSlug, buildSlug),
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
