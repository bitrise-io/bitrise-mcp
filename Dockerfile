FROM python:3.12-slim

# Set a working directory
WORKDIR /app

# Install curl and other dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install uv (install script installs it to /usr/local/bin)
RUN curl -LsSf https://astral.sh/uv/install.sh |env UV_INSTALL_DIR="/usr/local/bin" sh

# Copy project files
COPY . .

# Synchronize dependencies
RUN uv sync

# Command to run your app
ENTRYPOINT ["uv", "run", "main.py"]
