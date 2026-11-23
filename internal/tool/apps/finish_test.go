package apps

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

// wantCatalog lists the project_type/config pairs the finish API accepts.
// Update it, and projectConfigs, when the API changes; finish_bitrise_app gets
// 400 Bad Request for any pair that is not in it.
var wantCatalog = map[string][]string{
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

func TestFinish_EnumsMatchCatalog(t *testing.T) {
	props := Finish.Definition.InputSchema.Properties
	enumOf := func(name string) []string {
		prop, _ := props[name].(map[string]any)
		values, ok := prop["enum"].([]string)
		assert.True(t, ok, "property %s has no enum", name)
		return values
	}

	var wantTypes, wantConfigs []string
	for projectType, configs := range wantCatalog {
		wantTypes = append(wantTypes, projectType)
		wantConfigs = append(wantConfigs, configs...)
		assert.Equal(t, configs, configsFor(projectType), "configs of %s", projectType)
	}
	assert.ElementsMatch(t, wantTypes, enumOf("project_type"))
	assert.ElementsMatch(t, wantConfigs, enumOf("config"))

	desc := Finish.Definition.Description
	for _, config := range wantConfigs {
		assert.Contains(t, desc, config, "description should name every config")
	}
}

func TestFinish_Request(t *testing.T) {
	cases := map[string]struct {
		args       map[string]any
		wantConfig string // empty: the API must not be called
		wantError  string
	}{
		"flutter defaults to its only config": {
			args:       map[string]any{"project_type": "flutter"},
			wantConfig: "flutter-config-test-ios-android-web-0",
		},
		"explicit config of the project type": {
			args:       map[string]any{"project_type": "android", "config": "default-android-config-kts"},
			wantConfig: "default-android-config-kts",
		},
		"newly added project type": {
			args:       map[string]any{"project_type": "python", "config": "default-python-uv-config"},
			wantConfig: "default-python-uv-config",
		},
		"removed flutter config is rejected before the API call": {
			args:      map[string]any{"project_type": "flutter", "config": "flutter-config-test-both-0"},
			wantError: `config "flutter-config-test-both-0" is not valid for project_type "flutter"; valid configs for flutter: flutter-config-test-ios-android-web-0`,
		},
		"config of another project type is rejected": {
			args:      map[string]any{"project_type": "flutter", "config": "default-ios-config"},
			wantError: `config "default-ios-config" is not valid for project_type "flutter"`,
		},
		"unknown project type is rejected": {
			args:      map[string]any{"project_type": "xamarin"},
			wantError: `unknown project_type "xamarin"`,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var gotBody map[string]any
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.wantConfig == "" {
					t.Errorf("unexpected API call: %s %s", r.Method, r.URL.Path)
				}
				assert.Equal(t, "/apps/app-1/finish", r.URL.Path)
				assert.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
				_, _ = w.Write([]byte(`{"status":"ok"}`))
			}))
			defer srv.Close()

			orig := bitrise.APIBaseURL
			bitrise.APIBaseURL = srv.URL
			defer func() { bitrise.APIBaseURL = orig }()

			args := map[string]any{"app_slug": "app-1", "stack_id": "linux-docker-android-22.04"}
			for k, v := range tc.args {
				args[k] = v
			}
			req := mcp.CallToolRequest{}
			req.Params.Arguments = args
			res, err := Finish.Handler(bitrise.ContextWithPAT(context.Background(), "token"), req)
			if !assert.NoError(t, err) {
				return
			}

			if tc.wantError != "" {
				if assert.True(t, res.IsError) {
					assert.Contains(t, res.Content[0].(mcp.TextContent).Text, tc.wantError)
				}
				return
			}
			if !assert.False(t, res.IsError, "tool returned an error: %v", res.Content) {
				return
			}
			assert.Equal(t, map[string]any{
				"project_type": tc.args["project_type"],
				"stack_id":     "linux-docker-android-22.04",
				"mode":         "manual",
				"config":       tc.wantConfig,
			}, gotBody)
		})
	}
}

func TestFinish_APIErrorBodyReachesAgent(t *testing.T) {
	const apiMessage = `{"message":"Invalid project_type and config combination"}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(apiMessage))
	}))
	defer srv.Close()

	orig := bitrise.APIBaseURL
	bitrise.APIBaseURL = srv.URL
	defer func() { bitrise.APIBaseURL = orig }()

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"app_slug": "app-1", "project_type": "ios", "stack_id": "osx-xcode-26.0.x"}
	res, err := Finish.Handler(bitrise.ContextWithPAT(context.Background(), "token"), req)
	if !assert.NoError(t, err) || !assert.True(t, res.IsError) {
		return
	}
	text := res.Content[0].(mcp.TextContent).Text
	assert.Contains(t, text, "400")
	assert.Contains(t, text, apiMessage)
}
