package devenvironments

import (
	"context"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/devenv"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
)

func toolNames(b *Belt) map[string]bool {
	m := map[string]bool{}
	for _, t := range b.tools {
		m[t.Definition.Name] = true
	}
	return m
}

func newReq(name string) mcp.CallToolRequest {
	var r mcp.CallToolRequest
	r.Params.Name = name
	return r
}

// TestClassificationKeysAreRealTools guards against typos in the userScoped /
// localOnly maps: every key must correspond to a registered tool, otherwise a
// tool would be silently mis-classified (e.g. a user-scoped tool treated as
// workspace-scoped and triggering a spurious GET /organizations).
func TestClassificationKeysAreRealTools(t *testing.T) {
	b := NewBelt()
	names := toolNames(b)
	for k := range b.userScoped {
		assert.Truef(t, names[k], "userScoped key %q is not a registered tool", k)
	}
	for k := range b.localOnly {
		assert.Truef(t, names[k], "localOnly key %q is not a registered tool", k)
	}
}

// TestToolAnnotations enforces the Claude connectors directory requirement:
// every tool has a title, and read-only and destructive tools are never
// flagged as both.
func TestToolAnnotations(t *testing.T) {
	for _, tl := range NewBelt().tools {
		a := tl.Definition.Annotations
		assert.NotEmptyf(t, a.Title, "tool %q has no title annotation", tl.Definition.Name)
		if a.ReadOnlyHint != nil && *a.ReadOnlyHint {
			assert.Falsef(t, a.DestructiveHint != nil && *a.DestructiveHint, "read-only tool %q is also marked destructive", tl.Definition.Name)
		}
	}
}

func TestFilterTools(t *testing.T) {
	b := NewBelt()
	tools := []mcp.Tool{
		{Name: "bitrise_devenv_upload"},
		{Name: "bitrise_devenv_download"},
		{Name: "bitrise_devenv_list"},
		{Name: "bitrise_devenv_device_guide"},
	}

	t.Run("stdio lists every tool", func(t *testing.T) {
		got := b.FilterTools(context.Background(), tools)
		assert.Len(t, got, 4)
	})

	t.Run("hosted hides local-only tools", func(t *testing.T) {
		got := b.FilterTools(devenv.ContextWithHostedMode(context.Background()), tools)
		var got_names []string
		for _, tl := range got {
			got_names = append(got_names, tl.Name)
		}
		assert.ElementsMatch(t, []string{"bitrise_devenv_list", "bitrise_devenv_device_guide"}, got_names)
	})
}

func TestGateAndResolveWorkspace(t *testing.T) {
	b := NewBelt()

	t.Run("local-only tool is rejected on the hosted transport", func(t *testing.T) {
		ctx := devenv.ContextWithHostedMode(context.Background())
		_, errRes := b.GateAndResolveWorkspace(ctx, newReq("bitrise_devenv_upload"))
		assert.NotNil(t, errRes)
		assert.True(t, errRes.IsError)
	})

	t.Run("local-only tool is allowed in stdio with a workspace set", func(t *testing.T) {
		ctx := devenv.ContextWithWorkspace(context.Background(), "ws")
		_, errRes := b.GateAndResolveWorkspace(ctx, newReq("bitrise_devenv_upload"))
		assert.Nil(t, errRes)
	})

	t.Run("user-scoped tool skips workspace resolution", func(t *testing.T) {
		// No workspace configured and MainAPIBaseURL unset: if resolution ran it
		// would error. A nil result proves it was skipped.
		_, errRes := b.GateAndResolveWorkspace(context.Background(), newReq("bitrise_devenv_device_guide"))
		assert.Nil(t, errRes)
	})

	t.Run("workspace-scoped tool with a configured workspace does not auto-detect", func(t *testing.T) {
		ctx := devenv.ContextWithWorkspace(context.Background(), "ws-1")
		gotCtx, errRes := b.GateAndResolveWorkspace(ctx, newReq("bitrise_devenv_list"))
		assert.Nil(t, errRes)
		assert.Equal(t, "ws-1", devenv.WorkspaceFromCtx(gotCtx))
	})

	t.Run("workspace-scoped tool with no workspace and no discovery config errors gracefully", func(t *testing.T) {
		old := bitrise.APIBaseURL
		bitrise.APIBaseURL = ""
		defer func() { bitrise.APIBaseURL = old }()

		ctx := devenv.ContextWithPAT(context.Background(), "pat")
		_, errRes := b.GateAndResolveWorkspace(ctx, newReq("bitrise_devenv_list"))
		assert.NotNil(t, errRes)
		assert.True(t, errRes.IsError)
	})
}

func newReqWithArgs(name string, args map[string]any) mcp.CallToolRequest {
	r := newReq(name)
	r.Params.Arguments = args
	return r
}

// TestWorkspaceIDParamInjection asserts every workspace-scoped tool exposes the
// optional workspace_id parameter and user-scoped tools do not.
func TestWorkspaceIDParamInjection(t *testing.T) {
	b := NewBelt()
	for _, tl := range b.tools {
		name := tl.Definition.Name
		_, has := tl.Definition.InputSchema.Properties["workspace_id"]
		if b.userScoped[name] {
			assert.Falsef(t, has, "%s is user-scoped and must NOT expose workspace_id", name)
		} else {
			assert.Truef(t, has, "%s is workspace-scoped and must expose workspace_id", name)
		}
	}
}

func TestGateAndResolveWorkspace_ParamWins(t *testing.T) {
	b := NewBelt()
	// An explicit workspace_id param beats the per-connection default in ctx.
	ctx := devenv.ContextWithWorkspace(context.Background(), "default-ws")
	req := newReqWithArgs("bitrise_devenv_list", map[string]any{"workspace_id": "param-ws"})
	gotCtx, errRes := b.GateAndResolveWorkspace(ctx, req)
	assert.Nil(t, errRes)
	assert.Equal(t, "param-ws", devenv.WorkspaceFromCtx(gotCtx))
}

// TestGateAndResolveWorkspace_RemembersParamPerClientSession asserts an
// explicit workspace_id is remembered for the MCP client session that passed
// it, beats the per-connection default on later calls, is scoped to that
// session, and is not remembered when no client session is attached.
func TestGateAndResolveWorkspace_RemembersParamPerClientSession(t *testing.T) {
	b := NewBelt()
	srv := server.NewMCPServer("test", "0.0.0")
	sessA := server.NewInProcessSession("sess-a", nil)
	sessB := server.NewInProcessSession("sess-b", nil)
	base := devenv.ContextWithWorkspace(context.Background(), "default-ws")
	ctxA := srv.WithContext(base, sessA)
	ctxB := srv.WithContext(base, sessB)

	// First call in session A names the workspace explicitly.
	gotCtx, errRes := b.GateAndResolveWorkspace(ctxA, newReqWithArgs("bitrise_devenv_list", map[string]any{"workspace_id": "param-ws"}))
	assert.Nil(t, errRes)
	assert.Equal(t, "param-ws", devenv.WorkspaceFromCtx(gotCtx))

	// A later call in session A without the parameter inherits it.
	gotCtx, errRes = b.GateAndResolveWorkspace(ctxA, newReq("bitrise_devenv_get"))
	assert.Nil(t, errRes)
	assert.Equal(t, "param-ws", devenv.WorkspaceFromCtx(gotCtx))

	// Passing a different value later replaces the remembered one.
	_, _ = b.GateAndResolveWorkspace(ctxA, newReqWithArgs("bitrise_devenv_list", map[string]any{"workspace_id": "other-ws"}))
	gotCtx, _ = b.GateAndResolveWorkspace(ctxA, newReq("bitrise_devenv_get"))
	assert.Equal(t, "other-ws", devenv.WorkspaceFromCtx(gotCtx))

	// Session B never passed one and falls through to the connection default.
	gotCtx, errRes = b.GateAndResolveWorkspace(ctxB, newReq("bitrise_devenv_list"))
	assert.Nil(t, errRes)
	assert.Equal(t, "default-ws", devenv.WorkspaceFromCtx(gotCtx))

	// Without a client session nothing is remembered: the next call falls back.
	_, _ = b.GateAndResolveWorkspace(base, newReqWithArgs("bitrise_devenv_list", map[string]any{"workspace_id": "anon-ws"}))
	gotCtx, _ = b.GateAndResolveWorkspace(base, newReq("bitrise_devenv_list"))
	assert.Equal(t, "default-ws", devenv.WorkspaceFromCtx(gotCtx))
}

// The hosted transport is stateless: every request is a fresh client session,
// so only the caller's token can tie requests together.
func TestGateAndResolveWorkspace_HostedIgnoresClientSessionID(t *testing.T) {
	b := NewBelt()
	// Two hosted requests sharing a (client-invented) session id but different
	// tokens must not share a remembered workspace.
	alice := devenv.ContextWithHostedMode(devenv.ContextWithPAT(context.Background(), "pat-hosted-alice"))
	bob := devenv.ContextWithHostedMode(devenv.ContextWithPAT(context.Background(), "pat-hosted-bob"))
	gotCtx, errRes := b.GateAndResolveWorkspace(alice, newReqWithArgs("bitrise_devenv_list", map[string]any{"workspace_id": "ws-alice"}))
	assert.Nil(t, errRes)
	assert.Equal(t, "ws-alice", devenv.WorkspaceFromCtx(gotCtx))
	_, loaded := b.lastWorkspace.Load("")
	assert.False(t, loaded, "hosted requests must not populate the session-keyed memory")
	assert.Equal(t, "", b.callerWorkspaceFor(callerKey(bob), time.Now()))
	assert.Equal(t, "ws-alice", b.callerWorkspaceFor(callerKey(alice), time.Now()))
}

func TestGateAndResolveWorkspace_RemembersParamPerCallerToken(t *testing.T) {
	b := NewBelt()
	srv := server.NewMCPServer("test", "0.0.0")
	base := devenv.ContextWithWorkspace(context.Background(), "default-ws")
	alice := devenv.ContextWithPAT(base, "pat-alice")
	bob := devenv.ContextWithPAT(base, "pat-bob")

	// Alice names the workspace in one client session…
	ctx1 := srv.WithContext(alice, server.NewInProcessSession("req-1", nil))
	gotCtx, errRes := b.GateAndResolveWorkspace(ctx1, newReqWithArgs("bitrise_devenv_list", map[string]any{"workspace_id": "alice-ws"}))
	assert.Nil(t, errRes)
	assert.Equal(t, "alice-ws", devenv.WorkspaceFromCtx(gotCtx))

	// …and inherits it in a DIFFERENT client session (stateless HTTP shape)…
	ctx2 := srv.WithContext(alice, server.NewInProcessSession("req-2", nil))
	gotCtx, errRes = b.GateAndResolveWorkspace(ctx2, newReq("bitrise_devenv_get"))
	assert.Nil(t, errRes)
	assert.Equal(t, "alice-ws", devenv.WorkspaceFromCtx(gotCtx))

	// …and with no client session at all.
	gotCtx, errRes = b.GateAndResolveWorkspace(alice, newReq("bitrise_devenv_get"))
	assert.Nil(t, errRes)
	assert.Equal(t, "alice-ws", devenv.WorkspaceFromCtx(gotCtx))

	// Bob's token shares nothing with Alice's and falls through to the default.
	gotCtx, errRes = b.GateAndResolveWorkspace(bob, newReq("bitrise_devenv_list"))
	assert.Nil(t, errRes)
	assert.Equal(t, "default-ws", devenv.WorkspaceFromCtx(gotCtx))

	// A client-session memory still wins over the caller memory.
	sess := server.NewInProcessSession("sess-x", nil)
	ctxS := srv.WithContext(alice, sess)
	_, _ = b.GateAndResolveWorkspace(ctxS, newReqWithArgs("bitrise_devenv_list", map[string]any{"workspace_id": "session-ws"}))
	gotCtx, _ = b.GateAndResolveWorkspace(ctxS, newReq("bitrise_devenv_get"))
	assert.Equal(t, "session-ws", devenv.WorkspaceFromCtx(gotCtx))

	// The raw token is never used as the key.
	b.callerWorkspace.Range(func(k, _ any) bool {
		assert.NotContains(t, k.(string), "pat-")
		assert.Len(t, k.(string), 64)
		return true
	})
}

func TestCallerWorkspaceExpiresAndSweeps(t *testing.T) {
	b := NewBelt()
	now := time.Now()
	b.rememberCallerWorkspace("k-fresh", "ws-fresh")
	b.callerWorkspace.Store("k-old", callerWorkspaceEntry{workspace: "ws-old", storedAt: now.Add(-callerWorkspaceTTL - time.Minute)})

	assert.Equal(t, "ws-fresh", b.callerWorkspaceFor("k-fresh", now))
	// Expired on read: dropped and not returned.
	assert.Equal(t, "", b.callerWorkspaceFor("k-old", now))
	_, still := b.callerWorkspace.Load("k-old")
	assert.False(t, still)

	// The periodic sweep drops expired entries without a read.
	b.callerWorkspace.Store("k-old2", callerWorkspaceEntry{workspace: "ws-old", storedAt: now.Add(-2 * callerWorkspaceTTL)})
	b.sweepCallerWorkspaces(now)
	_, still = b.callerWorkspace.Load("k-old2")
	assert.False(t, still)
	assert.Equal(t, "ws-fresh", b.callerWorkspaceFor("k-fresh", now))

	// Empty keys and values are ignored.
	b.rememberCallerWorkspace("", "ws")
	b.rememberCallerWorkspace("k-empty", "")
	assert.Equal(t, "", b.callerWorkspaceFor("", now))
	assert.Equal(t, "", b.callerWorkspaceFor("k-empty", now))
}
