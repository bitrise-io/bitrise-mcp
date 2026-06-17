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
)

// extractPAT exchanges via OIDC (RFC 8693) when the token looks like a JWT;
// otherwise passes it through as a raw PAT. Returns ("", nil) with no auth header.
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

// jwtExchanger trades an external JWT for a Bitrise PAT via RFC 8693, caching by JWT hash.
type jwtExchanger struct {
	tokenEndpoint string
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

// isJWT is a heuristic: "eyJ" header prefix + two dots, no signature verification.
func isJWT(token string) bool {
	return strings.HasPrefix(token, "eyJ") && strings.Count(token, ".") == 2
}

// jwtTTL reads exp without signature verification; capped at 1h, falls back to 5m.
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

// cacheKey hashes the JWT so the full token is not kept in memory.
func cacheKey(jwt string) string {
	h := sha256.Sum256([]byte(jwt))
	return fmt.Sprintf("%x", h[:8])
}
