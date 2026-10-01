package devenvironments

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// The guides are embedded copies of the RDE backend's docs; an empty or
// truncated copy would silently ship an agent nothing to read.
func TestGuideResources(t *testing.T) {
	s := server.NewMCPServer("test", "0.0.0")
	NewBelt().RegisterResources(s)
	got := s.ListResources()
	if len(got) != len(guides) {
		t.Fatalf("registered %d resources, want %d", len(got), len(guides))
	}
	for _, g := range guides {
		res, ok := got[g.uri]
		if !ok {
			t.Fatalf("resource %s not registered", g.uri)
		}
		contents, err := res.Handler(context.Background(), mcp.ReadResourceRequest{Params: mcp.ReadResourceParams{URI: g.uri}})
		if err != nil {
			t.Fatalf("read %s: %v", g.uri, err)
		}
		if len(contents) != 1 {
			t.Fatalf("read %s: %d contents, want 1", g.uri, len(contents))
		}
		text, ok := contents[0].(mcp.TextResourceContents)
		if !ok {
			t.Fatalf("read %s: contents are %T, want TextResourceContents", g.uri, contents[0])
		}
		keyword := "device"
		if g.key == "macos-automation" {
			keyword = "osascript"
		}
		if len(text.Text) < 1000 || !strings.Contains(text.Text, keyword) {
			t.Errorf("guide %s looks truncated (%d bytes)", g.uri, len(text.Text))
		}
	}
	if !strings.Contains(guideDeviceSessions, "PREVIEW_DEVICE_STATE_READY") {
		t.Errorf("the main guide must document the readiness contract")
	}
}

// The mirror files carry a maintainer-only HTML comment on line 1; the guide
// an agent receives must start at its title, not at a note about syncing.
func TestGuidesDoNotServeTheMirrorHeader(t *testing.T) {
	for _, g := range guides {
		if strings.HasPrefix(g.body, "<!--") || strings.Contains(g.body, "do not edit here") {
			t.Errorf("guide %s still serves the mirror header:\n%.120s", g.key, g.body)
		}
		if !strings.HasPrefix(g.body, "# ") {
			t.Errorf("guide %s must start at its title, got %.80q", g.key, g.body)
		}
	}
	for _, tc := range []struct{ in, want string }{
		{"<!-- note -->\n\n# Title\n", "# Title\n"},
		{"<!-- note -->\n# Title\n", "# Title\n"},
		{"# Title\n", "# Title\n"},
		{"<!-- unterminated\n# Title\n", "<!-- unterminated\n# Title\n"},
	} {
		if got := stripMirrorHeader(tc.in); got != tc.want {
			t.Errorf("stripMirrorHeader(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// The tool is the resource for clients that cannot read resources: for every
// guide it must return byte-for-byte what the resource returns, and nothing
// else is accepted.
func TestDeviceGuideToolMatchesResources(t *testing.T) {
	for _, g := range guides {
		req := mcp.CallToolRequest{}
		req.Params.Name = DeviceGuide.Definition.Name
		req.Params.Arguments = map[string]any{"guide": g.key}
		res, err := DeviceGuide.Handler(context.Background(), req)
		if err != nil {
			t.Fatalf("guide %s: %v", g.key, err)
		}
		if res.IsError {
			t.Fatalf("guide %s: tool error %v", g.key, res.Content)
		}
		if len(res.Content) != 1 {
			t.Fatalf("guide %s: %d contents, want 1", g.key, len(res.Content))
		}
		text, ok := res.Content[0].(mcp.TextContent)
		if !ok {
			t.Fatalf("guide %s: content is %T, want TextContent", g.key, res.Content[0])
		}
		if text.Text != g.body {
			t.Errorf("guide %s: tool output differs from the resource body", g.key)
		}
	}

	req := mcp.CallToolRequest{}
	req.Params.Name = DeviceGuide.Definition.Name
	req.Params.Arguments = map[string]any{"guide": "windows"}
	res, err := DeviceGuide.Handler(context.Background(), req)
	if err != nil || !res.IsError {
		t.Fatalf("unknown guide must be a tool error, got err=%v isError=%v", err, res != nil && res.IsError)
	}
}
