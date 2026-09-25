package devenvironments

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/devenv"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// APIGroup is the tool-filtering group every Dev Environments tool belongs
// to (ENABLED_API_GROUPS / x-bitrise-enabled-api-groups). ReadOnlyAPIGroup
// holds its read-only subset. The Dev Environments tools deliberately stay
// out of the Bitrise API's shared "read-only" group so a pre-existing
// CI-only configuration does not silently gain them.
const (
	APIGroup         = "dev-environments"
	ReadOnlyAPIGroup = "dev-environments-read-only"
)

// maxCallerWorkspaces caps the per-caller workspace memory. Entries are
// stored on the bearer token hash before the token is ever validated by a
// backend, so without a cap an unauthenticated client could grow the map
// without bound on the hosted server.
const maxCallerWorkspaces = 10000

// Belt holds the Dev Environments tools plus the classification used to
// resolve a workspace per request and to hide host-dependent tools when
// hosted. It is composed into the server-wide tool belt (internal/tool).
type Belt struct {
	tools []bitrise.Tool
	// owned is the set of tool names this belt registers, so the workspace
	// gate leaves every other tool of the merged server untouched.
	owned map[string]bool
	// userScoped lists tools that are NOT workspace-scoped (their backend path
	// has no /v1/workspaces/{id} segment). Every other tool is workspace-scoped,
	// so a new session/template tool is treated as workspaced by default.
	userScoped map[string]bool
	// localOnly lists tools that depend on the host the server runs on (the
	// local filesystem). They are hidden and rejected on the hosted HTTP
	// transport, where "local" would mean the server, not the user's machine.
	localOnly map[string]bool
	// lastWorkspace remembers, per MCP client session (keyed by the session
	// id), the workspace_id most recently passed explicitly, so a
	// multi-workspace user states it once instead of on every call.
	lastWorkspace sync.Map
	// callerWorkspace is the same memory keyed by a hash of the caller's
	// access token. The hosted server runs the streamable HTTP transport
	// stateless, so every request is a fresh client session and lastWorkspace
	// never repeats there; the token is the only thing that identifies the
	// same caller across requests. Entries expire after callerWorkspaceTTL
	// and are swept opportunistically so the map cannot grow without bound.
	// The memory is per server instance: a replica that never saw the
	// explicit value still asks. Scripts and CI should pass workspace_id.
	callerWorkspace sync.Map
	callerSweeps    atomic.Uint32
	callerCount     atomic.Int64
}

// callerWorkspaceTTL bounds how long a caller's last explicit workspace_id is
// remembered. Long enough for a working day, short enough that a stale choice
// does not follow a token forever.
const callerWorkspaceTTL = 12 * time.Hour

// callerSweepEvery is how many stores pass between expiry sweeps.
const callerSweepEvery = 256

// callerWorkspaceEntry is a remembered workspace plus when it was stored.
type callerWorkspaceEntry struct {
	workspace string
	storedAt  time.Time
}

// workspaceIDParamDesc documents the optional workspace_id parameter injected
// onto every workspace-scoped tool.
const workspaceIDParamDesc = "Workspace ID (slug) to operate in; list_workspaces lists them. Optional: if omitted the server reuses the workspace_id you last passed (about 12 hours per access token), then BITRISE_WORKSPACE_ID / the x-bitrise-workspace-id header, then the sole workspace you belong to. Scripts and CI should always pass it."

// NewBelt creates a new tool belt with all tools registered.
func NewBelt() *Belt {
	b := &Belt{
		tools: []bitrise.Tool{
			// Sessions
			ListSessions,
			GetSession,
			CreateSession,
			UpdateSession,
			RestoreSession,
			TerminateSession,
			DeleteSession,
			DeleteTerminatedSessions,
			CompareSessionTemplate,

			// Templates
			ListTemplates,
			GetTemplate,
			CreateTemplate,
			UpdateTemplate,
			DeleteTemplate,

			// Warm Pools
			ListWarmPools,
			GetWarmPool,
			CreateWarmPool,
			UpdateWarmPool,
			DeleteWarmPool,

			// Saved Inputs
			ListSavedInputs,
			GetSavedInput,
			CreateSavedInput,
			UpdateSavedInput,
			DeleteSavedInput,

			// Stacks & Machine Types
			ListStacks,
			ListMachineTypes,

			// Workspace Usage
			GetWorkspaceUsage,

			// Session Notifications
			ListSessionNotifications,

			// Session Interaction
			Execute,
			Screenshot,
			Click,
			Type,
			Scroll,
			MouseDrag,

			// File Transfer
			Upload,
			Download,

			// Remote Access
			OpenRemoteAccess,

			// Device preview links
			CreatePreviewLink,

			// Guides (fallback for clients without MCP resource support)
			DeviceGuide,
		},
		// User-scoped tools hit /v1/me or /v1/saved-inputs (no workspace
		// segment) or no API at all (device_guide serves embedded text).
		// Everything else is workspace-scoped.
		userScoped: map[string]bool{
			"bitrise_devenv_device_guide":       true,
			"bitrise_devenv_list_saved_inputs":  true,
			"bitrise_devenv_get_saved_input":    true,
			"bitrise_devenv_create_saved_input": true,
			"bitrise_devenv_update_saved_input": true,
			"bitrise_devenv_delete_saved_input": true,
		},
		// upload reads the local filesystem; download writes it. Neither makes
		// sense on a hosted server.
		localOnly: map[string]bool{
			"bitrise_devenv_upload":   true,
			"bitrise_devenv_download": true,
		},
	}

	b.owned = make(map[string]bool, len(b.tools))
	for i := range b.tools {
		b.owned[b.tools[i].Definition.Name] = true

		// Every Dev Environments tool belongs to the dev-environments API
		// group; read-only ones also join dev-environments-read-only
		// (centralised here rather than repeated across ~40 definitions).
		name := b.tools[i].Definition.Name
		ann := &b.tools[i].Definition.Annotations
		readOnly := ann.ReadOnlyHint != nil && *ann.ReadOnlyHint
		groups := []string{APIGroup}
		if readOnly {
			groups = append(groups, ReadOnlyAPIGroup)
		}
		b.tools[i].APIGroups = groups

		// The tool files set readOnlyHint/destructiveHint per tool; the two
		// remaining hints follow from the operation: reads and
		// delete/terminate/restore/update calls are idempotent, creates and
		// interactions are not; only the embedded guide is closed-world.
		idempotent := readOnly || isIdempotentMutation(name)
		openWorld := name != "bitrise_devenv_device_guide"
		ann.IdempotentHint = &idempotent
		ann.OpenWorldHint = &openWorld

		// Inject an optional workspace_id parameter on every workspace-scoped
		// tool (centralised so it isn't repeated across ~27 tool definitions).
		// It is the top rung of the workspace-resolution ladder in
		// GateAndResolveWorkspace.
		if b.userScoped[b.tools[i].Definition.Name] {
			continue
		}
		if b.tools[i].Definition.InputSchema.Properties == nil {
			b.tools[i].Definition.InputSchema.Properties = map[string]any{}
		}
		b.tools[i].Definition.InputSchema.Properties["workspace_id"] = map[string]any{
			"type":        "string",
			"description": workspaceIDParamDesc,
		}
	}

	return b
}

// isIdempotentMutation reports whether repeating the mutating tool with the
// same arguments has no additional effect (deletes, terminate/restore,
// updates and idempotent bulk deletes).
func isIdempotentMutation(name string) bool {
	for _, verb := range []string{"_delete", "_terminate", "_restore", "_update"} {
		if strings.Contains(name, verb) {
			return true
		}
	}
	return false
}

// Tools returns the Dev Environments tools for registration by the
// server-wide belt.
func (b *Belt) Tools() []bitrise.Tool {
	return b.tools
}

// Owns reports whether name is a Dev Environments tool.
func (b *Belt) Owns(name string) bool {
	return b.owned[name]
}

// FilterTools hides host-dependent (local-only) tools when the request is
// served over the hosted HTTP transport. Wired via server.WithToolFilter; in
// stdio mode (no hosted marker) every tool is listed.
func (b *Belt) FilterTools(ctx context.Context, tools []mcp.Tool) []mcp.Tool {
	if !devenv.HostedModeFromCtx(ctx) {
		return tools
	}
	filtered := make([]mcp.Tool, 0, len(tools))
	for _, t := range tools {
		if !b.localOnly[t.Name] {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// GateAndResolveWorkspace enforces transport-level tool availability and
// resolves the workspace for workspace-scoped tools. It returns the (possibly
// updated) context to use for the handler, or a non-nil error result that the
// caller should return to the client.
//
// PAT (and any per-connection default workspace) must already be in ctx: the
// stdio middleware injects them from env, the HTTP context func from the bearer
// token and the x-bitrise-workspace-id header.
func (b *Belt) GateAndResolveWorkspace(ctx context.Context, request mcp.CallToolRequest) (context.Context, *mcp.CallToolResult) {
	name := request.Params.Name
	if !b.owned[name] {
		// Not a Dev Environments tool: nothing to gate or resolve.
		return ctx, nil
	}

	// Host-dependent tools are unavailable when hosted (defense in depth — they
	// are also hidden from the tool list by FilterTools).
	if devenv.HostedModeFromCtx(ctx) && b.localOnly[name] {
		return ctx, mcp.NewToolResultError("this tool needs a locally-run MCP server (it reads or writes your machine's filesystem) and is unavailable on the hosted server; run the MCP server locally to use it")
	}

	// Resolve the workspace for workspace-scoped tools. Ladder (highest first):
	//   1. an explicit workspace_id tool parameter (remembered for this client
	//      session and for this caller's token)
	//   2. the workspace_id last passed explicitly in this client session
	//   3. the workspace_id last passed explicitly by this caller (token hash),
	//      within callerWorkspaceTTL — the rung that works on the stateless
	//      hosted transport, where rung 2 never matches
	//   4. the per-connection default (BITRISE_WORKSPACE_ID env / x-bitrise-workspace-id header)
	//   5. auto-detection of the user's sole workspace (cached per PAT)
	if !b.userScoped[name] {
		// The hosted transport is stateless: the server never issues a
		// session id, so any id a client sends is its own invention. Do not
		// key memory on it there (it would be a client-controlled,
		// unauthenticated map key); the token-hash rung covers hosted callers.
		sessionKey := ""
		if !devenv.HostedModeFromCtx(ctx) {
			sessionKey = clientSessionKey(ctx)
		}
		callerKey := callerKey(ctx)
		ws := request.GetString("workspace_id", "")
		if ws != "" {
			if sessionKey != "" {
				b.lastWorkspace.Store(sessionKey, ws)
			}
			b.rememberCallerWorkspace(callerKey, ws)
		}
		if ws == "" && sessionKey != "" {
			if v, ok := b.lastWorkspace.Load(sessionKey); ok {
				ws, _ = v.(string)
			}
		}
		if ws == "" {
			ws = b.callerWorkspaceFor(callerKey, time.Now())
		}
		if ws == "" {
			ws = devenv.WorkspaceFromCtx(ctx)
		}
		if ws == "" {
			detected, err := devenv.ResolveSoleWorkspace(ctx)
			if err != nil {
				return ctx, mcp.NewToolResultError(err.Error())
			}
			ws = detected
		}
		ctx = devenv.ContextWithWorkspace(ctx, ws)
	}

	return ctx, nil
}

// callerKey identifies the caller across requests by a SHA-256 of the access
// token in ctx, or "" when the request carries none. The raw token is never
// stored or logged.
func callerKey(ctx context.Context) string {
	pat := devenv.PATFromCtx(ctx)
	if pat == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(pat))
	return hex.EncodeToString(sum[:])
}

// rememberCallerWorkspace stores ws for callerKey (no-op for an empty key) and
// every callerSweepEvery stores drops entries older than callerWorkspaceTTL.
func (b *Belt) rememberCallerWorkspace(callerKey, ws string) {
	if callerKey == "" || ws == "" {
		return
	}
	now := time.Now()
	if _, known := b.callerWorkspace.Load(callerKey); !known {
		if b.callerCount.Load() >= maxCallerWorkspaces {
			b.sweepCallerWorkspaces(now)
		}
		if b.callerCount.Load() >= maxCallerWorkspaces {
			// Still full of live entries: forget nothing, remember nothing.
			return
		}
		b.callerCount.Add(1)
	}
	b.callerWorkspace.Store(callerKey, callerWorkspaceEntry{workspace: ws, storedAt: now})
	if b.callerSweeps.Add(1)%callerSweepEvery == 0 {
		b.sweepCallerWorkspaces(now)
	}
}

// callerWorkspaceFor returns the workspace remembered for callerKey when it is
// still within callerWorkspaceTTL at now, else "" (an expired entry is dropped).
func (b *Belt) callerWorkspaceFor(callerKey string, now time.Time) string {
	if callerKey == "" {
		return ""
	}
	v, ok := b.callerWorkspace.Load(callerKey)
	if !ok {
		return ""
	}
	e, _ := v.(callerWorkspaceEntry)
	if now.Sub(e.storedAt) > callerWorkspaceTTL {
		b.callerWorkspace.Delete(callerKey)
		b.callerCount.Add(-1)
		return ""
	}
	return e.workspace
}

// sweepCallerWorkspaces drops every remembered caller workspace older than
// callerWorkspaceTTL at now.
func (b *Belt) sweepCallerWorkspaces(now time.Time) {
	b.callerWorkspace.Range(func(k, v any) bool {
		if e, ok := v.(callerWorkspaceEntry); !ok || now.Sub(e.storedAt) > callerWorkspaceTTL {
			b.callerWorkspace.Delete(k)
			b.callerCount.Add(-1)
		}
		return true
	})
}

// clientSessionKey identifies the MCP client session a call belongs to, or ""
// when the transport attached none (nothing is remembered then).
func clientSessionKey(ctx context.Context) string {
	cs := server.ClientSessionFromContext(ctx)
	if cs == nil {
		return ""
	}
	return cs.SessionID()
}
