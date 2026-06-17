package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
)

func TestExtractPAT(t *testing.T) {
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

		exchanger := &jwtExchanger{tokenEndpoint: srv.URL}
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

		exchanger := &jwtExchanger{tokenEndpoint: srv.URL}
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

func TestWithAuthContext(t *testing.T) {
	const metadataURL = "https://mcp.bitrise.io/.well-known/oauth-protected-resource"

	captureCtx := func() (http.Handler, *context.Context) {
		var captured context.Context
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			captured = r.Context()
			w.WriteHeader(http.StatusOK)
		})
		return h, &captured
	}

	t.Run("wraps writer as hijackable", func(t *testing.T) {
		next, ctx := captureCtx()
		mw := withAuthContext(next, metadataURL)

		mw.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil))

		hw, ok := (*ctx).Value(ctxWriterKey{}).(*hijackableWriter)
		assert.True(t, ok)
		assert.NotNil(t, hw)
	})

	t.Run("stores metadata URL in context verbatim", func(t *testing.T) {
		next, ctx := captureCtx()
		mw := withAuthContext(next, metadataURL)

		mw.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil))

		stored, _ := (*ctx).Value(ctxMetadataURLKey{}).(string)
		assert.Equal(t, metadataURL, stored)
	})

	t.Run("passes all requests through to next handler", func(t *testing.T) {
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})
		mw := withAuthContext(next, metadataURL)

		mw.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
		assert.True(t, called)
	})
}

func TestRequireAuthToolHandler(t *testing.T) {
	makeCtx := func(pat string) (context.Context, *httptest.ResponseRecorder) {
		rec := httptest.NewRecorder()
		hw := &hijackableWriter{ResponseWriter: rec}
		ctx := context.WithValue(context.Background(), ctxWriterKey{}, hw)
		ctx = context.WithValue(ctx, ctxMetadataURLKey{}, "http://mcp.example.com/.well-known/oauth-protected-resource")
		if pat != "" {
			ctx = bitrise.ContextWithPAT(ctx, pat)
		}
		return ctx, rec
	}

	passThroughFn := server.ToolHandlerFunc(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return &mcp.CallToolResult{}, nil
	})

	t.Run("unauthenticated call writes 401 and does not invoke fn", func(t *testing.T) {
		ctx, rec := makeCtx("")
		fnCalled := false
		fn := server.ToolHandlerFunc(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			fnCalled = true
			return &mcp.CallToolResult{}, nil
		})

		result, err := requireAuthToolHandler(fn)(ctx, mcp.CallToolRequest{})

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.False(t, fnCalled)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Equal(t,
			`Bearer resource_metadata="http://mcp.example.com/.well-known/oauth-protected-resource"`,
			rec.Header().Get("WWW-Authenticate"),
		)
		hw := ctx.Value(ctxWriterKey{}).(*hijackableWriter)
		assert.True(t, hw.done, "writer must be hijacked so mcp-go's error write is discarded")
	})

	t.Run("authenticated call passes through to fn", func(t *testing.T) {
		ctx, _ := makeCtx("valid-pat")

		result, err := requireAuthToolHandler(passThroughFn)(ctx, mcp.CallToolRequest{})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		hw := ctx.Value(ctxWriterKey{}).(*hijackableWriter)
		assert.False(t, hw.done)
	})

	t.Run("JWT exchange to valid PAT passes through", func(t *testing.T) {
		ctx, _ := makeCtx("exchanged-pat")

		result, err := requireAuthToolHandler(passThroughFn)(ctx, mcp.CallToolRequest{})

		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

