package releasemanagement

import (
	"context"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var CreateConnectedApp = bitrise.Tool{
	APIGroups: []string{"release-management"},
	Definition: mcp.NewTool("create_connected_app",
		mcp.WithDescription("Add a new Release Management connected app to Bitrise. "+
			"Adding an internal tester grants the App Tester role on the app; removing them later does not revoke it."),
		mcp.WithString("platform",
			mcp.Description("The mobile platform for the connected app. Available values are 'ios' and 'android'."),
			mcp.Required(),
			mcp.Enum("ios", "android"),
		),
		mcp.WithString("store_app_id",
			mcp.Description("The app store identifier for the connected app. In case of 'ios' platform it is the bundle id from App Store Connect. In case of Android platform it is the package name."),
			mcp.Required(),
		),
		mcp.WithString("workspace_slug",
			mcp.Description("Identifier of the Bitrise workspace for the Release Management connected app. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithString("id",
			mcp.Description("An uuidV4 identifier for your new connected app. If it is not given, one will be generated."),
		),
		mcp.WithBoolean("manual_connection",
			mcp.Description("If set to true it indicates a manual connection (bypassing using store api keys) and requires giving 'store_app_name' as well."),
			mcp.DefaultBool(false),
		),
		mcp.WithString("project_id",
			mcp.Description("Specifies which Bitrise Project you want to get the connected app to be associated with. If this field is not given a new project will be created alongside with the connected app."),
		),
		mcp.WithString("project_title",
			mcp.Description("Title for the project created when no 'project_id' is given. Defaults to the app name. Sending it together with 'project_id' is rejected with 400 and ERR_INVALID_PARAMS."),
		),
		mcp.WithString("store_app_name",
			mcp.Description("If you have no active app store API keys added on Bitrise, you can decide to add your app manually by giving the app's name as well while indicating manual connection."),
		),
		mcp.WithString("store_credential_id",
			mcp.Description("If you have credentials added on Bitrise, you can decide to select one for your app. In case of ios platform it will be an Apple API credential id. In case of android platform it will be a Google Service credential id."),
		),
		mcp.WithString("framework",
			mcp.Description("The framework used to build the app. Defaults to 'other' if not provided."),
			mcp.Enum("flutter", "react_native", "kotlin_multiplatform", "other", "native_ios", "native_android"),
		),
		mcp.WithArray("tester_groups",
			mcp.Description("Tester groups created together with the app, each with a 'name' and optionally 'type', 'auto_notify' and its members ('user_slugs' for internal groups, 'emails' for external ones). Created one by one, so check the response's 'errors' array."),
			mcp.Items(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name":        map[string]any{"type": "string"},
					"type":        map[string]any{"type": "string", "enum": []string{"internal", "external"}},
					"auto_notify": map[string]any{"type": "boolean"},
					"user_slugs":  map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"emails":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				},
				"required": []string{"name"},
			}),
		),
		mcp.WithArray("code_push_deployments",
			mcp.Description("CodePush deployments created together with the app, each with a 'name' and an optional 'key'. Created one by one, so check the response's 'errors' array."),
			mcp.Items(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
					"key":  map[string]any{"type": "string"},
				},
				"required": []string{"name"},
			}),
		),
		mcp.WithTitleAnnotation("Create Connected App"),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		platform, err := request.RequireString("platform")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		storeAppID, err := request.RequireString("store_app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		workspaceSlug, err := request.RequireString("workspace_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{
			"platform":       platform,
			"store_app_id":   storeAppID,
			"workspace_slug": workspaceSlug,
		}
		if v := request.GetString("id", ""); v != "" {
			body["id"] = v
		}
		if v := request.GetBool("manual_connection", false); v {
			body["manual_connection"] = v
		}
		if v := request.GetString("project_id", ""); v != "" {
			body["project_id"] = v
		}
		if v := request.GetString("project_title", ""); v != "" {
			body["project_title"] = v
		}
		if v := request.GetString("store_app_name", ""); v != "" {
			body["store_app_name"] = v
		}
		if v := request.GetString("store_credential_id", ""); v != "" {
			body["store_credential_id"] = v
		}
		if v := request.GetString("framework", ""); v != "" {
			body["framework"] = v
		}
		if v, ok := request.GetArguments()["tester_groups"]; ok {
			body["tester_groups"] = v
		}
		if v, ok := request.GetArguments()["code_push_deployments"]; ok {
			body["code_push_deployments"] = v
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APIRMAppsBaseURL,
			Path:    "/",
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
