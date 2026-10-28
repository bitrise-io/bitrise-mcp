package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool"
	"github.com/stretchr/testify/assert"
)

// TestDefaultEnabledAPIGroupsMatchToolGroups keeps the default
// ENABLED_API_GROUPS in step with the groups tools are tagged with: every
// group a tool uses is enabled by default, and every default group enables at
// least one tool, so naming it in ENABLED_API_GROUPS or the
// x-bitrise-enabled-api-groups header never returns an empty tool list.
func TestDefaultEnabledAPIGroupsMatchToolGroups(t *testing.T) {
	field, ok := reflect.TypeOf(config{}).FieldByName("EnabledAPIGroups")
	if !ok {
		t.Fatal("config has no EnabledAPIGroups field")
	}
	defaults := strings.Split(field.Tag.Get("default"), ",")

	toolGroups := map[string]bool{}
	for _, tl := range tool.NewBelt().Tools() {
		for _, group := range tl.APIGroups {
			toolGroups[group] = true
			assert.Containsf(t, defaults, group, "%s: group %q is missing from the default ENABLED_API_GROUPS", tl.Definition.Name, group)
		}
	}
	for _, group := range defaults {
		assert.Truef(t, toolGroups[group], "default group %q enables no tool", group)
	}
}
