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

func TestEveryAPIGroupHasADescription(t *testing.T) {
	for _, group := range NewBelt().groupNames() {
		assert.NotEmptyf(t, apiGroupDescriptions[group], "API group %q has no description in apiGroupDescriptions", group)
	}
}

func TestListAPIGroupsIsAlwaysEnabled(t *testing.T) {
	belt := NewBelt()
	for _, enabled := range [][]string{nil, {}, {"apps"}, {"no-such-group"}} {
		assert.Truef(t, belt.ToolEnabled("list_api_groups", enabled), "list_api_groups must be enabled with groups %v", enabled)
	}
}

func callListAPIGroups(t *testing.T, ctx context.Context) apiGroupsResult {
	t.Helper()
	var listTool bitrise.Tool
	for _, tl := range NewBelt().Tools() {
		if tl.Definition.Name == "list_api_groups" {
			listTool = tl
		}
	}
	if listTool.Handler == nil {
		t.Fatal("list_api_groups is not registered")
	}
	res, err := listTool.Handler(ctx, mcp.CallToolRequest{})
	mustOK(t, err)
	if res.IsError {
		t.Fatalf("list_api_groups returned an error result: %v", res.Content)
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
	assert.NoError(t, compileSchema(t, "list_api_groups", def.OutputSchema).Validate(instance), "result does not match the output schema")

	var out apiGroupsResult
	mustOK(t, json.Unmarshal(raw, &out))
	return out
}

func TestListAPIGroupsReportsWhatEnablingAGroupAdds(t *testing.T) {
	out := callListAPIGroups(t, bitrise.ContextWithEnabledGroups(context.Background(), []string{"apps", "read-only"}))

	assert.Equal(t, []string{"apps", "read-only"}, out.EnabledGroups)
	assert.Contains(t, out.HowToEnable, "x-bitrise-enabled-api-groups")
	groups := map[string]apiGroup{}
	for _, g := range out.Groups {
		groups[g.Name] = g
	}

	apps := groups["apps"]
	assert.True(t, apps.Enabled)
	assert.Empty(t, apps.UnavailableTools)

	builds := groups["builds"]
	assert.False(t, builds.Enabled)
	assert.Contains(t, builds.UnavailableTools, "trigger_bitrise_build")
	assert.NotContains(t, builds.UnavailableTools, "list_builds", "list_builds is available through read-only")

	for _, g := range out.Groups {
		assert.NotContains(t, g.UnavailableTools, "list_api_groups")
	}
}

func TestListAPIGroupsWithoutGroupsInContextTreatsAllAsEnabled(t *testing.T) {
	for _, g := range callListAPIGroups(t, context.Background()).Groups {
		assert.Truef(t, g.Enabled, "%s should be enabled", g.Name)
		assert.Emptyf(t, g.UnavailableTools, "%s should have no unavailable tools", g.Name)
	}
}

func TestInstructionsPointToListAPIGroups(t *testing.T) {
	for _, s := range []string{"list_api_groups", "x-bitrise-enabled-api-groups", "ENABLED_API_GROUPS"} {
		assert.Contains(t, Instructions, s)
	}
}
