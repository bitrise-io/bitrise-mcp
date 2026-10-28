# Install Bitrise MCP Server in Google Antigravity CLI

Google Antigravity CLI (`agy`) replaces Gemini CLI. If you used the Bitrise MCP server in Gemini CLI, see [Migrating from Gemini CLI](#migrating-from-gemini-cli).

## Prerequisites

1. Google Antigravity CLI installed and signed in (see the [official Antigravity CLI documentation](https://antigravity.google/docs/getting-started))
2. A Bitrise account
3. For local setup: [Go](https://go.dev/) (>=1.25) installed and a Bitrise PAT

## Bitrise MCP Server Configuration

Antigravity CLI reads MCP servers from a dedicated `mcp_config.json` file, under an `mcpServers` key:

- **Global configuration**: `~/.gemini/config/mcp_config.json` where `~` is your home directory
- **Project-specific**: `.agents/mcp_config.json` in your project directory

Create the file if it doesn't exist yet. Remote servers are declared with `serverUrl`: Antigravity CLI doesn't support the `url` and `httpUrl` keys that Gemini CLI used.

### Method 1: Remote Server (Recommended)

The remote Bitrise MCP server is hosted by Bitrise at `https://mcp.bitrise.io` and supports Streamable HTTP. It uses OAuth to authenticate you against your Bitrise account, so there's no token to copy or paste.

1. Add the server to `mcp_config.json`:

    ```json
    // ~/.gemini/config/mcp_config.json
    {
        "mcpServers": {
            "bitrise": {
                "serverUrl": "https://mcp.bitrise.io"
            }
        }
    }
    ```

2. Start Antigravity CLI with `agy`, then open the MCP manager:

    ```
    /mcp
    ```

3. Select `bitrise`, move to the **Authenticate** action, and press Enter.
4. Sign in to Bitrise in your browser (or confirm your existing session) and approve the consent screen.
5. Copy the authorization code that the browser shows, paste it into the terminal prompt, and press Enter.

Antigravity CLI stores the access token and refreshes it automatically when it expires.

#### Fallback: PAT-based authentication

If the OAuth sign-in doesn't complete, or you'd prefer to use a Personal Access Token, [create a Bitrise PAT](https://devcenter.bitrise.io/api/authentication) under [Account Settings → Security](https://app.bitrise.io/me/account/security) and pass it as a header:

```json
// ~/.gemini/config/mcp_config.json
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

Replace `YOUR_BITRISE_PAT` with your token. If you use the project-specific `.agents/mcp_config.json`, add it to `.gitignore` so that your token doesn't end up in your repository.

### Method 2: Local Server Setup (Go Required)

The local server runs in stdio mode and authenticates with a Personal Access Token (no browser OAuth flow in stdio mode):

```json
// ~/.gemini/config/mcp_config.json
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

## Verification

To verify that the Bitrise MCP server has been configured, start Antigravity CLI in your terminal with `agy`, then:

1. **Open the MCP manager**:

    ```
    /mcp
    ```

    The MCP manager lists your configured servers. An authenticated server shows **Authed** next to its name. Select `bitrise` to see its connection status and the list of its tools.

2. **Test with a prompt**:

    ```
    List my Bitrise apps
    ```

    By default, Antigravity CLI asks for your approval before it runs an MCP tool.

## Migrating from Gemini CLI

Antigravity CLI doesn't read the MCP servers declared in Gemini CLI's `~/.gemini/settings.json` or `.gemini/settings.json`.

- **Manual configuration**: Move the `bitrise` entry to `mcp_config.json` and rename its `httpUrl` (or `url`) key to `serverUrl`, as shown in [Method 1](#method-1-remote-server-recommended).
- **Bitrise Gemini extension**: If you installed the extension with `gemini extensions install`, you can convert your Gemini CLI extensions to Antigravity plugins with `agy plugin import gemini`. Afterwards, open `/mcp` and check that `bitrise` is listed and connects. If it doesn't, add the server manually as shown in [Method 1](#method-1-remote-server-recommended).

## Advanced configuration

See [Tools](/docs/tools.md) for enabling/disabling specific API groups.

Each server entry in `mcp_config.json` also accepts Antigravity's own options, such as `disabledTools` to withhold specific tools from the model, and `disabled` to turn the server off without removing it. See the [MCP configuration properties](https://antigravity.google/docs/mcp#mcp-configuration-properties).

To run Bitrise tools without an approval prompt, add `mcp(bitrise/*)` (or `mcp(bitrise/<tool_name>)` for a single tool) to the `permissions.allow` list in `~/.gemini/antigravity-cli/settings.json`. See [Permissions](https://antigravity.google/docs/permissions) for the rule syntax.

## Troubleshooting

### Authentication Issues

- **OAuth sign-in doesn't complete**: Fall back to PAT-based authentication
- **PAT expired or revoked**: Generate a new Bitrise token

### Configuration Issues

- **Server doesn't appear in `/mcp`**: Check that the configuration is in `~/.gemini/config/mcp_config.json` (or `.agents/mcp_config.json` in your project), not in Gemini CLI's `settings.json`. Then reload the server configurations from the `/mcp` manager.
- **Server is listed but doesn't connect**: Check that the entry uses `serverUrl`, not `url` or `httpUrl`
- **JSON syntax errors**: Validate your configuration:
    ```bash
    jq . ~/.gemini/config/mcp_config.json
    ```
- **MCP connection issues**: Open `/mcp`, select `bitrise`, and inspect its connection logs
