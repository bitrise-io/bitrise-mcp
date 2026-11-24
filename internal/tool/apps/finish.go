package apps

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strings"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

// projectConfigs lists the configs POST /apps/{slug}/finish accepts per project
// type with mode "manual". No API lists them, so keep this in sync by hand.
var projectConfigs = map[string][]string{
	"android":              {"default-android-config", "default-android-config-kts"},
	"cordova":              {"default-cordova-config"},
	"fastlane":             {"default-fastlane-android-config", "default-fastlane-ios-config"},
	"flutter":              {"flutter-config-test-ios-android-web-0"},
	"ionic":                {"default-ionic-config"},
	"ios":                  {"default-ios-config"},
	"java":                 {"default-java-gradle-config", "default-java-maven-config"},
	"kotlin-multiplatform": {"default-kotlin-multiplatform-config", "default-kotlin-multiplatform-config-android", "default-kotlin-multiplatform-config-android-ios", "default-kotlin-multiplatform-config-ios"},
	"macos":                {"default-macos-config"},
	"node-js":              {"default-node-js-npm-config", "default-node-js-yarn-config"},
	"other":                {"other-config"},
	"python":               {"default-python-pip-config", "default-python-poetry-config", "default-python-uv-config"},
	"react-native":         {"default-react-native-config", "default-react-native-expo-config"},
	"ruby":                 {"default-ruby-config"},
}

var projectTypes, allConfigs, configPairs = func() (types, configs []string, pairs string) {
	types = slices.Sorted(maps.Keys(projectConfigs))
	for _, t := range types {
		configs = append(configs, projectConfigs[t]...)
		pairs += fmt.Sprintf("\n- %s: %s", t, strings.Join(projectConfigs[t], ", "))
	}
	return types, configs, pairs
}()

var Finish = bitrise.Tool{
	APIGroups: []string{"apps"},
	Definition: mcp.NewTool("finish_bitrise_app",
		mcp.WithDescription("Finish the setup of a Bitrise app. If this is successful, a build can be triggered via trigger_bitrise_build. If you have access to the repository, decide the project type, the stack ID, and the config to use, based on https://stacks.bitrise.io/, and the config should be also based on the project type. Configs per project type (config is required if it has several):"+configPairs),
		mcp.WithString("app_slug",
			mcp.Description("The slug of the Bitrise app to finish setup for."),
			mcp.Required(),
		),
		mcp.WithString("project_type",
			mcp.Description("The type of project"),
			mcp.Required(),
			mcp.Enum(projectTypes...),
			mcp.DefaultString("other"),
		),
		mcp.WithString("stack_id",
			mcp.Description("The stack ID to use for the app."),
			mcp.DefaultString("linux-docker-android-22.04"),
			mcp.Required(),
		),
		mcp.WithString("config",
			mcp.Description("The configuration preset to use for the app."),
			mcp.Enum(allConfigs...),
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

		validConfigs := projectConfigs[projectType]
		config := request.GetString("config", "")
		if config == "" && len(validConfigs) == 1 {
			config = validConfigs[0]
		}
		if !slices.Contains(validConfigs, config) {
			return mcp.NewToolResultErrorf("config %q is not valid for project_type %q; use one of: %s", config, projectType, strings.Join(validConfigs, ", ")), nil
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
