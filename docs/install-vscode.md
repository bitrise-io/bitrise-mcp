# Install Bitrise MCP Server in VS Code

## Prerequisites
- [VS Code](https://code.visualstudio.com/Download) installed (recent version, with MCP OAuth support)
- A Bitrise account
- For local setup: [Go](https://go.dev/) (>=1.25) installed and a Bitrise PAT

## Remote Server Setup (Streamable HTTP) — Recommended

VS Code's MCP integration handles OAuth automatically. On first connection it'll open your browser to sign you in to Bitrise — no token needed.

Follow [VS Code | Add an MCP server](https://code.visualstudio.com/docs/copilot/customization/mcp-servers#_add-an-mcp-server) and add the following to your settings:

```json
{
  "servers": {
    "bitrise": {
      "type": "http",
      "url": "https://mcp.bitrise.io"
    }
  }
}
```

Save the configuration. VS Code will recognize the change, prompt you to sign in via your browser on first tool use, and load the tools into Copilot Chat.

### Fallback: PAT-based authentication

If your VS Code build doesn't support MCP OAuth yet, you can use a Personal Access Token:

```json
{
  "servers": {
    "bitrise": {
      "type": "http",
      "url": "https://mcp.bitrise.io",
      "headers": {
        "Authorization": "Bearer ${input:bitrise-token}"
      }
    }
  },
  "inputs": [
    {
      "id": "bitrise-token",
      "type": "promptString",
      "description": "Bitrise token",
      "password": true
    }
  ]
}
```

[Create a Bitrise API Token](https://devcenter.bitrise.io/api/authentication) under [Account Settings → Security](https://app.bitrise.io/me/account/security) when prompted.

## Local Server Setup (Go required)

Local stdio mode authenticates with a Personal Access Token:

```json
{
  "servers": {
    "bitrise-local": {
      "type": "stdio",
      "command": "go",
      "args": [
        "run",
        "github.com/bitrise-io/bitrise-mcp/v2@v2"
      ],
      "env": {
        "BITRISE_TOKEN": "${input:bitrise-token}"
      }
    }
  },
  "inputs": [
    {
      "id": "bitrise-token",
      "type": "promptString",
      "description": "Bitrise token",
      "password": true
    }
  ]
}
```

## Advanced configuration

VS Code allows at most 128 tools in a chat request, and its built-in tools count toward that limit. The default toolset exposes 88 Bitrise tools and `all` exposes 131, so together with the built-in tools and other MCP servers, chat requests can fail with "Cannot have more than 128 tools per request". GitHub Copilot doesn't load tools dynamically, so we suggest enabling only the toolsets you need with the `toolset` query parameter of the server URL. This example keeps the CI tools:

```json
{
  "servers": {
    "bitrise": {
      "type": "http",
      "url": "https://mcp.bitrise.io/?toolset=apps,builds,workspaces,outgoing-webhooks,artifacts,group-roles,cache-items,pipelines,account,configuration,insights"
    }
  }
}
```

For the local server, put the same comma-separated list in the `ENABLED_TOOLSET` environment variable instead. You can also turn off individual tools in the tool picker of the Chat view.

If a task needs a tool from a toolset you haven't enabled, such as the store release tools, the agent can call `list_toolset` to find out which toolset to add.

See [Toolsets](/docs/tools.md#advanced-configuration) for every toolset and combinations that stay under common tool limits.
