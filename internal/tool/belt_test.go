package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/outputschema"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
)

// noOutputSchema lists the tools whose result is not structured data (an
// image), so they carry no output schema. Every current tool returns
// structured data or text.
var noOutputSchema = map[string]bool{}

// TestEveryToolHasAnOutputSchema enforces the connector directory requirement
// (OpenAI asks for an outputSchema on every tool): each tool either declares
// one in code or has a generated schemas/<name>.json, and the schema is a
// valid JSON Schema describing an object.
func TestEveryToolHasAnOutputSchema(t *testing.T) {
	for _, tl := range NewBelt().Tools() {
		name := tl.Definition.Name
		if noOutputSchema[name] {
			assert.Falsef(t, outputschema.Declared(tl.Definition), "%s is listed as having no output schema but declares one", name)
			continue
		}
		if !assert.Truef(t, outputschema.Declared(tl.Definition), "tool %q has no output schema: declare one with mcp.WithOutputSchema or map it in internal/tool/outputschema/gen", name) {
			continue
		}
		raw, err := json.Marshal(tl.Definition)
		mustOK(t, err)
		var wire struct {
			OutputSchema map[string]any `json:"outputSchema"`
		}
		mustOK(t, json.Unmarshal(raw, &wire))
		if !assert.NotNilf(t, wire.OutputSchema, "%s: outputSchema missing from the wire format", name) {
			continue
		}
		assert.Equalf(t, "object", wire.OutputSchema["type"], "%s: output schema must describe an object", name)
		compileSchema(t, name, wire.OutputSchema)
	}
}

// openWorld lists the tools whose effect reaches beyond the authenticated
// Bitrise account, following the connector directory definition of
// openWorldHint (OpenAI's app review FAQ): true when a tool accesses the
// public internet or open-ended external entities (sends messages to external
// recipients, publishes content, writes to another service), false when it is
// limited to the bounded Bitrise workspace, even though Bitrise is externally
// hosted. Every other tool only reads or writes the user's own Bitrise data.
var openWorld = map[string]string{
	"register_webhook":                             "registers a webhook at the app's git provider",
	"register_ssh_key":                             "can register the key at the app's git provider",
	"invite_member_to_workspace":                   "emails an invitation to an arbitrary address",
	"create_outgoing_webhook":                      "makes Bitrise send build events to an arbitrary URL",
	"update_outgoing_webhook":                      "makes Bitrise send build events to an arbitrary URL",
	"set_installable_artifact_public_install_page": "publishes an install page on the public internet",
	"codepush_generate_update_upload_url":          "creates an update that end-user devices download once uploaded",
	"codepush_promote_deployment":                  "releases a package to the end-user devices of the target deployment",
	"codepush_rollback_deployment":                 "changes the package served to end-user devices",
	"codepush_patch_update":                        "changes whether and to how many end-user devices an update is served",
	"codepush_delete_update":                       "withdraws an update served to end-user devices",
}

// TestOpenWorldHintIsDeliberate keeps openWorldHint an explicit decision per
// tool: it must be set, and it is true exactly for the tools listed in
// openWorld.
func TestOpenWorldHintIsDeliberate(t *testing.T) {
	registered := map[string]bool{}
	for _, tl := range NewBelt().Tools() {
		name := tl.Definition.Name
		registered[name] = true
		hint := tl.Definition.Annotations.OpenWorldHint
		if !assert.NotNilf(t, hint, "%s: openWorldHint must be set explicitly", name) {
			continue
		}
		_, open := openWorld[name]
		assert.Equalf(t, open, *hint, "%s: openWorldHint is %v but the tool is%s listed in openWorld (belt_test.go)", name, *hint, map[bool]string{true: "", false: " not"}[open])
	}
	for name := range openWorld {
		assert.Truef(t, registered[name], "openWorld lists %s, which is not a registered tool", name)
	}
}

// TestGeneratedSchemasMatchTools guards against stale generated files: every
// schemas/<name>.json belongs to a registered tool.
func TestGeneratedSchemasMatchTools(t *testing.T) {
	registered := map[string]bool{}
	for _, tl := range NewBelt().Tools() {
		registered[tl.Definition.Name] = true
	}
	for _, name := range outputschema.Names() {
		assert.Truef(t, registered[name], "schemas/%s.json does not belong to a registered tool", name)
	}
}

func compileSchema(t *testing.T, name string, schema map[string]any) *jsonschema.Schema {
	t.Helper()
	b, err := json.Marshal(schema)
	mustOK(t, err)
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(b))
	mustOK(t, err)
	c := jsonschema.NewCompiler()
	mustOK(t, c.AddResource(name+".json", doc))
	sch, err := c.Compile(name + ".json")
	if !assert.NoErrorf(t, err, "%s: output schema does not compile", name) {
		t.FailNow()
	}
	return sch
}

// TestStructuredContentEnvelopes checks the belt's structuredContent
// envelopes validate against the schemas the generator emits for them.
func TestStructuredContentEnvelopes(t *testing.T) {
	textEnvelope := map[string]any{
		"type":       "object",
		"properties": map[string]any{"content": map[string]any{"type": "string"}},
	}
	sch := compileSchema(t, "text", textEnvelope)
	assert.NoError(t, sch.Validate(outputschema.StructuredFromText("a: b\n")))
	assert.NoError(t, sch.Validate(outputschema.StructuredFromText("")))

	obj := outputschema.StructuredFromText(`{"data": {"slug": "x"}}`)
	assert.Equal(t, map[string]any{"data": map[string]any{"slug": "x"}}, obj)

	arr := outputschema.StructuredFromText(`[1, 2]`)
	assert.Equal(t, map[string]any{"items": []any{1.0, 2.0}}, arr)
}

func TestWithStructuredContent(t *testing.T) {
	call := func(res *mcp.CallToolResult) *mcp.CallToolResult {
		h := outputschema.WithStructuredContent(func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return res, nil
		})
		out, err := h(context.Background(), mcp.CallToolRequest{})
		mustOK(t, err)
		return out
	}

	t.Run("text JSON object becomes structuredContent and keeps the text", func(t *testing.T) {
		out := call(mcp.NewToolResultText(`{"a": 1}`))
		assert.Equal(t, map[string]any{"a": 1.0}, out.StructuredContent)
		assert.Len(t, out.Content, 1)
	})
	t.Run("error results are untouched", func(t *testing.T) {
		out := call(mcp.NewToolResultError("boom"))
		assert.Nil(t, out.StructuredContent)
	})
	t.Run("existing structuredContent is kept", func(t *testing.T) {
		out := call(mcp.NewToolResultStructured(map[string]any{"k": "v"}, "fallback"))
		assert.Equal(t, map[string]any{"k": "v"}, out.StructuredContent)
	})
	t.Run("image results are untouched", func(t *testing.T) {
		res := &mcp.CallToolResult{Content: []mcp.Content{mcp.NewImageContent("AAAA", "image/png")}}
		assert.Nil(t, call(res).StructuredContent)
	})
}

func mustOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
