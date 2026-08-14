# Local MCP testing with Claude CLI

## 1. Start the server

From the repo root:

```bash
ADDR=":8765" go run .
```

## 2. Launch Claude with the local MCP server

```bash
BITRISE_TOKEN=... claude --strict-mcp-config --mcp-config devenv/claude/mcp.json
```

- `--strict-mcp-config` disables all other configured MCP servers for this session.
- `--mcp-config` loads the local server definition without persisting anything to your settings. This is a temporary MCP server definition that points Claude to `localhost:8765` and uses `$BITRISE_TOKEN` as the Bearer token for auth.
