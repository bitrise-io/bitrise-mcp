package tool

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
)

var createConnectedApp = Tool{
	APIGroups: []string{"release-management"},
	Definition: mcp.NewTool("create_connected_app",
		mcp.WithDescription("Add a new Release Management connected app to Bitrise."),
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
		mcp.WithString("store_app_name",
			mcp.Description("If you have no active app store API keys added on Bitrise, you can decide to add your app manually by giving the app's name as well while indicating manual connection."),
		),
		mcp.WithString("store_credential_id",
			mcp.Description("If you have credentials added on Bitrise, you can decide to select one for your app. In case of ios platform it will be an Apple API credential id. In case of android platform it will be a Google Service credential id."),
		),
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
		if v := request.GetString("store_app_name", ""); v != "" {
			body["store_app_name"] = v
		}
		if v := request.GetString("store_credential_id", ""); v != "" {
			body["store_credential_id"] = v
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodPost,
			baseURL: apiRMBaseURL,
			path:    "/v1/connected-apps",
			body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var listConnectedApps = Tool{
	APIGroups: []string{"release-management", "read-only"},
	Definition: mcp.NewTool("list_connected_apps",
		mcp.WithDescription("List Release Management connected apps available for the authenticated account within a workspace."),
		mcp.WithString("workspace_slug",
			mcp.Description("Identifier of the Bitrise workspace for the Release Management connected apps. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithString("project_id",
			mcp.Description("Specifies which Bitrise Project you want to get associated connected apps for"),
		),
		mcp.WithString("platform",
			mcp.Description("Filters for a specific mobile platform for the list of connected apps. Available values are: 'ios' and 'android'."),
			mcp.Enum("ios", "android"),
		),
		mcp.WithString("search",
			mcp.Description("Search by bundle ID (for ios), package name (for android), or app title (for both platforms). The filter is case-sensitive."),
		),
		mcp.WithNumber("items_per_page",
			mcp.Description("Specifies the maximum number of connected apps returned per page. Default value is 10."),
			mcp.DefaultNumber(10),
		),
		mcp.WithNumber("page",
			mcp.Description("Specifies which page should be returned from the whole result set in a paginated scenario. Default value is 1."),
			mcp.DefaultNumber(1),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		workspaceSlug, err := request.RequireString("workspace_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		params := map[string]string{
			"workspace_slug": workspaceSlug,
		}
		if v := request.GetString("project_id", ""); v != "" {
			params["project_id"] = v
		}
		if v := request.GetString("platform", ""); v != "" {
			params["platform"] = v
		}
		if v := request.GetString("search", ""); v != "" {
			params["search"] = v
		}
		if v := request.GetInt("items_per_page", 10); v != 10 {
			params["items_per_page"] = strconv.Itoa(v)
		}
		if page := request.GetInt("page", 1); page != 1 {
			params["page"] = strconv.Itoa(page)
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodGet,
			baseURL: apiRMBaseURL,
			path:    "/v1/connected-apps",
			params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var getConnectedApp = Tool{
	APIGroups: []string{"release-management", "read-only"},
	Definition: mcp.NewTool("get_connected_app",
		mcp.WithDescription("Gives back a Release Management connected app for the authenticated account."),
		mcp.WithString("id",
			mcp.Description("Identifier of the Release Management connected app"),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodGet,
			baseURL: apiRMBaseURL,
			path:    fmt.Sprintf("/v1/connected-apps/%s", id),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var updateConnectedApp = Tool{
	APIGroups: []string{"release-management"},
	Definition: mcp.NewTool("update_connected_app",
		mcp.WithDescription("Updates a connected app."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier for your connected app."),
			mcp.Required(),
		),
		mcp.WithBoolean("connect_to_store",
			mcp.Description("If true, will check connected app validity against the Apple App Store or Google Play Store (dependent on the platform of your connected app). This means, that the already set or just given store_app_id will be validated against the Store, using the already set or just given store credential id."),
			mcp.DefaultBool(false),
		),
		mcp.WithString("store_app_id",
			mcp.Description("The store identifier for your app. You can change the previously set store_app_id to match the one in the App Store or Google Play depending on the app platform. This is especially useful if you want to connect your app with the store as the system will validate the given store_app_id against the Store. In case of iOS platform it is the bundle id. In case of Android platform it is the package name."),
		),
		mcp.WithString("store_credential_id",
			mcp.Description("If you have credentials added on Bitrise, you can decide to select one for your app. In case of ios platform it will be an Apple API credential id. In case of android platform it will be a Google Service credential id."),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connectedAppID, err := request.RequireString("connected_app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{}
		if v := request.GetBool("connect_to_store", false); v {
			body["connect_to_store"] = v
		}
		if v := request.GetString("store_app_id", ""); v != "" {
			body["store_app_id"] = v
		}
		if v := request.GetString("store_credential_id", ""); v != "" {
			body["store_credential_id"] = v
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodPatch,
			baseURL: apiRMBaseURL,
			path:    fmt.Sprintf("/v1/connected-apps/%s", connectedAppID),
			body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var listInstallableArtifacts = Tool{
	APIGroups: []string{"release-management", "read-only"},
	Definition: mcp.NewTool("list_installable_artifacts",
		mcp.WithDescription("List Release Management installable artifacts of a connected app available for the authenticated account."),
		mcp.WithString("connected_app_id",
			mcp.Description("Identifier of the Release Management connected app for the installable artifacts. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithString("after_date",
			mcp.Description("A date in ISO 8601 string format specifying the start of the interval when the installable artifact to be returned was created or uploaded. This value will be defaulted to 1 month ago if distribution_ready filter is not set or set to false."),
		),
		mcp.WithString("artifact_type",
			mcp.Description("Filters for a specific artifact type or file extension for the list of installable artifacts. Available values are: 'aab' and 'apk' for android artifacts and 'ipa' for ios artifacts."),
			mcp.Enum("aab", "apk", "ipa"),
		),
		mcp.WithString("before_date",
			mcp.Description("A date in ISO 8601 string format specifying the end of the interval when the installable artifact to be returned was created or uploaded. This value will be defaulted to the current time if distribution_ready filter is not set or set to false."),
		),
		mcp.WithString("branch",
			mcp.Description("Filters for the Bitrise CI branch of the installable artifact on which it has been generated on."),
		),
		mcp.WithBoolean("distribution_ready",
			mcp.Description("Filters for distribution ready installable artifacts. This means .apk and .ipa (with distribution type ad-hoc, development, or enterprise) installable artifacts."),
		),
		mcp.WithNumber("items_per_page",
			mcp.Description("Specifies the maximum number of installable artifacts to be returned per page. Default value is 10."),
			mcp.DefaultNumber(10),
		),
		mcp.WithNumber("page",
			mcp.Description("Specifies which page should be returned from the whole result set in a paginated scenario. Default value is 1."),
			mcp.DefaultNumber(1),
		),
		mcp.WithString("platform",
			mcp.Description("Filters for a specific mobile platform for the list of installable artifacts. Available values are: 'ios' and 'android'."),
			mcp.Enum("ios", "android"),
		),
		mcp.WithString("search",
			mcp.Description("Search by version, filename or build number (Bitrise CI). The filter is case-sensitive."),
		),
		mcp.WithString("source",
			mcp.Description("Filters for the source of installable artifacts to be returned. Available values are 'api' and 'ci'."),
			mcp.Enum("api", "ci"),
		),
		mcp.WithBoolean("store_signed",
			mcp.Description("Filters for store ready installable artifacts. This means signed .aab and .ipa (with distribution type app-store) installable artifacts."),
		),
		mcp.WithString("version",
			mcp.Description("Filters for the version this installable artifact was created for. This field is required if the distribution_ready filter is set to true."),
		),
		mcp.WithString("workflow",
			mcp.Description("Filters for the Bitrise CI workflow of the installable artifact it has been generated by."),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connectedAppID, err := request.RequireString("connected_app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		params := map[string]string{}
		if v := request.GetString("after_date", ""); v != "" {
			params["after_date"] = v
		}
		if v := request.GetString("artifact_type", ""); v != "" {
			params["artifact_type"] = v
		}
		if v := request.GetString("before_date", ""); v != "" {
			params["before_date"] = v
		}
		if v := request.GetString("branch", ""); v != "" {
			params["branch"] = v
		}
		if _, ok := request.GetArguments()["distribution_ready"]; ok {
			v := request.GetBool("distribution_ready", false)
			params["distribution_ready"] = strconv.FormatBool(v)
		}
		if v := request.GetInt("items_per_page", 10); v != 10 {
			params["items_per_page"] = strconv.Itoa(v)
		}
		if v := request.GetInt("page", 1); v != 1 {
			params["page"] = strconv.Itoa(v)
		}
		if v := request.GetString("platform", ""); v != "" {
			params["platform"] = v
		}
		if v := request.GetString("search", ""); v != "" {
			params["search"] = v
		}
		if v := request.GetString("source", ""); v != "" {
			params["source"] = v
		}
		if _, ok := request.GetArguments()["store_signed"]; ok {
			v := request.GetBool("store_signed", false)
			params["store_signed"] = strconv.FormatBool(v)
		}
		if v := request.GetString("version", ""); v != "" {
			params["version"] = v
		}
		if v := request.GetString("workflow", ""); v != "" {
			params["workflow"] = v
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodGet,
			baseURL: apiRMBaseURL,
			path:    fmt.Sprintf("/v1/connected-apps/%s/installable-artifacts", connectedAppID),
			params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var generateInstallableArtifactUploadURL = Tool{
	APIGroups: []string{"release-management"},
	Definition: mcp.NewTool("generate_installable_artifact_upload_url",
		mcp.WithDescription("Generates a signed upload url valid for 1 hour for an installable artifact to be uploaded to Bitrise Release Management. The response will contain an url that can be used to upload an artifact to Bitrise Release Management using a simple curl request with the file data that should be uploaded. The necessary headers and http method will also be in the response. This artifact will need to be processed after upload to be usable. The status of processing can be checked by making another request to a different url giving back the processed status of an installable artifact."),
		mcp.WithString("connected_app_id",
			mcp.Description("Identifier of the Release Management connected app for the installable artifact. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithString("installable_artifact_id",
			mcp.Description("An uuidv4 identifier generated on the client side for the installable artifact. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithString("file_name",
			mcp.Description("The name of the installable artifact file (with extension) to be uploaded to Bitrise. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithString("file_size_bytes",
			mcp.Description("The byte size of the installable artifact file to be uploaded."),
			mcp.Required(),
		),
		mcp.WithString("branch",
			mcp.Description("Optionally you can add the name of the CI branch the installable artifact has been generated on."),
		),
		mcp.WithBoolean("with_public_page",
			mcp.Description("Optionally, you can enable public install page for your artifact. This can only be enabled by Bitrise Project Admins, Bitrise Project Owners and Bitrise Workspace Admins. Changing this value without proper permissions will result in an error. The default value is false."),
			mcp.DefaultBool(false),
		),
		mcp.WithString("workflow",
			mcp.Description("Optionally you can add the name of the CI workflow this installable artifact has been generated by."),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connectedAppID, err := request.RequireString("connected_app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		installableArtifactID, err := request.RequireString("installable_artifact_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		fileName, err := request.RequireString("file_name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		fileSizeBytes, err := request.RequireString("file_size_bytes")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		params := map[string]string{
			"file_name":       fileName,
			"file_size_bytes": fileSizeBytes,
		}
		if v := request.GetString("branch", ""); v != "" {
			params["branch"] = v
		}
		if v := request.GetBool("with_public_page", false); v {
			params["with_public_page"] = "true"
		}
		if v := request.GetString("workflow", ""); v != "" {
			params["workflow"] = v
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodGet,
			baseURL: apiRMBaseURL,
			path:    fmt.Sprintf("/v1/connected-apps/%s/installable-artifacts/%s/upload-url", connectedAppID, installableArtifactID),
			params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var getInstallableArtifactUploadAndProcessingStatus = Tool{
	APIGroups: []string{"release-management", "read-only"},
	Definition: mcp.NewTool("get_installable_artifact_upload_and_processing_status",
		mcp.WithDescription("Gets the processing and upload status of an installable artifact. An artifact will need to be processed after upload to be usable. This endpoint helps understanding when an uploaded installable artifacts becomes usable for later purposes."),
		mcp.WithString("connected_app_id",
			mcp.Description("Identifier of the Release Management connected app for the installable artifact. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithString("installable_artifact_id",
			mcp.Description("The uuidv4 identifier for the installable artifact. This field is mandatory."),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connectedAppID, err := request.RequireString("connected_app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		installableArtifactID, err := request.RequireString("installable_artifact_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodGet,
			baseURL: apiRMBaseURL,
			path:    fmt.Sprintf("/v1/connected-apps/%s/installable-artifacts/%s/status", connectedAppID, installableArtifactID),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var setInstallableArtifactPublicInstallPage = Tool{
	APIGroups: []string{"release-management"},
	Definition: mcp.NewTool("set_installable_artifact_public_install_page",
		mcp.WithDescription("Changes whether public install page should be available for the installable artifact or not."),
		mcp.WithString("connected_app_id",
			mcp.Description("Identifier of the Release Management connected app for the installable artifact. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithString("installable_artifact_id",
			mcp.Description("The uuidv4 identifier for the installable artifact. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithBoolean("with_public_page",
			mcp.Description("Boolean flag for enabling/disabling public install page for the installable artifact. This field is mandatory."),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connectedAppID, err := request.RequireString("connected_app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		installableArtifactID, err := request.RequireString("installable_artifact_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		withPublicPage, err := request.RequireBool("with_public_page")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{
			"with_public_page": withPublicPage,
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodPatch,
			baseURL: apiRMBaseURL,
			path:    fmt.Sprintf("/v1/connected-apps/%s/installable-artifacts/%s/public-install-page", connectedAppID, installableArtifactID),
			body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var listBuildDistributionVersions = Tool{
	APIGroups: []string{"release-management", "read-only"},
	Definition: mcp.NewTool("list_build_distribution_versions",
		mcp.WithDescription("Lists Build Distribution versions. Release Management offers a convenient, secure solution to distribute the builds of your mobile apps to testers without having to engage with either TestFlight or Google Play. Once you have installable artifacts, Bitrise can generate both private and public install links that testers or other stakeholders can use to install the app on real devices via over-the-air installation. Build distribution allows you to define tester groups that can receive notifications about installable artifacts. The email takes the notified testers to the test build page, from where they can install the app on their own device. Build distribution versions are the app versions available for testers."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier of the app the build distribution is connected to. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithNumber("items_per_page",
			mcp.Description("Specifies the maximum number of build distribution versions returned per page. Default value is 10."),
			mcp.DefaultNumber(10),
		),
		mcp.WithNumber("page",
			mcp.Description("Specifies which page should be returned from the whole result set in a paginated scenario. Default value is 1."),
			mcp.DefaultNumber(1),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connectedAppID, err := request.RequireString("connected_app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		params := map[string]string{}
		if v := request.GetInt("items_per_page", 10); v != 10 {
			params["items_per_page"] = strconv.Itoa(v)
		}
		if v := request.GetInt("page", 1); v != 1 {
			params["page"] = strconv.Itoa(v)
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodGet,
			baseURL: apiRMBaseURL,
			path:    fmt.Sprintf("/v1/connected-apps/%s/build-distributions", connectedAppID),
			params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var listBuildDistributionVersionTestBuilds = Tool{
	APIGroups: []string{"release-management", "read-only"},
	Definition: mcp.NewTool("list_build_distribution_version_test_builds",
		mcp.WithDescription("Gives back a list of test builds for the given build distribution version."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier of the app the build distribution is connected to. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithString("version",
			mcp.Description("The version of the build distribution. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithNumber("items_per_page",
			mcp.Description("Specifies the maximum number of test builds to return for a build distribution version per page. Default value is 10."),
			mcp.DefaultNumber(10),
		),
		mcp.WithNumber("page",
			mcp.Description("Specifies which page should be returned from the whole result set in a paginated scenario. Default value is 1."),
			mcp.DefaultNumber(1),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connectedAppID, err := request.RequireString("connected_app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		version, err := request.RequireString("version")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		params := map[string]string{
			"version": version,
		}
		if v := request.GetInt("items_per_page", 10); v != 10 {
			params["items_per_page"] = strconv.Itoa(v)
		}
		if v := request.GetInt("page", 1); v != 1 {
			params["page"] = strconv.Itoa(v)
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodGet,
			baseURL: apiRMBaseURL,
			path:    fmt.Sprintf("/v1/connected-apps/%s/build-distributions/test-builds", connectedAppID),
			params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var createTesterGroup = Tool{
	APIGroups: []string{"release-management"},
	Definition: mcp.NewTool("create_tester_group",
		mcp.WithDescription("Creates a tester group for a Release Management connected app. Tester groups can be used to distribute installable artifacts to testers automatically. When a new installable artifact is available, the tester groups can either automatically or manually be notified via email. The notification email will contain a link to the installable artifact page for the artifact within Bitrise Release Management. A Release Management connected app can have multiple tester groups. Project team members of the connected app can be selected to be testers and added to the tester group. This endpoint has an elevated access level requirement. Only the owner of the related Bitrise Workspace, a workspace manager or the related project's admin can manage tester groups."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier of the related Release Management connected app."),
			mcp.Required(),
		),
		mcp.WithString("name",
			mcp.Description("The name for the new tester group. Must be unique in the scope of the connected app."),
		),
		mcp.WithBoolean("auto_notify",
			mcp.Description("If set to true it indicates that the tester group will receive notifications automatically."),
			mcp.DefaultBool(false),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connectedAppID, err := request.RequireString("connected_app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{}
		if v := request.GetString("name", ""); v != "" {
			body["name"] = v
		}
		if v := request.GetBool("auto_notify", false); v {
			body["auto_notify"] = v
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodPost,
			baseURL: apiRMBaseURL,
			path:    fmt.Sprintf("/v1/connected-apps/%s/tester-groups", connectedAppID),
			body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var notifyTesterGroup = Tool{
	APIGroups: []string{"release-management"},
	Definition: mcp.NewTool("notify_tester_group",
		mcp.WithDescription("Notifies a tester group about a new test build."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier of the related Release Management connected app."),
			mcp.Required(),
		),
		mcp.WithString("id",
			mcp.Description("The uuidV4 identifier of the tester group whose members will be notified about the test build."),
			mcp.Required(),
		),
		mcp.WithString("test_build_id",
			mcp.Description("The unique identifier of the test build what will be sent in the notification of the tester group."),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connectedAppID, err := request.RequireString("connected_app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		testBuildID, err := request.RequireString("test_build_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{
			"test_build_id": testBuildID,
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodPost,
			baseURL: apiRMBaseURL,
			path:    fmt.Sprintf("/v1/connected-apps/%s/tester-groups/%s/notify", connectedAppID, id),
			body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var addTestersToTesterGroup = Tool{
	APIGroups: []string{"release-management"},
	Definition: mcp.NewTool("add_testers_to_tester_group",
		mcp.WithDescription("Adds testers to a tester group of a connected app."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier of the related Release Management connected app."),
			mcp.Required(),
		),
		mcp.WithString("id",
			mcp.Description("The uuidV4 identifier of the tester group to which testers will be added."),
			mcp.Required(),
		),
		mcp.WithArray("user_slugs",
			mcp.Description("The list of users identified by slugs that will be added to the tester group."),
			mcp.Required(),
			mcp.WithStringItems(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connectedAppID, err := request.RequireString("connected_app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		userSlugs, err := request.RequireStringSlice("user_slugs")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{
			"user_slugs": userSlugs,
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodPost,
			baseURL: apiRMBaseURL,
			path:    fmt.Sprintf("/v1/connected-apps/%s/tester-groups/%s/add-testers", connectedAppID, id),
			body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var updateTesterGroup = Tool{
	APIGroups: []string{"release-management"},
	Definition: mcp.NewTool("update_tester_group",
		mcp.WithDescription("Updates the given tester group. The name and the auto notification setting can be updated optionally."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier of the related Release Management connected app."),
			mcp.Required(),
		),
		mcp.WithString("id",
			mcp.Description("The uuidV4 identifier of the tester group to which testers will be added."),
			mcp.Required(),
		),
		mcp.WithString("name",
			mcp.Description("The new name for the tester group. Must be unique in the scope of the related connected app."),
		),
		mcp.WithBoolean("auto_notify",
			mcp.Description("If set to true it indicates the tester group will receive email notifications automatically from now on about new installable builds."),
			mcp.DefaultBool(false),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connectedAppID, err := request.RequireString("connected_app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{}
		if v := request.GetString("name", ""); v != "" {
			body["name"] = v
		}
		if v := request.GetBool("auto_notify", false); v {
			body["auto_notify"] = v
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodPut,
			baseURL: apiRMBaseURL,
			path:    fmt.Sprintf("/v1/connected-apps/%s/tester-groups/%s", connectedAppID, id),
			body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var listTesterGroups = Tool{
	APIGroups: []string{"release-management", "read-only"},
	Definition: mcp.NewTool("list_tester_groups",
		mcp.WithDescription("Gives back a list of tester groups related to a specific Release Management connected app."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier of the app the tester group is connected to. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithNumber("items_per_page",
			mcp.Description("Specifies the maximum number of tester groups to return related to a specific connected app. Default value is 10."),
			mcp.DefaultNumber(10),
		),
		mcp.WithNumber("page",
			mcp.Description("Specifies which page should be returned from the whole result set in a paginated scenario. Default value is 1."),
			mcp.DefaultNumber(1),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connectedAppID, err := request.RequireString("connected_app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		params := map[string]string{}
		if v := request.GetInt("items_per_page", 10); v != 10 {
			params["items_per_page"] = strconv.Itoa(v)
		}
		if v := request.GetInt("page", 1); v != 1 {
			params["page"] = strconv.Itoa(v)
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodGet,
			baseURL: apiRMBaseURL,
			path:    fmt.Sprintf("/v1/connected-apps/%s/tester-groups", connectedAppID),
			params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var getTesterGroup = Tool{
	APIGroups: []string{"release-management", "read-only"},
	Definition: mcp.NewTool("get_tester_group",
		mcp.WithDescription("Gives back the details of the selected tester group."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier of the app the tester group is connected to. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithString("id",
			mcp.Description("The uuidV4 identifier of the tester group. This field is mandatory."),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connectedAppID, err := request.RequireString("connected_app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodGet,
			baseURL: apiRMBaseURL,
			path:    fmt.Sprintf("/v1/connected-apps/%s/tester-groups/%s", connectedAppID, id),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var getPotentialTesters = Tool{
	APIGroups: []string{"release-management", "read-only"},
	Definition: mcp.NewTool("get_potential_testers",
		mcp.WithDescription("Gets a list of potential testers whom can be added as testers to a specific tester group. The list consists of Bitrise users having access to the related Release Management connected app."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier of the app the tester group is connected to. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithString("id",
			mcp.Description("The uuidV4 identifier of the tester group. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithNumber("items_per_page",
			mcp.Description("Specifies the maximum number of potential testers to return having access to a specific connected app. Default value is 10."),
			mcp.DefaultNumber(10),
		),
		mcp.WithNumber("page",
			mcp.Description("Specifies which page should be returned from the whole result set in a paginated scenario. Default value is 1."),
			mcp.DefaultNumber(1),
		),
		mcp.WithString("search",
			mcp.Description("Searches for potential testers based on email or username using a case-insensitive approach."),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connectedAppID, err := request.RequireString("connected_app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		params := map[string]string{}
		if v := request.GetInt("items_per_page", 10); v != 10 {
			params["items_per_page"] = strconv.Itoa(v)
		}
		if v := request.GetInt("page", 1); v != 1 {
			params["page"] = strconv.Itoa(v)
		}
		if v := request.GetString("search", ""); v != "" {
			params["search"] = v
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodGet,
			baseURL: apiRMBaseURL,
			path:    fmt.Sprintf("/v1/connected-apps/%s/tester-groups/%s/potential-testers", connectedAppID, id),
			params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var getTesters = Tool{
	APIGroups: []string{"release-management", "read-only"},
	Definition: mcp.NewTool("get_testers",
		mcp.WithDescription("Gives back a list of testers that has been associated with a tester group related to a specific connected app."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier of the app the tester group is connected to. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithString("tester_group_id",
			mcp.Description("The uuidV4 identifier of a tester group. If given, only testers within this specific tester group will be returned."),
		),
		mcp.WithNumber("items_per_page",
			mcp.Description("Specifies the maximum number of testers to be returned that have been added to a tester group related to the specific connected app. Default value is 10."),
			mcp.DefaultNumber(10),
		),
		mcp.WithNumber("page",
			mcp.Description("Specifies which page should be returned from the whole result set in a paginated scenario. Default value is 1."),
			mcp.DefaultNumber(1),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connectedAppID, err := request.RequireString("connected_app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		params := map[string]string{}
		if v := request.GetString("tester_group_id", ""); v != "" {
			params["tester_group_id"] = v
		}
		if v := request.GetInt("items_per_page", 10); v != 10 {
			params["items_per_page"] = strconv.Itoa(v)
		}
		if v := request.GetInt("page", 1); v != 1 {
			params["page"] = strconv.Itoa(v)
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodGet,
			baseURL: apiRMBaseURL,
			path:    fmt.Sprintf("/v1/connected-apps/%s/testers", connectedAppID),
			params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
