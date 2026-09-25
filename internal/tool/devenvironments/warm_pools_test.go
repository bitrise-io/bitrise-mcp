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

const testWarmPoolID = "7a6b5c4d-3e2f-4a1b-9c8d-7e6f5a4b3c2d"

// capturedRequest is what the fake backend saw: the method, path and query
// of the request plus its decoded JSON body (nil when the body was empty).
type capturedRequest struct {
	Method string
	Path   string
	Query  map[string][]string
	Body   map[string]any
	Called bool
}

// captureRequest points the API client at a fake backend that records the
// request and answers with response. It returns a context authenticated
// against workspace "ws" and the recorded request (filled once the backend
// is called).
func captureRequest(t *testing.T, response string) (context.Context, *capturedRequest) {
	t.Helper()
	got := &capturedRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.Called = true
		got.Method = r.Method
		got.Path = r.URL.Path
		got.Query = r.URL.Query()
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &got.Body); err != nil {
				t.Fatalf("unmarshal body %q: %v", raw, err)
			}
		}
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(srv.Close)
	old := devenv.BaseURL
	devenv.BaseURL = srv.URL
	t.Cleanup(func() { devenv.BaseURL = old })
	ctx := devenv.ContextWithWorkspace(devenv.ContextWithPAT(context.Background(), "pat"), "ws")
	return ctx, got
}

// callText runs the tool handler, fails the test unless it succeeds, and
// returns the text of the result.
func callText(t *testing.T, tool bitrise.Tool, ctx context.Context, args map[string]any) string {
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
	text, ok := res.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("content is %T, want TextContent", res.Content[0])
	}
	return text.Text
}

// assertRequest checks the method and path the backend saw.
func assertRequest(t *testing.T, got *capturedRequest, method, path string) {
	t.Helper()
	if !got.Called {
		t.Fatalf("backend was never called")
	}
	if got.Method != method || got.Path != path {
		t.Errorf("request %s %s, want %s %s", got.Method, got.Path, method, path)
	}
}

// The five warm pool tools are registered and classified workspace-scoped:
// every one of them hits /v1/workspaces/{id}/warm-pools.
func TestWarmPoolToolsRegistered(t *testing.T) {
	b := NewBelt()
	names := toolNames(b)
	for _, name := range []string{
		"bitrise_devenv_list_warm_pools",
		"bitrise_devenv_get_warm_pool",
		"bitrise_devenv_create_warm_pool",
		"bitrise_devenv_update_warm_pool",
		"bitrise_devenv_delete_warm_pool",
	} {
		if !names[name] {
			t.Errorf("%s is not registered in the belt", name)
		}
		if b.userScoped[name] {
			t.Errorf("%s is classified user-scoped, want workspace-scoped", name)
		}
	}
}

// The list is a GET whose only knobs are query parameters: template_id
// narrows to one template, all=true becomes the ALL scope by enum name, and
// neither is sent when not asked. The backend's JSON is passed through.
func TestListWarmPools(t *testing.T) {
	const path = "/v1/workspaces/ws/warm-pools"
	response := `{"warm_pools":[{"id":"` + testWarmPoolID + `","name":"ios lab","pool_size":2,"status":{"ready":1,"warming":1}}]}`

	t.Run("default lists visible pools", func(t *testing.T) {
		ctx, got := captureRequest(t, response)
		text := callText(t, ListWarmPools, ctx, map[string]any{})
		assertRequest(t, got, http.MethodGet, path)
		if len(got.Query) != 0 {
			t.Errorf("query = %v, want none", got.Query)
		}
		if text != response {
			t.Errorf("result = %q, want the backend response verbatim", text)
		}
		var decoded struct {
			WarmPools []struct {
				ID       string `json:"id"`
				PoolSize int    `json:"pool_size"`
				Status   struct {
					Ready int `json:"ready"`
				} `json:"status"`
			} `json:"warm_pools"`
		}
		if err := json.Unmarshal([]byte(text), &decoded); err != nil {
			t.Fatalf("result is not JSON: %v", err)
		}
		if len(decoded.WarmPools) != 1 || decoded.WarmPools[0].ID != testWarmPoolID || decoded.WarmPools[0].PoolSize != 2 || decoded.WarmPools[0].Status.Ready != 1 {
			t.Errorf("decoded %+v, want one pool %s with pool_size 2 and 1 ready", decoded.WarmPools, testWarmPoolID)
		}
	})
	t.Run("template_id and all become query parameters", func(t *testing.T) {
		ctx, got := captureRequest(t, response)
		callText(t, ListWarmPools, ctx, map[string]any{"template_id": testTemplateID, "all": true})
		assertRequest(t, got, http.MethodGet, path)
		want := map[string][]string{"template_id": {testTemplateID}, "scope": {"WARM_POOL_LIST_SCOPE_ALL"}}
		if !reflect.DeepEqual(got.Query, want) {
			t.Errorf("query = %v, want %v", got.Query, want)
		}
	})
	t.Run("all=false sends no scope", func(t *testing.T) {
		ctx, got := captureRequest(t, response)
		callText(t, ListWarmPools, ctx, map[string]any{"all": false})
		if _, present := got.Query["scope"]; present {
			t.Errorf("scope = %v present in query, want absent", got.Query["scope"])
		}
	})
	t.Run("malformed template_id is rejected before the API call", func(t *testing.T) {
		if msg := callErr(t, ListWarmPools, map[string]any{"template_id": "not-a-uuid"}); !strings.Contains(msg, "invalid template_id") {
			t.Errorf("unexpected error %q", msg)
		}
	})
}

// Get addresses the pool by id in the path and passes the pool — with its
// inventory — through unchanged.
func TestGetWarmPool(t *testing.T) {
	response := `{"warm_pool":{"id":"` + testWarmPoolID + `","status":{"ready":1,"sessions":[{"session_id":"s-1","state":"ready"}]}}}`
	ctx, got := captureRequest(t, response)
	text := callText(t, GetWarmPool, ctx, map[string]any{"warm_pool_id": testWarmPoolID})
	assertRequest(t, got, http.MethodGet, "/v1/workspaces/ws/warm-pools/"+testWarmPoolID)
	if got.Body != nil {
		t.Errorf("body = %v, want none on GET", got.Body)
	}
	var decoded struct {
		WarmPool struct {
			ID     string `json:"id"`
			Status struct {
				Sessions []struct {
					SessionID string `json:"session_id"`
					State     string `json:"state"`
				} `json:"sessions"`
			} `json:"status"`
		} `json:"warm_pool"`
	}
	if err := json.Unmarshal([]byte(text), &decoded); err != nil {
		t.Fatalf("result is not JSON: %v", err)
	}
	if decoded.WarmPool.ID != testWarmPoolID || len(decoded.WarmPool.Status.Sessions) != 1 || decoded.WarmPool.Status.Sessions[0].State != "ready" {
		t.Errorf("decoded %+v, want pool %s with one ready session", decoded.WarmPool, testWarmPoolID)
	}

	t.Run("missing id is rejected before the API call", func(t *testing.T) {
		if msg := callErr(t, GetWarmPool, map[string]any{}); !strings.Contains(msg, "warm_pool_id") {
			t.Errorf("unexpected error %q", msg)
		}
	})
	t.Run("malformed id is rejected before the API call", func(t *testing.T) {
		if msg := callErr(t, GetWarmPool, map[string]any{"warm_pool_id": "pool-1"}); !strings.Contains(msg, "invalid warm_pool_id") {
			t.Errorf("unexpected error %q", msg)
		}
	})
}

// Create posts the stored configuration: the scalars the backend cannot
// default are always sent, the optional ones only when given (an absent
// owner_type is the backend's default, an absent override is the template's
// value), and pool_size arrives as a JSON number.
func TestCreateWarmPool(t *testing.T) {
	const path = "/v1/workspaces/ws/warm-pools"
	response := `{"warm_pool":{"id":"` + testWarmPoolID + `","name":"ios lab","owner_type":"workspace","pool_size":3}}`

	t.Run("full configuration", func(t *testing.T) {
		ctx, got := captureRequest(t, response)
		inputs := []any{
			map[string]any{"key": "GITHUB_TOKEN", "value": "ghp_x", "is_secret": true},
			map[string]any{"key": "REPO", "value": "org/app"},
		}
		flags := []any{"install_deps", "start_agent"}
		text := callText(t, CreateWarmPool, ctx, map[string]any{
			"name":                       "ios lab",
			"template_id":                testTemplateID,
			"pool_size":                  float64(3),
			"owner_type":                 "workspace",
			"session_inputs":             inputs,
			"enabled_feature_flag_names": flags,
			"stack_id":                   "osx-xcode-26.0.x",
			"machine_type":               "g2.mac.m2pro.4c",
			"cluster":                    "mac-eu",
		})
		assertRequest(t, got, http.MethodPost, path)
		want := map[string]any{
			"name":                       "ios lab",
			"template_id":                testTemplateID,
			"pool_size":                  float64(3),
			"owner_type":                 "workspace",
			"session_inputs":             inputs,
			"enabled_feature_flag_names": flags,
			"stack_id":                   "osx-xcode-26.0.x",
			"machine_type":               "g2.mac.m2pro.4c",
			"cluster":                    "mac-eu",
		}
		if !reflect.DeepEqual(got.Body, want) {
			t.Errorf("body = %v, want %v", got.Body, want)
		}
		var decoded struct {
			WarmPool struct {
				ID        string `json:"id"`
				OwnerType string `json:"owner_type"`
				PoolSize  int    `json:"pool_size"`
			} `json:"warm_pool"`
		}
		if err := json.Unmarshal([]byte(text), &decoded); err != nil {
			t.Fatalf("result is not JSON: %v", err)
		}
		if decoded.WarmPool.ID != testWarmPoolID || decoded.WarmPool.OwnerType != "workspace" || decoded.WarmPool.PoolSize != 3 {
			t.Errorf("decoded %+v, want pool %s owned by workspace with pool_size 3", decoded.WarmPool, testWarmPoolID)
		}
	})
	t.Run("device fields are forwarded like a session create's", func(t *testing.T) {
		ctx, got := captureRequest(t, response)
		spec := map[string]any{"platform": "ios", "device_model": "iPhone 16"}
		callText(t, CreateWarmPool, ctx, map[string]any{"name": "ios lab", "template_id": testTemplateID, "pool_size": float64(1), "device_spec": spec})
		want := map[string]any{"name": "ios lab", "template_id": testTemplateID, "pool_size": float64(1), "device_spec": spec}
		if !reflect.DeepEqual(got.Body, want) {
			t.Errorf("body = %v, want %v", got.Body, want)
		}
		ctx, got = captureRequest(t, response)
		callText(t, CreateWarmPool, ctx, map[string]any{"name": "headless", "template_id": testTemplateID, "pool_size": float64(1), "no_device": true})
		if got.Body["no_device"] != true {
			t.Errorf("no_device = %v, want true", got.Body["no_device"])
		}
		if _, present := got.Body["device_spec"]; present {
			t.Errorf("device_spec present in body, want absent")
		}
	})
	t.Run("map_saved_to_session_inputs is forwarded only when true", func(t *testing.T) {
		ctx, got := captureRequest(t, response)
		callText(t, CreateWarmPool, ctx, map[string]any{"name": "mine", "template_id": testTemplateID, "pool_size": float64(1), "map_saved_to_session_inputs": true})
		want := map[string]any{"name": "mine", "template_id": testTemplateID, "pool_size": float64(1), "map_saved_to_session_inputs": true}
		if !reflect.DeepEqual(got.Body, want) {
			t.Errorf("body = %v, want %v", got.Body, want)
		}
		ctx, got = captureRequest(t, response)
		callText(t, CreateWarmPool, ctx, map[string]any{"name": "mine", "template_id": testTemplateID, "pool_size": float64(1), "map_saved_to_session_inputs": false})
		if _, present := got.Body["map_saved_to_session_inputs"]; present {
			t.Errorf("map_saved_to_session_inputs present in body, want absent")
		}
	})
	t.Run("minimal pool sends only the required fields", func(t *testing.T) {
		ctx, got := captureRequest(t, response)
		callText(t, CreateWarmPool, ctx, map[string]any{"name": "preset", "template_id": testTemplateID, "pool_size": float64(0)})
		assertRequest(t, got, http.MethodPost, path)
		want := map[string]any{"name": "preset", "template_id": testTemplateID, "pool_size": float64(0)}
		if !reflect.DeepEqual(got.Body, want) {
			t.Errorf("body = %v, want %v (pool_size 0 is a drained preset pool and must be sent)", got.Body, want)
		}
	})
	t.Run("rejected before the API call", func(t *testing.T) {
		cases := []struct {
			name string
			args map[string]any
			want string
		}{
			{"missing template_id", map[string]any{"name": "p", "pool_size": float64(1)}, "template_id"},
			{"malformed template_id", map[string]any{"name": "p", "template_id": "tpl", "pool_size": float64(1)}, "invalid template_id"},
			{"missing pool_size", map[string]any{"name": "p", "template_id": testTemplateID}, "pool_size is required"},
			{"negative pool_size", map[string]any{"name": "p", "template_id": testTemplateID, "pool_size": float64(-1)}, "pool_size must be >= 0"},
			{"pool_size not a number", map[string]any{"name": "p", "template_id": testTemplateID, "pool_size": "2"}, "pool_size must be a number"},
			{"device_spec with no_device", map[string]any{"name": "p", "template_id": testTemplateID, "pool_size": float64(1), "device_spec": map[string]any{"platform": "ios"}, "no_device": true}, "no_device cannot be combined"},
			{"device_spec not an object", map[string]any{"name": "p", "template_id": testTemplateID, "pool_size": float64(1), "device_spec": "ios"}, "device_spec must be an object"},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				if msg := callErr(t, CreateWarmPool, tc.args); !strings.Contains(msg, tc.want) {
					t.Errorf("error %q does not contain %q", msg, tc.want)
				}
			})
		}
	})
}

// Update is a PATCH carrying only what changes. Array fields switch on their
// update_* flag when given; override scalars are sent whenever present —
// including as "" to clear the override — and omitted otherwise.
func TestUpdateWarmPool(t *testing.T) {
	path := "/v1/workspaces/ws/warm-pools/" + testWarmPoolID
	response := `{"warm_pool":{"id":"` + testWarmPoolID + `","pool_size":0}}`

	t.Run("scale to zero sends pool_size alone", func(t *testing.T) {
		ctx, got := captureRequest(t, response)
		text := callText(t, UpdateWarmPool, ctx, map[string]any{"warm_pool_id": testWarmPoolID, "pool_size": float64(0)})
		assertRequest(t, got, http.MethodPatch, path)
		want := map[string]any{"pool_size": float64(0)}
		if !reflect.DeepEqual(got.Body, want) {
			t.Errorf("body = %v, want %v", got.Body, want)
		}
		if text != response {
			t.Errorf("result = %q, want the backend response verbatim", text)
		}
	})
	t.Run("arrays set their update switches", func(t *testing.T) {
		ctx, got := captureRequest(t, response)
		inputs := []any{map[string]any{"key": "REPO", "value": "org/other"}}
		callText(t, UpdateWarmPool, ctx, map[string]any{
			"warm_pool_id":               testWarmPoolID,
			"name":                       "renamed",
			"session_inputs":             inputs,
			"enabled_feature_flag_names": []any{},
		})
		want := map[string]any{
			"name":                              "renamed",
			"session_inputs":                    inputs,
			"update_session_inputs":             true,
			"enabled_feature_flag_names":        []any{},
			"update_enabled_feature_flag_names": true,
		}
		if !reflect.DeepEqual(got.Body, want) {
			t.Errorf("body = %v, want %v", got.Body, want)
		}
	})
	t.Run("overrides are sent when present, empty string clears", func(t *testing.T) {
		ctx, got := captureRequest(t, response)
		callText(t, UpdateWarmPool, ctx, map[string]any{
			"warm_pool_id": testWarmPoolID,
			"stack_id":     "osx-xcode-26.1.x",
			"machine_type": "",
		})
		want := map[string]any{"stack_id": "osx-xcode-26.1.x", "machine_type": ""}
		if !reflect.DeepEqual(got.Body, want) {
			t.Errorf("body = %v, want %v (cluster omitted, machine_type cleared)", got.Body, want)
		}
		for _, key := range []string{"update_session_inputs", "update_enabled_feature_flag_names", "session_inputs", "enabled_feature_flag_names"} {
			if v, present := got.Body[key]; present {
				t.Errorf("%s = %v present in body, want absent when the array was not given", key, v)
			}
		}
	})
	t.Run("device_spec switches update_device_spec on; {} clears the override", func(t *testing.T) {
		ctx, got := captureRequest(t, response)
		spec := map[string]any{"platform": "android", "system_image": "system-images;android-35;google_apis;x86_64"}
		callText(t, UpdateWarmPool, ctx, map[string]any{"warm_pool_id": testWarmPoolID, "device_spec": spec})
		want := map[string]any{"device_spec": spec, "update_device_spec": true}
		if !reflect.DeepEqual(got.Body, want) {
			t.Errorf("body = %v, want %v", got.Body, want)
		}
		ctx, got = captureRequest(t, response)
		callText(t, UpdateWarmPool, ctx, map[string]any{"warm_pool_id": testWarmPoolID, "device_spec": map[string]any{}})
		want = map[string]any{"update_device_spec": true}
		if !reflect.DeepEqual(got.Body, want) {
			t.Errorf("body = %v, want %v (the switch alone clears the override)", got.Body, want)
		}
	})
	t.Run("no_device is sent as given, false included", func(t *testing.T) {
		ctx, got := captureRequest(t, response)
		callText(t, UpdateWarmPool, ctx, map[string]any{"warm_pool_id": testWarmPoolID, "no_device": false})
		want := map[string]any{"no_device": false}
		if !reflect.DeepEqual(got.Body, want) {
			t.Errorf("body = %v, want %v", got.Body, want)
		}
	})
	t.Run("rejected before the API call", func(t *testing.T) {
		cases := []struct {
			name string
			args map[string]any
			want string
		}{
			{"missing id", map[string]any{"name": "x"}, "warm_pool_id"},
			{"device_spec with no_device true", map[string]any{"warm_pool_id": testWarmPoolID, "device_spec": map[string]any{"platform": "ios"}, "no_device": true}, "no_device cannot be true together"},
			{"device_spec not an object", map[string]any{"warm_pool_id": testWarmPoolID, "device_spec": "ios"}, "device_spec must be an object"},
			{"malformed id", map[string]any{"warm_pool_id": "pool", "name": "x"}, "invalid warm_pool_id"},
			{"nothing to update", map[string]any{"warm_pool_id": testWarmPoolID}, "nothing to update"},
			{"negative pool_size", map[string]any{"warm_pool_id": testWarmPoolID, "pool_size": float64(-2)}, "pool_size must be >= 0"},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				if msg := callErr(t, UpdateWarmPool, tc.args); !strings.Contains(msg, tc.want) {
					t.Errorf("error %q does not contain %q", msg, tc.want)
				}
			})
		}
	})
}

// Delete addresses the pool in the path with no body, and turns the
// backend's empty response into a sentence that says what happened.
func TestDeleteWarmPool(t *testing.T) {
	path := "/v1/workspaces/ws/warm-pools/" + testWarmPoolID

	t.Run("empty response is explained", func(t *testing.T) {
		ctx, got := captureRequest(t, `{}`)
		text := callText(t, DeleteWarmPool, ctx, map[string]any{"warm_pool_id": testWarmPoolID})
		assertRequest(t, got, http.MethodDelete, path)
		if got.Body != nil {
			t.Errorf("body = %v, want none on DELETE", got.Body)
		}
		if !strings.Contains(text, testWarmPoolID) || !strings.Contains(text, "deleted") {
			t.Errorf("result %q should name the deleted pool", text)
		}
	})
	t.Run("malformed id is rejected before the API call", func(t *testing.T) {
		if msg := callErr(t, DeleteWarmPool, map[string]any{"warm_pool_id": "pool"}); !strings.Contains(msg, "invalid warm_pool_id") {
			t.Errorf("unexpected error %q", msg)
		}
	})
}

// A claim: bitrise_devenv_create with warm_pool_id sends the pool id and the
// per-session fields only, needs neither template nor stack/machine type, and
// never injects a configuration field the backend would reject.
func TestCreateSessionClaimsFromWarmPool(t *testing.T) {
	const path = "/v1/workspaces/ws/sessions"
	response := `{"session":{"id":"sess-1","warm_pool_id":"` + testWarmPoolID + `","warm_state":"claimed"}}`

	t.Run("per-session fields pass through", func(t *testing.T) {
		ctx, got := captureRequest(t, response)
		labels := map[string]any{"team": "mobile"}
		text := callText(t, CreateSession, ctx, map[string]any{
			"name":                   "task 42",
			"description":            "claimed",
			"warm_pool_id":           testWarmPoolID,
			"labels":                 labels,
			"auto_terminate_minutes": float64(120),
			"owner":                  "workspace",
		})
		assertRequest(t, got, http.MethodPost, path)
		want := map[string]any{
			"name":                   "task 42",
			"description":            "claimed",
			"warm_pool_id":           testWarmPoolID,
			"labels":                 labels,
			"auto_terminate_minutes": float64(120),
			"owner_type":             "workspace",
		}
		if !reflect.DeepEqual(got.Body, want) {
			t.Errorf("body = %v, want %v", got.Body, want)
		}
		var decoded struct {
			Session struct {
				WarmState string `json:"warm_state"`
			} `json:"session"`
		}
		if err := json.Unmarshal([]byte(text), &decoded); err != nil {
			t.Fatalf("result is not JSON: %v", err)
		}
		if decoded.Session.WarmState != "claimed" {
			t.Errorf("warm_state = %q, want claimed", decoded.Session.WarmState)
		}
	})
	t.Run("name alone is enough", func(t *testing.T) {
		ctx, got := captureRequest(t, response)
		callText(t, CreateSession, ctx, map[string]any{"name": "quick", "warm_pool_id": testWarmPoolID})
		want := map[string]any{"name": "quick", "warm_pool_id": testWarmPoolID}
		if !reflect.DeepEqual(got.Body, want) {
			t.Errorf("body = %v, want %v (no defaults injected)", got.Body, want)
		}
	})
	t.Run("artifact rides the pool's device", func(t *testing.T) {
		ctx, got := captureRequest(t, response)
		artifact := map[string]any{"url": "https://example.com/app.apk"}
		callText(t, CreateSession, ctx, map[string]any{"name": "with build", "warm_pool_id": testWarmPoolID, "artifact": artifact})
		if !reflect.DeepEqual(got.Body["artifact"], artifact) {
			t.Errorf("artifact = %v, want %v", got.Body["artifact"], artifact)
		}
		if v, present := got.Body["device_spec"]; present {
			t.Errorf("device_spec = %v present in body, want absent", v)
		}
	})
	t.Run("blank configuration values are not conflicts", func(t *testing.T) {
		ctx, got := captureRequest(t, response)
		callText(t, CreateSession, ctx, map[string]any{
			"name":                       "tolerant",
			"warm_pool_id":               testWarmPoolID,
			"template_id":                "",
			"stack_id":                   "",
			"enabled_feature_flag_names": []any{},
			"no_device":                  false,
		})
		want := map[string]any{"name": "tolerant", "warm_pool_id": testWarmPoolID}
		if !reflect.DeepEqual(got.Body, want) {
			t.Errorf("body = %v, want %v", got.Body, want)
		}
	})
	t.Run("configuration fields are rejected before the API call", func(t *testing.T) {
		cases := []struct {
			name string
			args map[string]any
			want string
		}{
			{"template_id", map[string]any{"template_id": testTemplateID}, "template_id"},
			{"stack and machine", map[string]any{"stack_id": "s", "machine_type": "m"}, "stack_id, machine_type"},
			{"session_inputs", map[string]any{"session_inputs": []any{map[string]any{"key": "k", "value": "v"}}}, "session_inputs"},
			{"map_saved_to_session_inputs", map[string]any{"map_saved_to_session_inputs": true}, "map_saved_to_session_inputs"},
			{"feature flags", map[string]any{"enabled_feature_flag_names": []any{"x"}}, "enabled_feature_flag_names"},
			{"cluster", map[string]any{"cluster": "c"}, "cluster"},
			{"device_spec", map[string]any{"device_spec": map[string]any{"platform": "ios"}}, "device_spec"},
			{"no_device", map[string]any{"no_device": true}, "no_device"},
			{"ai_prompt", map[string]any{"ai_prompt": "do things"}, "ai_prompt"},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				args := map[string]any{"name": "x", "warm_pool_id": testWarmPoolID}
				for k, v := range tc.args {
					args[k] = v
				}
				msg := callErr(t, CreateSession, args)
				if !strings.Contains(msg, "warm_pool_id cannot be combined with") || !strings.Contains(msg, tc.want) {
					t.Errorf("error %q should name warm_pool_id and %q", msg, tc.want)
				}
			})
		}
	})
	t.Run("malformed warm_pool_id is rejected before the API call", func(t *testing.T) {
		if msg := callErr(t, CreateSession, map[string]any{"name": "x", "warm_pool_id": "pool"}); !strings.Contains(msg, "invalid warm_pool_id") {
			t.Errorf("unexpected error %q", msg)
		}
	})
}

// A preview link served from a warm pool: warm_pool_id goes on the wire, the
// device_spec requirement is lifted (the pool names the device), and the
// machine knobs the backend rejects alongside a pool are caught first.
func TestCreatePreviewLinkFromWarmPool(t *testing.T) {
	const path = "/v1/workspaces/ws/preview-links"
	response := `{"url":"https://app.bitrise.io/dev-environments/ws/device-preview/tok","jti":"link-1"}`
	artifact := map[string]any{"url": "https://example.com/app.apk", "app_name": "Demo"}

	t.Run("pool and artifact alone", func(t *testing.T) {
		ctx, got := captureRequest(t, response)
		text := callText(t, CreatePreviewLink, ctx, map[string]any{"warm_pool_id": testWarmPoolID, "artifact": artifact, "ttl_seconds": float64(3600)})
		assertRequest(t, got, http.MethodPost, path)
		want := map[string]any{"warm_pool_id": testWarmPoolID, "artifact": artifact, "ttl_seconds": float64(3600)}
		if !reflect.DeepEqual(got.Body, want) {
			t.Errorf("body = %v, want %v (no device_spec, stack_id or machine_type)", got.Body, want)
		}
		var decoded struct {
			URL string `json:"url"`
			JTI string `json:"jti"`
		}
		if err := json.Unmarshal([]byte(text), &decoded); err != nil {
			t.Fatalf("result is not JSON: %v", err)
		}
		if decoded.JTI != "link-1" || decoded.URL == "" {
			t.Errorf("decoded %+v, want jti link-1 and a url", decoded)
		}
	})
	t.Run("matching device_spec may accompany the pool", func(t *testing.T) {
		ctx, got := captureRequest(t, response)
		deviceSpec := map[string]any{"platform": "android"}
		callText(t, CreatePreviewLink, ctx, map[string]any{"warm_pool_id": testWarmPoolID, "device_spec": deviceSpec, "artifact": artifact})
		if !reflect.DeepEqual(got.Body["device_spec"], deviceSpec) {
			t.Errorf("device_spec = %v, want %v", got.Body["device_spec"], deviceSpec)
		}
		if got.Body["warm_pool_id"] != testWarmPoolID {
			t.Errorf("warm_pool_id = %v, want %s", got.Body["warm_pool_id"], testWarmPoolID)
		}
	})
	t.Run("without a pool device_spec stays required", func(t *testing.T) {
		if msg := callErr(t, CreatePreviewLink, map[string]any{"artifact": artifact}); !strings.Contains(msg, "device_spec is required") {
			t.Errorf("unexpected error %q", msg)
		}
	})
	t.Run("artifact stays required with a pool", func(t *testing.T) {
		if msg := callErr(t, CreatePreviewLink, map[string]any{"warm_pool_id": testWarmPoolID}); !strings.Contains(msg, "artifact is required") {
			t.Errorf("unexpected error %q", msg)
		}
	})
	t.Run("machine knobs are rejected alongside a pool", func(t *testing.T) {
		for _, key := range []string{"stack_id", "machine_type"} {
			args := map[string]any{"warm_pool_id": testWarmPoolID, "artifact": artifact, key: "x"}
			if msg := callErr(t, CreatePreviewLink, args); !strings.Contains(msg, "warm_pool_id cannot be combined with stack_id or machine_type") {
				t.Errorf("%s: unexpected error %q", key, msg)
			}
		}
	})
	t.Run("malformed warm_pool_id is rejected before the API call", func(t *testing.T) {
		if msg := callErr(t, CreatePreviewLink, map[string]any{"warm_pool_id": "pool", "artifact": artifact}); !strings.Contains(msg, "invalid warm_pool_id") {
			t.Errorf("unexpected error %q", msg)
		}
	})
}
