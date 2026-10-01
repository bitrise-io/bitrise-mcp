package devenv

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/stretchr/testify/assert"
)

func TestWsPath(t *testing.T) {
	t.Run("builds workspace-scoped path from context", func(t *testing.T) {
		ctx := ContextWithWorkspace(context.Background(), "ws-123")
		assert.Equal(t, "/v1/workspaces/ws-123/sessions", WsPath(ctx, "/sessions"))
	})

	t.Run("escapes the workspace id", func(t *testing.T) {
		ctx := ContextWithWorkspace(context.Background(), "a b")
		assert.Equal(t, "/v1/workspaces/a%20b/x", WsPath(ctx, "/x"))
	})

	t.Run("empty when no workspace in context", func(t *testing.T) {
		assert.Equal(t, "/v1/workspaces//sessions", WsPath(context.Background(), "/sessions"))
	})
}

func TestWorkspaceFromCtx(t *testing.T) {
	assert.Equal(t, "", WorkspaceFromCtx(context.Background()))
	assert.Equal(t, "ws", WorkspaceFromCtx(ContextWithWorkspace(context.Background(), "ws")))
}

func TestHostedModeFromCtx(t *testing.T) {
	assert.False(t, HostedModeFromCtx(context.Background()))
	assert.True(t, HostedModeFromCtx(ContextWithHostedMode(context.Background())))
}

func TestAuthFromCtx(t *testing.T) {
	t.Run("errors when no PAT", func(t *testing.T) {
		_, _, err := AuthFromCtx(context.Background())
		assert.Error(t, err)
	})

	t.Run("returns Bearer header", func(t *testing.T) {
		h, v, err := AuthFromCtx(ContextWithPAT(context.Background(), "pat"))
		assert.NoError(t, err)
		assert.Equal(t, "Authorization", h)
		assert.Equal(t, "Bearer pat", v)
	})
}

func TestSoleWorkspace(t *testing.T) {
	t.Run("zero workspaces errors", func(t *testing.T) {
		_, err := SoleWorkspace(nil)
		assert.Error(t, err)
	})

	t.Run("exactly one is returned", func(t *testing.T) {
		ws, err := SoleWorkspace([]Organization{{Slug: "only", Name: "Only"}})
		assert.NoError(t, err)
		assert.Equal(t, "only", ws.Slug)
	})

	t.Run("multiple errors and lists them", func(t *testing.T) {
		_, err := SoleWorkspace([]Organization{{Slug: "a", Name: "Alpha"}, {Slug: "b", Name: "Beta"}})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ASK THE USER") // steer the model to ask, not enumerate
		assert.Contains(t, err.Error(), "workspace_id")
		assert.Contains(t, err.Error(), "Alpha")
		assert.Contains(t, err.Error(), "b")
	})
}

func TestListOrganizations(t *testing.T) {
	t.Run("parses the data array", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "pat", r.Header.Get("Authorization")) // main API: raw PAT, no "Bearer "
			assert.Equal(t, "/organizations", r.URL.Path)
			_, _ = w.Write([]byte(`{"data":[{"slug":"ws-1","name":"One"},{"slug":"ws-2","name":"Two"}]}`))
		}))
		defer srv.Close()

		old := bitrise.APIBaseURL
		bitrise.APIBaseURL = srv.URL
		defer func() { bitrise.APIBaseURL = old }()

		orgs, err := ListOrganizations(ContextWithPAT(context.Background(), "pat"))
		assert.NoError(t, err)
		assert.Len(t, orgs, 2)
		assert.Equal(t, "ws-1", orgs[0].Slug)
		assert.Equal(t, "Two", orgs[1].Name)
	})

	t.Run("requires authentication", func(t *testing.T) {
		_, err := ListOrganizations(context.Background())
		assert.Error(t, err)
	})

	t.Run("surfaces an HTTP error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("boom"))
		}))
		defer srv.Close()

		old := bitrise.APIBaseURL
		bitrise.APIBaseURL = srv.URL
		defer func() { bitrise.APIBaseURL = old }()

		_, err := ListOrganizations(ContextWithPAT(context.Background(), "pat"))
		assert.Error(t, err)
	})
}

func TestResolveSoleWorkspace(t *testing.T) {
	t.Run("caches the sole workspace per PAT", func(t *testing.T) {
		calls := 0
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			calls++
			_, _ = w.Write([]byte(`{"data":[{"slug":"only-ws","name":"Only"}]}`))
		}))
		defer srv.Close()

		old := bitrise.APIBaseURL
		bitrise.APIBaseURL = srv.URL
		defer func() { bitrise.APIBaseURL = old }()

		// Unique PAT keeps this subtest isolated from the package-global cache.
		ctx := ContextWithPAT(context.Background(), "pat-resolve-cache")
		s1, err := ResolveSoleWorkspace(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "only-ws", s1)

		s2, err := ResolveSoleWorkspace(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "only-ws", s2)
		assert.Equal(t, 1, calls, "second call should be served from the per-PAT cache")
	})

	t.Run("does not cache the multiple-workspace error", func(t *testing.T) {
		calls := 0
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			calls++
			_, _ = w.Write([]byte(`{"data":[{"slug":"a","name":"A"},{"slug":"b","name":"B"}]}`))
		}))
		defer srv.Close()

		old := bitrise.APIBaseURL
		bitrise.APIBaseURL = srv.URL
		defer func() { bitrise.APIBaseURL = old }()

		ctx := ContextWithPAT(context.Background(), "pat-resolve-multi")
		_, err1 := ResolveSoleWorkspace(ctx)
		_, err2 := ResolveSoleWorkspace(ctx)
		assert.Error(t, err1)
		assert.Error(t, err2)
		assert.Equal(t, 2, calls, "the multi-workspace error must not be cached")
	})

	t.Run("separate PATs are cached independently", func(t *testing.T) {
		var seen []string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			seen = append(seen, r.Header.Get("Authorization"))
			_, _ = w.Write([]byte(`{"data":[{"slug":"ws","name":"WS"}]}`))
		}))
		defer srv.Close()

		old := bitrise.APIBaseURL
		bitrise.APIBaseURL = srv.URL
		defer func() { bitrise.APIBaseURL = old }()

		_, _ = ResolveSoleWorkspace(ContextWithPAT(context.Background(), "pat-A"))
		_, _ = ResolveSoleWorkspace(ContextWithPAT(context.Background(), "pat-B"))
		assert.Len(t, seen, 2, "different PATs must not share a cache entry")
	})
}
