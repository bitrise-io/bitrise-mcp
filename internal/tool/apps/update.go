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
		mcp.WithDescription("Update an app. Only app_slug is required. Omit all other fields you don't wish to update"),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithBoolean("is_public",
			mcp.Description("Whether the app's builds visibility is \"public\""),
		),
		mcp.WithString("project_type",
			mcp.Description("Type of project"),
		),
		mcp.WithString("provider",
			mcp.Description("Repository provider"),
		),
		mcp.WithString("repo_url",
			mcp.Description("Repository URL"),
		),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(true),
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
		if v := request.GetString("project_type", ""); v != "" {
			body["project_type"] = v
		}
		if v := request.GetString("provider", ""); v != "" {
			body["provider"] = v
		}
		if v := request.GetString("repo_url", ""); v != "" {
			body["repo_url"] = v
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
