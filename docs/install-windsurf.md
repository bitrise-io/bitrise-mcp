# Install Bitrise MCP Server in Windsurf

## Prerequisites
1. [Windsurf IDE](https://windsurf.com/) installed (latest version)
2. A Bitrise account
3. For local setup: [Go](https://go.dev/) (>=1.25) installed and a Bitrise PAT

## Remote Server Setup (Recommended)

The remote Bitrise MCP server is hosted by Bitrise at `https://mcp.bitrise.io` and supports Streamable HTTP. Recent Windsurf builds support MCP OAuth — the first tool use opens your browser to sign in to Bitrise; no token to paste.

### Streamable HTTP Configuration

```json
{
  "mcpServers": {
    "bitrise": {
      "serverUrl": "https://mcp.bitrise.io"
    }
  }
}
```

### Fallback: PAT-based authentication

If your Windsurf build doesn't yet support MCP OAuth, [create a Bitrise PAT](https://devcenter.bitrise.io/api/authentication) under [Account Settings → Security](https://app.bitrise.io/me/account/security) and add it as a header:

```json
{
  "mcpServers": {
    "bitrise": {
      "serverUrl": "https://mcp.bitrise.io",
      "headers": {
        "Authorization": "Bearer YOUR_BITRISE_PAT"
      }
    }
  }
}
```

## Local Server Setup (Go required)

```json
{
  "mcpServers": {
    "bitrise": {
      "command": "go",
      "args": [
        "run",
        "github.com/bitrise-io/bitrise-mcp/v2@v2"
      ],
      "env": {
        "BITRISE_TOKEN": "YOUR_BITRISE_PAT"
      }
    }
  }
}
```

## Installation Steps

### Manual Configuration
1. Click the hammer icon (🔨) in Cascade
2. Click **Configure** to open `~/.codeium/windsurf/mcp_config.json`
3. Add your chosen configuration from above
4. Save the file
5. Click **Refresh** (🔄) in the MCP toolbar
6. On the first tool call, complete the browser-based sign-in flow (OAuth setup only)

## Configuration Details

- **File path**: `~/.codeium/windsurf/mcp_config.json`
- **Scope**: Global configuration only (no per-project support)
- **Format**: Must be valid JSON (use a linter to verify)

## Verification

After installation:
1. Look for "1 available MCP server" in the MCP toolbar
2. Click the hammer icon to see available Bitrise tools
3. Test with: "List my Bitrise apps" (first call triggers OAuth sign-in if using the remote server)
4. Check for green dot next to the server name

## Advanced configuration

See [Tools](/docs/tools.md) for enabling/disabling specific API groups.

## Troubleshooting

### Remote Server Issues
- **OAuth flow doesn't open browser**: Update Windsurf to a recent build. Fall back to PAT-based auth in older versions.
- **Authentication failures (PAT)**: Verify the PAT hasn't expired or been revoked
- **Connection errors**: Check firewall/proxy settings for HTTPS connections
- **Streamable HTTP not working**: Ensure you're using the correct `serverUrl` field format

### General Issues
- **Invalid JSON**: Validate with [jsonlint.com](https://jsonlint.com)
- **Tools not appearing**: Restart Windsurf completely
- **Check logs**: `~/.codeium/windsurf/logs/`

## Important Notes
- **Windsurf limitations**: No environment variable interpolation, global config only
