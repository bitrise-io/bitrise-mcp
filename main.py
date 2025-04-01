import os
from typing import Any
import httpx
from mcp.server.fastmcp import FastMCP

mcp = FastMCP("bitrise")

BITRISE_API_BASE = "https://api.bitrise.io/v0.1"
USER_AGENT = "bitrise-mcp/1.0"

async def call_api(method, url: str, body = None) -> str:
    headers = {
        "User-Agent": USER_AGENT,
        "Accept": "application/json",
        "Authorization": os.environ.get("BITRISE_TOKEN"),
    }
    async with httpx.AsyncClient() as client:
        response = await client.request(method, url, headers=headers, json=body, timeout=30.0)
        response.raise_for_status()
        return response.text

@mcp.tool()
async def list_builds() -> str:
    """List all the Bitrise builds that can be accessed with the authenticated account."""
    url = f"{BITRISE_API_BASE}/builds"
    return await call_api("GET", url)

@mcp.tool()
async def get_build_log(app_slug: str, build_slug: str) -> str:
    """Get the build log of a specified build of a Bitrise app.

    Args:
        app_slug: Identifier of the Bitrise app (e.g., "d8db74e2675d54c4" or "8eb495d0-f653-4eed-910b-8d6b56cc0ec7")
        build_slug: Identifier of the Bitrise build
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/builds/{build_slug}/log"
    return await call_api("GET", url)

@mcp.tool()
async def get_bitrise_yml(app_slug: str) -> str:
    """Get the current Bitrise YML config file of a specified Bitrise app.

    Args:
        app_slug: Identifier of the Bitrise app (e.g., "d8db74e2675d54c4" or "8eb495d0-f653-4eed-910b-8d6b56cc0ec7")
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/bitrise.yml"
    return await call_api("GET", url)

@mcp.tool()
async def update_bitrise_yml(app_slug: str, bitrise_yml_as_json: dict) -> str:
    """Update the Bitrise YML config file of a specified Bitrise app.

    Args:
        app_slug: Identifier of the Bitrise app (e.g., "d8db74e2675d54c4" or "8eb495d0-f653-4eed-910b-8d6b56cc0ec7")
        bitrise_yml_as_json: The new Bitrise YML config file content to be updated. It must be a JSON string.
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/bitrise.yml"
    return await call_api("POST", url, {
        "app_config_datastore_yaml": bitrise_yml_as_json,
    })

@mcp.tool()
async def trigger_bitrise_build(app_slug: str) -> str:
    """Trigger a new build/pipeline for a specified Bitrise app.

    Args:
        app_slug: Identifier of the Bitrise app (e.g., "d8db74e2675d54c4" or "8eb495d0-f653-4eed-910b-8d6b56cc0ec7")
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/builds"
    return await call_api("POST", url, {
        "build_params": {
            "branch": "main",
        },
        "hook_info": {
            "type": "bitrise"
        },
    })

if __name__ == "__main__":
    mcp.run(transport='stdio')
