package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool"
	"github.com/stretchr/testify/assert"
)

// TestDefaultToolsetsMatchToolGroups keeps defaultToolsets in step with the
// toolsets tools are tagged with: every toolset a tool uses is enabled by
// default, and every default toolset enables at least one tool, so naming it
// in ENABLED_TOOLSETS, the toolset query parameter or a header never returns
// an empty tool list.
func TestDefaultToolsetsMatchToolGroups(t *testing.T) {
	defaults := strings.Split(defaultToolsets, ",")

	toolGroups := map[string]bool{}
	for _, tl := range tool.NewBelt().Tools() {
		for _, group := range tl.APIGroups {
			toolGroups[group] = true
			assert.Containsf(t, defaults, group, "%s: toolset %q is missing from defaultToolsets", tl.Definition.Name, group)
		}
	}
	for _, group := range defaults {
		assert.Truef(t, toolGroups[group], "default toolset %q enables no tool", group)
	}
}

func TestConfigToolsets(t *testing.T) {
	assert.Equal(t, defaultToolsets, config{}.toolsets())
	assert.Equal(t, "apps", config{EnabledAPIGroups: "apps"}.toolsets(), "ENABLED_API_GROUPS is the fallback")
	assert.Equal(t, "builds", config{EnabledToolsets: "builds", EnabledAPIGroups: "apps"}.toolsets(), "ENABLED_TOOLSETS wins")
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
	newHeader := "x-bitrise-enabled-toolsets"

	assert.Equal(t, "default", requestToolsets(request("/", nil), "default"))
	assert.Equal(t, "apps", requestToolsets(request("/", map[string]string{oldHeader: "apps"}), "default"))
	assert.Equal(t, "builds", requestToolsets(request("/", map[string]string{oldHeader: "apps", newHeader: "builds"}), "default"))
	assert.Equal(t, "all", requestToolsets(request("/?toolset=all", map[string]string{oldHeader: "apps", newHeader: "builds"}), "default"))
	assert.Equal(t, "builds,rde", requestToolsets(request("/?toolset=builds,rde", nil), "default"))
}
