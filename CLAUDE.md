# Bitrise MCP Project Guidelines

## Environment Setup
- Python 3.12.6 required (pyenv)
- Use uv for dependency management

## Development Commands
```bash
# Development server
python main.py

# Install dependencies 
uv pip install -e .

# Linting (ruff recommended)
ruff check .

# Type checking
mypy --python-version 3.12 .

# Testing (pytest recommended)
pytest
pytest tests/test_specific.py -v  # Run single test
```

## Code Style Guidelines
- **Imports**: Group standard lib, third-party, and local imports
- **Formatting**: Follow PEP 8, max line length 100
- **Types**: Use Python type hints for all function parameters and returns
- **Error Handling**: Use try-except blocks with specific exceptions
- **Async Patterns**: Use async/await with proper error handling
- **Naming**: snake_case for functions/variables, PascalCase for classes
- **Documentation**: Docstrings for all functions with Args and Returns sections
- **Environment Variables**: Access via os.environ.get with default values