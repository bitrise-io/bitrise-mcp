package tool

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
)

var listApps = Tool{
	APIGroups: []string{"apps", "read-only"},
	Definition: mcp.NewTool("list_apps",
		mcp.WithDescription("List all apps for the currently authenticated user account"),
		mcp.WithString("sort_by",
			mcp.Description("Order of the apps: last_build_at (default) or created_at. If set, you should accept the response as sorted"),
			mcp.Enum("last_build_at", "created_at"),
			mcp.DefaultString("last_build_at"),
		),
		mcp.WithString("next",
			mcp.Description("Slug of the first app in the response"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Max number of elements per page (default: 50)"),
			mcp.DefaultNumber(50),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodGet,
			baseURL: apiBaseURL,
			path:    "/apps",
			params: map[string]string{
				"sort_by": request.GetString("sort_by", "last_build_at"),
				"next":    request.GetString("next", ""),
				"limit":   strconv.Itoa(request.GetInt("limit", 50)),
			},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var registerApp = Tool{
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
		mcp.WithString("project_type",
			mcp.Description("Type of project (ios, android, etc.)"),
			mcp.DefaultString("other"),
		),
		mcp.WithString("provider",
			mcp.Description("Repository provider"),
			mcp.DefaultString("github"),
		),
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

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodPost,
			baseURL: apiBaseURL,
			path:    "/apps/register",
			body: map[string]any{
				"repo_url":          repoURL,
				"is_public":         isPublic,
				"organization_slug": orgSlug,
				"project_type":      request.GetString("project_type", "other"),
				"provider":          request.GetString("provider", "github"),
			},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var finishBitriseApp = Tool{
	APIGroups: []string{"apps"},
	Definition: mcp.NewTool("finish_bitrise_app",
		mcp.WithDescription("Finish the setup of a Bitrise app. If this is successful, a build can be triggered via trigger_bitrise_build. If you have access to the repository, decide the project type, the stack ID, and the config to use, based on https://stacks.bitrise.io/, and the config should be also based on the project type."),
		mcp.WithString("app_slug",
			mcp.Description("The slug of the Bitrise app to finish setup for."),
			mcp.Required(),
		),
		mcp.WithString("project_type",
			mcp.Description("The type of project (e.g., android, ios, flutter, etc.)."),
			mcp.DefaultString("other"),
		),
		mcp.WithString("stack_id",
			mcp.Description("The stack ID to use for the app."),
			mcp.DefaultString("linux-docker-android-22.04"),
		),
		mcp.WithString("mode",
			mcp.Description("The mode of setup."),
			mcp.DefaultString("manual"),
		),
		mcp.WithString("config",
			mcp.Description("The configuration to use for the app (default is \"other-config\", other valid values are \"default-android-config\", \"default-ios-config\", \"default-macos-config\", etc)."),
			mcp.DefaultString("other-config"),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodPost,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/apps/%s/finish", appSlug),
			body: map[string]any{
				"project_type": request.GetString("project_type", "other"),
				"stack_id":     request.GetString("stack_id", "linux-docker-android-22.04"),
				"mode":         request.GetString("mode", "manual"),
				"config":       request.GetString("config", "other-config"),
			},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var getApp = Tool{
	APIGroups: []string{"apps", "read-only"},
	Definition: mcp.NewTool("get_app",
		mcp.WithDescription("Get the details of a specific app."),
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
			path:    fmt.Sprintf("/apps/%s", appSlug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var deleteApp = Tool{
	APIGroups: []string{"apps"},
	Definition: mcp.NewTool("delete_app",
		mcp.WithDescription("Delete an app from Bitrise. When deleting apps belonging to multiple workspaces always confirm that which workspaces' apps the user wants to delete."),
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
			method:  http.MethodDelete,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/apps/%s", appSlug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var updateApp = Tool{
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

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodPatch,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/apps/%s", appSlug),
			body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var getBitriseYML = Tool{
	APIGroups: []string{"apps", "read-only"},
	Definition: mcp.NewTool("get_bitrise_yml",
		mcp.WithDescription("Get the current Bitrise YML config file of a specified Bitrise app."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app (e.g., \"d8db74e2675d54c4\" or \"8eb495d0-f653-4eed-910b-8d6b56cc0ec7\")"),
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
			path:    fmt.Sprintf("/apps/%s/bitrise.yml", appSlug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var updateBitriseYML = Tool{
	APIGroups: []string{"apps"},
	Definition: mcp.NewTool("update_bitrise_yml",
		mcp.WithDescription("Update the Bitrise YML config file of a specified Bitrise app."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app (e.g., \"d8db74e2675d54c4\" or \"8eb495d0-f653-4eed-910b-8d6b56cc0ec7\")"),
			mcp.Required(),
		),
		mcp.WithString("bitrise_yml_as_json",
			mcp.Description("The new Bitrise YML config file content to be updated. It must be a string."),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		ymlContent, err := request.RequireString("bitrise_yml_as_json")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodPost,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/apps/%s/bitrise.yml", appSlug),
			body: map[string]any{
				"app_config_datastore_yaml": ymlContent,
			},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var listBranches = Tool{
	APIGroups: []string{"apps", "read-only"},
	Definition: mcp.NewTool("list_branches",
		mcp.WithDescription("List the branches with existing builds of an app's repository."),
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
			path:    fmt.Sprintf("/apps/%s/branches", appSlug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var registerSSHKey = Tool{
	APIGroups: []string{"apps"},
	Definition: mcp.NewTool("register_ssh_key",
		mcp.WithDescription("Add an SSH-key to a specific app."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("auth_ssh_private_key",
			mcp.Description("Private SSH key"),
			mcp.Required(),
		),
		mcp.WithString("auth_ssh_public_key",
			mcp.Description("Public SSH key"),
			mcp.Required(),
		),
		mcp.WithBoolean("is_register_key_into_provider_service",
			mcp.Description("Register the key in the provider service"),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		privateKey, err := request.RequireString("auth_ssh_private_key")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		publicKey, err := request.RequireString("auth_ssh_public_key")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		body := map[string]any{
			"auth_ssh_private_key": privateKey,
			"auth_ssh_public_key":  publicKey,
		}
		regKey := "is_register_key_into_provider_service"
		if _, ok := request.GetArguments()[regKey]; ok {
			body[regKey] = request.GetBool(regKey, false)
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodPost,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/apps/%s/register-ssh-key", appSlug),
			body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var registerWebhook = Tool{
	APIGroups: []string{"apps"},
	Definition: mcp.NewTool("register_webhook",
		mcp.WithDescription("Register an incoming webhook for a specific application."),
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
			method:  http.MethodPost,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/apps/%s/register-webhook", appSlug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
