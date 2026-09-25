package devenv

import (
	"context"
	"fmt"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
)

type ctxKey int

const (
	keyWorkspace ctxKey = iota
	keyHostedMode
)

// ContextWithPAT stores the Bitrise token in the context. The token lives in
// the shared bitrise package's context slot, so the transports inject it once
// and both the Bitrise API tools and the Dev Environments tools read the same
// value.
func ContextWithPAT(ctx context.Context, pat string) context.Context {
	return bitrise.ContextWithPAT(ctx, pat)
}

// AuthFromCtx returns the Authorization header name and value the Dev
// Environments backend expects ("Bearer <token>", unlike the main Bitrise API
// which takes the raw token).
func AuthFromCtx(ctx context.Context) (headerName, headerValue string, err error) {
	if v := bitrise.PATFromCtx(ctx); v != "" {
		return "Authorization", "Bearer " + v, nil
	}
	return "", "", fmt.Errorf("missing authentication - complete the OAuth flow, or set the BITRISE_TOKEN env var (stdio) / Authorization header (http) to a Bitrise personal access token")
}

// PATFromCtx returns the caller's raw Bitrise token from context, or "" when
// none is attached. Callers must never log or persist the value; hash it when
// a per-caller key is needed (see devenvironments.callerKey).
func PATFromCtx(ctx context.Context) string {
	return bitrise.PATFromCtx(ctx)
}

// ContextWithWorkspace returns a new context with the resolved workspace ID
// (slug) stored. Tools read it via WorkspaceFromCtx / WsPath.
func ContextWithWorkspace(ctx context.Context, workspaceID string) context.Context {
	return context.WithValue(ctx, keyWorkspace, workspaceID)
}

// WorkspaceFromCtx returns the resolved workspace ID from context, or "" if none.
func WorkspaceFromCtx(ctx context.Context) string {
	if v, ok := ctx.Value(keyWorkspace).(string); ok {
		return v
	}
	return ""
}

// ContextWithHostedMode marks the request as served over the hosted HTTP
// transport. The tool filter uses this to hide host-dependent (local-only)
// tools that only work when the server runs on the user's own machine.
func ContextWithHostedMode(ctx context.Context) context.Context {
	return context.WithValue(ctx, keyHostedMode, true)
}

// HostedModeFromCtx reports whether the request is served over the hosted HTTP
// transport.
func HostedModeFromCtx(ctx context.Context) bool {
	v, ok := ctx.Value(keyHostedMode).(bool)
	return ok && v
}
