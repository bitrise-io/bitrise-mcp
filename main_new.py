import os
from typing import Any, Dict, List, Optional
import httpx
from mcp.server.fastmcp import FastMCP

mcp = FastMCP("bitrise")

BITRISE_API_BASE = "https://api-staging.bitrise.io/v0.1"
USER_AGENT = "bitrise-mcp/1.0"

async def call_api(method, url: str, body = None) -> str:
    headers = {
        "User-Agent": USER_AGENT,
        "Accept": "application/json",
        "Content-Type": "application/json",
        "Authorization": os.environ.get("BITRISE_TOKEN"),
    }
    async with httpx.AsyncClient() as client:
        response = await client.request(method, url, headers=headers, json=body, timeout=30.0)
        response.raise_for_status()
        return response.text

# ===== Apps =====

@mcp.tool()
async def list_apps(sort_by: Optional[str] = None, next: Optional[str] = None, 
                   limit: Optional[int] = None) -> str:
    """List all the apps available for the authenticated account.
    
    Args:
        sort_by: Order of the apps: last_build_at (default) or created_at
        next: Slug of the first app in the response
        limit: Max number of elements per page (default: 50)
    """
    params = {}
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
            "Authorization": os.environ.get("BITRISE_TOKEN"),
        }
        response = await client.get(url, headers=headers, params=params, timeout=30.0)
        response.raise_for_status()
        return response.text

@mcp.tool()
async def register_app(repo_url: str, is_public: bool, 
                      organization_slug: str,
                      project_type: Optional[str] = "other",
                      provider: Optional[str] = "github" ) -> str:
    """Add a new app to Bitrise. After this app should be finished on order to be registered coompletely on Bitrise.
    
    Args:
        repo_url: Repository URL
        is_public: Whether the app's builds visibility is "public"
        organization_slug: The organization (aka workspace) the app to add to
        project_type: Type of project (ios, android, etc.)
        provider: github
    """
    url = f"{BITRISE_API_BASE}/apps/register"
    body = {
        "repo_url": repo_url,
        "is_public": is_public,
        "organization_slug": repo_provider,
        "project_type": project_type,
        "provider": provider
    }
    return await call_api("POST", url, body)

@mcp.tool()
async def finish_bitrise_app(app_slug: str, stack_id: str, organization_slug: str, project_type: str) -> str:
    """Finish the setup of a Bitrise app.

    Args:
        app_slug: The slug of the Bitrise app to finish setup for.
        stack_id: The stack ID to use for the app (default is "linux-docker-22.04").
        organization_slug: the workspace's slug
        project_type: Type of project (ios, android, etc.) If it's not given use the app's project type.

    Returns:
        The response from the Bitrise API after finishing the app setup.
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/finish"
    payload = {
        "mode": "manual",
        "stack_id": stack_id,
        "organizaion_slug": organization_slug,
        "project_type": project_type
    }
    return await call_api("POST", url, payload)

@mcp.tool()
async def get_app(app_slug: str) -> str:
    """Get the details of a specific app.
    
    Args:
        app_slug: Identifier of the Bitrise app
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}"
    return await call_api("GET", url)

@mcp.tool()
async def delete_app(app_slug: str) -> str:
    """Delete an app from Bitrise.
    
    Args:
        app_slug: Identifier of the Bitrise app
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}"
    return await call_api("DELETE", url)

@mcp.tool()
async def update_app(app_slug: str, is_public: bool, project_type: str, 
                    provider: str, repo_url: str) -> str:
    """Update an app.
    
    Args:
        app_slug: Identifier of the Bitrise app
        is_public: Whether the app's builds visibility is "public"
        project_type: Type of project
        provider: Repository provider
        repo_url: Repository URL
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}"
    body = {
        "is_public": is_public,
        "project_type": project_type,
        "provider": provider,
        "repo_url": repo_url
    }
    return await call_api("PATCH", url, body)

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
async def list_branches(app_slug: str) -> str:
    """List the branches with existing builds of an app's repository.
    
    Args:
        app_slug: Identifier of the Bitrise app
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/branches"
    return await call_api("GET", url)

@mcp.tool()
async def register_ssh_key(app_slug: str, auth_ssh_private_key: str, 
                         auth_ssh_public_key: str, is_register_key_into_provider_service: bool) -> str:
    """Add an SSH-key to a specific app.
    
    Args:
        app_slug: Identifier of the Bitrise app
        auth_ssh_private_key: Private SSH key
        auth_ssh_public_key: Public SSH key
        is_register_key_into_provider_service: Register the key in the provider service
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/register-ssh-key"
    body = {
        "auth_ssh_private_key": auth_ssh_private_key,
        "auth_ssh_public_key": auth_ssh_public_key,
        "is_register_key_into_provider_service": is_register_key_into_provider_service
    }
    return await call_api("POST", url, body)

@mcp.tool()
async def register_webhook(app_slug: str) -> str:
    """Register an incoming webhook for a specific application.
    
    Args:
        app_slug: Identifier of the Bitrise app
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/register-webhook"
    return await call_api("POST", url)

# ===== Builds =====

@mcp.tool()
async def list_builds(app_slug: Optional[str] = None, sort_by: Optional[str] = None, 
                     branch: Optional[str] = None, workflow: Optional[str] = None,
                     status: Optional[int] = None, next: Optional[str] = None, 
                     limit: Optional[int] = None) -> str:
    """List all the builds of a specified Bitrise app or all accessible builds.
    
    Args:
        app_slug: Identifier of the Bitrise app (optional)
        sort_by: Order of builds: created_at (default), running_first
        branch: Filter builds by branch
        workflow: Filter builds by workflow
        status: Filter builds by status (0: not finished, 1: successful, 2: failed, 3: aborted, 4: in-progress)
        next: Slug of the first build in the response
        limit: Max number of elements per page (default: 50)
    """
    params = {}
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
            "Authorization": os.environ.get("BITRISE_TOKEN"),
        }
        response = await client.get(url, headers=headers, params=params, timeout=30.0)
        response.raise_for_status()
        return response.text

@mcp.tool()
async def trigger_bitrise_build(app_slug: str, branch: str = "main", 
                               workflow_id: Optional[str] = None, 
                               commit_message: Optional[str] = None,
                               commit_hash: Optional[str] = None) -> str:
    """Trigger a new build/pipeline for a specified Bitrise app.

    Args:
        app_slug: Identifier of the Bitrise app (e.g., "d8db74e2675d54c4" or "8eb495d0-f653-4eed-910b-8d6b56cc0ec7")
        branch: The branch to build (default: main)
        workflow_id: The workflow to build (optional)
        commit_message: The commit message for the build (optional)
        commit_hash: The commit hash for the build (optional)
    """
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
        "hook_info": {
            "type": "bitrise"
        },
    }
    
    return await call_api("POST", url, body)

@mcp.tool()
async def get_build(app_slug: str, build_slug: str) -> str:
    """Get a specific build of a given app.
    
    Args:
        app_slug: Identifier of the Bitrise app
        build_slug: Identifier of the build
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/builds/{build_slug}"
    return await call_api("GET", url)

@mcp.tool()
async def abort_build(app_slug: str, build_slug: str, reason: Optional[str] = None) -> str:
    """Abort a specific build.
    
    Args:
        app_slug: Identifier of the Bitrise app
        build_slug: Identifier of the build
        reason: Reason for aborting the build
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/builds/{build_slug}/abort"
    body = {}
    if reason:
        body["abort_reason"] = reason
    return await call_api("POST", url, body)

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
async def get_build_bitrise_yml(app_slug: str, build_slug: str) -> str:
    """Get the bitrise.yml of a build.
    
    Args:
        app_slug: Identifier of the Bitrise app
        build_slug: Identifier of the build
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/builds/{build_slug}/bitrise.yml"
    return await call_api("GET", url)

@mcp.tool()
async def list_build_workflows(app_slug: str) -> str:
    """List the workflows of an app.
    
    Args:
        app_slug: Identifier of the Bitrise app
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/build-workflows"
    return await call_api("GET", url)

# ===== Build Artifacts =====

@mcp.tool()
async def list_artifacts(app_slug: str, build_slug: str, next: Optional[str] = None, 
                        limit: Optional[int] = None) -> str:
    """Get a list of all build artifacts.
    
    Args:
        app_slug: Identifier of the Bitrise app
        build_slug: Identifier of the build
        next: Slug of the first artifact in the response
        limit: Max number of elements per page (default: 50)
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/builds/{build_slug}/artifacts"
    params = {}
    if next:
        params["next"] = next
    if limit:
        params["limit"] = limit
        
    async with httpx.AsyncClient() as client:
        headers = {
            "User-Agent": USER_AGENT,
            "Accept": "application/json",
            "Authorization": os.environ.get("BITRISE_TOKEN"),
        }
        response = await client.get(url, headers=headers, params=params, timeout=30.0)
        response.raise_for_status()
        return response.text

@mcp.tool()
async def get_artifact(app_slug: str, build_slug: str, artifact_slug: str) -> str:
    """Get a specific build artifact.
    
    Args:
        app_slug: Identifier of the Bitrise app
        build_slug: Identifier of the build
        artifact_slug: Identifier of the artifact
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/builds/{build_slug}/artifacts/{artifact_slug}"
    return await call_api("GET", url)

@mcp.tool()
async def delete_artifact(app_slug: str, build_slug: str, artifact_slug: str) -> str:
    """Delete a build artifact.
    
    Args:
        app_slug: Identifier of the Bitrise app
        build_slug: Identifier of the build
        artifact_slug: Identifier of the artifact
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/builds/{build_slug}/artifacts/{artifact_slug}"
    return await call_api("DELETE", url)

@mcp.tool()
async def update_artifact(app_slug: str, build_slug: str, artifact_slug: str, 
                         is_public_page_enabled: bool) -> str:
    """Update a build artifact.
    
    Args:
        app_slug: Identifier of the Bitrise app
        build_slug: Identifier of the build
        artifact_slug: Identifier of the artifact
        is_public_page_enabled: Enable public page for the artifact
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/builds/{build_slug}/artifacts/{artifact_slug}"
    body = {"is_public_page_enabled": is_public_page_enabled}
    return await call_api("PATCH", url, body)

# ===== Addons =====

@mcp.tool()
async def list_addons() -> str:
    """List all the available Bitrise addons."""
    url = f"{BITRISE_API_BASE}/addons"
    return await call_api("GET", url)

@mcp.tool()
async def get_addon(addon_id: str) -> str:
    """Show details of a specific Bitrise addon.
    
    Args:
        addon_id: Identifier of the addon
    """
    url = f"{BITRISE_API_BASE}/addons/{addon_id}"
    return await call_api("GET", url)

@mcp.tool()
async def list_app_addons(app_slug: str) -> str:
    """List all the provisioned addons for an app.
    
    Args:
        app_slug: Identifier of the Bitrise app
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/addons"
    return await call_api("GET", url)

# ===== iOS Code Signing =====

@mcp.tool()
async def list_build_certificates(app_slug: str) -> str:
    """Get a list of the build certificates.
    
    Args:
        app_slug: Identifier of the Bitrise app
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/build-certificates"
    return await call_api("GET", url)

@mcp.tool()
async def create_build_certificate(app_slug: str, upload_file_name: str, 
                                 upload_file_size: int, upload_content_type: str,
                                 certificate_password: str) -> str:
    """Create a build certificate.
    
    Args:
        app_slug: Identifier of the Bitrise app
        upload_file_name: Name of the certificate file
        upload_file_size: Size of the certificate file
        upload_content_type: Content type of the certificate file
        certificate_password: Password of the certificate
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/build-certificates"
    body = {
        "upload_file_name": upload_file_name,
        "upload_file_size": upload_file_size,
        "upload_content_type": upload_content_type,
        "certificate_password": certificate_password
    }
    return await call_api("POST", url, body)

@mcp.tool()
async def list_provisioning_profiles(app_slug: str) -> str:
    """Get a list of the provisioning profiles.
    
    Args:
        app_slug: Identifier of the Bitrise app
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/provisioning-profiles"
    return await call_api("GET", url)

@mcp.tool()
async def create_provisioning_profile(app_slug: str, upload_file_name: str, 
                                    upload_file_size: int, 
                                    upload_content_type: str) -> str:
    """Create a provisioning profile.
    
    Args:
        app_slug: Identifier of the Bitrise app
        upload_file_name: Name of the provisioning profile file
        upload_file_size: Size of the provisioning profile file
        upload_content_type: Content type of the provisioning profile file
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/provisioning-profiles"
    body = {
        "upload_file_name": upload_file_name,
        "upload_file_size": upload_file_size,
        "upload_content_type": upload_content_type
    }
    return await call_api("POST", url, body)

# ===== Android Keystore =====

@mcp.tool()
async def list_android_keystore_files(app_slug: str) -> str:
    """Get a list of the android keystore files.
    
    Args:
        app_slug: Identifier of the Bitrise app
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/android-keystore-files"
    return await call_api("GET", url)

@mcp.tool()
async def create_android_keystore_file(app_slug: str, upload_file_name: str, 
                                     upload_file_size: int, upload_content_type: str,
                                     alias: str, password: str, private_key_password: str) -> str:
    """Create an Android keystore file.
    
    Args:
        app_slug: Identifier of the Bitrise app
        upload_file_name: Name of the keystore file
        upload_file_size: Size of the keystore file
        upload_content_type: Content type of the keystore file
        alias: Alias of the keystore
        password: Password of the keystore
        private_key_password: Private key password of the keystore
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/android-keystore-files"
    body = {
        "upload_file_name": upload_file_name,
        "upload_file_size": upload_file_size,
        "upload_content_type": upload_content_type,
        "alias": alias,
        "password": password,
        "private_key_password": private_key_password
    }
    return await call_api("POST", url, body)

# ===== Webhooks =====

@mcp.tool()
async def list_outgoing_webhooks(app_slug: str) -> str:
    """List the outgoing webhooks of an app.
    
    Args:
        app_slug: Identifier of the Bitrise app
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/outgoing-webhooks"
    return await call_api("GET", url)

@mcp.tool()
async def create_outgoing_webhook(app_slug: str, events: List[str], url: str, 
                                 headers: Dict[str, str] = None) -> str:
    """Create an outgoing webhook for an app.
    
    Args:
        app_slug: Identifier of the Bitrise app
        events: List of events to trigger the webhook
        url: URL of the webhook
        headers: Headers to be sent with the webhook
    """
    api_url = f"{BITRISE_API_BASE}/apps/{app_slug}/outgoing-webhooks"
    body = {
        "events": events,
        "url": url
    }
    if headers:
        body["headers"] = headers
    return await call_api("POST", api_url, body)

# ===== Cache Items =====

@mcp.tool()
async def list_cache_items(app_slug: str) -> str:
    """List the key-value cache items belonging to an app.
    
    Args:
        app_slug: Identifier of the Bitrise app
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/cache-items"
    return await call_api("GET", url)

@mcp.tool()
async def delete_all_cache_items(app_slug: str) -> str:
    """Delete all key-value cache items belonging to an app.
    
    Args:
        app_slug: Identifier of the Bitrise app
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/cache-items"
    return await call_api("DELETE", url)

@mcp.tool()
async def delete_cache_item(app_slug: str, cache_item_id: str) -> str:
    """Delete a key-value cache item.
    
    Args:
        app_slug: Identifier of the Bitrise app
        cache_item_id: Identifier of the cache item
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/cache-items/{cache_item_id}"
    return await call_api("DELETE", url)

@mcp.tool()
async def get_cache_item_download_url(app_slug: str, cache_item_id: str) -> str:
    """Get the download URL of a key-value cache item.
    
    Args:
        app_slug: Identifier of the Bitrise app
        cache_item_id: Identifier of the cache item
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/cache-items/{cache_item_id}/download"
    return await call_api("GET", url)

# ===== Pipelines =====

@mcp.tool()
async def list_pipelines(app_slug: str) -> str:
    """List all pipelines and standalone builds of an app.
    
    Args:
        app_slug: Identifier of the Bitrise app
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/pipelines"
    return await call_api("GET", url)

@mcp.tool()
async def get_pipeline(app_slug: str, pipeline_id: str) -> str:
    """Get a pipeline of a given app.
    
    Args:
        app_slug: Identifier of the Bitrise app
        pipeline_id: Identifier of the pipeline
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/pipelines/{pipeline_id}"
    return await call_api("GET", url)

@mcp.tool()
async def abort_pipeline(app_slug: str, pipeline_id: str, reason: Optional[str] = None) -> str:
    """Abort a pipeline.
    
    Args:
        app_slug: Identifier of the Bitrise app
        pipeline_id: Identifier of the pipeline
        reason: Reason for aborting the pipeline
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/pipelines/{pipeline_id}/abort"
    body = {}
    if reason:
        body["abort_reason"] = reason
    return await call_api("POST", url, body)

@mcp.tool()
async def rebuild_pipeline(app_slug: str, pipeline_id: str) -> str:
    """Rebuild a pipeline.
    
    Args:
        app_slug: Identifier of the Bitrise app
        pipeline_id: Identifier of the pipeline
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/pipelines/{pipeline_id}/rebuild"
    return await call_api("POST", url, {})

# ===== App Roles =====

@mcp.tool()
async def list_group_roles(app_slug: str, role_name: str) -> str:
    """List group roles for an app.
    
    Args:
        app_slug: Identifier of the Bitrise app
        role_name: Name of the role
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/roles/{role_name}"
    return await call_api("GET", url)

@mcp.tool()
async def replace_group_roles(app_slug: str, role_name: str, group_slugs: List[str]) -> str:
    """Replace group roles for an app.
    
    Args:
        app_slug: Identifier of the Bitrise app
        role_name: Name of the role
        group_slugs: List of group slugs
    """
    url = f"{BITRISE_API_BASE}/apps/{app_slug}/roles/{role_name}"
    body = group_slugs
    return await call_api("PUT", url, body)


# ==== Workspaces ====
@mcp.tool()
async def list_workspaces() -> str:
    """List the workspaces the user has access to"""
    url = f"{BITRISE_API_BASE}/organizations"
    return await call_api("GET", url)

@mcp.tool()
async def get_workspace(workspace_slug: str) -> str:
    """Get details for one workspace

    Args:
        workspace_slug: Slug of the Bitrise workspace
    """
    url = f"{BITRISE_API_BASE}/organizations/{workspace_slug}"
    return await call_api("GET", url)

@mcp.tool()
async def get_workspace_groups(workspace_slug: str) -> str:
    """Get the groups in a workspace

    Args:
        workspace_slug: Slug of the Bitrise workspace
    """
    url = f"{BITRISE_API_BASE}/organizations/{workspace_slug}/groups"
    return await call_api("GET", url)

@mcp.tool()
async def create_workspace_group(workspace_slug: str, group_name: str) -> str:
    """Get the groups in a workspace

    Args:
        workspace_slug: Slug of the Bitrise workspace
        group_name: Name of the group
    """
    url = f"{BITRISE_API_BASE}/organizations/{workspace_slug}/groups"
    return await call_api("POST", url, {"name": group_name})

@mcp.tool()
async def get_workspace_members(workspace_slug: str) -> str:
    """Get the groups in a workspace

    Args:
        workspace_slug: Slug of the Bitrise workspace
    """
    url = f"{BITRISE_API_BASE}/organizations/{workspace_slug}/members"
    return await call_api("GET", url)

@mcp.tool()
async def invite_member_to_workspace(workspace_slug: str, email: str) -> str:
    """Get the groups in a workspace

    Args:
        workspace_slug: Slug of the Bitrise workspace
        email: Email address of the user
    """
    url = f"{BITRISE_API_BASE}/organizations/{workspace_slug}/members"
    return await call_api("POST", url, {"email": email})

@mcp.tool()
async def add_member_to_group(group_slug: str, user_slug: str) -> str:
    """Get the groups in a workspace

    Args:
        workspace_slug: Slug of the Bitrise workspace
        user_slug: Slug of the user
    """
    url = f"{BITRISE_API_BASE}/groups/{group_slug}/add_member"
    return await call_api("POST", url, {"user_id": user_slug})

@mcp.tool()
async def me() -> str:
    """Get info from the currently authenticated user account

    Args:
    """
    url = f"{BITRISE_API_BASE}/me"
    return await call_api("GET", url)

if __name__ == "__main__":
    mcp.run(transport='stdio')
