FROM ghcr.io/astral-sh/uv:python3.12-bookworm-slim

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PYTHONFAULTHANDLER=1 \
    PIP_NO_CACHE_DIR=off \
    PIP_DISABLE_PIP_VERSION_CHECK=on \
    UV_COMPILE_BYTECODE=1 \
    UV_LINK_MODE=copy \
    UV_SYSTEM_PYTHON=true

WORKDIR /app

# Create a non-root user
RUN adduser --disabled-password --gecos "" bitrise

# Copy dependency files first to leverage Docker cache
COPY pyproject.toml uv.lock ./

# Install dependencies before copying application code
# This improves build caching - dependencies change less frequently than code
RUN uv sync --frozen --no-dev

# Make sure the virtual environment is in the PATH
ENV PATH="/app/.venv/bin:$PATH"

# Copy application code
COPY . . 

# Set ownership to the non-root user
RUN chown -R bitrise:bitrise /app

# Switch to non-root user for security
USER bitrise

# Expose the port the app runs on
EXPOSE 8000

# Command to run the API with optimized parameters
CMD ["python3", "main.py"]