# Install Bitrise MCP Server in Cursor

## Prerequisites

1. [Cursor](https://cursor.com/download) IDE installed (latest version)
2. [Create a Bitrise API Token](https://devcenter.bitrise.io/api/authentication):
   - Go to your [Bitrise Account Settings/Security](https://app.bitrise.io/me/account/security).
   - Navigate to the "Personal access tokens" section.
   - Copy the generated token.
3. For local setup: [Go](https://go.dev/) (>=1.23) installed

## Remote Server Setup (Recommended)

[![Install MCP Server](https://cursor.com/deeplink/mcp-install-dark.svg)](https://cursor.com/en-US/install-mcp?name=bitrise&config=eyJ1cmwiOiJodHRwczovL21jcC5iaXRyaXNlLmlvIiwiaGVhZGVycyI6eyJBdXRob3JpemF0aW9uIjoiQmVhcmVyIFlPVVJfQklUUklTRV9QQVQifX0%3D%0A)

Uses Bitrise's hosted server at https://mcp.bitrise.io. Requires Cursor v0.48.0+ for Streamable HTTP support. While Cursor supports OAuth for some MCP servers, the Bitrise server currently requires a Personal Access Token.

### Install steps

1. Click the install button above and follow the flow, or go directly to your global MCP configuration file at `~/.cursor/mcp.json` and enter the code block below
2. In Tools & Integrations > MCP tools, click the pencil icon next to "bitrise"
3. Replace `YOUR_BITRISE_PAT` with your actual [Bitrise Personal Access Token](https://devcenter.bitrise.io/api/authentication)
4. Save the file
5. Restart Cursor

### Streamable HTTP Configuration

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

[![Install MCP Server](https://cursor.com/deeplink/mcp-install-dark.svg)](https://cursor.com/en-US/install-mcp?name=bitrise&config=eyJlbnYiOnsiQklUUklTRV9UT0tFTiI6IllPVVJfQklUUklTRV9QQVQifSwiY29tbWFuZCI6ImdvIHJ1biBnaXRodWIuY29tL2JpdHJpc2UtaW8vYml0cmlzZS1tY3BAbGF0ZXN0In0%3D)

The local Bitrise MCP server runs via Go and requires Go to be installed.

### Install steps

1. Click the install button above and follow the flow, or go directly to your global MCP configuration file at `~/.cursor/mcp.json` and enter the code block below
2. In Tools & Integrations > MCP tools, click the pencil icon next to "bitrise"
3. Replace `YOUR_BITRISE_PAT` with your actual [Bitrise Personal Access Token](https://devcenter.bitrise.io/api/authentication)
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
        "github.com/bitrise-io/bitrise-mcp@v2"
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
4. Test with: "List my Bitrise apps"

## Advanced configuration

See [Tools](tools.md) for enabling/disabling specific API groups.

## Troubleshooting

### Remote Server Issues

- **Streamable HTTP not working**: Ensure you're using Cursor v0.48.0 or later
- **Connection errors**: Check firewall/proxy settings

### General Issues

- **MCP not loading**: Restart Cursor completely after configuration
- **Invalid JSON**: Validate that json format is correct
- **Tools not appearing**: Check server shows green dot in MCP settings
- **Check logs**: Look for MCP-related errors in Cursor logs

## Important Notes

- **Cursor specifics**: Supports both project and global configurations, uses `mcpServers` key
