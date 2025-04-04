# Bitrise MCP Project Guidelines

## Environment Setup
- Python 3.12.6 required (pyenv)
- Use uv for dependency management

## Development Commands
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

# Devcontainer support
For editing the code you can use the devcontainer setup, which installs the dependencies and configures vscode for you automatically.
However you still need to setup your host to be able to run uv, as setting up the claude code desktop with the devcontainer instance is not supported.

## Code Style Guidelines
- **Imports**: Group standard lib, third-party, and local imports
- **Formatting**: Follow PEP 8, max line length 100
- **Types**: Use Python type hints for all function parameters and returns
- **Error Handling**: Use try-except blocks with specific exceptions
- **Async Patterns**: Use async/await with proper error handling
- **Naming**: snake_case for functions/variables, PascalCase for classes
- **Documentation**: Docstrings for all functions with Args and Returns sections
- **Environment Variables**: Access via os.environ.get with default values