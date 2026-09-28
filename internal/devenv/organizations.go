package devenv

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
)

// Organization is a workspace the authenticated user belongs to. Field names
// match v0.OrganizationResponseModel in the Bitrise API.
type Organization struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// ListOrganizations returns the workspaces (organizations) the authenticated
// user can access via GET {bitrise.APIBaseURL}/organizations. The Dev
// Environments backend has no list-workspaces endpoint (only /v1/me,
// /v1/saved-inputs and /v1/workspaces/{id}/...), so the main Bitrise API
// serves workspace discovery.
func ListOrganizations(ctx context.Context) ([]Organization, error) {
	body, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
		Method:  http.MethodGet,
		BaseURL: bitrise.APIBaseURL,
		Path:    "/organizations",
	})
	if err != nil {
		return nil, fmt.Errorf("list workspaces: %w", err)
	}

	// Bitrise v0.1 list responses wrap items under a top-level "data" array.
	var page struct {
		Data []Organization `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &page); err != nil {
		return nil, fmt.Errorf("parse organizations response: %w", err)
	}
	return page.Data, nil
}

// SoleWorkspace returns the user's workspace when they have exactly one. With
// zero or 2+ workspaces it returns a friendly error listing them, since no
// default can be picked unambiguously.
func SoleWorkspace(orgs []Organization) (Organization, error) {
	switch len(orgs) {
	case 0:
		return Organization{}, fmt.Errorf("no workspaces found for this account — create one in the Bitrise dashboard, then pass workspace_id")
	case 1:
		return orgs[0], nil
	default:
		var b strings.Builder
		for _, o := range orgs {
			if o.Name != "" {
				fmt.Fprintf(&b, "\n  %s (%s)", o.Name, o.Slug)
			} else {
				fmt.Fprintf(&b, "\n  %s", o.Slug)
			}
		}
		return Organization{}, fmt.Errorf("you belong to multiple Bitrise workspaces, so one can't be chosen automatically. ASK THE USER which workspace to use, then pass its slug as the workspace_id argument once — this server remembers the last value you passed for about 12 hours (per access token, per server instance); if a later call still asks, pass it again. Scripts and CI should always pass it explicitly. Do NOT retry across workspaces or list them one by one, and do NOT guess. Available workspaces:%s", b.String())
	}
}

type wsCacheEntry struct {
	slug      string
	expiresAt time.Time
}

// The sole-workspace cache is process-global and keyed by token hash. Entries
// expire after workspaceCacheTTL; expired entries are swept on every
// workspaceCacheSweepEvery stores and whenever the cache is full, and a cache
// that is still full of live entries admits no new key, so it cannot grow with
// the lifetime number of users of a hosted server.
const (
	workspaceCacheTTL        = 5 * time.Minute
	workspaceCacheMaxEntries = 10000
	workspaceCacheSweepEvery = 256
)

var (
	workspaceCacheMu     sync.Mutex
	workspaceCache       = map[string]wsCacheEntry{} // wsCacheKey(pat) -> entry
	workspaceCacheStores uint32
)

func cachedSoleWorkspace(key string, now time.Time) (string, bool) {
	workspaceCacheMu.Lock()
	defer workspaceCacheMu.Unlock()
	e, ok := workspaceCache[key]
	if !ok {
		return "", false
	}
	if !now.Before(e.expiresAt) {
		delete(workspaceCache, key)
		return "", false
	}
	return e.slug, true
}

func storeSoleWorkspace(key, slug string, now time.Time) {
	workspaceCacheMu.Lock()
	defer workspaceCacheMu.Unlock()
	if _, known := workspaceCache[key]; !known && len(workspaceCache) >= workspaceCacheMaxEntries {
		sweepWorkspaceCacheLocked(now)
		if len(workspaceCache) >= workspaceCacheMaxEntries {
			return
		}
	}
	workspaceCache[key] = wsCacheEntry{slug: slug, expiresAt: now.Add(workspaceCacheTTL)}
	workspaceCacheStores++
	if workspaceCacheStores%workspaceCacheSweepEvery == 0 {
		sweepWorkspaceCacheLocked(now)
	}
}

func sweepWorkspaceCacheLocked(now time.Time) {
	for k, e := range workspaceCache {
		if !now.Before(e.expiresAt) {
			delete(workspaceCache, k)
		}
	}
}

// ResolveSoleWorkspace returns the slug of the user's only workspace,
// auto-detecting it via GET /organizations. The result is cached per PAT for a
// short window so that, when no workspace is configured, repeated
// workspace-scoped tool calls don't each issue a discovery request.
//
// The zero- and multiple-workspace cases (SoleWorkspace's errors) are never
// cached, so a freshly created workspace is picked up promptly and the
// multi-workspace error always reflects current state.
func ResolveSoleWorkspace(ctx context.Context) (string, error) {
	pat := bitrise.PATFromCtx(ctx)
	key := wsCacheKey(pat)
	if pat != "" {
		if slug, ok := cachedSoleWorkspace(key, time.Now()); ok {
			return slug, nil
		}
	}

	orgs, err := ListOrganizations(ctx)
	if err != nil {
		return "", err
	}
	ws, err := SoleWorkspace(orgs)
	if err != nil {
		return "", err
	}
	if pat != "" {
		storeSoleWorkspace(key, ws.Slug, time.Now())
	}
	return ws.Slug, nil
}

// wsCacheKey hashes the PAT so the full token is not kept in memory.
func wsCacheKey(pat string) string {
	h := sha256.Sum256([]byte(pat))
	return fmt.Sprintf("%x", h[:8])
}
