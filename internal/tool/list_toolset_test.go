package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
)

func TestEveryToolsetHasADescription(t *testing.T) {
	for _, name := range NewBelt().toolsetNames() {
		assert.NotEmptyf(t, toolsetDescriptions[name], "toolset %q has no description in toolsetDescriptions", name)
	}
}

func TestListToolsetIsAlwaysEnabled(t *testing.T) {
	belt := NewBelt()
	for _, enabled := range [][]string{nil, {}, {"apps"}, {"no-such-toolset"}} {
		assert.Truef(t, belt.ToolEnabled("list_toolset", enabled), "list_toolset must be enabled with toolsets %v", enabled)
	}
}

func TestEnabledToolsets(t *testing.T) {
	belt := NewBelt()

	assert.Equal(t, []string{"apps", "builds"}, belt.EnabledToolsets("apps,builds"))
	assert.Equal(t, belt.toolsetNames(), belt.EnabledToolsets(AllToolsets))
	assert.Equal(t, belt.toolsetNames(), belt.EnabledToolsets("apps,all"), "all enables every toolset whatever else is listed")
	assert.NotContains(t, belt.toolsetNames(), AllToolsets, "no tool may be tagged with the reserved all toolset")
}

func callListToolset(t *testing.T, ctx context.Context) toolsetsResult {
	t.Helper()
	var listTool bitrise.Tool
	for _, tl := range NewBelt().Tools() {
		if tl.Definition.Name == "list_toolset" {
			listTool = tl
		}
	}
	if listTool.Handler == nil {
		t.Fatal("list_toolset is not registered")
	}
	res, err := listTool.Handler(ctx, mcp.CallToolRequest{})
	mustOK(t, err)
	if res.IsError {
		t.Fatalf("list_toolset returned an error result: %v", res.Content)
	}
	raw, err := json.Marshal(res.StructuredContent)
	mustOK(t, err)

	wire, err := json.Marshal(listTool.Definition)
	mustOK(t, err)
	var def struct {
		OutputSchema map[string]any `json:"outputSchema"`
	}
	mustOK(t, json.Unmarshal(wire, &def))
	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	mustOK(t, err)
	assert.NoError(t, compileSchema(t, "list_toolset", def.OutputSchema).Validate(instance), "result does not match the output schema")

	var out toolsetsResult
	mustOK(t, json.Unmarshal(raw, &out))
	return out
}

func TestListToolsetReportsWhatEnablingAToolsetAdds(t *testing.T) {
	out := callListToolset(t, bitrise.ContextWithEnabledGroups(context.Background(), []string{"apps", "read-only"}))

	assert.Equal(t, []string{"apps", "read-only"}, out.EnabledToolsets)
	assert.Contains(t, out.HowToEnable, "x-bitrise-enabled-toolset")
	sets := map[string]toolset{}
	for _, s := range out.Toolsets {
		sets[s.Name] = s
	}

	apps := sets["apps"]
	assert.True(t, apps.Enabled)
	assert.Empty(t, apps.UnavailableTools)

	builds := sets["builds"]
	assert.False(t, builds.Enabled)
	assert.Contains(t, builds.UnavailableTools, "trigger_bitrise_build")
	assert.NotContains(t, builds.UnavailableTools, "list_builds", "list_builds is available through read-only")

	for _, s := range out.Toolsets {
		assert.NotContains(t, s.UnavailableTools, "list_toolset")
	}
}

func TestListToolsetWithAllEnabledReportsNothingUnavailable(t *testing.T) {
	belt := NewBelt()
	out := callListToolset(t, bitrise.ContextWithEnabledGroups(context.Background(), belt.EnabledToolsets(AllToolsets)))
	for _, s := range out.Toolsets {
		assert.Truef(t, s.Enabled, "%s should be enabled", s.Name)
		assert.Emptyf(t, s.UnavailableTools, "%s should have no unavailable tools", s.Name)
	}
}

func TestListToolsetWithoutToolsetsInContextTreatsAllAsEnabled(t *testing.T) {
	for _, s := range callListToolset(t, context.Background()).Toolsets {
		assert.Truef(t, s.Enabled, "%s should be enabled", s.Name)
		assert.Emptyf(t, s.UnavailableTools, "%s should have no unavailable tools", s.Name)
	}
}

func TestInstructionsPointToListToolset(t *testing.T) {
	for _, s := range []string{"list_toolset", "toolset query parameter", "x-bitrise-enabled-toolset", "ENABLED_TOOLSET"} {
		assert.Contains(t, Instructions, s)
	}
}
