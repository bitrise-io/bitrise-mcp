# Bitrise MCP Server

## Tool definitions

- Tools live one-per-file under `internal/tool/<group>/`, each as a package-level `var` of type `bitrise.Tool`, and are registered in `internal/tool/belt.go`.
- Every tool sets all four behavioral annotations (`readOnlyHint`, `destructiveHint`, `idempotentHint`, `openWorldHint`) plus a human-readable `title`.
- `docs/tools.md` is hand-maintained (not generated) — update it manually when adding or changing tools.
