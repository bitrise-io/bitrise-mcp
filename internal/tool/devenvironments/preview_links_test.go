package devenvironments

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/devenv"
	"github.com/mark3labs/mcp-go/mcp"
)

// Unlike a device session, a preview link can default neither the device nor
// the app: there is nothing to preview without both. mcp-go only marks
// top-level parameters as required, so the nested lists are set by hand and
// easy to lose.
func TestCreatePreviewLinkNestedRequired(t *testing.T) {
	raw, err := json.Marshal(CreatePreviewLink.Definition.InputSchema)
	if err != nil {
		t.Fatalf("marshal input schema: %v", err)
	}
	var schema struct {
		Properties map[string]struct {
			Type     string   `json:"type"`
			Required []string `json:"required"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("unmarshal input schema: %v", err)
	}
	for name, want := range map[string][]string{
		"device_spec": {"platform"},
		"artifact":    {"url"},
	} {
		prop, ok := schema.Properties[name]
		if !ok {
			t.Fatalf("property %s missing from schema", name)
		}
		if prop.Type != "object" {
			t.Errorf("property %s: type %q, want object", name, prop.Type)
		}
		if !reflect.DeepEqual(prop.Required, want) {
			t.Errorf("property %s: required %v, want %v", name, prop.Required, want)
		}
	}
}

// Everything the backend is certain to reject is rejected before the round
// trip, so the agent gets the reason instead of a 400 it has to interpret.
func TestCreatePreviewLinkRejectsIncompleteRequest(t *testing.T) {
	artifact := map[string]any{"url": "https://example.com/app.apk"}
	cases := []struct {
		name string
		args map[string]any
		want string
	}{
		{"no device_spec", map[string]any{"artifact": artifact}, "device_spec is required"},
		{"device_spec not an object", map[string]any{"device_spec": "ios", "artifact": artifact}, "device_spec must be an object"},
		{"device_spec missing platform", map[string]any{"device_spec": map[string]any{"device_model": "pixel_7"}, "artifact": artifact}, "device_spec.platform is required"},
		{"device_spec empty platform", map[string]any{"device_spec": map[string]any{"platform": "  "}, "artifact": artifact}, "device_spec.platform is required"},
		{"no artifact", map[string]any{"device_spec": map[string]any{"platform": "ios"}}, "artifact is required"},
		{"artifact missing url", map[string]any{"device_spec": map[string]any{"platform": "ios"}, "artifact": map[string]any{"app_name": "Demo"}}, "artifact.url is required"},
		{"artifact empty url", map[string]any{"device_spec": map[string]any{"platform": "ios"}, "artifact": map[string]any{"url": ""}}, "artifact.url is required"},
		{"ttl_seconds not a number", map[string]any{"device_spec": map[string]any{"platform": "ios"}, "artifact": artifact, "ttl_seconds": "86400"}, "ttl_seconds must be a number"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := mcp.CallToolRequest{}
			req.Params.Arguments = tc.args
			res, err := CreatePreviewLink.Handler(context.Background(), req)
			if err != nil {
				t.Fatalf("handler returned error: %v", err)
			}
			if !res.IsError {
				t.Fatalf("expected a tool error result, got success")
			}
			text, ok := res.Content[0].(mcp.TextContent)
			if !ok {
				t.Fatalf("content is %T, want TextContent", res.Content[0])
			}
			if !strings.Contains(text.Text, tc.want) {
				t.Errorf("error %q does not contain %q", text.Text, tc.want)
			}
		})
	}
}

// The happy path: device_spec and artifact are forwarded verbatim, and the
// optional knobs are omitted rather than sent as zeros — 0 means "deployment
// default" on ttl_seconds but is rejected outright on stack_id/machine_type,
// so sending them unasked would change the link.
func TestCreatePreviewLinkForwardsMintRequest(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workspaces/ws/preview-links" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("unmarshal body %q: %v", raw, err)
		}
		_, _ = w.Write([]byte(`{"url":"https://app.bitrise.io/dev-environments/ws/device-preview/tok","jti":"link-1"}`))
	}))
	defer srv.Close()
	old := devenv.BaseURL
	devenv.BaseURL = srv.URL
	defer func() { devenv.BaseURL = old }()

	deviceSpec := map[string]any{"platform": "android", "device_model": "pixel_7"}
	artifact := map[string]any{"url": "https://example.com/app.apk", "app_name": "Demo", "build_number": "42"}
	ctx := devenv.ContextWithWorkspace(devenv.ContextWithPAT(context.Background(), "pat"), "ws")
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"device_spec": deviceSpec, "artifact": artifact}
	res, err := CreatePreviewLink.Handler(ctx, req)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %v", res.Content)
	}
	if got == nil {
		t.Fatalf("backend was never called")
	}
	if !reflect.DeepEqual(got["device_spec"], deviceSpec) {
		t.Errorf("device_spec = %v, want %v", got["device_spec"], deviceSpec)
	}
	if !reflect.DeepEqual(got["artifact"], artifact) {
		t.Errorf("artifact = %v, want %v", got["artifact"], artifact)
	}
	for _, key := range []string{"ttl_seconds", "session_auto_terminate_minutes", "stack_id", "machine_type"} {
		if v, present := got[key]; present {
			t.Errorf("%s = %v present in body, want absent", key, v)
		}
	}
}

// The optional knobs reach the backend when given. ttl_seconds arrives as a
// JSON number even though the proto field is an int64 (which proto3 JSON
// would also accept as a string).
func TestCreatePreviewLinkForwardsOptionalKnobs(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("unmarshal body %q: %v", raw, err)
		}
		_, _ = w.Write([]byte(`{"jti":"link-1"}`))
	}))
	defer srv.Close()
	old := devenv.BaseURL
	devenv.BaseURL = srv.URL
	defer func() { devenv.BaseURL = old }()

	ctx := devenv.ContextWithWorkspace(devenv.ContextWithPAT(context.Background(), "pat"), "ws")
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"device_spec":                    map[string]any{"platform": "ios"},
		"artifact":                       map[string]any{"url": "https://example.com/App.zip"},
		"ttl_seconds":                    float64(7200),
		"session_auto_terminate_minutes": float64(15),
		"stack_id":                       "osx-27-edge",
		"machine_type":                   "g2.mac.m2pro.4c",
	}
	if _, err := CreatePreviewLink.Handler(ctx, req); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	for key, want := range map[string]any{
		"ttl_seconds":                    float64(7200),
		"session_auto_terminate_minutes": float64(15),
		"stack_id":                       "osx-27-edge",
		"machine_type":                   "g2.mac.m2pro.4c",
	} {
		if got[key] != want {
			t.Errorf("%s = %v (%T), want %v", key, got[key], got[key], want)
		}
	}
}
