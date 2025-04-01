# Bitrise MCP Server

MCP Server for the Bitrise API, enabling app management, build operations, artifact management, iOS/Android code signing, and more.

### Features

- **Comprehensive API Access**: Full access to Bitrise APIs including apps, builds, artifacts, and more
- **Authentication Support**: Secure API token-based access to Bitrise resources
- **Detailed Documentation**: Well-documented tools with parameter descriptions

## Tools

1. `list_apps`
   - List all the apps available for the authenticated account
   - Inputs:
     - `sort_by` (optional string): Order of the apps: last_build_at (default) or created_at
     - `next` (optional string): Slug of the first app in the response
     - `limit` (optional number): Max number of elements per page (default: 50)
   - Returns: List of Bitrise apps

2. `register_app`
   - Add a new app to Bitrise
   - Inputs:
     - `repo_url` (string): Repository URL
     - `is_public` (boolean): Whether the app's builds visibility is "public"
     - `repo_provider` (string): Repository provider (github, gitlab, etc.)
     - `git_owner` (string): Owner of the repository
     - `git_repo_slug` (string): Repository slug
     - `project_type` (optional string): Type of project (ios, android, etc.)
   - Returns: Details of the created app

3. `get_app`
   - Get the details of a specific app
   - Inputs:
     - `app_slug` (string): Identifier of the Bitrise app
   - Returns: App details

4. `delete_app`
   - Delete an app from Bitrise
   - Inputs:
     - `app_slug` (string): Identifier of the Bitrise app
   - Returns: Deletion status

5. `update_app`
   - Update an app
   - Inputs:
     - `app_slug` (string): Identifier of the Bitrise app
     - `is_public` (boolean): Whether the app's builds visibility is "public"
     - `project_type` (string): Type of project
     - `provider` (string): Repository provider
     - `repo_url` (string): Repository URL
   - Returns: Updated app details

6. `get_bitrise_yml`
   - Get the current Bitrise YML config file of a specified Bitrise app
   - Inputs:
     - `app_slug` (string): Identifier of the Bitrise app
   - Returns: Bitrise YML configuration

7. `update_bitrise_yml`
   - Update the Bitrise YML config file of a specified Bitrise app
   - Inputs:
     - `app_slug` (string): Identifier of the Bitrise app
     - `bitrise_yml_as_json` (object): The new Bitrise YML config file content as JSON
   - Returns: Update status

8. `list_branches`
   - List the branches with existing builds of an app's repository
   - Inputs:
     - `app_slug` (string): Identifier of the Bitrise app
   - Returns: List of branches

9. `register_ssh_key`
   - Add an SSH-key to a specific app
   - Inputs:
     - `app_slug` (string): Identifier of the Bitrise app
     - `auth_ssh_private_key` (string): Private SSH key
     - `auth_ssh_public_key` (string): Public SSH key
     - `is_register_key_into_provider_service` (boolean): Register the key in the provider service
   - Returns: Registration status

10. `register_webhook`
    - Register an incoming webhook for a specific application
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
    - Returns: Webhook registration details

11. `list_builds`
    - List all the builds of a specified Bitrise app or all accessible builds
    - Inputs:
      - `app_slug` (optional string): Identifier of the Bitrise app
      - `sort_by` (optional string): Order of builds: created_at (default), running_first
      - `branch` (optional string): Filter builds by branch
      - `workflow` (optional string): Filter builds by workflow
      - `status` (optional number): Filter builds by status (0: not finished, 1: successful, 2: failed, 3: aborted, 4: in-progress)
      - `next` (optional string): Slug of the first build in the response
      - `limit` (optional number): Max number of elements per page (default: 50)
    - Returns: List of builds

12. `trigger_bitrise_build`
    - Trigger a new build/pipeline for a specified Bitrise app
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
      - `branch` (optional string): The branch to build (default: main)
      - `workflow_id` (optional string): The workflow to build
      - `commit_message` (optional string): The commit message for the build
      - `commit_hash` (optional string): The commit hash for the build
    - Returns: Build trigger details

13. `get_build`
    - Get a specific build of a given app
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
      - `build_slug` (string): Identifier of the build
    - Returns: Build details

14. `abort_build`
    - Abort a specific build
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
      - `build_slug` (string): Identifier of the build
      - `reason` (optional string): Reason for aborting the build
    - Returns: Abort status

15. `get_build_log`
    - Get the build log of a specified build of a Bitrise app
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
      - `build_slug` (string): Identifier of the Bitrise build
    - Returns: Build log contents

16. `get_build_bitrise_yml`
    - Get the bitrise.yml of a build
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
      - `build_slug` (string): Identifier of the build
    - Returns: Build-specific bitrise.yml

17. `list_build_workflows`
    - List the workflows of an app
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
    - Returns: List of available workflows

18. `list_artifacts`
    - Get a list of all build artifacts
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
      - `build_slug` (string): Identifier of the build
      - `next` (optional string): Slug of the first artifact in the response
      - `limit` (optional number): Max number of elements per page (default: 50)
    - Returns: List of build artifacts

19. `get_artifact`
    - Get a specific build artifact
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
      - `build_slug` (string): Identifier of the build
      - `artifact_slug` (string): Identifier of the artifact
    - Returns: Artifact details

20. `delete_artifact`
    - Delete a build artifact
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
      - `build_slug` (string): Identifier of the build
      - `artifact_slug` (string): Identifier of the artifact
    - Returns: Deletion status

21. `update_artifact`
    - Update a build artifact
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
      - `build_slug` (string): Identifier of the build
      - `artifact_slug` (string): Identifier of the artifact
      - `is_public_page_enabled` (boolean): Enable public page for the artifact
    - Returns: Updated artifact details

22. `list_addons`
    - List all the available Bitrise addons
    - Inputs: None
    - Returns: List of available addons

23. `get_addon`
    - Show details of a specific Bitrise addon
    - Inputs:
      - `addon_id` (string): Identifier of the addon
    - Returns: Addon details

24. `list_app_addons`
    - List all the provisioned addons for an app
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
    - Returns: List of provisioned addons

25. `list_build_certificates`
    - Get a list of the build certificates
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
    - Returns: List of build certificates

26. `create_build_certificate`
    - Create a build certificate
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
      - `upload_file_name` (string): Name of the certificate file
      - `upload_file_size` (number): Size of the certificate file
      - `upload_content_type` (string): Content type of the certificate file
      - `certificate_password` (string): Password of the certificate
    - Returns: Creation status

27. `list_provisioning_profiles`
    - Get a list of the provisioning profiles
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
    - Returns: List of provisioning profiles

28. `create_provisioning_profile`
    - Create a provisioning profile
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
      - `upload_file_name` (string): Name of the provisioning profile file
      - `upload_file_size` (number): Size of the provisioning profile file
      - `upload_content_type` (string): Content type of the provisioning profile file
    - Returns: Creation status

29. `list_android_keystore_files`
    - Get a list of the android keystore files
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
    - Returns: List of Android keystore files

30. `create_android_keystore_file`
    - Create an Android keystore file
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
      - `upload_file_name` (string): Name of the keystore file
      - `upload_file_size` (number): Size of the keystore file
      - `upload_content_type` (string): Content type of the keystore file
      - `alias` (string): Alias of the keystore
      - `password` (string): Password of the keystore
      - `private_key_password` (string): Private key password of the keystore
    - Returns: Creation status

31. `list_outgoing_webhooks`
    - List the outgoing webhooks of an app
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
    - Returns: List of outgoing webhooks

32. `create_outgoing_webhook`
    - Create an outgoing webhook for an app
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
      - `events` (array): List of events to trigger the webhook
      - `url` (string): URL of the webhook
      - `headers` (optional object): Headers to be sent with the webhook
    - Returns: Creation status

33. `list_cache_items`
    - List the key-value cache items belonging to an app
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
    - Returns: List of cache items

34. `delete_all_cache_items`
    - Delete all key-value cache items belonging to an app
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
    - Returns: Deletion status

35. `delete_cache_item`
    - Delete a key-value cache item
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
      - `cache_item_id` (string): Identifier of the cache item
    - Returns: Deletion status

36. `get_cache_item_download_url`
    - Get the download URL of a key-value cache item
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
      - `cache_item_id` (string): Identifier of the cache item
    - Returns: Download URL

37. `list_pipelines`
    - List all pipelines and standalone builds of an app
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
    - Returns: List of pipelines

38. `get_pipeline`
    - Get a pipeline of a given app
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
      - `pipeline_id` (string): Identifier of the pipeline
    - Returns: Pipeline details

39. `abort_pipeline`
    - Abort a pipeline
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
      - `pipeline_id` (string): Identifier of the pipeline
      - `reason` (optional string): Reason for aborting the pipeline
    - Returns: Abort status

40. `rebuild_pipeline`
    - Rebuild a pipeline
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
      - `pipeline_id` (string): Identifier of the pipeline
    - Returns: Rebuild status

41. `list_group_roles`
    - List group roles for an app
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
      - `role_name` (string): Name of the role
    - Returns: List of group roles

42. `replace_group_roles`
    - Replace group roles for an app
    - Inputs:
      - `app_slug` (string): Identifier of the Bitrise app
      - `role_name` (string): Name of the role
      - `group_slugs` (array): List of group slugs
    - Returns: Replacement status

## Setup

### Bitrise API Token
[Create a Bitrise API Token](https://devcenter.bitrise.io/api/authentication) with appropriate permissions:
   - Go to your [Bitrise Profile Settings](https://app.bitrise.io/me/profile#/security)
   - Navigate to the "API access tokens" section
   - Generate a new API token with the required scopes
   - Copy the generated token

### Usage with Claude Desktop
To use this with Claude Desktop, add the following to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "bitrise": {
      "command": "uv",
      "env": {
        "BITRISE_TOKEN": "<your Bitrise API token>"
      },
      "args": [
        "--directory",
        "<full path to>/bitrise-mcp",
        "run",
        "main_new.py"
      ]
    }
  }
}
```

## License

This MCP server is licensed under the MIT License. This means you are free to use, modify, and distribute the software, subject to the terms and conditions of the MIT License. For more details, please see the LICENSE file in the project repository.
