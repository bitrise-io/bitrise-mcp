package apps

import (
	"context"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var Register = bitrise.Tool{
	APIGroups: []string{"apps"},
	Definition: mcp.NewTool("register_app",
		mcp.WithDescription("Add a new app to Bitrise. After this app should be finished on order to be registered completely on Bitrise (via the finish_bitrise_app tool). Before doing this step, try understanding the repository details from the repository URL. This is a two-step process. First, you register the app with the Bitrise API, and then you finish the setup. The first step creates a new app in Bitrise, and the second step configures it with the necessary settings. If the user has multiple workspaces, always prompt the user to choose which one you should use. Don't prompt the user for finishing the app, just do it automatically."),
		mcp.WithString("repo_url",
			mcp.Description("Repository URL"),
			mcp.Required(),
		),
		mcp.WithBoolean("is_public",
			mcp.Description("Whether the app's builds visibility is \"public\""),
			mcp.Required(),
		),
		mcp.WithString("organization_slug",
			mcp.Description("The organization (aka workspace) the app to add to"),
			mcp.Required(),
		),
		mcp.WithString("title",
			mcp.Description("The title of the application (if not specified, will use the git repository's name)"),
		),
		mcp.WithString("default_branch_name",
			mcp.Description("The default branch of the repository"),
			mcp.DefaultString("master"),
		),
		mcp.WithBoolean("manual_approval_enabled",
			mcp.Description("Toggles whether manual approval should be enabled for the app's builds"),
			mcp.DefaultBool(true),
		),
		mcp.WithString("provider",
			mcp.Description("The git provider of the repository"),
			mcp.Enum("bitbucket", "bitbucket-server", "custom", "github", "github-app", "github-self-hosted", "gitlab", "gitlab-self-hosted"),
			mcp.DefaultString("github"),
		),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		repoURL, err := request.RequireString("repo_url")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		orgSlug, err := request.RequireString("organization_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		isPublic, err := request.RequireBool("is_public")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{
			"repo_url":          repoURL,
			"is_public":         isPublic,
			"organization_slug": orgSlug,
			"provider":          request.GetString("provider", "github"),
		}
		if v := request.GetString("title", ""); v != "" {
			body["title"] = v
		}
		if v := request.GetString("default_branch_name", ""); v != "" {
			body["default_branch_name"] = v
		}
		if _, ok := request.GetArguments()["manual_approval_enabled"]; ok {
			body["manual_approval_enabled"] = request.GetBool("manual_approval_enabled", true)
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APIBaseURL,
			Path:    "/apps/register",
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
