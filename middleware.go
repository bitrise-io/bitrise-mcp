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

// requireAuthMiddleware gates tool-executing MCP requests behind a valid bearer
// token. When a tools/call arrives without usable credentials it responds with
// an RFC 6750 401 whose WWW-Authenticate header points at this server's RFC 9728
// protected resource metadata. That 401 is the signal a spec-compliant reactive
// OAuth client waits for before starting its authorization flow; without it such
// clients connect and list tools fine but can never authenticate to run them.
//
// initialize and tools/list are intentionally left open so clients can connect
// and enumerate tools before authorizing. The middleware is only installed when
// an external OAuth issuer is configured; otherwise the server keeps its prior
// behaviour of surfacing missing auth as an in-band tool error.
func requireAuthMiddleware(next http.Handler, exchanger *jwtExchanger, logger *zap.SugaredLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only POST carries JSON-RPC requests; GET/DELETE manage the SSE stream
		// and session and never execute tools.
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
			// A bearer token was supplied but the JWT→PAT exchange rejected it
			// (e.g. expired or invalid token): treat it as unauthenticated.
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

// peekMCPMethod reads the JSON-RPC method from the request body without
// consuming it, returning the method and a fresh body for downstream handlers.
// A body that is missing, unreadable, or not a single JSON-RPC object yields an
// empty method, which is treated as not requiring auth so the streamable HTTP
// server can produce its own parse error.
func peekMCPMethod(r *http.Request) (mcp.MCPMethod, io.ReadCloser) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return "", io.NopCloser(bytes.NewReader(nil))
	}
	var msg struct {
		Method mcp.MCPMethod `json:"method"`
	}
	_ = json.Unmarshal(raw, &msg)
	return msg.Method, io.NopCloser(bytes.NewReader(raw))
}

// methodRequiresAuth reports whether an MCP method executes a tool and therefore
// needs Bitrise credentials. Only tools/call does.
func methodRequiresAuth(method mcp.MCPMethod) bool {
	return method == mcp.MethodToolsCall
}

// writeUnauthorized emits an RFC 6750 401 whose WWW-Authenticate header points
// OAuth clients at this server's RFC 9728 protected resource metadata.
func writeUnauthorized(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", fmt.Sprintf("Bearer resource_metadata=%q", resourceMetadataURL(r)))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"unauthorized","error_description":"authentication required to call tools"}`))
}
