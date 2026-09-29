# Bitrise MCP Server

## Tool definitions

- Tools live one-per-file under `internal/tool/<group>/`, each as a package-level `var` of type `bitrise.Tool`, and are registered in `internal/tool/belt.go`.
- Every tool sets all four behavioral annotations (`readOnlyHint`, `destructiveHint`, `idempotentHint`, `openWorldHint`) plus a human-readable `title`.
- `docs/tools.md` is hand-maintained (not generated) — update it manually when adding or changing tools.
- Every tool exposes an output schema (`TestEveryToolHasAnOutputSchema` enforces it). A tool that builds its result in Go declares it with `mcp.WithOutputSchema[T]()`; a tool that passes an API response through gets `internal/tool/outputschema/schemas/<tool>.json`, generated from the API's OpenAPI document by `go run ./internal/tool/outputschema/gen` — add the tool's operation to the `mappings` table there (and the fields the handler strips to `Drop`), then regenerate. The belt wraps such handlers so a JSON text result is also returned as `structuredContent`.
