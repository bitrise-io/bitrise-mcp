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

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/devenv"
	"github.com/mark3labs/mcp-go/mcp"
)

const testTemplateID = "0f1e2d3c-4b5a-4978-8877-665544332211"

// captureBody points the API client at a fake backend for the duration of the
// test and returns a context authenticated against workspace "ws" plus a
// pointer that holds the JSON body of the first request matching method and
// path (nil until the backend is called).
func captureBody(t *testing.T, method, path string) (context.Context, *map[string]any) {
	t.Helper()
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method || r.URL.Path != path {
			t.Errorf("unexpected request %s %s, want %s %s", r.Method, r.URL.Path, method, path)
		}
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("unmarshal body %q: %v", raw, err)
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	old := devenv.BaseURL
	devenv.BaseURL = srv.URL
	t.Cleanup(func() { devenv.BaseURL = old })
	ctx := devenv.ContextWithWorkspace(devenv.ContextWithPAT(context.Background(), "pat"), "ws")
	return ctx, &got
}

// callOK runs the tool handler and fails the test unless it returns a
// successful result.
func callOK(t *testing.T, tool bitrise.Tool, ctx context.Context, args map[string]any) {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args
	res, err := tool.Handler(ctx, req)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %v", res.Content)
	}
}

// callErr runs the tool handler and returns the text of the tool error it
// produced, failing the test if it succeeded instead.
func callErr(t *testing.T, tool bitrise.Tool, args map[string]any) string {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args
	res, err := tool.Handler(context.Background(), req)
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
	return text.Text
}

// The device_spec object is declared on every tool that accepts one through
// the same helper, so its properties must match bitrise_devenv_create's. The
// advisory platform requirement differs on purpose: a template's device is a
// complete device (platform required), a session's device_spec may omit the
// platform to tweak the template's device per field.
func TestTemplateToolsShareDeviceSpecSchema(t *testing.T) {
	type prop struct {
		Type       string         `json:"type"`
		Required   []string       `json:"required"`
		Properties map[string]any `json:"properties"`
	}
	schemaOf := func(tool bitrise.Tool) prop {
		raw, err := json.Marshal(tool.Definition.InputSchema)
		if err != nil {
			t.Fatalf("marshal input schema: %v", err)
		}
		var schema struct {
			Properties map[string]prop `json:"properties"`
		}
		if err := json.Unmarshal(raw, &schema); err != nil {
			t.Fatalf("unmarshal input schema: %v", err)
		}
		p, ok := schema.Properties["device_spec"]
		if !ok {
			t.Fatalf("%s: device_spec missing from schema", tool.Definition.Name)
		}
		return p
	}
	want := schemaOf(CreateSession)
	if want.Type != "object" || want.Required != nil {
		t.Fatalf("bitrise_devenv_create device_spec: type %q required %v (a session's platform must stay optional for template overrides)", want.Type, want.Required)
	}
	for _, tool := range []bitrise.Tool{CreateTemplate, UpdateTemplate} {
		got := schemaOf(tool)
		if got.Type != want.Type || !reflect.DeepEqual(got.Properties, want.Properties) {
			t.Errorf("%s: device_spec schema differs from bitrise_devenv_create's", tool.Definition.Name)
		}
		if !reflect.DeepEqual(got.Required, []string{"platform"}) {
			t.Errorf("%s: device_spec required %v, want [platform] — a template's device is a complete device", tool.Definition.Name, got.Required)
		}
	}
}

// A template created with a device_spec forwards it verbatim; one created
// without sends no device_spec key at all (absent = the template declares no
// device).
func TestCreateTemplateForwardsDeviceSpec(t *testing.T) {
	deviceSpec := map[string]any{"platform": "android", "device_model": "pixel_7", "cold_boot": true}
	base := map[string]any{"name": "android lab", "stack_id": "ubuntu-android", "machine_type": "g2.linux.large"}

	t.Run("with device", func(t *testing.T) {
		ctx, got := captureBody(t, http.MethodPost, "/v1/workspaces/ws/templates")
		args := map[string]any{"device_spec": deviceSpec}
		for k, v := range base {
			args[k] = v
		}
		callOK(t, CreateTemplate, ctx, args)
		if *got == nil {
			t.Fatalf("backend was never called")
		}
		if !reflect.DeepEqual((*got)["device_spec"], deviceSpec) {
			t.Errorf("device_spec = %v, want %v", (*got)["device_spec"], deviceSpec)
		}
	})
	t.Run("without device", func(t *testing.T) {
		ctx, got := captureBody(t, http.MethodPost, "/v1/workspaces/ws/templates")
		callOK(t, CreateTemplate, ctx, base)
		if v, present := (*got)["device_spec"]; present {
			t.Errorf("device_spec = %v present in body, want absent", v)
		}
	})
	t.Run("missing platform is rejected before the API call", func(t *testing.T) {
		args := map[string]any{"device_spec": map[string]any{"device_model": "pixel_7"}}
		for k, v := range base {
			args[k] = v
		}
		if msg := callErr(t, CreateTemplate, args); !strings.Contains(msg, "device_spec.platform is required") {
			t.Errorf("error %q does not name device_spec.platform", msg)
		}
	})
}

// bitrise_devenv_update_template mirrors the array-field convention for the
// device: device_spec sets update_device_spec alongside the new object,
// clear_device_spec sends the flag alone, and omitting both leaves the wire
// free of either key.
func TestUpdateTemplateDeviceSpecFlag(t *testing.T) {
	path := "/v1/workspaces/ws/templates/" + testTemplateID
	deviceSpec := map[string]any{"platform": "ios", "device_model": "iPhone 16", "os_version": "18.2"}

	t.Run("set", func(t *testing.T) {
		ctx, got := captureBody(t, http.MethodPatch, path)
		callOK(t, UpdateTemplate, ctx, map[string]any{"template_id": testTemplateID, "device_spec": deviceSpec})
		if *got == nil {
			t.Fatalf("backend was never called")
		}
		if !reflect.DeepEqual((*got)["device_spec"], deviceSpec) {
			t.Errorf("device_spec = %v, want %v", (*got)["device_spec"], deviceSpec)
		}
		if (*got)["update_device_spec"] != true {
			t.Errorf("update_device_spec = %v, want true", (*got)["update_device_spec"])
		}
	})
	t.Run("clear", func(t *testing.T) {
		ctx, got := captureBody(t, http.MethodPatch, path)
		callOK(t, UpdateTemplate, ctx, map[string]any{"template_id": testTemplateID, "clear_device_spec": true, "name": "renamed"})
		if (*got)["update_device_spec"] != true {
			t.Errorf("update_device_spec = %v, want true", (*got)["update_device_spec"])
		}
		if v, present := (*got)["device_spec"]; present {
			t.Errorf("device_spec = %v present in body, want absent so the backend clears the device", v)
		}
		if (*got)["name"] != "renamed" {
			t.Errorf("name = %v, want %q", (*got)["name"], "renamed")
		}
	})
	t.Run("untouched", func(t *testing.T) {
		ctx, got := captureBody(t, http.MethodPatch, path)
		callOK(t, UpdateTemplate, ctx, map[string]any{"template_id": testTemplateID, "name": "renamed"})
		for _, key := range []string{"device_spec", "update_device_spec"} {
			if v, present := (*got)[key]; present {
				t.Errorf("%s = %v present in body, want absent", key, v)
			}
		}
	})
	t.Run("set and clear together are rejected", func(t *testing.T) {
		args := map[string]any{"template_id": testTemplateID, "device_spec": deviceSpec, "clear_device_spec": true}
		if msg := callErr(t, UpdateTemplate, args); !strings.Contains(msg, "clear_device_spec cannot be combined with device_spec") {
			t.Errorf("unexpected error %q", msg)
		}
	})
	t.Run("missing platform is rejected before the API call", func(t *testing.T) {
		args := map[string]any{"template_id": testTemplateID, "device_spec": map[string]any{"os_version": "18.2"}}
		if msg := callErr(t, UpdateTemplate, args); !strings.Contains(msg, "device_spec.platform is required") {
			t.Errorf("error %q does not name device_spec.platform", msg)
		}
	})
}
