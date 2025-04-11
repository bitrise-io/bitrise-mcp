import argparse
import os
import httpx
import sys
from functools import partial
from typing import Any, Dict, List, Optional, Union
from mcp.server.fastmcp import FastMCP
from pydantic import Field

mcp = FastMCP("bitrise")


BITRISE_API_BASE = "https://api.bitrise.io/v0.1"
USER_AGENT = "bitrise-mcp/1.0"


parser = argparse.ArgumentParser()
parser.add_argument(
    "--enabled-api-groups",
    help="The list of enabled API groups, comma separated",
    type=partial(str.split, sep=","),
    default="apps,builds,workspaces,webhooks,build-artifacts,group-roles,cache-items,pipelines,account,read-only",
)
args = parser.parse_args()
print(f"Enabled API groups {args.enabled_api_groups}", file=sys.stderr)


def mcp_tool(
    api_groups: List[str] = [],
    name: str | None = None,
    description: str | None = None,
):
    def decorator(fn):
        if set(api_groups) & set(args.enabled_api_groups):
            mcp.add_tool(fn, name=name, description=description)
        return fn

    return decorator


async def call_api(method, url: str, body=None) -> str:
    headers = {
        "User-Agent": USER_AGENT,
        "Accept": "application/json",
        "Content-Type": "application/json",
        "Authorization": os.environ.get("BITRISE_TOKEN") or "",
    }
    async with httpx.AsyncClient() as client:
        response = await client.request(
            method, url, headers=headers, json=body, timeout=30.0
        )
        return response.text


# ===== Apps =====


@mcp_tool(
    api_groups=["apps", "read-only"],
    description="List all the apps available for the authenticated account.",
)
async def list_apps(
    sort_by: str = Field(
        default="last_build_at",
        description="Order of the apps: last_build_at (default) or created_at. If set, you should accept the response as sorted.",
    ),
    next: str = Field(
        default=None,
        description="Slug of the first app in the response",
    ),
    limit: int = Field(
        default=50,
        description="Max number of elements per page (default: 50)",
    ),
) -> str:
    params: Dict[str, Union[str, int]] = {}
    if sort_by:
        params["sort_by"] = sort_by
    if next:
        params["next"] = next
    if limit:
        params["limit"] = limit

    url = f"{BITRISE_API_BASE}/apps"
    async with httpx.AsyncClient() as client:
        headers = {
            "User-Agent": USER_AGENT,
            "Accept": "application/json",
            "Authorization": os.environ.get("BITRISE_TOKEN") or "",
        }
        response = await client.get(url, headers=headers, params=params, timeout=30.0)
        return response.text


@mcp_tool(
    api_groups=["apps"],
    description="Add a new app to Bitrise. After this app should be finished on order to be registered completely on Bitrise (via the finish_bitrise_app tool). "
    "Before doing this step, try understanding the repository details from the repository URL. "
    "This is a two-step process. First, you register the app with the Bitrise API, and then you finish the setup. "
    "The first step creates a new app in Bitrise, and the second step configures it with the necessary settings. "
    "If the user has multiple workspaces, always prompt the user to choose which one you should use. "
    "Don't prompt the user for finishing the app, just do it automatically.",
)
async def register_app(
    repo_url: str = Field(
        description="Repository URL",
    ),
    is_public: bool = Field(
        description='Whether the app\'s builds visibility is "public"',
    ),
    organization_slug: str = Field(
        description="The organization (aka workspace) the app to add to",
    ),
    project_type: str = Field(
        default="other",
        description="Type of project (ios, android, etc.)",
    ),
    provider: str = Field(
        default="github",
        description="Repository provider",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/register"
    body = {
        "repo_url": repo_url,
        "is_public": is_public,
        "organization_slug": organization_slug,
        "project_type": project_type,
        "provider": provider,
    }
    return await call_api("POST", url, body)


@mcp_tool(
    api_groups=["apps"],
    description="Finish the setup of a Bitrise app. If this is successful, a build can be triggered via trigger_bitrise_build. "
    "If you have access to the repository, decide the project type, the stack ID, and the config to use, based on https://stacks.bitrise.io/, "
    "and the config should be also based on the projec type.",
)
async def finish_bitrise_app(
    app_slug: str = Field(
        description="The slug of the Bitrise app to finish setup for.",
    ),
    project_type: str = Field(
        default="other",
        description="The type of project (e.g., android, ios, flutter, etc.).",
    ),
    stack_id: str = Field(
        default="linux-docker-android-22.04",
        description="The stack ID to use for the app.",
    ),
    mode: str = Field(
        default="manual",
        description="The mode of setup.",
    ),
    config: str = Field(
        default="other-config",
        description='The configuration to use for the app (default is "default-android-config", other valid values are "other-config", "default-ios-config", "default-macos-config", etc).',
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/finish"
    payload = {
        "project_type": project_type,
        "stack_id": stack_id,
        "mode": mode,
        "config": config,
    }
    return await call_api("POST", url, payload)


@mcp_tool(
    api_groups=["apps", "read-only"],
    description="Get the details of a specific app.",
)
async def get_app(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}"
    return await call_api("GET", url)


@mcp_tool(
    api_groups=["apps"],
    description="Delete an app from Bitrise. When deleting apps belonging to multiple workspaces always confirm that which workspaces' apps the user wants to delete.",
)
async def delete_app(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}"
    return await call_api("DELETE", url)


@mcp_tool(
    api_groups=["apps"],
    description="Update an app.",
)
async def update_app(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
    is_public: bool = Field(
        description='Whether the app\'s builds visibility is "public"',
    ),
    project_type: str = Field(
        description="Type of project",
    ),
    provider: str = Field(
        description="Repository provider",
    ),
    repo_url: str = Field(
        description="Repository URL",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}"
    body = {
        "is_public": is_public,
        "project_type": project_type,
        "provider": provider,
        "repo_url": repo_url,
    }
    return await call_api("PATCH", url, body)


@mcp_tool(
    api_groups=["apps", "read-only"],
    description="Get the current Bitrise YML config file of a specified Bitrise app.",
)
async def get_bitrise_yml(
    app_slug: str = Field(
        description='Identifier of the Bitrise app (e.g., "d8db74e2675d54c4" or "8eb495d0-f653-4eed-910b-8d6b56cc0ec7")',
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/bitrise.yml"
    return await call_api("GET", url)


@mcp_tool(
    api_groups=["apps"],
    description="Update the Bitrise YML config file of a specified Bitrise app.",
)
async def update_bitrise_yml(
    app_slug: str = Field(
        description='Identifier of the Bitrise app (e.g., "d8db74e2675d54c4" or "8eb495d0-f653-4eed-910b-8d6b56cc0ec7")',
    ),
    bitrise_yml_as_json: str = Field(
        description="The new Bitrise YML config file content to be updated. It must be a string.",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/bitrise.yml"
    return await call_api(
        "POST",
        url,
        {
            "app_config_datastore_yaml": bitrise_yml_as_json,
        },
    )


@mcp_tool(
    api_groups=["apps", "read-only"],
    description="List the branches with existing builds of an app's repository.",
)
async def list_branches(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/branches"
    return await call_api("GET", url)


@mcp_tool(
    api_groups=["apps"],
    description="Add an SSH-key to a specific app.",
)
async def register_ssh_key(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
    auth_ssh_private_key: str = Field(
        description="Private SSH key",
    ),
    auth_ssh_public_key: str = Field(
        description="Public SSH key",
    ),
    is_register_key_into_provider_service: bool = Field(
        description="Register the key in the provider service",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/register-ssh-key"
    body = {
        "auth_ssh_private_key": auth_ssh_private_key,
        "auth_ssh_public_key": auth_ssh_public_key,
        "is_register_key_into_provider_service": is_register_key_into_provider_service,
    }
    return await call_api("POST", url, body)


@mcp_tool(
    api_groups=["apps"],
    description="Register an incoming webhook for a specific application.",
)
async def register_webhook(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/register-webhook"
    return await call_api("POST", url)


# ===== Builds =====


@mcp_tool(
    api_groups=["builds", "read-only"],
    description="List all the builds of a specified Bitrise app or all accessible builds.",
)
async def list_builds(
    app_slug: str = Field(
        default=None,
        description="Identifier of the Bitrise app",
    ),
    sort_by: str = Field(
        default="created_at",
        description="Order of builds: created_at (default), running_first",
    ),
    branch: str = Field(
        default=None,
        description="Filter builds by branch",
    ),
    workflow: str = Field(
        default=None,
        description="Filter builds by workflow",
    ),
    status: int = Field(
        default=None,
        description="Filter builds by status (0: not finished, 1: successful, 2: failed, 3: aborted, 4: in-progress)",
    ),
    next: str = Field(
        default=None,
        description="Slug of the first build in the response",
    ),
    limit: int = Field(
        default=None,
        description="Max number of elements per page (default: 50)",
    ),
) -> str:
    params: Dict[str, Union[str, int]] = {}
    if sort_by:
        params["sort_by"] = sort_by
    if branch:
        params["branch"] = branch
    if workflow:
        params["workflow"] = workflow
    if status is not None:
        params["status"] = status
    if next:
        params["next"] = next
    if limit:
        params["limit"] = limit

    if app_slug:
        url = f"{BITRISE_API_BASE}/apps/{app_slug}/builds"
    else:
        url = f"{BITRISE_API_BASE}/builds"

    async with httpx.AsyncClient() as client:
        headers = {
            "User-Agent": USER_AGENT,
            "Accept": "application/json",
            "Authorization": os.environ.get("BITRISE_TOKEN") or "",
        }
        response = await client.get(url, headers=headers, params=params, timeout=30.0)
        response.raise_for_status()
        return response.text


@mcp_tool(
    api_groups=["builds"],
    description="Trigger a new build/pipeline for a specified Bitrise app.",
)
async def trigger_bitrise_build(
    app_slug: str = Field(
        description='Identifier of the Bitrise app (e.g., "d8db74e2675d54c4" or "8eb495d0-f653-4eed-910b-8d6b56cc0ec7")',
    ),
    branch: str = Field(
        default="main",
        description="The branch to build",
    ),
    workflow_id: str = Field(
        default=None,
        description="The workflow to build",
    ),
    commit_message: str = Field(
        default=None,
        description="The commit message for the build",
    ),
    commit_hash: str = Field(
        default=None,
        description="The commit hash for the build",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/builds"
    build_params = {"branch": branch}

    if workflow_id:
        build_params["workflow_id"] = workflow_id
    if commit_message:
        build_params["commit_message"] = commit_message
    if commit_hash:
        build_params["commit_hash"] = commit_hash

    body = {
        "build_params": build_params,
        "hook_info": {"type": "bitrise"},
    }

    return await call_api("POST", url, body)


@mcp_tool(
    api_groups=["builds", "read-only"],
    description="Get a specific build of a given app.",
)
async def get_build(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
    build_slug: str = Field(
        description="Identifier of the build",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/builds/{build_slug}"
    return await call_api("GET", url)


@mcp_tool(
    api_groups=["builds"],
    description="Abort a specific build.",
)
async def abort_build(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
    build_slug: str = Field(
        description="Identifier of the build",
    ),
    reason: str = Field(
        default=None,
        description="Reason for aborting the build",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/builds/{build_slug}/abort"
    body = {}
    if reason:
        body["abort_reason"] = reason
    return await call_api("POST", url, body)


@mcp_tool(
    api_groups=["builds", "read-only"],
    description="Get the build log of a specified build of a Bitrise app.",
)
async def get_build_log(
    app_slug: str = Field(
        description='Identifier of the Bitrise app (e.g., "d8db74e2675d54c4" or "8eb495d0-f653-4eed-910b-8d6b56cc0ec7")',
    ),
    build_slug: str = Field(
        description="Identifier of the Bitrise build",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/builds/{build_slug}/log"
    return await call_api("GET", url)


@mcp_tool(
    api_groups=["builds", "read-only"],
    description="Get the bitrise.yml of a build.",
)
async def get_build_bitrise_yml(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
    build_slug: str = Field(
        description="Identifier of the build",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/builds/{build_slug}/bitrise.yml"
    return await call_api("GET", url)


@mcp_tool(
    api_groups=["builds", "read-only"],
    description="List the workflows of an app.",
)
async def list_build_workflows(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/build-workflows"
    return await call_api("GET", url)


# ===== Build Artifacts =====


@mcp_tool(
    api_groups=["artifacts", "read-only"],
    description="Get a list of all build artifacts.",
)
async def list_artifacts(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
    build_slug: str = Field(
        description="Identifier of the build",
    ),
    next: str = Field(
        default=None,
        description="Slug of the first artifact in the response",
    ),
    limit: int = Field(
        default=None,
        description="Max number of elements per page (default: 50)",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/builds/{build_slug}/artifacts"
    params: Dict[str, Union[str, int]] = {}
    if next:
        params["next"] = next
    if limit:
        params["limit"] = limit

    async with httpx.AsyncClient() as client:
        headers = {
            "User-Agent": USER_AGENT,
            "Accept": "application/json",
            "Authorization": os.environ.get("BITRISE_TOKEN") or "",
        }
        response = await client.get(url, headers=headers, params=params, timeout=30.0)
        response.raise_for_status()
        return response.text


@mcp_tool(
    api_groups=["artifacts", "read-only"],
    description="Get a specific build artifact.",
)
async def get_artifact(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
    build_slug: str = Field(
        description="Identifier of the build",
    ),
    artifact_slug: str = Field(
        description="Identifier of the artifact",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/builds/{build_slug}/artifacts/{artifact_slug}"
    return await call_api("GET", url)


@mcp_tool(
    api_groups=["artifacts"],
    description="Delete a build artifact.",
)
async def delete_artifact(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
    build_slug: str = Field(
        description="Identifier of the build",
    ),
    artifact_slug: str = Field(
        description="Identifier of the artifact",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/builds/{build_slug}/artifacts/{artifact_slug}"
    return await call_api("DELETE", url)


@mcp_tool(
    api_groups=["artifacts"],
    description="Update a build artifact.",
)
async def update_artifact(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
    build_slug: str = Field(
        description="Identifier of the build",
    ),
    artifact_slug: str = Field(
        description="Identifier of the artifact",
    ),
    is_public_page_enabled: bool = Field(
        description="Enable public page for the artifact",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/builds/{build_slug}/artifacts/{artifact_slug}"
    body = {"is_public_page_enabled": is_public_page_enabled}
    return await call_api("PATCH", url, body)


# ===== Webhooks =====


@mcp_tool(
    api_groups=["outgoing-webhooks", "read-only"],
    description="List the outgoing webhooks of an app.",
)
async def list_outgoing_webhooks(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/outgoing-webhooks"
    return await call_api("GET", url)


@mcp_tool(
    api_groups=["outgoing-webhooks"],
    description="Delete the outgoing webhook of an app.",
)
async def delete_outgoing_webhook(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
    webhook_slug: str = Field(
        description="Identifier of the webhook",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/outgoing-webhooks/{webhook_slug}"
    return await call_api("DELETE", url)


@mcp_tool(
    api_groups=["outgoing-webhooks"],
    description="Update an outgoing webhook for an app.",
)
async def update_outgoing_webhook(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
    webhook_slug: str = Field(
        description="Identifier of the webhook",
    ),
    events: List[str] = Field(
        description="List of events to trigger the webhook",
    ),
    url: str = Field(
        description="URL of the webhook",
    ),
    headers: Dict[str, str] = Field(
        default=None,
        description="Headers to be sent with the webhook",
    ),
) -> str:
    api_url = f"{BITRISE_API_BASE}/apps/{app_slug}/outgoing-webhooks/{webhook_slug}"
    body = {"events": events, "url": url, "headers": headers}

    return await call_api("PUT", api_url, body)


@mcp_tool(
    api_groups=["outgoing-webhooks"],
    description="Create an outgoing webhook for an app.",
)
async def create_outgoing_webhook(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
    events: List[str] = Field(
        description="List of events to trigger the webhook",
    ),
    url: str = Field(
        description="URL of the webhook",
    ),
    headers: Dict[str, str] = Field(
        default=None,
        description="Headers to be sent with the webhook",
    ),
) -> str:
    api_url = f"{BITRISE_API_BASE}/apps/{app_slug}/outgoing-webhooks"
    body: Dict[str, Any] = {"events": events, "url": url}
    if headers:
        body["headers"] = headers
    return await call_api("POST", api_url, body)


# ===== Cache Items =====


@mcp_tool(
    api_groups=["cache-items", "read-only"],
    description="List the key-value cache items belonging to an app.",
)
async def list_cache_items(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/cache-items"
    return await call_api("GET", url)


@mcp_tool(
    api_groups=["cache-items"],
    description="Delete all key-value cache items belonging to an app.",
)
async def delete_all_cache_items(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/cache"
    return await call_api("DELETE", url)


@mcp_tool(
    api_groups=["cache-items"],
    description="Delete a key-value cache item.",
)
async def delete_cache_item(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
    cache_item_id: str = Field(
        description="Key of the cache item",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/cache/{cache_item_id}"
    return await call_api("DELETE", url)


@mcp_tool(api_groups=["cache-items", "read-only"])
async def get_cache_item_download_url(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
    cache_item_id: str = Field(
        description="Key of the cache item",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/cache-items/{cache_item_id}/download"
    return await call_api("GET", url)


# ===== Pipelines =====


@mcp_tool(
    api_groups=["pipelines", "read-only"],
    description="List all pipelines and standalone builds of an app.",
)
async def list_pipelines(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/pipelines"
    return await call_api("GET", url)


@mcp_tool(
    api_groups=["pipelines", "read-only"],
    description="Get a pipeline of a given app.",
)
async def get_pipeline(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
    pipeline_id: str = Field(
        description="Identifier of the pipeline",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/pipelines/{pipeline_id}"
    return await call_api("GET", url)


@mcp_tool(
    api_groups=["pipelines"],
    description="Abort a pipeline.",
)
async def abort_pipeline(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
    pipeline_id: str = Field(
        description="Identifier of the pipeline",
    ),
    reason: str = Field(
        default=None,
        description="Reason for aborting the pipeline",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/pipelines/{pipeline_id}/abort"
    body = {}
    if reason:
        body["abort_reason"] = reason
    return await call_api("POST", url, body)


@mcp_tool(
    api_groups=["pipelines"],
    description="Rebuild a pipeline.",
)
async def rebuild_pipeline(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
    pipeline_id: str = Field(
        description="Identifier of the pipeline",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/pipelines/{pipeline_id}/rebuild"
    return await call_api("POST", url, {})


# ===== Group Roles =====


@mcp_tool(
    api_groups=["group-roles", "read-only"], description="List group roles for an app"
)
async def list_group_roles(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
    role_name: str = Field(
        description="Name of the role",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/roles/{role_name}"
    return await call_api("GET", url)


@mcp_tool(
    api_groups=["group-roles"],
    description="Replace group roles for an app.",
)
async def replace_group_roles(
    app_slug: str = Field(
        description="Identifier of the Bitrise app",
    ),
    role_name: str = Field(
        description="Name of the role",
    ),
    group_slugs: List[str] = Field(
        description="List of group slugs",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/roles/{role_name}"
    body = {"groups": group_slugs}
    return await call_api("PUT", url, body)


# ==== Workspaces ====


@mcp_tool(
    api_groups=["workspaces", "read-only"],
    description="List the workspaces the user has access to",
)
async def list_workspaces() -> str:
    url = f"{BITRISE_API_BASE}/organizations"
    return await call_api("GET", url)


@mcp_tool(
    api_groups=["workspaces", "read-only"],
    description="Get details for one workspace",
)
async def get_workspace(
    workspace_slug: str = Field(
        description="Slug of the Bitrise workspace",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/organizations/{workspace_slug}"
    return await call_api("GET", url)


@mcp_tool(
    api_groups=["workspaces", "read-only"],
    description="Get the groups in a workspace",
)
async def get_workspace_groups(
    workspace_slug: str = Field(
        description="Slug of the Bitrise workspace",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/organizations/{workspace_slug}/groups"
    return await call_api("GET", url)


@mcp_tool(
    api_groups=["workspaces"],
    description="Create a new group in a workspace.",
)
async def create_workspace_group(
    workspace_slug: str = Field(
        description="Slug of the Bitrise workspace",
    ),
    group_name: str = Field(
        description="Name of the group",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/organizations/{workspace_slug}/groups"
    return await call_api("POST", url, {"name": group_name})


@mcp_tool(
    api_groups=["workspaces", "read-only"],
    description="Get the members of a workspace",
)
async def get_workspace_members(
    workspace_slug: str = Field(
        description="Slug of the Bitrise workspace",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/organizations/{workspace_slug}/members"
    return await call_api("GET", url)


@mcp_tool(
    api_groups=["workspaces"],
    description="Invite new Bitrise users to a workspace.",
)
async def invite_member_to_workspace(
    workspace_slug: str = Field(
        description="Slug of the Bitrise workspace",
    ),
    email: str = Field(
        description="Email address of the user",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/organizations/{workspace_slug}/members"
    return await call_api("POST", url, {"email": email})


@mcp_tool(
    api_groups=["workspaces"],
    description="Add a member to a group.",
)
async def add_member_to_group(
    group_slug: str = Field(
        description="Slug of the group",
    ),
    user_slug: str = Field(
        description="Slug of the user",
    ),
) -> str:
    url = f"{BITRISE_API_BASE}/groups/{group_slug}/members/{user_slug}"
    return await call_api("PUT", url)


@mcp_tool(
    api_groups=["user", "read-only"],
    description="Get user info for the currently authenticated user account",
)
async def me() -> str:
    url = f"{BITRISE_API_BASE}/me"
    return await call_api("GET", url)


def main():
    mcp.run(transport="stdio")


if __name__ == "__main__":
    main()
