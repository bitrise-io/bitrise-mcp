# Install Bitrise MCP Server in Claude Applications

## Claude Code CLI

### Prerequisites
- Claude Code CLI installed
- [Create a Bitrise API Token](https://devcenter.bitrise.io/api/authentication):
   - Go to your [Bitrise Account Settings/Security](https://app.bitrise.io/me/account/security).
   - Navigate to the "Personal access tokens" section.
   - Copy the generated token.
- For local setup: [Go](https://go.dev/) (>=1.25) installed
- Open Claude Code inside the directory for your project (recommended for best experience and clear scope of configuration)

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

</details>

### Remote Server Setup (Streamable HTTP)

1. Run the following command in the Claude Code CLI
```bash
claude mcp add --transport http bitrise https://mcp.bitrise.io -H "Authorization: Bearer YOUR_BITRISE_PAT"
```

With an environment variable:
```bash
claude mcp add --transport http bitrise https://mcp.bitrise.io -H "Authorization: Bearer $(grep BITRISE_PAT .env | cut -d '=' -f2)"
```
2. Restart Claude Code
3. Run `claude mcp list` to see if the Bitrise server is configured

### Local Server Setup (Go required)

1. Run the following command in the Claude Code CLI:
```bash
claude mcp add bitrise -e BITRISE_TOKEN=YOUR_BITRISE_PAT -- go run github.com/bitrise-io/bitrise-mcp/v2@v2
```

With an environment variable:
```bash
claude mcp add bitrise -e BITRISE_TOKEN=$(grep BITRISE_PAT .env | cut -d '=' -f2) -- go run github.com/bitrise-io/bitrise-mcp/v2@v2
```
2. Restart Claude Code
3. Run `claude mcp list` to see if the Bitrise server is configured

### Verification
```bash
claude mcp list
claude mcp get bitrise
```

## Claude Desktop

### Prerequisites
- Claude Desktop installed (latest version)
- [Create a Bitrise API Token](https://devcenter.bitrise.io/api/authentication):
   - Go to your [Bitrise Account Settings/Security](https://app.bitrise.io/me/account/security).
   - Navigate to the "Personal access tokens" section.
   - Copy the generated token.
- For local setup: [Go](https://go.dev/) (>=1.25) installed

### Configuration File Location
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

### Remote Server setup (Streamable HTTP)

See [Claude | Connecting to a Remote MCP Server
](https://modelcontextprotocol.io/docs/develop/connect-remote-servers#connecting-to-a-remote-mcp-server) for more details.

In case this feature is not available in your Claude Desktop version, you can use [mcp-remote](https://www.npmjs.com/package/mcp-remote) as an adapter. Add this codeblock to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "bitrise": {
      "command": "npx",
      "args": [
        "mcp-remote",
        "https://mcp.bitrise.io",
        "--header",
        "Authorization: YOUR_BITRISE_PAT"
      ]
    }
  }
}
```

Save the config file and restart Claude Desktop. If everything is set up correctly, you should see a hammer icon next to the message composer.

In case `npx` is not found by Claude (`ENOENT`), you can specify the path to the `npx` binary in the `env` section of the configuration like this:

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

Add this codeblock to your `claude_desktop_config.json`:

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
7. Replace `YOUR_BITRISE_PAT` with your actual token or $BITRISE_PAT environment variable
8. Save the file
9. Restart Claude Desktop

### Advanced configuration

See [Tools](/docs/tools.md) for enabling/disabling specific API groups.

## Troubleshooting

**Authentication Failed:**
- Check token hasn't expired

**Remote Server:**
- Verify URL: `https://mcp.bitrise.io`

**Server Not Starting / Tools Not Showing:**
- Run `claude mcp list` to view currently configured MCP servers
- Validate JSON syntax
- If using an environment variable to store your PAT, make sure you're properly sourcing your PAT using the environment variable
- Restart Claude Code and check `/mcp` command
- Delete the Bitrise server by running `claude mcp remove bitrise` and repeating the setup process with a different method
- Make sure you're running Claude Code within the project you're currently working on to ensure the MCP configuration is properly scoped to your project
- Check logs:
  - Claude Code: Use `/mcp` command
  - Claude Desktop: `ls ~/Library/Logs/Claude/` and `cat ~/Library/Logs/Claude/mcp-server-*.log` (macOS) or `%APPDATA%\Claude\logs\` (Windows)

## Important Notes

- Remote server requires Streamable HTTP support (check your Claude version)
- Configuration scopes for Claude Code:
  - `-s user`: Available across all projects
  - `-s project`: Shared via `.mcp.json` file
  - Default: `local` (current project only)
