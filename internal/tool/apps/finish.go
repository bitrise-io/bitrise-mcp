package apps

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var Finish = bitrise.Tool{
	APIGroups: []string{"apps"},
	Definition: mcp.NewTool("finish_bitrise_app",
		mcp.WithDescription("Finish the setup of a Bitrise app. If this is successful, a build can be triggered via trigger_bitrise_build. If you have access to the repository, decide the project type, the stack ID, and the config to use, based on https://stacks.bitrise.io/, and the config should be also based on the project type."),
		mcp.WithString("app_slug",
			mcp.Description("The slug of the Bitrise app to finish setup for."),
			mcp.Required(),
		),
		mcp.WithString("project_type",
			mcp.Description("The type of project"),
			mcp.Required(),
			mcp.Enum("android", "cordova", "fastlane", "flutter", "ios", "ionic", "java", "kotlin-multiplatform", "macos", "node-js", "react-native", "other"),
			mcp.DefaultString("other"),
		),
		mcp.WithString("stack_id",
			mcp.Description("The stack ID to use for the app."),
			mcp.DefaultString("linux-docker-android-22.04"),
			mcp.Required(),
		),
		mcp.WithString("config",
			mcp.Description("The configuration preset to use for the app."),
			mcp.Enum(
				"default-android-config",
				"default-android-config-kts",
				"default-cordova-config",
				"default-fastlane-android-config",
				"default-fastlane-ios-config",
				"flutter-config-test-android-2",
				"flutter-config-test-both-0",
				"flutter-config-test-ios-1",
				"default-ionic-config",
				"default-ios-config",
				"default-java-gradle-config",
				"default-java-maven-config",
				"default-kotlin-multiplatform-config",
				"default-kotlin-multiplatform-config-ios",
				"default-kotlin-multiplatform-config-android",
				"default-kotlin-multiplatform-config-android-ios",
				"default-macos-config",
				"default-node-js-npm-config",
				"default-node-js-yarn-config",
				"default-react-native-config",
				"default-react-native-expo-config",
			),
			mcp.DefaultString("other-config"),
		),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(true),
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

		body := map[string]any{
			"project_type": projectType,
			"stack_id":     stackID,
			"mode":         "manual",
			"config":       request.GetString("config", "other-config"),
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
