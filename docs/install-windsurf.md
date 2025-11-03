# Install Bitrise MCP Server in Windsurf

## Prerequisites
1. [Windsurf IDE](https://windsurf.com/) installed (latest version)
2. [Create a Bitrise API Token](https://devcenter.bitrise.io/api/authentication):
   - Go to your [Bitrise Account Settings/Security](https://app.bitrise.io/me/account/security).
   - Navigate to the "Personal access tokens" section.
   - Copy the generated token.
3. For local setup: [Go](https://go.dev/) (>=1.23) installed

## Remote Server Setup (Recommended)

The remote Bitrise MCP server is hosted by Bitrise at `https://mcp.bitrise.io` and supports Streamable HTTP protocol.

### Streamable HTTP Configuration
Windsurf supports Streamable HTTP servers with a `serverUrl` field:

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

## Configuration Details

- **File path**: `~/.codeium/windsurf/mcp_config.json`
- **Scope**: Global configuration only (no per-project support)
- **Format**: Must be valid JSON (use a linter to verify)

## Verification

After installation:
1. Look for "1 available MCP server" in the MCP toolbar
2. Click the hammer icon to see available Bitrise tools
3. Test with: "List my Bitrise apps"
4. Check for green dot next to the server name

## Advanced configuration

See [Tools](tools.md) for enabling/disabling specific API groups.

## Troubleshooting

### Remote Server Issues
- **Authentication failures**: Verify PAT hasn't expired
- **Connection errors**: Check firewall/proxy settings for HTTPS connections
- **Streamable HTTP not working**: Ensure you're using the correct `serverUrl` field format

### General Issues
- **Invalid JSON**: Validate with [jsonlint.com](https://jsonlint.com)
- **Tools not appearing**: Restart Windsurf completely
- **Check logs**: `~/.codeium/windsurf/logs/`

## Important Notes
- **Windsurf limitations**: No environment variable interpolation, global config only
