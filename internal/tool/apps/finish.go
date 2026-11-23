package apps

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

// projectConfigs lists the project types and config presets that
// POST /apps/{slug}/finish accepts with mode "manual", the first config of
// each project type being the default. The API answers 400 Bad Request when
// the project_type/config pair is not one of these.
//
// No public API or OpenAPI document lists the accepted pairs (the swagger
// description of the config field is out of date), so keep this table in sync
// with the API by hand.
var projectConfigs = []struct {
	ProjectType string
	Configs     []string
}{
	{"android", []string{"default-android-config", "default-android-config-kts"}},
	{"cordova", []string{"default-cordova-config"}},
	{"fastlane", []string{"default-fastlane-android-config", "default-fastlane-ios-config"}},
	{"flutter", []string{"flutter-config-test-ios-android-web-0"}},
	{"ionic", []string{"default-ionic-config"}},
	{"ios", []string{"default-ios-config"}},
	{"java", []string{"default-java-gradle-config", "default-java-maven-config"}},
	{"kotlin-multiplatform", []string{
		"default-kotlin-multiplatform-config",
		"default-kotlin-multiplatform-config-android",
		"default-kotlin-multiplatform-config-android-ios",
		"default-kotlin-multiplatform-config-ios",
	}},
	{"macos", []string{"default-macos-config"}},
	{"node-js", []string{"default-node-js-npm-config", "default-node-js-yarn-config"}},
	{"other", []string{"other-config"}},
	{"python", []string{"default-python-pip-config", "default-python-poetry-config", "default-python-uv-config"}},
	{"react-native", []string{"default-react-native-config", "default-react-native-expo-config"}},
	{"ruby", []string{"default-ruby-config"}},
}

func projectTypes() []string {
	types := make([]string, 0, len(projectConfigs))
	for _, pc := range projectConfigs {
		types = append(types, pc.ProjectType)
	}
	return types
}

func allConfigs() []string {
	var configs []string
	for _, pc := range projectConfigs {
		configs = append(configs, pc.Configs...)
	}
	return configs
}

func configsFor(projectType string) []string {
	for _, pc := range projectConfigs {
		if pc.ProjectType == projectType {
			return pc.Configs
		}
	}
	return nil
}

func configCatalogDescription() string {
	lines := make([]string, 0, len(projectConfigs))
	for _, pc := range projectConfigs {
		lines = append(lines, fmt.Sprintf("- %s: %s", pc.ProjectType, strings.Join(pc.Configs, ", ")))
	}
	return strings.Join(lines, "\n")
}

var Finish = bitrise.Tool{
	APIGroups: []string{"apps"},
	Definition: mcp.NewTool("finish_bitrise_app",
		mcp.WithDescription("Finish the setup of a Bitrise app. If this is successful, a build can be triggered via trigger_bitrise_build. "+
			"If you have access to the repository, decide the project type from it, pick a stack ID from https://stacks.bitrise.io/, and pick a config that belongs to that project type. "+
			"The config must be one listed for the chosen project type, otherwise the API rejects the request. Valid project_type: config pairs (the first config is the default):\n"+
			configCatalogDescription()),
		mcp.WithString("app_slug",
			mcp.Description("The slug of the Bitrise app to finish setup for."),
			mcp.Required(),
		),
		mcp.WithString("project_type",
			mcp.Description("The type of project. Determines which config values are valid."),
			mcp.Required(),
			mcp.Enum(projectTypes()...),
			mcp.DefaultString("other"),
		),
		mcp.WithString("stack_id",
			mcp.Description("The stack ID to use for the app."),
			mcp.DefaultString("linux-docker-android-22.04"),
			mcp.Required(),
		),
		mcp.WithString("config",
			mcp.Description("The configuration preset to use for the app. Must belong to the chosen project_type (see the tool description). Defaults to the first config of the project_type."),
			mcp.Enum(allConfigs()...),
		),
		mcp.WithTitleAnnotation("Finish App Registration"),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		projectType, err := request.RequireString("project_type")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		stackID, err := request.RequireString("stack_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		validConfigs := configsFor(projectType)
		if validConfigs == nil {
			return mcp.NewToolResultErrorf("unknown project_type %q; valid values: %s", projectType, strings.Join(projectTypes(), ", ")), nil
		}
		config := request.GetString("config", "")
		if config == "" {
			config = validConfigs[0]
		}
		if !slices.Contains(validConfigs, config) {
			return mcp.NewToolResultErrorf("config %q is not valid for project_type %q; valid configs for %s: %s",
				config, projectType, projectType, strings.Join(validConfigs, ", ")), nil
		}

		body := map[string]any{
			"project_type": projectType,
			"stack_id":     stackID,
			"mode":         "manual",
			"config":       config,
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/finish", appSlug),
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
