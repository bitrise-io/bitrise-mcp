package main

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// requireAuthMiddleware gates credential-less JSON-RPC requests at the HTTP layer
// with an RFC 6750 401 whose WWW-Authenticate header points at this server's RFC
// 9728 protected resource metadata. That 401 is the signal a spec-compliant
// reactive OAuth client waits for before starting its authorization flow.
//
// Unlike gating only tools/call, this challenges the initialize handshake too.
// Claude Code (and other clients) only surface the interactive, URL-showing auth
// flow once they have flagged a server as needing auth; that flag is set from a
// connection-level 401. If initialize succeeds anonymously, the client connects,
// lists tools, and only discovers the 401 reactively on the first tool call —
// where it refuses to surface the authorization URL. Challenging the connection
// itself flags the server as needing auth up front, so the very first tool
// invocation drives the auth flow instead of failing silently.
//
// The trade-off is that unauthenticated clients can no longer connect and
// enumerate tools; for an OAuth-gated resource that does nothing useful without
// credentials this is acceptable. PAT resolution for authenticated requests
// still happens in the WithHTTPContextFunc so it applies across transports.
//
// Only installed when an external OAuth issuer is configured; otherwise the
// server keeps its prior behaviour of surfacing missing auth as an in-band tool
// error.
func requireAuthMiddleware(next http.Handler, exchanger *jwtExchanger, metadataURL string, logger *zap.SugaredLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only POST carries JSON-RPC requests (initialize, tools/*). GET/DELETE
		// manage the SSE stream and session and never need credentials here.
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		pat, err := extractPAT(r, exchanger)
		if err != nil {
			// A bearer token was supplied but the JWT→PAT exchange rejected it
			// (e.g. expired or invalid token): treat it as unauthenticated.
			logger.Warnw("JWT→PAT exchange failed", "error", err)
		}
		if pat == "" {
			writeUnauthorized(w, metadataURL)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeUnauthorized(w http.ResponseWriter, metadataURL string) {
	w.Header().Set("WWW-Authenticate", fmt.Sprintf("Bearer resource_metadata=%q", metadataURL))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"unauthorized","error_description":"authentication required"}`))
}
