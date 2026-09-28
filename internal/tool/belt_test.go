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
