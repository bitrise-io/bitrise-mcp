# Install Bitrise MCP Server in Google Gemini CLI

## Prerequisites

1. The latest version of Google Gemini CLI installed (see [official Gemini CLI documentation](https://github.com/google-gemini/gemini-cli))
2. [Create a Bitrise API Token](https://devcenter.bitrise.io/api/authentication):
   - Go to your [Bitrise Account Settings/Security](https://app.bitrise.io/me/account/security).
   - Navigate to the "Personal access tokens" section.
   - Copy the generated token.
3. For local setup: [Go](https://go.dev/) (>=1.23) installed

<details>
<summary><b>Storing Your PAT Securely</b></summary>
<br>

For security, avoid hardcoding your token. Create or update `~/.gemini/.env` (where `~` is your home or project directory) with your PAT:

```bash
# ~/.gemini/.env
BITRISE_PAT=your_token_here
```

</details>

## Bitrise MCP Server Configuration

MCP servers for Gemini CLI are configured in its settings JSON under an `mcpServers` key.

- **Global configuration**: `~/.gemini/settings.json` where `~` is your home directory
- **Project-specific**: `.gemini/settings.json` in your project directory

After securely storing your PAT, you can add the Bitrise MCP server configuration to your settings file using one of the methods below. You may need to restart the Gemini CLI for changes to take effect.

### Method 1: Gemini Extension (Recommended)

The simplest way is to use Bitrise's hosted MCP server via our gemini extension.

`gemini extensions install https://github.com/bitrise-io/bitrise-mcp`

> [!NOTE]
> You will still need to have a personal access token called `BITRISE_PAT` in your environment.

### Method 2: Remote Server

You can also connect to the hosted MCP server directly. After securely storing your PAT, configure Gemini CLI with:

```json
// ~/.gemini/settings.json
{
    "mcpServers": {
        "bitrise": {
            "httpUrl": "https://mcp.bitrise.io",
            "headers": {
                "Authorization": "Bearer $BITRISE_PAT"
            }
        }
    }
}
```

### Method 3: Local Server Setup (Go Required)

```json
// ~/.gemini/settings.json
{
    "mcpServers": {
        "bitrise": {
            "command": "go",
            "args": [
                "run",
                "github.com/bitrise-io/bitrise-mcp@v2"
            ],
            "env": {
                "BITRISE_TOKEN": "$BITRISE_PAT"
            }
        }
    }
}
```

## Verification

To verify that the Bitrise MCP server has been configured, start Gemini CLI in your terminal with `gemini`, then:

1. **Check MCP server status**:

    ```
    /mcp list
    ```

    ```
    ℹ Configured MCP servers:

    🟢 bitrise - Ready (62 tools)
        - abort_build
        - abort_pipeline
        - add_member_to_group
        ...
    ```

2. **Test with a prompt**
    ```
    List my Bitrise apps
    ```

## Advanced configuration

See [Tools](tools.md) for enabling/disabling specific API groups.

You can find more MCP configuration options for Gemini CLI here: [MCP Configuration Structure](https://google-gemini.github.io/gemini-cli/docs/tools/mcp-server.html#configuration-structure). For example, bypassing tool confirmations or excluding specific tools.

## Troubleshooting

### Authentication Issues

- **Token expired**: Generate a new Bitrise token

### Configuration Issues

- **Invalid JSON**: Validate your configuration:
    ```bash
    cat ~/.gemini/settings.json | jq .
    ```
- **MCP connection issues**: Check logs for connection errors:
    ```bash
    gemini --debug "test command"
    ```
