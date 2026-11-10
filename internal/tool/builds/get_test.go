package builds

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

func TestGet_StackImageVersion(t *testing.T) {
	const build = `{"data":{"slug":"build-1","stack_identifier":"osx-xcode-26.0.x"}}`
	cases := map[string]struct {
		summaryStatus int
		summaryBody   string
		verbose       bool
		want          any // nil means the field is absent
		wantWarning   string
	}{
		"summary has image version": {
			summaryStatus: http.StatusOK,
			summaryBody:   `{"app_id":"app-1","agent_info":{"stack_image_version":"2-159-0","machine_type":"g2.mac.medium"}}`,
			want:          "2-159-0",
		},
		"summary has image version, verbose": {
			summaryStatus: http.StatusOK,
			summaryBody:   `{"agent_info":{"stack_image_version":"2-159-0"}}`,
			verbose:       true,
			want:          "2-159-0",
		},
		"summary not found": {
			summaryStatus: http.StatusNotFound,
			summaryBody:   `{"message":"not found"}`,
			wantWarning:   "could not fetch stack image version: unexpected status code 404",
		},
		"summary without agent_info": {
			summaryStatus: http.StatusOK,
			summaryBody:   `{"app_id":"app-1"}`,
		},
		"agent_info with empty image version": {
			summaryStatus: http.StatusOK,
			summaryBody:   `{"agent_info":{"stack_image_version":""}}`,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/apps/app-1/builds/build-1":
					_, _ = w.Write([]byte(build))
				case "/apps/app-1/builds/build-1/log/summary":
					w.WriteHeader(tc.summaryStatus)
					_, _ = w.Write([]byte(tc.summaryBody))
				default:
					t.Errorf("unexpected request: %s", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer srv.Close()

			orig := bitrise.APIBaseURL
			bitrise.APIBaseURL = srv.URL
			defer func() { bitrise.APIBaseURL = orig }()

			req := mcp.CallToolRequest{}
			req.Params.Arguments = map[string]any{"app_slug": "app-1", "build_slug": "build-1", "verbose": tc.verbose}
			res, err := Get.Handler(bitrise.ContextWithPAT(context.Background(), "token"), req)
			if !assert.NoError(t, err) || !assert.False(t, res.IsError, "tool returned an error: %v", res.Content) {
				return
			}

			result := res.StructuredContent.(map[string]any)
			if tc.wantWarning == "" {
				assert.NotContains(t, result, "warnings")
			} else if assert.Len(t, result["warnings"], 1) {
				assert.Contains(t, result["warnings"].([]string)[0], tc.wantWarning)
			}

			data := result["data"].(map[string]any)
			assert.Equal(t, "osx-xcode-26.0.x", data["stack_identifier"])
			v, ok := data["stack_image_version"]
			if tc.want == nil {
				assert.False(t, ok, "stack_image_version should be absent, got %v", v)
			} else {
				assert.Equal(t, tc.want, v)
			}
		})
	}
}

func TestGet_BuildErrorStillFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	orig := bitrise.APIBaseURL
	bitrise.APIBaseURL = srv.URL
	defer func() { bitrise.APIBaseURL = orig }()

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"app_slug": "app-1", "build_slug": "build-1"}
	res, err := Get.Handler(bitrise.ContextWithPAT(context.Background(), "token"), req)
	assert.NoError(t, err)
	assert.True(t, res.IsError)
}
