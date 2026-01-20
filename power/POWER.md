---
name: "bitrise-ci"
displayName: "Bitrise CI/CD Platform"
description: "Control Bitrise CI/CD platform with 63 tools for app management, builds, artifacts, workspaces, pipelines, and more"
keywords: ["bitrise", "ci", "cd", "build", "mobile", "ios", "android", "pipeline", "workflow", "artifact", "test", "release", "app", "workspace"]
author: "Bitrise"
---

# Bitrise CI/CD Platform

MCP Server for the Bitrise API, enabling app management, build operations, artifact management, and more.

## Features

- **Comprehensive API Access**: Access to Bitrise APIs including apps, builds, artifacts, and more.
- **Authentication Support**: Secure API token-based access to Bitrise resources.
- **Detailed Documentation**: Well-documented tools with parameter descriptions.

## Authentication Setup

### Prerequisites
- [Create a Bitrise API Token](https://devcenter.bitrise.io/api/authentication):
  - Go to your [Bitrise Account Settings/Security](https://app.bitrise.io/me/account/security).
  - Navigate to the "Personal access tokens" section.
  - Copy the generated token.

### Configuration
The Bitrise MCP server requires a Personal Access Token (PAT) for authentication. This token should be configured as `BITRISE_TOKEN` environment variable.

## Available Tools

The Bitrise MCP server provides 63 tools organized into the following categories:

### Apps (11 tools)

1. **list_apps** - List all the apps available for the authenticated account
   - `sort_by` (optional): Order of the apps: last_build_at (default) or created_at
   - `next` (optional): Slug of the first app in the response
   - `limit` (optional): Max number of elements per page (default: 50)

2. **register_app** - Add a new app to Bitrise
   - `repo_url`: Repository URL
   - `is_public`: Whether the app's builds visibility is "public"
   - `organization_slug`: The organization (aka workspace) the app to add to
   - `project_type` (optional): Type of project (ios, android, etc.)
   - `provider` (optional): github

3. **finish_bitrise_app** - Finish the setup of a Bitrise app
   - `app_slug`: The slug of the Bitrise app to finish setup for
   - `project_type` (optional): The type of project (e.g., android, ios, flutter, etc.)
   - `stack_id` (optional): The stack ID to use for the app
   - `mode` (optional): The mode of setup
   - `config` (optional): The configuration to use for the app

4. **get_app** - Get the details of a specific app
   - `app_slug`: Identifier of the Bitrise app

5. **delete_app** - Delete an app from Bitrise
   - `app_slug`: Identifier of the Bitrise app

6. **update_app** - Update an app
   - `app_slug`: Identifier of the Bitrise app
   - `is_public`: Whether the app's builds visibility is "public"
   - `project_type`: Type of project
   - `provider`: Repository provider
   - `repo_url`: Repository URL

7. **get_bitrise_yml** - Get the current Bitrise YML config file of a specified Bitrise app
   - `app_slug`: Identifier of the Bitrise app

8. **update_bitrise_yml** - Update the Bitrise YML config file of a specified Bitrise app
   - `app_slug`: Identifier of the Bitrise app
   - `bitrise_yml_as_json`: The new Bitrise YML config file content

9. **list_branches** - List the branches with existing builds of an app's repository
   - `app_slug`: Identifier of the Bitrise app

10. **register_ssh_key** - Add an SSH-key to a specific app
    - `app_slug`: Identifier of the Bitrise app
    - `auth_ssh_private_key`: Private SSH key
    - `auth_ssh_public_key`: Public SSH key
    - `is_register_key_into_provider_service`: Register the key in the provider service

11. **register_webhook** - Register an incoming webhook for a specific application
    - `app_slug`: Identifier of the Bitrise app

### Builds (8 tools)

12. **list_builds** - List all the builds of a specified Bitrise app or all accessible builds
    - `app_slug` (optional): Identifier of the Bitrise app
    - `sort_by` (optional): Order of builds: created_at (default), running_first
    - `branch` (optional): Filter builds by branch
    - `workflow` (optional): Filter builds by workflow
    - `status` (optional): Filter builds by status (0: not finished, 1: successful, 2: failed, 3: aborted, 4: in-progress)
    - `next` (optional): Slug of the first build in the response
    - `limit` (optional): Max number of elements per page (default: 50)

13. **trigger_bitrise_build** - Trigger a new build/pipeline for a specified Bitrise app
    - `app_slug`: Identifier of the Bitrise app
    - `branch` (optional): The branch to build (default: main)
    - `pipeline_id` (optional): The pipeline to build
    - `workflow_id` (optional): The workflow to build
    - `commit_message` (optional): The commit message for the build
    - `commit_hash` (optional): The commit hash for the build
    - `environments` (optional): Custom environment variables for the build (array of objects with `mapped_to`, `value`, and optional `is_expand` properties)

14. **get_build** - Get a specific build of a given app
    - `app_slug`: Identifier of the Bitrise app
    - `build_slug`: Identifier of the build

15. **abort_build** - Abort a specific build
    - `app_slug`: Identifier of the Bitrise app
    - `build_slug`: Identifier of the build
    - `reason` (optional): Reason for aborting the build

16. **get_build_log** - Get the build log of a specified build of a Bitrise app
    - `app_slug`: Identifier of the Bitrise app
    - `build_slug`: Identifier of the Bitrise build

17. **get_build_bitrise_yml** - Get the bitrise.yml of a build
    - `app_slug`: Identifier of the Bitrise app
    - `build_slug`: Identifier of the build

18. **list_build_workflows** - List the workflows of an app
    - `app_slug`: Identifier of the Bitrise app

19. **get_build_steps** - Get step statuses of a specific build of a given app
    - `app_slug`: Identifier of the Bitrise app
    - `build_slug`: Identifier of the build

### Artifacts (4 tools)

20. **list_artifacts** - Get a list of all build artifacts
    - `app_slug`: Identifier of the Bitrise app
    - `build_slug`: Identifier of the build
    - `next` (optional): Slug of the first artifact in the response
    - `limit` (optional): Max number of elements per page (default: 50)

21. **get_artifact** - Get a specific build artifact
    - `app_slug`: Identifier of the Bitrise app
    - `build_slug`: Identifier of the build
    - `artifact_slug`: Identifier of the artifact

22. **delete_artifact** - Delete a build artifact
    - `app_slug`: Identifier of the Bitrise app
    - `build_slug`: Identifier of the build
    - `artifact_slug`: Identifier of the artifact

23. **update_artifact** - Update a build artifact
    - `app_slug`: Identifier of the Bitrise app
    - `build_slug`: Identifier of the build
    - `artifact_slug`: Identifier of the artifact
    - `is_public_page_enabled`: Enable public page for the artifact

### Outgoing Webhooks (4 tools)

24. **list_outgoing_webhooks** - List the outgoing webhooks of an app
    - `app_slug`: Identifier of the Bitrise app

25. **delete_outgoing_webhook** - Delete the outgoing webhook of an app
    - `app_slug`: Identifier of the Bitrise app
    - `webhook_slug`: Identifier of the webhook

26. **update_outgoing_webhook** - Update an outgoing webhook for an app
    - `app_slug`: Identifier of the Bitrise app
    - `webhook_slug`: Identifier of the webhook
    - `events`: List of events to trigger the webhook
    - `url`: URL of the webhook
    - `headers` (optional): Headers to be sent with the webhook

27. **create_outgoing_webhook** - Create an outgoing webhook for an app
    - `app_slug`: Identifier of the Bitrise app
    - `events`: List of events to trigger the webhook
    - `url`: URL of the webhook
    - `headers` (optional): Headers to be sent with the webhook

### Cache Items (4 tools)

28. **list_cache_items** - List the key-value cache items belonging to an app
    - `app_slug`: Identifier of the Bitrise app

29. **delete_all_cache_items** - Delete all key-value cache items belonging to an app
    - `app_slug`: Identifier of the Bitrise app

30. **delete_cache_item** - Delete a key-value cache item
    - `app_slug`: Identifier of the Bitrise app
    - `cache_item_id`: Identifier of the cache item

31. **get_cache_item_download_url** - Get the download URL of a key-value cache item
    - `app_slug`: Identifier of the Bitrise app
    - `cache_item_id`: Identifier of the cache item

### Pipelines (4 tools)

32. **list_pipelines** - List all pipelines and standalone builds of an app
    - `app_slug`: Identifier of the Bitrise app

33. **get_pipeline** - Get a pipeline of a given app
    - `app_slug`: Identifier of the Bitrise app
    - `pipeline_id`: Identifier of the pipeline

34. **abort_pipeline** - Abort a pipeline
    - `app_slug`: Identifier of the Bitrise app
    - `pipeline_id`: Identifier of the pipeline
    - `reason` (optional): Reason for aborting the pipeline

35. **rebuild_pipeline** - Rebuild a pipeline
    - `app_slug`: Identifier of the Bitrise app
    - `pipeline_id`: Identifier of the pipeline

### Group Roles (2 tools)

36. **list_group_roles** - List group roles for an app
    - `app_slug`: Identifier of the Bitrise app
    - `role_name`: Name of the role

37. **replace_group_roles** - Replace group roles for an app
    - `app_slug`: Identifier of the Bitrise app
    - `role_name`: Name of the role
    - `group_slugs`: List of group slugs

### Workspaces (7 tools)

38. **list_workspaces** - List the workspaces the user has access to

39. **get_workspace** - Get details for one workspace
    - `workspace_slug`: Slug of the Bitrise workspace

40. **get_workspace_groups** - Get the groups in a workspace
    - `workspace_slug`: Slug of the Bitrise workspace

41. **create_workspace_group** - Create a group in a workspace
    - `workspace_slug`: Slug of the Bitrise workspace
    - `group_name`: Name of the group

42. **get_workspace_members** - Get the members in a workspace
    - `workspace_slug`: Slug of the Bitrise workspace

43. **invite_member_to_workspace** - Invite a member to a workspace
    - `workspace_slug`: Slug of the Bitrise workspace
    - `email`: Email address of the user

44. **add_member_to_group** - Add a member to a group
    - `group_slug`: Slug of the group
    - `user_slug`: Slug of the user

### Account (1 tool)

45. **me** - Get info from the currently authenticated user account

### Release Management (18 tools)

46. **create_connected_app** - Add a new Release Management connected app to Bitrise
    - `platform`: The mobile platform for the connected app (ios/android)
    - `store_app_id`: The app store identifier for the connected app
    - `workspace_slug`: Identifier of the Bitrise workspace
    - `id` (optional): An uuidV4 identifier for your new connected app
    - `manual_connection` (optional): Indicates a manual connection
    - `project_id` (optional): Specifies which Bitrise Project to associate with
    - `store_app_name` (optional): App name for manual connections
    - `store_credential_id` (optional): Selection of credentials added on Bitrise

47. **list_connected_apps** - List Release Management connected apps available for the authenticated account within a workspace
    - `workspace_slug`: Identifier of the Bitrise workspace
    - `items_per_page` (optional): Maximum number of connected apps per page
    - `page` (optional): Page number to return
    - `platform` (optional): Filter for a specific mobile platform
    - `project_id` (optional): Filter for a specific Bitrise Project
    - `search` (optional): Search by bundle ID, package name, or app title

48. **get_connected_app** - Gives back a Release Management connected app for the authenticated account
    - `id`: Identifier of the Release Management connected app

49. **update_connected_app** - Updates a connected app
    - `connected_app_id`: The uuidV4 identifier for your connected app
    - `store_app_id`: The store identifier for your app
    - `connect_to_store` (optional): Check validity against the App Store or Google Play
    - `store_credential_id` (optional): Selection of credentials added on Bitrise

50. **list_installable_artifacts** - List Release Management installable artifacts of a connected app
    - `connected_app_id`: Identifier of the Release Management connected app
    - `after_date` (optional): Start of the interval for artifact creation/upload
    - `artifact_type` (optional): Filter for a specific artifact type
    - `before_date` (optional): End of the interval for artifact creation/upload
    - `branch` (optional): Filter for the Bitrise CI branch
    - `distribution_ready` (optional): Filter for distribution ready artifacts
    - `items_per_page` (optional): Maximum number of artifacts per page
    - `page` (optional): Page number to return
    - `platform` (optional): Filter for a specific mobile platform
    - `search` (optional): Search by version, filename or build number
    - `source` (optional): Filter for the source of installable artifacts
    - `store_signed` (optional): Filter for store ready installable artifacts
    - `version` (optional): Filter for a specific version
    - `workflow` (optional): Filter for a specific Bitrise CI workflow

51. **generate_installable_artifact_upload_url** - Generates a signed upload URL for an installable artifact to be uploaded to Bitrise
    - `connected_app_id`: Identifier of the Release Management connected app
    - `installable_artifact_id`: An uuidv4 identifier for the installable artifact
    - `file_name`: The name of the installable artifact file
    - `file_size_bytes`: The byte size of the installable artifact file
    - `branch` (optional): Name of the CI branch
    - `with_public_page` (optional): Enable public install page
    - `workflow` (optional): Name of the CI workflow

52. **get_installable_artifact_upload_and_proc_status** - Gets the processing and upload status of an installable artifact
    - `connected_app_id`: Identifier of the Release Management connected app
    - `installable_artifact_id`: The uuidv4 identifier for the installable artifact

53. **set_installable_artifact_public_install_page** - Changes whether public install page should be available for the installable artifact
    - `connected_app_id`: Identifier of the Release Management connected app
    - `installable_artifact_id`: The uuidv4 identifier for the installable artifact
    - `with_public_page`: Boolean flag for enabling/disabling public install page

54. **list_build_distribution_versions** - Lists Build Distribution versions available for testers
    - `connected_app_id`: The uuidV4 identifier of the connected app
    - `items_per_page` (optional): Maximum number of versions per page
    - `page` (optional): Page number to return

55. **list_build_distribution_version_test_builds** - Gives back a list of test builds for the given build distribution version
    - `connected_app_id`: The uuidV4 identifier of the connected app
    - `version`: The version of the build distribution
    - `items_per_page` (optional): Maximum number of test builds per page
    - `page` (optional): Page number to return

56. **create_tester_group** - Creates a tester group for a Release Management connected app
    - `connected_app_id`: The uuidV4 identifier of the connected app
    - `name`: The name for the new tester group
    - `auto_notify` (optional): Indicates automatic notifications for the group

57. **notify_tester_group** - Notifies a tester group about a new test build
    - `connected_app_id`: The uuidV4 identifier of the connected app
    - `id`: The uuidV4 identifier of the tester group
    - `test_build_id`: The unique identifier of the test build

58. **add_testers_to_tester_group** - Adds testers to a tester group of a connected app
    - `connected_app_id`: The uuidV4 identifier of the connected app
    - `id`: The uuidV4 identifier of the tester group
    - `user_slugs`: The list of users identified by slugs to be added

59. **update_tester_group** - Updates the given tester group settings
    - `connected_app_id`: The uuidV4 identifier of the connected app
    - `id`: The uuidV4 identifier of the tester group
    - `auto_notify` (optional): Setting for automatic email notifications
    - `name` (optional): The new name for the tester group

60. **list_tester_groups** - Gives back a list of tester groups related to a specific connected app
    - `connected_app_id`: The uuidV4 identifier of the connected app
    - `items_per_page` (optional): Maximum number of tester groups per page
    - `page` (optional): Page number to return

61. **get_tester_group** - Gives back the details of the selected tester group
    - `connected_app_id`: The uuidV4 identifier of the connected app
    - `id`: The uuidV4 identifier of the tester group

62. **get_potential_testers** - Gets a list of potential testers who can be added to a specific tester group
    - `connected_app_id`: The uuidV4 identifier of the connected app
    - `id`: The uuidV4 identifier of the tester group
    - `items_per_page` (optional): Maximum number of potential testers per page
    - `page` (optional): Page number to return
    - `search` (optional): Search for testers by email or username

63. **get_testers** - Gets a list of testers that have been associated with a tester group related to a specific connected app
    - `connected_app_id`: The uuidV4 identifier of the connected app
    - `tester_group_id` (optional): The uuidV4 identifier of a tester group. If given, only testers within this specific tester group will be returned
    - `items_per_page` (optional): Maximum number of testers per page (default: 10)
    - `page` (optional): Page number to return (default: 1)

## Advanced Configuration

You can limit the number of tools exposed to the MCP client. This is useful if you want to optimize token usage or your MCP client has a limit on the number of tools.

Tools are grouped by their "API group", and you can pass the groups you want to expose as tools. Possible values: `apps, builds, workspaces, outgoing-webhooks, artifacts, group-roles, cache-items, pipelines, account, read-only, release-management`.

We recommend using the `release-management` API group separately to avoid any confusion with the `apps` API group.

By default, all API groups are enabled. You can specify which groups to enable using the `ENABLED_API_GROUPS` environment variable for local (stdio) servers or the `x-bitrise-enabled-api-groups` HTTP header for remote (Streamable HTTP) servers with a comma-separated list of group names.

## Resources

- [Bitrise API Documentation](https://devcenter.bitrise.io/api/api-index/)
- [Bitrise Account Settings](https://app.bitrise.io/me/account/security)
- [MCP Server Repository](https://github.com/bitrise-io/bitrise-mcp)