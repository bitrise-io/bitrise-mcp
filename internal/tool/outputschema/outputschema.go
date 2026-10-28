// Package outputschema holds the generated JSON Schemas of the tools' results
// (see gen/main.go) and applies them to tool definitions.
//
// MCP clients use a tool's outputSchema to understand its result; the OpenAI
// connector directory asks for one on every tool. A tool that builds its
// result in Go declares the schema in its definition with
// mcp.WithOutputSchema; every other tool gets the schema of the API response
// it passes through from schemas/<tool name>.json. Regenerate the files with:
//
//	go run ./internal/tool/outputschema/gen
package outputschema

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"path"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

//go:embed schemas/*.json
var files embed.FS

// Lookup returns the generated schema of the named tool.
func Lookup(toolName string) (json.RawMessage, bool) {
	b, err := files.ReadFile(path.Join("schemas", toolName+".json"))
	if err != nil {
		return nil, false
	}
	return json.RawMessage(b), true
}

// Names lists the tools that have a generated schema.
func Names() []string {
	entries, _ := fs.ReadDir(files, "schemas")
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, strings.TrimSuffix(e.Name(), ".json"))
	}
	return names
}

// Declared reports whether the tool definition already carries an output
// schema (set in code with mcp.WithOutputSchema / mcp.WithRawOutputSchema).
func Declared(t mcp.Tool) bool {
	return t.RawOutputSchema != nil || t.OutputSchema.Type != ""
}

// Apply sets the generated schema on the tool unless it declares its own.
// It reports whether the tool ends up with an output schema.
func Apply(t *mcp.Tool) bool {
	if Declared(*t) {
		return true
	}
	schema, ok := Lookup(t.Name)
	if !ok {
		return false
	}
	t.RawOutputSchema = schema
	return true
}

// StructuredFromText builds the structuredContent of a tool result from the
// text the handler returned, mirroring the envelopes the generated schemas
// describe: a JSON object is used as is, a JSON array is wrapped as
// {"items": [...]}, and any other text (YAML, markdown, a status message, an
// empty body) as {"content": "<text>"}.
func StructuredFromText(text string) any {
	trimmed := strings.TrimSpace(text)
	switch {
	case strings.HasPrefix(trimmed, "{"):
		var obj map[string]any
		if err := json.Unmarshal([]byte(trimmed), &obj); err == nil {
			return obj
		}
	case strings.HasPrefix(trimmed, "["):
		var arr []any
		if err := json.Unmarshal([]byte(trimmed), &arr); err == nil {
			return map[string]any{"items": arr}
		}
	}
	return map[string]any{"content": text}
}

// WithStructuredContent wraps a tool handler so a successful result whose
// only content is text also carries structuredContent (per the MCP spec a
// tool with an outputSchema must return structuredContent). Results that
// already have it, error results, and non-text results are left untouched.
func WithStructuredContent(handler func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := handler(ctx, request)
		if err != nil || result == nil || result.IsError || result.StructuredContent != nil || len(result.Content) != 1 {
			return result, err
		}
		text, ok := result.Content[0].(mcp.TextContent)
		if !ok {
			return result, err
		}
		result.StructuredContent = StructuredFromText(text.Text)
		return result, nil
	}
}
