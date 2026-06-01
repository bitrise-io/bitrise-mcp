package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// makeTestJWT builds a minimal unsigned JWT with the given exp claim.
func makeTestJWT(expUnix int64) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	claims, _ := json.Marshal(map[string]any{"sub": "user123", "exp": expUnix})
	payload := base64.RawURLEncoding.EncodeToString(claims)
	return header + "." + payload + ".fakesignature"
}

func TestIsJWT(t *testing.T) {
	cases := map[string]struct {
		token string
		want  bool
	}{
		"valid JWT":                 {token: "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJ1c2VyIn0.sig", want: true},
		"plain PAT":                 {token: "some-bitrise-personal-access-token", want: false},
		"empty string":              {token: "", want: false},
		"only two dot-separated":    {token: "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJ1c2VyIn0", want: false},
		"four parts":                {token: "eyJa.eyJb.sig.extra", want: false},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, isJWT(tc.token))
		})
	}
}

func TestJWTTTL(t *testing.T) {
	t.Run("future expiry within 1 hour", func(t *testing.T) {
		exp := time.Now().Add(30 * time.Minute).Unix()
		ttl := jwtTTL(makeTestJWT(exp))
		assert.InDelta(t, (30 * time.Minute).Seconds(), ttl.Seconds(), 5)
	})

	t.Run("expiry beyond 1 hour is capped at 1 hour", func(t *testing.T) {
		exp := time.Now().Add(3 * time.Hour).Unix()
		ttl := jwtTTL(makeTestJWT(exp))
		assert.Equal(t, time.Hour, ttl)
	})

	t.Run("already expired token returns 0", func(t *testing.T) {
		exp := time.Now().Add(-5 * time.Minute).Unix()
		ttl := jwtTTL(makeTestJWT(exp))
		assert.Equal(t, time.Duration(0), ttl)
	})

	t.Run("non-JWT falls back to 5 minutes", func(t *testing.T) {
		ttl := jwtTTL("not.a.jwt")
		assert.Equal(t, 5*time.Minute, ttl)
	})

	t.Run("missing exp claim falls back to 5 minutes", func(t *testing.T) {
		header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256"}`))
		payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"user"}`))
		token := header + "." + payload + ".sig"
		ttl := jwtTTL(token)
		assert.Equal(t, 5*time.Minute, ttl)
	})
}

func TestCacheKey(t *testing.T) {
	k1 := cacheKey("token-a")
	k2 := cacheKey("token-a")
	k3 := cacheKey("token-b")
	assert.Equal(t, k1, k2, "same token should produce same key")
	assert.NotEqual(t, k1, k3, "different tokens should produce different keys")
	assert.Len(t, k1, 16, "key should be 8 bytes as hex")
}

func TestJwtExchangerExchange(t *testing.T) {
	nopLogger := zap.NewNop().Sugar()

	t.Run("successful exchange returns PAT", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
			assert.NoError(t, r.ParseForm())
			assert.Equal(t, "urn:ietf:params:oauth:grant-type:token-exchange", r.FormValue("grant_type"))
			body, _ := json.Marshal(map[string]string{"access_token": "my-bitrise-pat"})
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(body)
		}))
		defer srv.Close()

		exchanger := &jwtExchanger{tokenEndpoint: srv.URL, logger: nopLogger}
		jwt := makeTestJWT(time.Now().Add(10 * time.Minute).Unix())
		pat, err := exchanger.exchange(t.Context(), jwt)
		assert.NoError(t, err)
		assert.Equal(t, "my-bitrise-pat", pat)
	})

	t.Run("result is cached on second call", func(t *testing.T) {
		calls := 0
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			body, _ := json.Marshal(map[string]string{"access_token": fmt.Sprintf("pat-call-%d", calls)})
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(body)
		}))
		defer srv.Close()

		exchanger := &jwtExchanger{tokenEndpoint: srv.URL, logger: nopLogger}
		jwt := makeTestJWT(time.Now().Add(10 * time.Minute).Unix())

		pat1, err := exchanger.exchange(t.Context(), jwt)
		assert.NoError(t, err)
		pat2, err := exchanger.exchange(t.Context(), jwt)
		assert.NoError(t, err)

		assert.Equal(t, pat1, pat2, "second call should return cached value")
		assert.Equal(t, 1, calls, "endpoint should only be called once")
	})

	t.Run("non-200 response returns error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"invalid_token"}`))
		}))
		defer srv.Close()

		exchanger := &jwtExchanger{tokenEndpoint: srv.URL, logger: nopLogger}
		jwt := makeTestJWT(time.Now().Add(10 * time.Minute).Unix())
		_, err := exchanger.exchange(t.Context(), jwt)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "401")
	})

	t.Run("response missing access_token returns error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := json.Marshal(map[string]string{"token_type": "Bearer"})
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(body)
		}))
		defer srv.Close()

		exchanger := &jwtExchanger{tokenEndpoint: srv.URL, logger: nopLogger}
		jwt := makeTestJWT(time.Now().Add(10 * time.Minute).Unix())
		_, err := exchanger.exchange(t.Context(), jwt)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access_token")
	})
}

func TestOauthProtectedResourceHandler(t *testing.T) {
	handler := oauthProtectedResourceHandler("https://auth.example.com/")

	t.Run("returns correct metadata with https forwarded proto", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/.well-known/oauth-protected-resource", nil)
		req.Host = "mcp.example.com"
		req.Header.Set("X-Forwarded-Proto", "https")
		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var body map[string]any
		if assert.NoError(t, json.NewDecoder(w.Body).Decode(&body)) {
			assert.Equal(t, "https://mcp.example.com", body["resource"])
			authServers, ok := body["authorization_servers"].([]any)
			if assert.True(t, ok) {
				assert.Equal(t, "https://auth.example.com", authServers[0], "trailing slash on issuer should be stripped")
			}
		}
	})

	t.Run("uses http when no TLS and no forwarded proto", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/.well-known/oauth-protected-resource", nil)
		req.Host = "localhost:8080"
		w := httptest.NewRecorder()
		handler(w, req)

		var body map[string]any
		if assert.NoError(t, json.NewDecoder(w.Body).Decode(&body)) {
			assert.Equal(t, "http://localhost:8080", body["resource"])
		}
	})
}
