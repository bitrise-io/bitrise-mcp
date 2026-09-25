# Bitrise MCP Server

## Tool definitions

- Tools live one-per-file under `internal/tool/<group>/`, each as a package-level `var` of type `bitrise.Tool`, and are registered in `internal/tool/belt.go`.
- Every tool sets all four behavioral annotations (`readOnlyHint`, `destructiveHint`, `idempotentHint`, `openWorldHint`) plus a human-readable `title`. The Dev Environments tools set `readOnlyHint`/`destructiveHint` per tool; `idempotentHint`/`openWorldHint` are derived centrally in `devenvironments.NewBelt`. The belt copies the annotation title to the spec-level tool `title`.
- `docs/tools.md` is hand-maintained (not generated) — update it manually when adding or changing tools.
- `main.go` sends server instructions at initialize time describing the two product areas; keep them in sync when tools move between products.
- One Bitrise account covers both product areas: identity and workspace lookup are the shared `me` and `list_workspaces` tools (also in the `dev-environments` groups). The Dev Environments user UUID (`ownerId` on sessions) is an internal identifier and is deliberately not exposed: listings are scoped to the caller and warm pools carry `createdByEmail`. Do not add product-specific identity tools or fields.
- Every tool exposes an output schema (`TestEveryToolHasAnOutputSchema` enforces it). A tool that builds its result in Go declares it with `mcp.WithOutputSchema[T]()`; a tool that passes an API response through gets `internal/tool/outputschema/schemas/<tool>.json`, generated from the API's OpenAPI document by `go run ./internal/tool/outputschema/gen` — add the tool's operation to the `mappings` table there (and the fields the handler strips to `Drop`), then regenerate. The belt wraps such handlers so a JSON text result is also returned as `structuredContent`.

## Dev Environments (RDE) tools

- The `bitrise_devenv_*` tools live in `internal/tool/devenvironments/` (grouped by domain, several tools per file) with their shared client in `internal/devenv/`; they were merged in from `bitrise-mcp-dev-environments` and keep that layout. They all belong to the `dev-environments` API group (read-only ones also to `dev-environments-read-only`, never to the shared `read-only`), assigned centrally in `devenvironments.NewBelt`.
- `devenvironments.Belt.GateAndResolveWorkspace` runs as a tool middleware for these tools only: it rejects the local-only file-transfer tools on the HTTP transport and resolves the workspace (`workspace_id` argument → remembered value → `BITRISE_WORKSPACE_ID` env / `x-bitrise-workspace-id` header → auto-detect via the main API). The PAT is shared with the Bitrise API tools through `bitrise.ContextWithPAT`; the Dev Environments backend takes it as `Bearer <token>`, the main API as the raw token.
- The Dev Environments backend URL is `BITRISE_DEVENV_API_BASE_URL`; `BITRISE_API_BASE_URL` stays the main Bitrise API (also used for workspace discovery).

## Running locally

```bash
# stdio (what local installs use)
BITRISE_TOKEN=<pat> BITRISE_WORKSPACE_ID=<workspace-slug> go run .

# HTTP + OAuth, the hosted setup on localhost (sign in through the browser like the hosted server)
ADDR=127.0.0.1:8000 \
  EXTERNAL_OAUTH_ISSUER=https://oauth.bitrise.io \
  OIDC_TOKEN_ENDPOINT=https://app.bitrise.io/oidc/token \
  SERVER_BASE_URL=http://127.0.0.1:8000 go run .
# then: claude mcp add --transport http bitrise-local http://127.0.0.1:8000
```

The issuer and token-exchange endpoint are the production ones the Bitrise CLI and VS Code extension use; `BITRISE_TOKEN` must be unset in HTTP mode. Per-request options go in headers: `x-bitrise-enabled-api-groups`, `x-bitrise-workspace-id`.
