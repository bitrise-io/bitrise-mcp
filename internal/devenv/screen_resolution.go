package devenv

import "sync"

// DefaultResolution is the resolution assumed when no screenshot has been
// captured yet for a session. It matches the standard macOS session image.
var DefaultResolution = Resolution{Width: 1920, Height: 1080}

// Resolution is the pixel dimensions of a session's display screenshot.
type Resolution struct {
	Width  int
	Height int
}

var (
	screenResMu    sync.RWMutex
	screenResCache = map[string]Resolution{}
)

// SetScreenResolution stores the screenshot resolution for a session.
func SetScreenResolution(sessionID string, r Resolution) {
	screenResMu.Lock()
	defer screenResMu.Unlock()
	screenResCache[sessionID] = r
}

// GetScreenResolution returns the cached resolution for a session, falling
// back to DefaultResolution when the screenshot tool hasn't run yet for it.
// The boolean reports whether the value came from the cache.
func GetScreenResolution(sessionID string) (Resolution, bool) {
	screenResMu.RLock()
	defer screenResMu.RUnlock()
	r, ok := screenResCache[sessionID]
	if !ok {
		return DefaultResolution, false
	}
	return r, true
}
