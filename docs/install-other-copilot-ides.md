# Install Bitrise MCP Server in Copilot IDEs

Quick setup guide for the Bitrise MCP server in GitHub Copilot across different IDEs. For VS Code instructions, refer to the [Install Bitrise MCP Server in VS Code](/docs/install-vscode.md)

### Requirements:
1. GitHub Copilot License: Any Copilot plan (Free, Pro, Pro+, Business, Enterprise) for Copilot access
2. Bitrise Account: Bitrise account for Bitrise MCP server access
3. MCP Servers in Copilot Policy: Organizations assigning Copilot seats must enable this policy for all MCP access in Copilot for VS Code and Copilot Coding Agent – all other Copilot IDEs will migrate to this policy in the coming months
4. [Create a Bitrise API Token](https://devcenter.bitrise.io/api/authentication):
   - Go to your [Bitrise Account Settings/Security](https://app.bitrise.io/me/account/security).
   - Navigate to the "Personal access tokens" section.
   - Copy the generated token.
5. For local setup: [Go](https://go.dev/) (>=1.23) installed

## Visual Studio

Requires Visual Studio 2022 version 17.14.9 or later.

### Remote Server (Recommended)

The remote Bitrise MCP server is hosted by Bitrise and provides automatic updates with no local setup required.

#### Configuration
1. Create an `.mcp.json` file in your solution or %USERPROFILE% directory.
2. Add this configuration:
```json
{
  "servers": {
    "bitrise": {
      "url": "https://mcp.bitrise.io",
      "headers": {
        "Authorization": "Bearer YOUR_BITRISE_PAT"
      }
    }
  }
}
```
3. Save the file. Wait for CodeLens to update to offer a way to authenticate to the new server, activate that and pick the Bitrise account to authenticate with.
4. In the GitHub Copilot Chat window, switch to Agent mode.
5. Activate the tool picker in the Chat window and enable one or more tools from the "bitrise" MCP server.

### Local Server (Go required)

For users who prefer to run the Bitrise MCP server locally.

#### Configuration
1. Create an `.mcp.json` file in your solution or %USERPROFILE% directory.
2. Add this configuration:
```json
{
  "servers": {
    "bitrise": {
      "type": "stdio",
      "command": "go",
      "args": ["run", "github.com/bitrise-io/bitrise-mcp/v2@v2"],
      "env": {
        "BITRISE_TOKEN": "YOUR_BITRISE_PAT"
      }
    }
  }
}
```
3. Save the file. Wait for CodeLens to update to offer a way to provide user inputs, activate that and paste in a PAT you generate from your [Bitrise Account Settings/Security](https://app.bitrise.io/me/account/security).
4. In the GitHub Copilot Chat window, switch to Agent mode.
5. Activate the tool picker in the Chat window and enable one or more tools from the "bitrise" MCP server.

**Documentation:** [Visual Studio MCP Guide](https://learn.microsoft.com/visualstudio/ide/mcp-servers)

## JetBrains IDEs

Agent mode and MCP support available in public preview across IntelliJ IDEA, PyCharm, WebStorm, and other JetBrains IDEs.

### Remote Server (Recommended)

The remote Bitrise MCP server is hosted by Bitrise and provides automatic updates with no local setup required.

#### Configuration Steps
1. Install/update the GitHub Copilot plugin
2. Click **GitHub Copilot icon in the status bar** → **Edit Settings** → **Model Context Protocol** → **Configure**
3. Add configuration:
```json
{
  "servers": {
    "bitrise": {
      "url": "https://mcp.bitrise.io",
      "requestInit": {
        "headers": {
          "Authorization": "Bearer YOUR_BITRISE_PAT"
        }
      }
    }
  }
}
```
4. Press `Ctrl + S` or `Command + S` to save, or close the `mcp.json` file. The configuration should take effect immediately and restart all the MCP servers defined. You can restart the IDE if needed.

### Local Server (Go required)

For users who prefer to run the Bitrise MCP server locally.

#### Configuration
```json
{
  "servers": {
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

**Documentation:** [JetBrains Copilot Guide](https://plugins.jetbrains.com/plugin/17718-github-copilot)

## Xcode

Agent mode and MCP support now available in public preview for Xcode.

### Remote Server (Recommended)

The remote Bitrise MCP server is hosted by Bitrise and provides automatic updates with no local setup required.

#### Configuration Steps
1. Install/update [GitHub Copilot for Xcode](https://github.com/github/CopilotForXcode)
2. Open **GitHub Copilot for Xcode app** → **Agent Mode** → **🛠️ Tool Picker** → **Edit Config**
3. Configure your MCP servers:
```json
{
  "servers": {
    "bitrise": {
      "url": "https://mcp.bitrise.io",
      "requestInit": {
        "headers": {
          "Authorization": "Bearer YOUR_BITRISE_PAT"
        }
      }
    }
  }
}
```

### Local Server (Go required)

For users who prefer to run the Bitrise MCP server locally.

#### Configuration
```json
{
  "servers": {
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

**Documentation:** [Xcode Copilot Guide](https://devblogs.microsoft.com/xcode/github-copilot-exploring-agent-mode-and-mcp-support-in-public-preview-for-xcode/)

## Eclipse

MCP support available with Eclipse 2024-03+ and latest version of the GitHub Copilot plugin.

### Remote Server (Recommended)

The remote Bitrise MCP server is hosted by Bitrise and provides automatic updates with no local setup required.

#### Configuration Steps
1. Install GitHub Copilot extension from Eclipse Marketplace
2. Click the **GitHub Copilot icon** → **Edit Preferences** → **MCP** (under **GitHub Copilot**)
3. Add Bitrise MCP server configuration:
```json
{
  "servers": {
    "bitrise": {
      "url": "https://mcp.bitrise.io",
      "requestInit": {
        "headers": {
          "Authorization": "Bearer YOUR_BITRISE_PAT"
        }
      }
    }
  }
}
```
4. Click the "Apply and Close" button in the preference dialog and the configuration will take effect automatically.

### Local Server (Go required)

For users who prefer to run the Bitrise MCP server locally.

#### Configuration
```json
{
  "servers": {
    "bitrise": {
      "command": "go",
      "args": [
        "run",
        "github.com/bitrise-io/bitrise-mcp/v2@v2"
      ],
      "env": {
        "BITRISE_TOKEN": "YOUR_BITRISE_PAT",
        "PATH": "PATH to bin directory of go:PATH to directory of git"
      }
    }
  }
}
```

**Documentation:** [Eclipse Copilot plugin](https://marketplace.eclipse.org/content/github-copilot)

## Usage

After setup:
1. Restart your IDE completely
2. Open Agent mode in Copilot Chat
3. Try: *"List my Bitrise apps"*
4. Copilot can now access Bitrise data and perform operations

## Advanced configuration

See [Tools](/docs/tools.md) for enabling/disabling specific API groups.

## Troubleshooting

- **Connection issues**: Verify IDE version compatibility
- **Authentication errors**: Check if your organization has enabled the MCP policy for Copilot
- **Tools not appearing**: Restart IDE after configuration changes and check error logs
