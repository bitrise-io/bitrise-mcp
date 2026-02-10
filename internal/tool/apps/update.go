package apps

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var Update = bitrise.Tool{
	APIGroups: []string{"apps"},
	Definition: mcp.NewTool("update_app",
		mcp.WithDescription("Update an app. Only app_slug is required, add only fields you wish to update"),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("title",
			mcp.Description("The new title of the application"),
		),
		mcp.WithString("default_branch",
			mcp.Description("The new default branch for the application"),
		),
		mcp.WithString("repository_url",
			mcp.Description("The new repository URL for the application"),
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
		body := map[string]any{}
		if _, ok := request.GetArguments()["is_public"]; ok {
			body["is_public"] = request.GetBool("is_public", false)
		}
		if v := request.GetString("title", ""); v != "" {
			body["title"] = v
		}
		if v := request.GetString("default_branch", ""); v != "" {
			body["default_branch"] = v
		}
		if v := request.GetString("repository_url", ""); v != "" {
			body["repository_url"] = v
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPatch,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s", appSlug),
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
