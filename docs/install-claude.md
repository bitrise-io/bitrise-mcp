# Install Bitrise MCP Server in Claude Applications

## Claude Code CLI

### Prerequisites
- Claude Code CLI installed (recent version, with MCP OAuth support)
- A Bitrise account
- Open Claude Code inside the directory for your project (recommended for best experience and clear scope of configuration)

### Remote Server Setup (Streamable HTTP) — Recommended

The remote server uses OAuth to authenticate you against your Bitrise account. No token to copy or paste.

1. Add the server:
   ```bash
   claude mcp add --transport http bitrise https://mcp.bitrise.io
   ```
2. Restart Claude Code.
3. On the first tool use, Claude Code will open your browser to sign in. Log in to Bitrise (or confirm your existing session) and approve the consent screen.
4. Run `claude mcp list` to confirm the server is configured.

#### Fallback: PAT-based authentication

If you're on a Claude Code version without MCP OAuth support, or you'd prefer to use a Personal Access Token:

1. [Create a Bitrise API Token](https://devcenter.bitrise.io/api/authentication) under [Account Settings → Security](https://app.bitrise.io/me/account/security).
2. Add the server with the token as a Bearer header:
   ```bash
   claude mcp add --transport http bitrise https://mcp.bitrise.io -H "Authorization: Bearer YOUR_BITRISE_PAT"
   ```

<details>
<summary><b>Storing Your PAT Securely</b></summary>
<br>

For security, avoid hardcoding your token. One common approach:

1. Store your token in `.env` file
```
BITRISE_PAT=your_token_here
```

2. Add to .gitignore
```bash
echo -e ".env\n.mcp.json" >> .gitignore
```

3. Reference it when adding:
```bash
claude mcp add --transport http bitrise https://mcp.bitrise.io -H "Authorization: Bearer $(grep BITRISE_PAT .env | cut -d '=' -f2)"
```

</details>

### Local Server Setup (Go required)

The local server runs in stdio mode and authenticates with a Personal Access Token (no browser OAuth flow in stdio mode).

Prerequisites: [Go](https://go.dev/) (>=1.25) installed and a Bitrise PAT.

1. Run:
   ```bash
   claude mcp add bitrise -e BITRISE_TOKEN=YOUR_BITRISE_PAT -- go run github.com/bitrise-io/bitrise-mcp/v2@v2
   ```

   With an environment variable:
   ```bash
   claude mcp add bitrise -e BITRISE_TOKEN=$(grep BITRISE_PAT .env | cut -d '=' -f2) -- go run github.com/bitrise-io/bitrise-mcp/v2@v2
   ```
2. Restart Claude Code.
3. Run `claude mcp list` to see if the Bitrise server is configured.

### Verification
```bash
claude mcp list
claude mcp get bitrise
```

## Claude Desktop

### Prerequisites
- Claude Desktop installed (latest version)
- A Bitrise account

### Configuration File Location
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

### Remote Server setup (Streamable HTTP) — Recommended

Recent Claude Desktop versions support MCP OAuth natively. See [Claude | Connecting to a Remote MCP Server](https://modelcontextprotocol.io/docs/develop/connect-remote-servers#connecting-to-a-remote-mcp-server). On first connection, Claude Desktop opens your browser to authenticate with Bitrise — no token needed.

If your Claude Desktop version doesn't yet support remote MCP servers natively, you can use [mcp-remote](https://www.npmjs.com/package/mcp-remote) as an adapter:

```json
{
  "mcpServers": {
    "bitrise": {
      "command": "npx",
      "args": [
        "mcp-remote",
        "https://mcp.bitrise.io"
      ]
    }
  }
}
```

`mcp-remote` will handle the OAuth flow on your behalf. If your version of `mcp-remote` doesn't yet support OAuth, you can fall back to providing a PAT:

```json
{
  "mcpServers": {
    "bitrise": {
      "command": "npx",
      "args": [
        "mcp-remote",
        "https://mcp.bitrise.io",
        "--header",
        "Authorization: Bearer YOUR_BITRISE_PAT"
      ]
    }
  }
}
```

Save the config file and restart Claude Desktop. If everything is set up correctly, you should see a hammer icon next to the message composer.

In case `npx` is not found by Claude (`ENOENT`), specify the path to the `npx` binary in the `env` section:

```json
{
  "mcpServers": {
    "bitrise": {
      ...
      "env": {
        "PATH": "PATH to bin of npx"
      }
    }
  }
}
```

### Local Server Setup (Go required)

The local server uses stdio with a Personal Access Token:

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
        "BITRISE_TOKEN": "YOUR_BITRISE_PAT",
        "PATH": "PATH to bin directory of go:PATH to directory of git",
        "GOPATH": "your GOPATH",
        "GOCACHE": "your GOCACHE"
      }
    }
  }
}
```

### Manual Setup Steps
1. Open Claude Desktop
2. Go to Settings → Developer → Edit Config
3. Paste the code block above in your configuration file
4. If you're navigating to the configuration file outside of the app:
   - **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
5. Open the file in a text editor
6. Paste one of the code blocks above, based on your chosen configuration (remote or local)
7. Save the file
8. Restart Claude Desktop
9. If using OAuth: complete the sign-in flow in your browser the first time you use a tool

### Advanced configuration

See [Tools](/docs/tools.md) for enabling/disabling specific API groups.

## Troubleshooting

**OAuth flow doesn't open browser:**
- Make sure your Claude version supports MCP OAuth (recent builds only)
- Try falling back to PAT-based authentication

**Authentication Failed:**
- For OAuth: re-authenticate via `/mcp` command in Claude Code, or by deleting and re-adding the server
- For PAT: check token hasn't expired or been revoked

**Remote Server:**
- Verify URL: `https://mcp.bitrise.io`

**Server Not Starting / Tools Not Showing:**
- Run `claude mcp list` to view currently configured MCP servers
- Validate JSON syntax
- Restart Claude Code and check `/mcp` command
- Delete the Bitrise server by running `claude mcp remove bitrise` and repeating the setup process with a different method
- Check logs:
  - Claude Code: Use `/mcp` command
  - Claude Desktop: `ls ~/Library/Logs/Claude/` and `cat ~/Library/Logs/Claude/mcp-server-*.log` (macOS) or `%APPDATA%\Claude\logs\` (Windows)

## Important Notes

- Remote server requires Streamable HTTP support (check your Claude version). OAuth requires a more recent build than basic Streamable HTTP.
- Configuration scopes for Claude Code:
  - `-s user`: Available across all projects
  - `-s project`: Shared via `.mcp.json` file
  - Default: `local` (current project only)
