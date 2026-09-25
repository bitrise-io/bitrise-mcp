# Bitrise MCP Server

[![Build Status](https://app.bitrise.io/app/37d1538c-d2d6-4b20-a2cc-78862be22342/status.svg?token=QQWdHofHMTWA2EIkXT4rcA&branch=main)](https://app.bitrise.io/app/37d1538c-d2d6-4b20-a2cc-78862be22342)

MCP Server for Bitrise: mobile CI/CD (apps, builds, pipelines, artifacts), Release Management, Insights and Dev Environments from one server.

## Features

- **Comprehensive API Access**: Access to Bitrise APIs including apps, builds, artifacts, and more.
- **Release Management & Insights**: Connected apps, installable artifacts, tester groups, CodePush, and build/test analytics.
- **Dev Environments (RDE)**: Create and manage remote development sessions, run commands on them, transfer files and drive macOS GUIs (`bitrise_devenv_*` tools, the `dev-environments` API group).
- **OAuth-based Authentication**: Sign in once with your Bitrise account — no need to copy a Personal Access Token. (PAT-based auth still supported for clients that don't speak MCP OAuth.)
- **Detailed Documentation**: [Well-documented tools with parameter descriptions and output schemas](/docs/tools.md).

## Installation

- **[VS Code](/docs/install-vscode.md)** - Installation for VS Code IDE
- **[GitHub Copilot in other IDEs](/docs/install-other-copilot-ides.md)** - Installation for JetBrains, Visual Studio, Eclipse, and Xcode with GitHub Copilot
- **[Claude Applications](/docs/install-claude.md)** - Installation guide for Claude Desktop and Claude Code CLI
- **[Cursor](/docs/install-cursor.md)** - Installation guide for Cursor IDE
- **[Windsurf](/docs/install-windsurf.md)** - Installation guide for Windsurf IDE
- **[Gemini CLI](/docs/install-gemini-cli.md)** - Installation guide for Gemini CLI
- **[AWS Kiro](/docs/install-kiro.md)** - Installation guide for AWS Kiro IDE

## Configuration (local install)

| Variable | Required | Description |
|---|---|---|
| `BITRISE_TOKEN` | Yes | Personal access token. (A Workspace API Token works for the Dev Environments tools only.) |
| `BITRISE_WORKSPACE_ID` | No | Default workspace ID (slug) for the Dev Environments tools. If omitted, pass `workspace_id` on the call, or it is auto-detected when you belong to exactly one workspace (required with a Workspace API Token). |
| `ENABLED_API_GROUPS` | No | Comma-separated API groups to expose. Default: every group, including `dev-environments`. See [Advanced configuration](/docs/tools.md#advanced-configuration). |
| `BITRISE_API_BASE_URL` | No | Bitrise API base URL (default: `https://api.bitrise.io/v0.1`) |
| `BITRISE_DEVENV_API_BASE_URL` | No | Dev Environments backend base URL (default: `https://codespaces-api.services.bitrise.io`) |
| `LOG_LEVEL` | No | `debug`, `info` (default), `warn`, `error` |

On the hosted server the same choices are made per request with the `x-bitrise-enabled-api-groups` and `x-bitrise-workspace-id` headers.

## Example prompts

Once connected, you can ask your AI assistant things like:

- "Show me my last failed iOS build and explain why it failed."
- "What's the success rate of my main branch builds over the last week?"
- "Add a new team member to my workspace as an admin."
- "Start a dev environment from my iOS template and run the unit tests in it."

## Support

For help, questions, or to report an issue, visit the [Bitrise support portal](https://support.bitrise.io/en/).

## Privacy

Use of the Bitrise MCP Server is governed by the [Bitrise Privacy Policy](https://bitrise.io/legal/privacy-policy).
