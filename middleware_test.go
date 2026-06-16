package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMethodRequiresAuth(t *testing.T) {
	cases := map[string]struct {
		method mcp.MCPMethod
		want   bool
	}{
		"tools/call requires auth":      {method: mcp.MethodToolsCall, want: true},
		"initialize is open":            {method: mcp.MethodInitialize, want: false},
		"tools/list is open":            {method: mcp.MethodToolsList, want: false},
		"ping is open":                  {method: mcp.MethodPing, want: false},
		"empty/unparseable method open": {method: "", want: false},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, methodRequiresAuth(tc.method))
		})
	}
}

func TestExtractPAT(t *testing.T) {
	nopLogger := zap.NewNop().Sugar()

	t.Run("no Authorization header returns empty PAT and no error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		pat, err := extractPAT(req, nil)
		assert.NoError(t, err)
		assert.Empty(t, pat)
	})

	t.Run("raw PAT bearer token is returned verbatim", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Authorization", "Bearer my-bitrise-pat")
		pat, err := extractPAT(req, nil)
		assert.NoError(t, err)
		assert.Equal(t, "my-bitrise-pat", pat)
	})

	t.Run("JWT bearer token is exchanged for a PAT", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			body, _ := json.Marshal(map[string]string{"access_token": "exchanged-pat"})
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(body)
		}))
		defer srv.Close()

		exchanger := &jwtExchanger{tokenEndpoint: srv.URL, logger: nopLogger}
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Authorization", "Bearer "+makeTestJWT(time.Now().Add(10*time.Minute).Unix()))
		pat, err := extractPAT(req, exchanger)
		assert.NoError(t, err)
		assert.Equal(t, "exchanged-pat", pat)
	})

	t.Run("failed JWT exchange surfaces an error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer srv.Close()

		exchanger := &jwtExchanger{tokenEndpoint: srv.URL, logger: nopLogger}
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Authorization", "Bearer "+makeTestJWT(time.Now().Add(10*time.Minute).Unix()))
		_, err := extractPAT(req, exchanger)
		assert.Error(t, err)
	})

	t.Run("JWT without configured exchanger is passed through as-is", func(t *testing.T) {
		jwt := makeTestJWT(time.Now().Add(10 * time.Minute).Unix())
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Authorization", "Bearer "+jwt)
		pat, err := extractPAT(req, nil)
		assert.NoError(t, err)
		assert.Equal(t, jwt, pat)
	})
}

// jsonRPCBody builds a minimal JSON-RPC request body for the given method.
func jsonRPCBody(method mcp.MCPMethod) string {
	return fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":%q,"params":{}}`, string(method))
}

// newMCPRequest builds an MCP POST request for the given method with an optional
// bearer token.
func newMCPRequest(method mcp.MCPMethod, bearer string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonRPCBody(method)))
	req.Header.Set("Content-Type", "application/json")
	req.Host = "mcp.example.com"
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	return req
}

func TestRequireAuthMiddleware(t *testing.T) {
	nopLogger := zap.NewNop().Sugar()

	// spyHandler records whether it was reached and what body it observed.
	type spy struct {
		called bool
		body   string
	}
	newSpy := func(s *spy) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s.called = true
			b, _ := io.ReadAll(r.Body)
			s.body = string(b)
			w.WriteHeader(http.StatusOK)
		})
	}

	t.Run("unauthenticated tools/call returns 401 with WWW-Authenticate", func(t *testing.T) {
		s := &spy{}
		mw := requireAuthMiddleware(newSpy(s), nil, nopLogger)

		req := newMCPRequest(mcp.MethodToolsCall, "")
		w := httptest.NewRecorder()
		mw.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Equal(t,
			`Bearer resource_metadata="http://mcp.example.com/.well-known/oauth-protected-resource"`,
			w.Header().Get("WWW-Authenticate"),
		)
		assert.False(t, s.called, "downstream handler must not run for unauthenticated tool calls")
	})

	t.Run("WWW-Authenticate uses https behind a TLS-terminating proxy", func(t *testing.T) {
		s := &spy{}
		mw := requireAuthMiddleware(newSpy(s), nil, nopLogger)

		req := newMCPRequest(mcp.MethodToolsCall, "")
		req.Header.Set("X-Forwarded-Proto", "https")
		w := httptest.NewRecorder()
		mw.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Equal(t,
			`Bearer resource_metadata="https://mcp.example.com/.well-known/oauth-protected-resource"`,
			w.Header().Get("WWW-Authenticate"),
		)
	})

	t.Run("tools/call with a valid PAT reaches the handler", func(t *testing.T) {
		s := &spy{}
		mw := requireAuthMiddleware(newSpy(s), nil, nopLogger)

		req := newMCPRequest(mcp.MethodToolsCall, "valid-pat")
		w := httptest.NewRecorder()
		mw.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, s.called, "downstream handler should run for authenticated tool calls")
		assert.Equal(t, jsonRPCBody(mcp.MethodToolsCall), s.body, "request body must be preserved for the handler")
	})

	t.Run("tools/call with a valid JWT is exchanged then reaches the handler", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			body, _ := json.Marshal(map[string]string{"access_token": "exchanged-pat"})
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(body)
		}))
		defer srv.Close()

		s := &spy{}
		exchanger := &jwtExchanger{tokenEndpoint: srv.URL, logger: nopLogger}
		mw := requireAuthMiddleware(newSpy(s), exchanger, nopLogger)

		req := newMCPRequest(mcp.MethodToolsCall, makeTestJWT(time.Now().Add(10*time.Minute).Unix()))
		w := httptest.NewRecorder()
		mw.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, s.called)
	})

	t.Run("tools/call with an invalid JWT returns 401", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer srv.Close()

		s := &spy{}
		exchanger := &jwtExchanger{tokenEndpoint: srv.URL, logger: nopLogger}
		mw := requireAuthMiddleware(newSpy(s), exchanger, nopLogger)

		req := newMCPRequest(mcp.MethodToolsCall, makeTestJWT(time.Now().Add(10*time.Minute).Unix()))
		w := httptest.NewRecorder()
		mw.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.NotEmpty(t, w.Header().Get("WWW-Authenticate"))
		assert.False(t, s.called)
	})

	t.Run("initialize is allowed without auth", func(t *testing.T) {
		s := &spy{}
		mw := requireAuthMiddleware(newSpy(s), nil, nopLogger)

		req := newMCPRequest(mcp.MethodInitialize, "")
		w := httptest.NewRecorder()
		mw.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, s.called)
		assert.Empty(t, w.Header().Get("WWW-Authenticate"))
	})

	t.Run("tools/list is allowed without auth", func(t *testing.T) {
		s := &spy{}
		mw := requireAuthMiddleware(newSpy(s), nil, nopLogger)

		req := newMCPRequest(mcp.MethodToolsList, "")
		w := httptest.NewRecorder()
		mw.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, s.called)
	})

	t.Run("non-POST requests pass through untouched", func(t *testing.T) {
		s := &spy{}
		mw := requireAuthMiddleware(newSpy(s), nil, nopLogger)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		mw.ServeHTTP(w, req)

		assert.True(t, s.called)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
