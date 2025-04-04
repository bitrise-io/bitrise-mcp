## Environment Setup
- Python 3.12.6 required (pyenv)
- Use uv for dependency management

## Initialize dependencies in this project
```bash
# Activate python virtual environment
source .venv/bin/activate

# Development server
uv run main.py

# Add new dependencies
uv add "mcp[cli]" httpx

# Sync dependencies
uv sync
```

## Usage in another project

- Use [VSCode insiders](https://code.visualstudio.com/insiders/)
- Set up Github Copilot with your GitHub account
- [The agent mode](https://code.visualstudio.com/blogs/2025/02/24/introducing-copilot-agent-mode) should be available 
- If you have set up MCP tools for Claude desktop, they should already be usable within Copilot
- If you want to use MCP servers separately, [install this extension](https://marketplace.visualstudio.com/items?itemName=SemanticWorkbenchTeam.mcp-server-vscode) and configure it within the project's `.vscode/settings.json`:

```json
    "mcpManager.servers": [
        {
            "name": "bitrise",
            "command": "uv",
            "enabled": true,
            "env": {
                "BITRISE_TOKEN": "<YOUR_PAT>"
            },
            "args": [
                "--directory",
                "<full path of the bitrise-mcp folder>",
                "run",
                "main_new.py"
            ]
        }
    ]
```
- The server should show up in the MCP Server Manager extension, and it should be available as a bunch of tools in the Copilot chat pane (in Agent mode).