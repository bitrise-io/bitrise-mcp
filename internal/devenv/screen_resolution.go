package devenv

import (
	"sync"
	"time"
)

// DefaultResolution is the resolution assumed when no screenshot has been
// captured yet for a session. It matches the standard macOS session image.
var DefaultResolution = Resolution{Width: 1920, Height: 1080}

// Resolution is the pixel dimensions of a session's display screenshot.
type Resolution struct {
	Width  int
	Height int
}

// The cache is process-global and keyed by session id. Entries expire after
// screenResTTL (a session's display does not change size, but the session
// does not live forever) and the map is capped so a long-running hosted
// server cannot accumulate one entry per session it ever screenshotted.
const (
	screenResTTL        = 24 * time.Hour
	screenResMaxEntries = 4096
)

type screenResEntry struct {
	res      Resolution
	storedAt time.Time
}

var (
	screenResMu    sync.RWMutex
	screenResCache = map[string]screenResEntry{}
)

// SetScreenResolution stores the screenshot resolution for a session.
func SetScreenResolution(sessionID string, r Resolution) {
	now := time.Now()
	screenResMu.Lock()
	defer screenResMu.Unlock()
	if len(screenResCache) >= screenResMaxEntries {
		sweepScreenResolutionsLocked(now)
	}
	for k := range screenResCache {
		if len(screenResCache) < screenResMaxEntries {
			break
		}
		delete(screenResCache, k)
	}
	screenResCache[sessionID] = screenResEntry{res: r, storedAt: now}
}

// ForgetScreenResolution drops the cached resolution of a session that no
// longer exists.
func ForgetScreenResolution(sessionID string) {
	screenResMu.Lock()
	defer screenResMu.Unlock()
	delete(screenResCache, sessionID)
}

// GetScreenResolution returns the cached resolution for a session, falling
// back to DefaultResolution when the screenshot tool hasn't run yet for it
// or the entry has expired. The boolean reports whether the value came from
// the cache.
func GetScreenResolution(sessionID string) (Resolution, bool) {
	screenResMu.RLock()
	defer screenResMu.RUnlock()
	e, ok := screenResCache[sessionID]
	if !ok || time.Since(e.storedAt) > screenResTTL {
		return DefaultResolution, false
	}
	return e.res, true
}

// sweepScreenResolutionsLocked drops expired entries; the caller holds the
// write lock.
func sweepScreenResolutionsLocked(now time.Time) {
	for k, e := range screenResCache {
		if now.Sub(e.storedAt) > screenResTTL {
			delete(screenResCache, k)
		}
	}
}
