package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
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

func TestRequireAuthMiddleware(t *testing.T) {
	const metadataURL = "https://mcp.bitrise.io/.well-known/oauth-protected-resource"
	logger := zap.NewNop().Sugar()

	newNext := func() (http.Handler, *bool) {
		called := false
		h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})
		return h, &called
	}

	t.Run("credential-less POST is challenged with 401 + WWW-Authenticate", func(t *testing.T) {
		next, called := newNext()
		mw := requireAuthMiddleware(next, nil, metadataURL, logger)

		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))

		assert.False(t, *called, "next handler must not be reached without credentials")
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Equal(t,
			`Bearer resource_metadata="`+metadataURL+`"`,
			rec.Header().Get("WWW-Authenticate"),
		)
	})

	t.Run("POST with a valid PAT passes through", func(t *testing.T) {
		next, called := newNext()
		mw := requireAuthMiddleware(next, nil, metadataURL, logger)

		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Authorization", "Bearer my-bitrise-pat")
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, req)

		assert.True(t, *called)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("POST whose JWT exchange fails is treated as unauthenticated", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer srv.Close()

		next, called := newNext()
		exchanger := &jwtExchanger{tokenEndpoint: srv.URL}
		mw := requireAuthMiddleware(next, exchanger, metadataURL, logger)

		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Authorization", "Bearer "+makeTestJWT(time.Now().Add(10*time.Minute).Unix()))
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, req)

		assert.False(t, *called)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("non-POST requests are not challenged", func(t *testing.T) {
		next, called := newNext()
		mw := requireAuthMiddleware(next, nil, metadataURL, logger)

		mw.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
		assert.True(t, *called)
	})
}

