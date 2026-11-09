package main

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool"
	"github.com/stretchr/testify/assert"
)

// optInToolsets are left out of defaultToolsets on purpose: a client gets
// their tools only by naming the toolset or all. The default toolset is not a
// stable contract (see the toolset configuration RFC), so a toolset can be
// moved here from defaultToolsets.
var optInToolsets = []string{}

// TestDefaultToolsetsMatchToolGroups keeps defaultToolsets in step with the
// toolsets tools are tagged with: every toolset a tool uses is either enabled
// by default or opt-in, so a new toolset can't land without a decision, and
// every default or opt-in toolset enables at least one tool, so naming it in
// ENABLED_TOOLSET, the toolset query parameter or a header never returns an
// empty tool list.
func TestDefaultToolsetsMatchToolGroups(t *testing.T) {
	defaults := strings.Split(defaultToolsets, ",")

	toolGroups := map[string]bool{}
	for _, tl := range tool.NewBelt().Tools() {
		for _, group := range tl.APIGroups {
			toolGroups[group] = true
			if !slices.Contains(optInToolsets, group) {
				assert.Containsf(t, defaults, group, "%s: toolset %q is missing from defaultToolsets (or from optInToolsets, if it is opt-in)", tl.Definition.Name, group)
			}
		}
	}
	for _, group := range defaults {
		assert.Truef(t, toolGroups[group], "default toolset %q enables no tool", group)
	}
	for _, group := range optInToolsets {
		assert.Truef(t, toolGroups[group], "opt-in toolset %q enables no tool", group)
		assert.NotContainsf(t, defaults, group, "opt-in toolset %q is in defaultToolsets", group)
	}
}

// TestOptInToolsAreNotEnabledByDefault catches a tool of an opt-in toolset
// that is also tagged with a default toolset, such as read-only or
// release-management, which would enable it by default.
func TestOptInToolsAreNotEnabledByDefault(t *testing.T) {
	belt := tool.NewBelt()
	defaults := belt.EnabledToolsets(defaultToolsets)
	for _, tl := range belt.Tools() {
		for _, group := range tl.APIGroups {
			if slices.Contains(optInToolsets, group) {
				assert.Falsef(t, belt.ToolEnabled(tl.Definition.Name, defaults), "%s is in opt-in toolset %q, but the default toolsets enable it", tl.Definition.Name, group)
			}
		}
	}
}

func TestConfigToolsets(t *testing.T) {
	assert.Equal(t, defaultToolsets, config{}.toolsets())
	assert.Equal(t, "apps", config{EnabledAPIGroups: "apps"}.toolsets(), "ENABLED_API_GROUPS is the fallback")
	assert.Equal(t, "builds", config{EnabledToolsets: "builds", EnabledAPIGroups: "apps"}.toolsets(), "ENABLED_TOOLSET wins")
}

func TestRequestToolsets(t *testing.T) {
	request := func(target string, headers map[string]string) *http.Request {
		r := httptest.NewRequest(http.MethodPost, target, nil)
		for k, v := range headers {
			r.Header.Set(k, v)
		}
		return r
	}
	oldHeader := "x-bitrise-enabled-api-groups"
	newHeader := "x-bitrise-enabled-toolset"

	assert.Equal(t, "default", requestToolsets(request("/", nil), "default"))
	assert.Equal(t, "apps", requestToolsets(request("/", map[string]string{oldHeader: "apps"}), "default"))
	assert.Equal(t, "builds", requestToolsets(request("/", map[string]string{oldHeader: "apps", newHeader: "builds"}), "default"))
	assert.Equal(t, "all", requestToolsets(request("/?toolset=all", map[string]string{oldHeader: "apps", newHeader: "builds"}), "default"))
	assert.Equal(t, "builds,rde", requestToolsets(request("/?toolset=builds,rde", nil), "default"))
}
