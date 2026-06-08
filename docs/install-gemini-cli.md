# Install Bitrise MCP Server in Google Gemini CLI

## Prerequisites

1. The latest version of Google Gemini CLI installed (see [official Gemini CLI documentation](https://github.com/google-gemini/gemini-cli))
2. A Bitrise account
3. For local setup: [Go](https://go.dev/) (>=1.25) installed and a Bitrise PAT

## Bitrise MCP Server Configuration

MCP servers for Gemini CLI are configured in its settings JSON under an `mcpServers` key.

- **Global configuration**: `~/.gemini/settings.json` where `~` is your home directory
- **Project-specific**: `.gemini/settings.json` in your project directory

You may need to restart the Gemini CLI for changes to take effect.

### Method 1: Gemini Extension (Recommended)

The simplest way is to use Bitrise's hosted MCP server via our Gemini extension:

```
gemini extensions install https://github.com/bitrise-io/bitrise-mcp
```

The extension handles the OAuth flow automatically on first use — you'll be prompted to sign in to Bitrise in your browser.

### Method 2: Remote Server

You can also connect to the hosted MCP server directly:

```json
// ~/.gemini/settings.json
{
    "mcpServers": {
        "bitrise": {
            "httpUrl": "https://mcp.bitrise.io"
        }
    }
}
```

On first tool use, Gemini CLI will open your browser for the Bitrise OAuth sign-in.

#### Fallback: PAT-based authentication

For Gemini CLI builds that don't yet support MCP OAuth, [create a Bitrise PAT](https://devcenter.bitrise.io/api/authentication) and pass it as a header:

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

<details>
<summary><b>Storing Your PAT Securely</b></summary>
<br>

For security, avoid hardcoding your token. Create or update `~/.gemini/.env` (where `~` is your home or project directory) with your PAT:

```bash
# ~/.gemini/.env
BITRISE_PAT=your_token_here
```

</details>

### Method 3: Local Server Setup (Go Required)

The local server uses stdio with a Personal Access Token:

```json
// ~/.gemini/settings.json
{
    "mcpServers": {
        "bitrise": {
            "command": "go",
            "args": [
                "run",
                "github.com/bitrise-io/bitrise-mcp/v2@v2"
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

    The first tool call will trigger the OAuth sign-in flow if you're using the remote server without a PAT.

## Advanced configuration

See [Tools](/docs/tools.md) for enabling/disabling specific API groups.

You can find more MCP configuration options for Gemini CLI here: [MCP Configuration Structure](https://google-gemini.github.io/gemini-cli/docs/tools/mcp-server.html#configuration-structure). For example, bypassing tool confirmations or excluding specific tools.

## Troubleshooting

### Authentication Issues

- **OAuth flow doesn't open**: Update Gemini CLI to a recent build, or fall back to PAT-based auth
- **Token expired (PAT)**: Generate a new Bitrise token

### Configuration Issues

- **Invalid JSON**: Validate your configuration:
    ```bash
    cat ~/.gemini/settings.json | jq .
    ```
- **MCP connection issues**: Check logs for connection errors:
    ```bash
    gemini --debug "test command"
    ```
