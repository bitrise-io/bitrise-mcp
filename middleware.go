package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ctxWriterKey struct{}
type ctxMetadataURLKey struct{}

// hijackableWriter lets requireAuthToolHandler write a 401 and then suppress
// mcp-go's own error write that follows — mcp-go has no hook to return an HTTP
// error status from a tool handler, so we intercept at the transport layer.
type hijackableWriter struct {
	http.ResponseWriter
	done bool
}

func (h *hijackableWriter) Write(b []byte) (int, error) {
	if h.done {
		return len(b), nil
	}
	return h.ResponseWriter.Write(b)
}

func (h *hijackableWriter) WriteHeader(code int) {
	if h.done {
		return
	}
	h.ResponseWriter.WriteHeader(code)
}

// withAuthContext wraps the response writer in a hijackableWriter and stores it
// alongside the metadata URL in the context. requireAuthToolHandler reads both to
// write a 401 directly to the wire and discard mcp-go's subsequent error write.
// PAT extraction happens in WithHTTPContextFunc so it applies to all transports.
func withAuthContext(next http.Handler, metadataURL string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hw := &hijackableWriter{ResponseWriter: w}
		ctx := context.WithValue(r.Context(), ctxWriterKey{}, hw)
		ctx = context.WithValue(ctx, ctxMetadataURLKey{}, metadataURL)
		next.ServeHTTP(hw, r.WithContext(ctx))
	})
}

// requireAuthToolHandler is a WithToolHandlerMiddleware that returns RFC 6750
// 401 + WWW-Authenticate for unauthenticated tool calls. It writes the 401
// directly to the response writer stored in context by withAuthContext, then
// hijacks that writer so mcp-go's subsequent error response is silently discarded.
func requireAuthToolHandler(fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		pat, err := bitrise.PATFromCtx(ctx)
		if pat == "" || err != nil {
			if hw, ok := ctx.Value(ctxWriterKey{}).(*hijackableWriter); ok {
				metadataURL, _ := ctx.Value(ctxMetadataURLKey{}).(string)
				writeUnauthorized(hw, metadataURL)
				hw.done = true
			}
			return nil, errors.New("unauthorized")
		}
		return fn(ctx, request)
	}
}

func writeUnauthorized(w http.ResponseWriter, metadataURL string) {
	w.Header().Set("WWW-Authenticate", fmt.Sprintf("Bearer resource_metadata=%q", metadataURL))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"unauthorized","error_description":"authentication required to call tools"}`))
}
