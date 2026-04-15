# bitrise-mcp

## Repository overview

Go MCP server that wraps the Bitrise API and exposes tools to AI agents via the MCP protocol.

## Key paths

| What | Where |
|---|---|
| Tool implementations | `internal/tool/<feature>/` — one `.go` file per operation |
| Release Management tools | `internal/tool/releasemanagement/` — includes `codepush/` subfolder |
| Tool registration | `internal/tool/belt.go` — add to `NewBelt()` and import the package |
| API client + base URLs | `internal/bitrise/call_api.go` |
| Tool documentation | `docs/tools.md` |

## Base URLs

```go
APIBaseURL         = "https://api.bitrise.io/v0.1"
APIRMBaseURL       = "https://api.bitrise.io/release-management/v1"
APICodePushBaseURL = "https://api.bitrise.io/release-management/v2/code-push/v1"
```

## Adding a new tool — checklist

1. Create `internal/tool/<feature>/<operation>.go` following the pattern below
2. Add a base URL constant to `internal/bitrise/call_api.go` if needed
3. Register the tool var in `internal/tool/belt.go` under `NewBelt()`
4. Document in `docs/tools.md` — new section + API groups table column (if new group)

## Package structure note

CodePush is part of Release Management and lives at `internal/tool/releasemanagement/codepush/`.
Future Release Management sub-features should follow the same subfolder pattern.

## Tool file pattern

```go
package <feature>

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var MyTool = bitrise.Tool{
	APIGroups: []string{"my-group", "read-only"}, // omit "read-only" for write ops
	Definition: mcp.NewTool("my_tool_name",
		mcp.WithDescription("What this tool does."),
		mcp.WithString("param", mcp.Description("..."), mcp.Required()),
		mcp.WithReadOnlyHintAnnotation(true),      // true for GET only
		mcp.WithDestructiveHintAnnotation(false),  // true for DELETE and irreversible writes
		mcp.WithOpenWorldHintAnnotation(true),     // always true for API calls
		mcp.WithIdempotentHintAnnotation(true),    // true for GET/PUT; false for POST/trigger ops
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		param, err := request.RequireString("param")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/some/path/%s", param),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
```

## Parameter helpers

- Required string: `request.RequireString("name")`
- Optional string: `request.GetString("name", "")`
- Optional bool: `request.GetBool("name", false)`
- Optional int: `request.GetInt("name", defaultValue)`
- Query params: `CallAPIParams{Params: map[string]any{"key": value}}`
- Request body: `CallAPIParams{Body: map[string]any{"key": value}}`
- Booleans in query params: pass as string `"true"` (not bool)
- Integers in query params: use `strconv.Itoa(v)`

## Boolean sentinel pattern for optional booleans in PATCH bodies

`GetBool` cannot distinguish "not provided" from `false`. When a caller must be able to
explicitly set a boolean to `false`, use `mcp.WithString` and check for empty string:

```go
mcp.WithString("disabled",
    mcp.Description("Set to 'true' to disable or 'false' to re-enable. Omit to leave unchanged."),
)
// In handler:
if v := request.GetString("disabled", ""); v != "" {
    body["disabled"] = v == "true"
}
```

## Rollout sentinel pattern

`rollout=0` (0% rollout) is a valid value — never use `0` as a default sentinel.
Use `-1` as sentinel and guard with `v >= 0`:

```go
if v := request.GetInt("rollout", -1); v >= 0 {
    body["rollout"] = v
}
```

## Verify after changes

```bash
go build ./...
go test ./...
make lint
```

## API groups

Valid values for `ENABLED_API_GROUPS`:
`apps, builds, workspaces, outgoing-webhooks, artifacts, group-roles, cache-items, pipelines, account, read-only, release-management, configuration, code-push`
