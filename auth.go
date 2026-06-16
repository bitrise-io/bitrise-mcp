package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// protectedResourceMetadataPath is the well-known location (RFC 9728) where the
// server advertises which authorization server issues tokens for this resource.
const protectedResourceMetadataPath = "/.well-known/oauth-protected-resource"

// serverBaseURL reconstructs this server's externally visible base URL from the
// incoming request, honouring the X-Forwarded-Proto header set by TLS-terminating
// proxies.
func serverBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

// resourceMetadataURL returns the absolute URL of this server's RFC 9728
// protected resource metadata document, derived from the incoming request so it
// stays correct regardless of the host the server is reached on.
func resourceMetadataURL(r *http.Request) string {
	return serverBaseURL(r) + protectedResourceMetadataPath
}

// oauthProtectedResourceHandler serves RFC 9728 Protected Resource Metadata,
// telling OAuth clients which authorization server issues tokens for this resource.
func oauthProtectedResourceHandler(issuer string) http.HandlerFunc {
	issuer = strings.TrimRight(issuer, "/")
	return func(w http.ResponseWriter, r *http.Request) {
		metadata := map[string]any{
			"resource":                 serverBaseURL(r),
			"authorization_servers":    []string{issuer},
			"bearer_methods_supported": []string{"header"},
		}
		body, _ := json.Marshal(metadata)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}
}

// extractPAT resolves the Authorization bearer token on r into a Bitrise PAT,
// exchanging it via the OIDC token endpoint (RFC 8693) when it looks like a JWT
// and an exchanger is configured. It returns an empty PAT and a nil error when
// no bearer token is present.
func extractPAT(r *http.Request, exchanger *jwtExchanger) (string, error) {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token == "" {
		return "", nil
	}
	if exchanger != nil && isJWT(token) {
		return exchanger.exchange(r.Context(), token)
	}
	return token, nil
}

type cacheEntry struct {
	pat       string
	expiresAt time.Time
}

// jwtExchanger calls an OIDC token exchange endpoint (RFC 8693) to trade an
// external JWT for a Bitrise PAT, caching results until the JWT expires.
type jwtExchanger struct {
	tokenEndpoint string
	logger        *zap.SugaredLogger
	cache         sync.Map
}

func (e *jwtExchanger) exchange(ctx context.Context, jwt string) (string, error) {
	key := cacheKey(jwt)
	if v, ok := e.cache.Load(key); ok {
		entry := v.(cacheEntry) //nolint:forcetypeassert
		if time.Now().Before(entry.expiresAt) {
			return entry.pat, nil
		}
		e.cache.Delete(key)
	}

	pat, err := e.callExchangeEndpoint(ctx, jwt)
	if err != nil {
		return "", err
	}

	e.cache.Store(key, cacheEntry{
		pat:       pat,
		expiresAt: time.Now().Add(jwtTTL(jwt)),
	})
	return pat, nil
}

func (e *jwtExchanger) callExchangeEndpoint(ctx context.Context, jwt string) (string, error) {
	body := url.Values{
		"grant_type":         {"urn:ietf:params:oauth:grant-type:token-exchange"},
		"subject_token":      {jwt},
		"subject_token_type": {"urn:ietf:params:oauth:token-type:access_token"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.tokenEndpoint, strings.NewReader(body.Encode()))
	if err != nil {
		return "", fmt.Errorf("create exchange request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("exchange request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read exchange response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("exchange returned %d: %s", resp.StatusCode, respBody)
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse exchange response: %w", err)
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("exchange response missing access_token")
	}
	return result.AccessToken, nil
}

// isJWT returns true when the token looks like a JWT (three base64url parts
// with an "eyJ" header prefix).
func isJWT(token string) bool {
	return strings.HasPrefix(token, "eyJ") && strings.Count(token, ".") == 2
}

// jwtTTL decodes the exp claim from a JWT (without verification) and returns
// the remaining lifetime, capped at 1 hour. Falls back to 5 minutes on any error.
func jwtTTL(jwt string) time.Duration {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return 5 * time.Minute
	}
	payload := parts[1]
	if p := len(payload) % 4; p != 0 {
		payload += strings.Repeat("=", 4-p)
	}
	data, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return 5 * time.Minute
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(data, &claims); err != nil || claims.Exp == 0 {
		return 5 * time.Minute
	}
	ttl := time.Until(time.Unix(claims.Exp, 0))
	if ttl <= 0 {
		return 0
	}
	if ttl > time.Hour {
		return time.Hour
	}
	return ttl
}

// cacheKey returns a short stable identifier for a JWT without storing the
// full token value.
func cacheKey(jwt string) string {
	h := sha256.Sum256([]byte(jwt))
	return fmt.Sprintf("%x", h[:8])
}
