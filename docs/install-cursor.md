# Install Bitrise MCP Server in Cursor

## Prerequisites

1. [Cursor](https://cursor.com/download) IDE installed (latest version, with MCP OAuth support — Cursor v0.48.0+ for Streamable HTTP, recent builds for OAuth)
2. A Bitrise account
3. For local setup: [Go](https://go.dev/) (>=1.25) installed and a Bitrise PAT

## Remote Server Setup (Recommended)

Recent Cursor versions support MCP OAuth — on first tool use Cursor opens your browser to sign in to Bitrise; no token to paste.

### Install steps

1. Open your global MCP configuration file at `~/.cursor/mcp.json` (or use a project-local `.cursor/mcp.json`) and add the configuration below
2. Save the file
3. Restart Cursor
4. On first tool invocation, complete the browser-based sign-in flow

### Streamable HTTP Configuration

```json
{
  "mcpServers": {
    "bitrise": {
      "url": "https://mcp.bitrise.io"
    }
  }
}
```

### Fallback: PAT-based authentication

If your Cursor version doesn't yet support MCP OAuth, you can use a Personal Access Token. [Create one](https://devcenter.bitrise.io/api/authentication) under [Account Settings → Security](https://app.bitrise.io/me/account/security), then:

```json
{
  "mcpServers": {
    "bitrise": {
      "url": "https://mcp.bitrise.io",
      "headers": {
        "Authorization": "Bearer YOUR_BITRISE_PAT"
      }
    }
  }
}
```

## Local Server Setup

The local Bitrise MCP server runs via Go and uses a Personal Access Token (stdio mode, no OAuth).

[![Install MCP Server](https://cursor.com/deeplink/mcp-install-dark.svg)](https://cursor.com/en-US/install-mcp?name=bitrise&config=eyJlbnYiOnsiQklUUklTRV9UT0tFTiI6IllPVVJfQklUUklTRV9QQVQifSwiY29tbWFuZCI6ImdvIHJ1biBnaXRodWIuY29tL2JpdHJpc2UtaW8vYml0cmlzZS1tY3AvdjJAdjIifQo%3D)

### Install steps

1. Click the install button above and follow the flow, or open `~/.cursor/mcp.json` and add the configuration below
2. In Tools & Integrations > MCP tools, click the pencil icon next to "bitrise"
3. Replace `YOUR_BITRISE_PAT` with your actual Personal Access Token
4. Save the file
5. Restart Cursor

### Local Configuration

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

## Configuration Files

- **Global (all projects)**: `~/.cursor/mcp.json`
- **Project-specific**: `.cursor/mcp.json` in project root

## Verify Installation

1. Restart Cursor completely
2. Check for green dot in Settings → Tools & Integrations → MCP Tools
3. In chat/composer, check "Available Tools"
4. Test with: "List my Bitrise apps" (the first tool call will trigger the OAuth flow if you're not already signed in)

## Advanced configuration

See [Tools](/docs/tools.md) for enabling/disabling specific API groups.

## Troubleshooting

### Remote Server Issues

- **OAuth flow doesn't open browser**: Update Cursor to a recent build. Older builds with Streamable HTTP but without OAuth support can use the PAT-based fallback above.
- **Streamable HTTP not working**: Ensure you're using Cursor v0.48.0 or later
- **Connection errors**: Check firewall/proxy settings

### General Issues

- **MCP not loading**: Restart Cursor completely after configuration
- **Invalid JSON**: Validate that json format is correct
- **Tools not appearing**: Check server shows green dot in MCP settings
- **Check logs**: Look for MCP-related errors in Cursor logs

## Important Notes

- **Cursor specifics**: Supports both project and global configurations, uses `mcpServers` key
