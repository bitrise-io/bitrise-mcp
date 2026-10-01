package devenvironments

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Device-session guides, served as MCP resources so an agent that creates a
// session with a device_spec can read the know-how (readiness, connectivity,
// accessibility tree, input, screenshots, do-nots, recovery) instead of
// re-deriving it. The markdown mirrors the RDE backend's device-session guide
// verbatim; the backend is the source of truth — keep these in sync with the
// backend release this server targets rather than editing here.
//
// Served both as resources (they cost no context until an agent asks for
// them) and as the DeviceGuide tool, sharing this slice so the two cannot
// drift. The tool is the door the create/get descriptions point at first: in
// practice many MCP clients cannot read resources at all, and a pointer that
// only works sometimes is not a pointer.

//go:embed guides/device-sessions.md
var guideDeviceSessionsRaw string

//go:embed guides/ios.md
var guideIOSRaw string

//go:embed guides/android.md
var guideAndroidRaw string

// guides/macos-automation.md is maintained here (it is not a backend mirror).
//
//go:embed guides/macos-automation.md
var guideMacOSAutomation string

// The mirror files open with a one-line HTML comment addressed to whoever
// edits them ("Mirror of the RDE device-session guide … do not edit here").
// That note is for the maintainer, not the agent: served verbatim it was the
// first line every agent read. Strip it here so the file keeps its warning
// and the guide starts at its title.
var (
	guideDeviceSessions = stripMirrorHeader(guideDeviceSessionsRaw)
	guideIOS            = stripMirrorHeader(guideIOSRaw)
	guideAndroid        = stripMirrorHeader(guideAndroidRaw)
)

// stripMirrorHeader drops a leading HTML comment line (and the blank lines
// after it) from a mirrored guide; any other text is returned unchanged.
func stripMirrorHeader(s string) string {
	if !strings.HasPrefix(s, "<!--") {
		return s
	}
	end := strings.Index(s, "-->")
	if end < 0 {
		return s
	}
	return strings.TrimLeft(s[end+len("-->"):], "\r\n")
}

// GuideURIDeviceSessions is the platform-independent device session guide.
const GuideURIDeviceSessions = "bitrise-devenv://guides/device-sessions"

// GuideURIIOS / GuideURIAndroid are the per-platform driving guides;
// GuideURIMacOSAutomation is the desktop (session VM) automation cookbook.
const (
	GuideURIIOS             = "bitrise-devenv://guides/device-sessions/ios"
	GuideURIAndroid         = "bitrise-devenv://guides/device-sessions/android"
	GuideURIMacOSAutomation = "bitrise-devenv://guides/macos-automation"
)

type guide struct {
	// key selects the guide in the DeviceGuide tool ("device-sessions", "ios", "android").
	key                          string
	uri, name, description, body string
}

// guideByKey returns the guide the DeviceGuide tool serves for key.
func guideByKey(key string) (guide, bool) {
	for _, g := range guides {
		if g.key == key {
			return g, true
		}
	}
	return guide{}, false
}

// guideKeys lists the DeviceGuide tool's accepted keys in guide order.
func guideKeys() []string {
	keys := make([]string, 0, len(guides))
	for _, g := range guides {
		keys = append(keys, g.key)
	}
	return keys
}

// Guides lists the resources RegisterResources serves (exported for tests and
// the README table).
func Guides() []struct{ URI, Name string } {
	out := make([]struct{ URI, Name string }, 0, len(guides))
	for _, g := range guides {
		out = append(out, struct{ URI, Name string }{g.uri, g.name})
	}
	return out
}

var guides = []guide{
	{
		key:         "device-sessions",
		uri:         GuideURIDeviceSessions,
		name:        "Device sessions — agent guide",
		description: "How to create an RDE session that boots an iOS simulator or Android emulator (bitrise_devenv_create with device_spec), wait for the device to be READY (and what a FAILED device really means), connect (bitrise_devenv_execute or an SSH tunnel), drive it efficiently (accessibility tree first, then input), get files off the VM, let a human watch from the session page in the RDE web UI, and what never to do. Read this before creating or driving a device session.",
		body:        guideDeviceSessions,
	},
	{
		key:         "ios",
		uri:         GuideURIIOS,
		name:        "Device sessions — iOS simulator specifics",
		description: "Driving the iOS simulator on a device session: simctl, serve-sim's CLI (tap/type/button/gesture — run with TMPDIR=/tmp) and its /helper/<UDID>/ax accessibility JSON endpoint, screenshots, logs, and recovery.",
		body:        guideIOS,
	},
	{
		key:         "android",
		uri:         GuideURIAndroid,
		name:        "Device sessions — Android emulator specifics",
		description: "Driving the Android emulator on a device session: adb over the tunnel or in-band, uiautomator dump for the accessibility tree, input, install/launch, screenshots, logs, and recovery.",
		body:        guideAndroid,
	},
	{
		key:         "macos-automation",
		uri:         GuideURIMacOSAutomation,
		name:        "macOS session automation — shell recipes",
		description: "Driving a macOS session's desktop from bitrise_devenv_execute instead of screenshots and clicks: opening System Settings panes, launching apps, menus and keystrokes with osascript, defaults, and the osascript timeout safety net.",
		body:        guideMacOSAutomation,
	},
}

// RegisterResources registers the guide resources with the MCP server.
func (b *Belt) RegisterResources(s *server.MCPServer) {
	for _, g := range guides {
		g := g
		s.AddResource(
			mcp.NewResource(g.uri, g.name, mcp.WithResourceDescription(g.description), mcp.WithMIMEType("text/markdown")),
			func(_ context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
				return []mcp.ResourceContents{mcp.TextResourceContents{
					URI:      req.Params.URI,
					MIMEType: "text/markdown",
					Text:     g.body,
				}}, nil
			},
		)
	}
}

// DeviceGuide serves the device-session guides as a tool, for MCP clients that
// cannot read resources. It returns exactly the markdown the resource of the
// same guide returns — both read the guides slice — so an agent on a
// tools-only client gets the same know-how, just paid for in context up front.
var DeviceGuide = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_device_guide",
		mcp.WithTitleAnnotation("Read device session guide"),
		mcp.WithDescription(`Returns the device-session agent guide as markdown. Read "device-sessions" BEFORE creating or driving a session that boots an iOS simulator / Android emulator (bitrise_devenv_create with device_spec, or a template that declares one), then "ios" or "android" for the platform you boot. It is the know-how the tool descriptions cannot hold: create rules, the readiness contract ("running" is not "ready"; what a FAILED device really means), connecting, the accessibility tree, input, screenshots, getting files off the VM, optionally letting a human watch, recovery, and what never to do. It applies equally to unattended/batch runs — the device is driven from execute (simctl / adb / serve-sim); nobody needs to watch.

"macos-automation" is the shell cookbook for driving a macOS session's desktop (open, osascript, defaults) instead of screenshots and clicks.

The same text is also served as the MCP resources bitrise-devenv://guides/device-sessions, .../ios, .../android and bitrise-devenv://guides/macos-automation for clients that read resources; this tool works everywhere.`),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithString("guide",
			mcp.Description(`Which guide to return: "device-sessions" (platform-independent, read first), "ios" (simulator specifics: simctl, serve-sim CLI, /ax accessibility endpoint), "android" (emulator specifics: adb, uiautomator dump, input, install) or "macos-automation" (driving a macOS session's desktop from the shell instead of screenshots and clicks).`),
			mcp.Required(),
			mcp.Enum(guideKeys()...),
		),
	),
	Handler: func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		key := request.GetString("guide", "")
		g, ok := guideByKey(key)
		if !ok {
			return mcp.NewToolResultError(fmt.Sprintf("unknown guide %q; one of %s", key, strings.Join(guideKeys(), ", "))), nil
		}
		return mcp.NewToolResultText(g.body), nil
	},
}
