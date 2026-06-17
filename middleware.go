package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"
)

// requireAuthMiddleware returns RFC 6750 401 + WWW-Authenticate on unauthenticated
// tools/call so reactive OAuth clients can start their authorization flow without
// pre-probing the metadata endpoint. initialize and tools/list stay open.
func requireAuthMiddleware(next http.Handler, exchanger *jwtExchanger, logger *zap.SugaredLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		method, body := peekMCPMethod(r)
		r.Body = body
		if !methodRequiresAuth(method) {
			next.ServeHTTP(w, r)
			return
		}

		pat, err := extractPAT(r, exchanger)
		if err != nil {
			logger.Warnw("JWT→PAT exchange failed", "error", err)
			writeUnauthorized(w, r)
			return
		}
		if pat == "" {
			writeUnauthorized(w, r)
			return
		}

		next.ServeHTTP(w, r.WithContext(bitrise.ContextWithPAT(r.Context(), pat)))
	})
}

// peekMCPMethod restores the body after reading so downstream handlers still see it.
// Unparseable bodies yield "" (no auth required); the MCP server handles the error.
// JSON arrays are treated as tools/call so a batch can't sneak past the auth check.
func peekMCPMethod(r *http.Request) (mcp.MCPMethod, io.ReadCloser) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return "", io.NopCloser(bytes.NewReader(nil))
	}
	body := io.NopCloser(bytes.NewReader(raw))
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) > 0 && trimmed[0] == '[' {
		return mcp.MethodToolsCall, body
	}
	var msg struct {
		Method mcp.MCPMethod `json:"method"`
	}
	_ = json.Unmarshal(raw, &msg)
	return msg.Method, body
}

func methodRequiresAuth(method mcp.MCPMethod) bool {
	return method == mcp.MethodToolsCall
}

func writeUnauthorized(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", fmt.Sprintf("Bearer resource_metadata=%q", resourceMetadataURL(r)))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"unauthorized","error_description":"authentication required to call tools"}`))
}
