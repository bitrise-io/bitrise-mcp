import argparse
import os
import httpx
import sys
from functools import partial
from typing import Any, Dict, List, Union
from mcp.server.fastmcp import FastMCP
from pydantic import Field

mcp = FastMCP("bitrise")


BITRISE_API_BASE = "https://api.bitrise.io/v0.1"
BITRISE_RM_API_BASE = "https://api.bitrise.io/release-management/v1"
USER_AGENT = "bitrise-mcp/1.0"


parser = argparse.ArgumentParser()
parser.add_argument(
    "--enabled-api-groups",
    help="The list of enabled API groups, comma separated",
    type=partial(str.split, sep=","),
    default="apps,builds,workspaces,webhooks,build-artifacts,group-roles,cache-items,pipelines,account,read-only,release-management",
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


async def call_api(method, url: str, body=None, params=None) -> str:
    headers = {
        "User-Agent": USER_AGENT,
        "Accept": "application/json",
        "Content-Type": "application/json",
        "Authorization": os.environ.get("BITRISE_TOKEN") or "",
    }
    async with httpx.AsyncClient() as client:
        response = await client.request(
            method, url, headers=headers, json=body, params=params, timeout=30.0
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
    return await call_api("GET", url, params=params)


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
    pipeline_id: str = Field(
        default=None,
        description="The pipeline to build",
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

    if pipeline_id:
        build_params["pipeline_id"] = pipeline_id
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


# ===== Release Management =====


@mcp_tool(
    api_groups=["release-management"],
    description="Add a new Release Management connected app to Bitrise."
)
async def create_connected_app(
    platform: str = Field(
        description="The mobile platform for the connected app. Available values are 'ios' and 'android'.",
    ),
    store_app_id: str = Field(
        description="The app store identifier for the connected app. In case of 'ios' platform it is the bundle id "
                    "from App Store Connect. For additional context you can check the property description: "
                    "https://developer.apple.com/documentation/bundleresources/information-property-list/cfbundleidentifier"
                    "In case of Android platform it is the package name. Check the documentation: "
                    "https://developer.android.com/build/configure-app-module#set_the_application_id",
    ),
    workspace_slug: str = Field(
        description="Identifier of the Bitrise workspace for the Release Management connected app. This field is mandatory.",
    ),
    id: str = Field(
        default=None,
        description="An uuidV4 identifier for your new connected app. If it is not given, one will be generated. It is "
                    "useful for making the request idempotent or if the id is triggered outside of Bitrise and needs "
                    "to be stored separately as well.",
    ),
    manual_connection: bool = Field(
        default=False,
        description="If set to true it indicates a manual connection (bypassing using store api keys) and requires "
                    "giving 'store_app_name' as well. This can be especially useful for enterprise apps.",
    ),
    project_id: str = Field(
        default=None,
        description="Specifies which Bitrise Project you want to get the connected app to be associated with. If this field is not given a new project will be created alongside with the connected app.",
    ),
    store_app_name: str = Field(
        default=None,
        description="If you have no active app store API keys added on Bitrise, you can decide to add your app manually by giving the app's name as well while indicating manual connection with the similarly named boolean flag.",
    ),
    store_credential_id: str = Field(
        default=None,
        description="If you have credentials added on Bitrise, you can decide to select one for your app. In case of "
                    "ios platform it will be an Apple API credential id. In case of android platform it will be a "
                    "Google Service credential id.",
    ),
) -> str:
    url = f"{BITRISE_RM_API_BASE}/connected-apps"

    body = {
        "platform": platform,
        "store_app_id": store_app_id,
        "workspace_slug": workspace_slug,
    }
    if id:
        body["id"] = id
    if manual_connection:
        body["manual_connection"] = manual_connection
    if project_id:
        body["project_id"] = project_id
    if store_app_name:
        body["store_app_name"] = store_app_name
    if store_credential_id:
        body["store_credential_id"] = store_credential_id

    return await call_api("POST", url, body=body)

@mcp_tool(
    api_groups=["release-management"],
    description="List Release Management connected apps available for the authenticated account within a workspace.",
)
async def list_connected_apps(
    workspace_slug: str = Field(
        description="Identifier of the Bitrise workspace for the Release Management connected apps. This field is mandatory.",
    ),
    project_id: str = Field(
        default=None,
        description="Specifies which Bitrise Project you want to get associated connected apps for",
    ),
    platform: str = Field(
        default=None,
        description="Filters for a specific mobile platform for the list of connected apps. Available values are: 'ios' and 'android'.",
    ),
    search: str = Field(
        default=None,
        description="Search by bundle ID (for ios), package name (for android), or app title (for both platforms). The filter is case-sensitive.",
    ),
    items_per_page: int = Field(
        default=10,
        description="Specifies the maximum number of connected apps returned per page. Default value is 10.",
    ),
    page: int = Field(
        default=1,
        description="Specifies which page should be returned from the whole result set in a paginated scenario. Default value is 1.",
    ),
) -> str:
    params: Dict[str, Union[str, int]] = {
        "workspace_slug": workspace_slug
    }
    if project_id:
        params["project_id"] = project_id
    if platform:
        params["platform"] = platform
    if search:
        params["search"] = search
    if items_per_page:
        params["items_per_page"] = items_per_page
    if page:
        params["page"] = page

    url = f"{BITRISE_RM_API_BASE}/connected-apps"
    return await call_api("GET", url, params=params)

@mcp_tool(
    api_groups=["release-management"],
    description="Gives back a Release Management connected app for the authenticated account.",
)
async def get_connected_app(
    id: str = Field(
        description="Identifier of the Release Management connected app",
    ),
) -> str:
    url = f"{BITRISE_RM_API_BASE}/connected-apps/{id}"
    return await call_api("GET", url)

@mcp_tool(
    api_groups=["release-management"],
    description="Updates a connected app."
)
async def update_connected_app(
    connected_app_id: str = Field(
        description="The uuidV4 identifier for your connected app.",
    ),
    connect_to_store: bool = Field(
        default=False,
        description="If true, will check connected app validity against the Apple App Store or Google Play Store "
                    "(dependent on the platform of your connected app). This means, that the already set or just given "
                    "store_app_id will be validated against the Store, using the already set or just given store "
                    "credential id.",
    ),
    store_app_id: str = Field(
        description="The store identifier for your app. You can change the previously set store_app_id to match the "
                "one in the App Store or Google Play depending on the app platform. This is especially useful if "
                "you want to connect your app with the store as the system will validate the given store_app_id "
                "against the Store. In case of iOS platform it is the bundle id. In case of Android platform it is "
                "the package name.",
    ),
    store_credential_id: str = Field(
        default=None,
        description="If you have credentials added on Bitrise, you can decide to select one for your app. In case of "
                    "ios platform it will be an Apple API credential id. In case of android platform it will be a "
                    "Google Service credential id.",
    ),
) -> str:
    url = f"{BITRISE_RM_API_BASE}/connected-apps/{connected_app_id}"

    body = {}
    if connect_to_store:
        body["connect_to_store"] = connect_to_store
    if store_app_id:
        body["store_app_id"] = store_app_id
    if store_credential_id:
        body["store_credential_id"] = store_credential_id

    return await call_api("PATCH", url, body=body)

@mcp_tool(
    api_groups=["release-management"],
    description="List Release Management installable artifacts of a connected app available for the authenticated account.",
)
async def list_installable_artifacts(
    connected_app_id: str = Field(
        description="Identifier of the Release Management connected app for the installable artifacts. This field is mandatory.",
    ),

    after_date: str = Field(
        default=None,
        description="A date in ISO 8601 string format specifying the start of the interval when the installable "
                    "artifact to be returned was created or uploaded. This value will be defaulted to 1 month ago if "
                    "distribution_ready filter is not set or set to false."
    ),
    artifact_type: str = Field(
        default=None,
        description="Filters for a specific artifact type or file extension for the list of installable artifacts. "
                    "Available values are: 'aab' and 'apk' for android artifacts and 'ipa' for ios artifacts."
    ),
    before_date: str = Field(
        default=None,
        description="A date in ISO 8601 string format specifying the end of the interval when the installable artifact "
                    "to be returned was created or uploaded. This value will be defaulted to the current time if "
                    "distribution_ready filter is not set or set to false."
    ),
    branch: str = Field(
        default=None,
        description="Filters for the Bitrise CI branch of the installable artifact on which it has been generated on.",
    ),
    distribution_ready: bool = Field(
        default=None,
        description="Filters for distribution ready installable artifacts. This means .apk and .ipa (with "
                    "distribution type ad-hoc, development, or enterprise) installable artifacts.",
    ),
    items_per_page: int = Field(
        default=10,
        description="Specifies the maximum number of installable artifacts to be returned per page. Default value is 10.",
    ),
    page: int = Field(
        default=1,
        description="Specifies which page should be returned from the whole result set in a paginated scenario. Default value is 1.",
    ),
    platform: str = Field(
        default=None,
        description="Filters for a specific mobile platform for the list of installable artifacts. Available values are: 'ios' and 'android'.",
    ),
    search: str = Field(
        default=None,
        description="Search by version, filename or build number (Bitrise CI). The filter is case-sensitive.",
    ),
    source: str = Field(
        default=None,
        description="Filters for the source of installable artifacts to be returned. Available values are 'api' and "
                    "'ci'."
    ),
    store_signed: bool = Field(
        default=None,
        description="Filters for store ready installable artifacts. This means signed .aab and .ipa (with distribution type app-store) installable artifacts.",
    ),
    version: str = Field(
        default=None,
        description="Filters for the version this installable artifact was created for. This field is required if the "
                    "distribution_ready filter is set to true."
    ),
    workflow: str = Field(
        default=None,
        description="Filters for the Bitrise CI workflow of the installable artifact it has been generated by.",
    ),
) -> str:
    params: Dict[str, Union[str, int, bool]] = {}
    if after_date:
        params["after_date"] = after_date
    if artifact_type:
        params["artifact_type"] = artifact_type
    if before_date:
        params["before_date"] = before_date
    if branch:
        params["branch"] = branch
    if distribution_ready:
        params["distribution_ready"] = distribution_ready
    if items_per_page:
        params["items_per_page"] = items_per_page
    if page:
        params["page"] = page
    if platform:
        params["platform"] = platform
    if search:
        params["search"] = search
    if source:
        params["source"] = source
    if store_signed:
        params["store_signed"] = store_signed
    if version:
        params["version"] = version
    if workflow:
        params["workflow"] = workflow

    url = f"{BITRISE_RM_API_BASE}/connected-apps/{connected_app_id}/installable-artifacts"
    return await call_api("GET", url, params=params)

@mcp_tool(
    api_groups=["release-management"],
    description="Generates a signed upload url valid for 1 hour for an installable artifact to be uploaded to Bitrise "
                "Release Management. The response will contain an url that can be used to upload an artifact to Bitrise "
                "Release Management using a simple curl request with the file data that should be uploaded. The "
                "necessary headers and http method will also be in the response. This artifact will need to be "
                "processed after upload to be usable. The status of processing can be checked by making another request"
                "to a different url giving back the processed status of an installable artifact.",
)
async def generate_installable_artifact_upload_url(
    connected_app_id: str = Field(
        description="Identifier of the Release Management connected app for the installable artifact. This field is mandatory.",
    ),
    installable_artifact_id: str = Field(
        description="An uuidv4 identifier generated on the client side for the installable artifact. This field is "
                    "mandatory.",
    ),
    file_name: str = Field(
        description="The name of the installable artifact file (with extension) to be uploaded to Bitrise. This field "
                    "is mandatory.",
    ),
    file_size_bytes: str = Field(
        description="The byte size of the installable artifact file to be uploaded.",
    ),
    branch: str = Field(
        default=None,
        description="Optionally you can add the name of the CI branch the installable artifact has been generated on.",
    ),
    with_public_page: bool = Field(
        default=None,
        description="Optionally, you can enable public install page for your artifact. This can only be enabled by "
                    "Bitrise Project Admins, Bitrise Project Owners and Bitrise Workspace Admins. Changing this value "
                    "without proper permissions will result in an error. The default value is false.",
    ),
    workflow: str = Field(
        default=None,
        description="Optionally you can add the name of the CI workflow this installable artifact has been generated by.",
    ),
) -> str:
    params: Dict[str, Union[str, int, bool]] = {
        "file_name": file_name,
        "file_size_bytes": file_size_bytes
    }

    if branch:
        params["branch"] = branch
    if with_public_page:
        params["with_public_page"] = with_public_page
    if workflow:
        params["workflow"] = workflow

    url = f"{BITRISE_RM_API_BASE}/connected-apps/{connected_app_id}/installable-artifacts/{installable_artifact_id}/upload-url"
    return await call_api("GET", url, params=params)

@mcp_tool(
    api_groups=["release-management"],
    description="Gets the processing and upload status of an installable artifact. An artifact will need to be "
                "processed after upload to be usable. This endpoint helps understanding when an uploaded installable "
                "artifacts becomes usable for later purposes.",
)
async def get_installable_artifact_upload_and_processing_status(
    connected_app_id: str = Field(
        description="Identifier of the Release Management connected app for the installable artifact. This field is mandatory.",
    ),
    installable_artifact_id: str = Field(
        description="The uuidv4 identifier for the installable artifact. This field is mandatory.",
    ),
) -> str:
    url = f"{BITRISE_RM_API_BASE}/connected-apps/{connected_app_id}/installable-artifacts/{installable_artifact_id}/status"
    return await call_api("GET", url)

@mcp_tool(
    api_groups=["release-management"],
    description="Changes whether public install page should be available for the installable artifact or not."
)
async def set_installable_artifact_public_install_page(
    connected_app_id: str = Field(
        description="Identifier of the Release Management connected app for the installable artifact. This field is mandatory.",
    ),
    installable_artifact_id: str = Field(
        description="The uuidv4 identifier for the installable artifact. This field is mandatory.",
    ),
    with_public_page: bool = Field(
        description="Boolean flag for enabling/disabling public install page for the installable artifact. This field is mandatory.",
    ),
) -> str:
    url = f"{BITRISE_RM_API_BASE}/connected-apps/{connected_app_id}/installable-artifacts/{installable_artifact_id}/public-install-page"
    body = {
        "with_public_page": with_public_page,
    }
    return await call_api("PATCH", url, body=body)

@mcp_tool(
    api_groups=["release-management"],
    description="Lists Build Distribution versions. Release Management offers a convenient, secure solution to "
                "distribute the builds of your mobile apps to testers without having to engage with either TestFlight "
                "or Google Play. Once you have installable artifacts, Bitrise can generate both private and public "
                "install links that testers or other stakeholders can use to install the app on real devices via "
                "over-the-air installation. Build distribution allows you to define tester groups that can receive "
                "notifications about installable artifacts. The email takes the notified testers to the test build "
                "page, from where they can install the app on their own device. Build distribution versions are the "
                " app versions available for testers.",
)
async def list_build_distribution_versions(
    connected_app_id: str = Field(
        description="The uuidV4 identifier of the app the build distribution is connected to. This field is mandatory.",
    ),
    items_per_page: int = Field(
        default=10,
        description="Specifies the maximum number of build distribution versions returned per page. Default value is 10.",
    ),
    page: int = Field(
        default=1,
        description="Specifies which page should be returned from the whole result set in a paginated scenario. Default value is 1.",
    ),
) -> str:
    params: Dict[str, Union[str, int]] = {}
    if items_per_page:
        params["items_per_page"] = items_per_page
    if page:
        params["page"] = page

    url = f"{BITRISE_RM_API_BASE}/connected-apps/{connected_app_id}/build-distributions"
    return await call_api("GET", url, params=params)

@mcp_tool(
    api_groups=["release-management"],
    description="Gives back a list of test builds for the given build distribution version.",
)
async def list_build_distribution_version_test_builds(
    connected_app_id: str = Field(
        description="The uuidV4 identifier of the app the build distribution is connected to. This field is mandatory.",
    ),
    version: str = Field(
        description="The version of the build distribution. This field is mandatory.",
    ),

    items_per_page: int = Field(
        default=10,
        description="Specifies the maximum number of test builds to return for a build distribution version per page. Default value is 10.",
    ),
    page: int = Field(
        default=1,
        description="Specifies which page should be returned from the whole result set in a paginated scenario. Default value is 1.",
    ),
) -> str:
    params: Dict[str, Union[str, int]] = {
        "version": version
    }

    if items_per_page:
        params["items_per_page"] = items_per_page
    if page:
        params["page"] = page

    url = f"{BITRISE_RM_API_BASE}/connected-apps/{connected_app_id}/build-distributions/test-builds"
    return await call_api("GET", url, params=params)

@mcp_tool(
    api_groups=["release-management"],
    description="Creates a tester group for a Release Management connected app. Tester groups can be used to distribute "
                "installable artifacts to testers automatically. When a new installable artifact is available, the "
                "tester groups can either automatically or manually be notified via email. The notification email will "
                "contain a link to the installable artifact page for the artifact within Bitrise Release "
                "Management. A Release Management connected app can have multiple tester groups. Project team members "
                "of the connected app can be selected to be testers and added to the tester group. This endpoint has "
                "an elevated access level requirement. Only the owner of the related Bitrise Workspace, a workspace "
                "manager or the related project's admin can manage tester groups."
)
async def create_tester_group(
    connected_app_id: str = Field(
        description="The uuidV4 identifier of the related Release Management connected app.",
    ),
    name: str = Field(
        description="The name for the new tester group. Must be unique in the scope of the connected app.",
    ),
    auto_notify: bool = Field(
        default=False,
        description="If set to true it indicates that the tester group will receive notifications automatically.",
    ),
) -> str:
    url = f"{BITRISE_RM_API_BASE}/connected-apps/{connected_app_id}/tester-groups"

    body = {}
    if name:
        body["name"] = name
    if auto_notify:
        body["auto_notify"] = auto_notify

    return await call_api("POST", url, body=body)

@mcp_tool(
    api_groups=["release-management"],
    description="Notifies a tester group about a new test build."
)
async def notify_tester_group(
    connected_app_id: str = Field(
        description="The uuidV4 identifier of the related Release Management connected app.",
    ),
    id: str = Field(
        description="The uuidV4 identifier of the tester group whose members will be notified about the test build.",
    ),
    test_build_id: str = Field(
        description="The unique identifier of the test build what will be sent in the notification of the tester group.",
    ),
) -> str:
    url = f"{BITRISE_RM_API_BASE}/connected-apps/{connected_app_id}/tester-groups/{id}/notify"
    body = {
        "test_build_id": test_build_id,
    }
    return await call_api("POST", url, body=body)

@mcp_tool(
    api_groups=["release-management"],
    description="Adds testers to a tester group of a connected app."
)
async def add_testers_to_tester_group(
    connected_app_id: str = Field(
        description="The uuidV4 identifier of the related Release Management connected app.",
    ),
    id: str = Field(
        description="The uuidV4 identifier of the tester group to which testers will be added.",
    ),
    user_slugs: list[str] = Field(
        description="The list of users identified by slugs that will be added to the tester group.",
    ),
) -> str:
    url = f"{BITRISE_RM_API_BASE}/connected-apps/{connected_app_id}/tester-groups/{id}/add-testers"
    body = {
        "user_slugs": user_slugs,
    }
    return await call_api("POST", url, body=body)

@mcp_tool(
    api_groups=["release-management"],
    description="Updates the given tester group. The name and the auto notification setting can be updated optionally."
)
async def update_tester_group(
    connected_app_id: str = Field(
        description="The uuidV4 identifier of the related Release Management connected app.",
    ),
    id: str = Field(
        description="The uuidV4 identifier of the tester group to which testers will be added.",
    ),
    name: str = Field(
        default=None,
        description="The new name for the tester group. Must be unique in the scope of the related connected app.",
    ),
    auto_notify: bool = Field(
        default=False,
        description="If set to true it indicates the tester group will receive email notifications automatically from "
                    "now on about new installable builds.",
    ),
) -> str:
    url = f"{BITRISE_RM_API_BASE}/connected-apps/{connected_app_id}/tester-groups/{id}"

    body = {}
    if name:
        body["name"] = name
    if auto_notify:
        body["auto_notify"] = auto_notify

    return await call_api("PUT", url, body=body)

@mcp_tool(
    api_groups=["release-management"],
    description="Gives back a list of tester groups related to a specific Release Management connected app.",
)
async def list_tester_groups(
    connected_app_id: str = Field(
        description="The uuidV4 identifier of the app the tester group is connected to. This field is mandatory.",
    ),
    items_per_page: int = Field(
        default=10,
        description="Specifies the maximum number of tester groups to return related to a specific connected app. Default value is 10.",
    ),
    page: int = Field(
        default=1,
        description="Specifies which page should be returned from the whole result set in a paginated scenario. Default value is 1.",
    ),
) -> str:
    params: Dict[str, Union[str, int]] = {}
    if items_per_page:
        params["items_per_page"] = items_per_page
    if page:
        params["page"] = page

    url = f"{BITRISE_RM_API_BASE}/connected-apps/{connected_app_id}/tester-groups"
    return await call_api("GET", url, params=params)

@mcp_tool(
    api_groups=["release-management"],
    description="Gives back the details of the selected tester group.",
)
async def get_tester_group(
    connected_app_id: str = Field(
        description="The uuidV4 identifier of the app the tester group is connected to. This field is mandatory.",
    ),
    id: str = Field(
        description="The uuidV4 identifier of the tester group. This field is mandatory.",
    ),
) -> str:
    url = f"{BITRISE_RM_API_BASE}/connected-apps/{connected_app_id}/tester-groups/{id}"
    return await call_api("GET", url)

@mcp_tool(
    api_groups=["release-management"],
    description="Gets a list of potential testers whom can be added as testers to a specific tester group. The list "
                "consists of Bitrise users having access to the related Release Management connected app.",
)
async def get_potential_testers(
    connected_app_id: str = Field(
        description="The uuidV4 identifier of the app the tester group is connected to. This field is mandatory.",
    ),
    id: str = Field(
        description="The uuidV4 identifier of the tester group. This field is mandatory.",
    ),
    items_per_page: int = Field(
        default=10,
        description="Specifies the maximum number of potential testers to return having access to a specific connected app. Default value is 10.",
    ),
    page: int = Field(
        default=1,
        description="Specifies which page should be returned from the whole result set in a paginated scenario. Default value is 1.",
    ),
    search: str = Field(
        default=None,
        description="Searches for potential testers based on email or username using a case-insensitive approach.",
    ),
) -> str:
    params: Dict[str, Union[str, int]] = {}
    if items_per_page:
        params["items_per_page"] = items_per_page
    if page:
        params["page"] = page
    if search:
        params["search"] = search

    url = f"{BITRISE_RM_API_BASE}/connected-apps/{connected_app_id}/tester-groups/{id}/potential-testers"
    return await call_api("GET", url, params=params)

@mcp_tool(
    api_groups=["release-management"],
    description="Gives back a list of testers that has been associated with a tester group related to a specific "
                "connected app.",
)
async def get_testers(
    connected_app_id: str = Field(
        description="The uuidV4 identifier of the app the tester group is connected to. This field is mandatory.",
    ),
    tester_group_id: str = Field(
        description="The uuidV4 identifier of a tester group. If given, only testers within this specific tester group "
                    "will be returned.",
    ),
    items_per_page: int = Field(
        default=10,
        description="Specifies the maximum number of testers to be returned that have been added to a tester group "
                    "related to the specific connected app.. Default value is 10.",
    ),
    page: int = Field(
        default=1,
        description="Specifies which page should be returned from the whole result set in a paginated scenario. Default value is 1.",
    ),
) -> str:
    params: Dict[str, Union[str, int]] = {}
    if tester_group_id:
        params["tester_group_id"] = tester_group_id
    if items_per_page:
        params["items_per_page"] = items_per_page
    if page:
        params["page"] = page

    url = f"{BITRISE_RM_API_BASE}/connected-apps/{connected_app_id}/testers"
    return await call_api("GET", url, params=params)


def main():
    mcp.run(transport="stdio")


if __name__ == "__main__":
    main()
