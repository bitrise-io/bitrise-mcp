## Advanced configuration

You can limit the number of tools exposed to the MCP client. This is useful if you want to optimize token usage or your MCP client has a limit on the number of tools.

Tools are grouped by their "API group", and you can pass the groups you want to expose as tools. Possible values: `apps, builds, workspaces, outgoing-webhooks, artifacts, group-roles, cache-items, pipelines, account, read-only, release-management, configuration, release-management-code-push, release-management-store-releases, insights`.

We recommend using the `release-management` API group separately to avoid any confusion with the `apps` API group.

By default, all API groups are enabled. You can specify which groups to enable using the `ENABLED_API_GROUPS` environment variable for local (stdio) servers or the `x-bitrise-enabled-api-groups` HTTP header for remote (Streamable HTTP) servers with a comma-separated list of group names.

## Tools

### Apps

1. `list_apps`
   - List all the apps available for the authenticated account
   - Arguments:
     - `sort_by` (optional): Order of the apps: last_build_at (default) or created_at
     - `next` (optional): Slug of the first app in the response
     - `limit` (optional): Max number of elements per page (default: 50)

2. `register_app`
   - Add a new app to Bitrise
   - Arguments:
     - `repo_url`: Repository URL
     - `is_public`: Whether the app's builds visibility is "public"
     - `organization_slug`: The organization (aka workspace) the app to add to
     - `project_type` (optional): Type of project (ios, android, etc.)
     - `provider` (optional): github

3. `finish_bitrise_app`
   - Finish the setup of a Bitrise app
   - Arguments:
     - `app_slug`: The slug of the Bitrise app to finish setup for
     - `project_type` (optional): The type of project (e.g., android, ios, flutter, etc.)
     - `stack_id` (optional): The stack ID to use for the app
     - `mode` (optional): The mode of setup
     - `config` (optional): The configuration to use for the app

4. `get_app`
   - Get the details of a specific app
   - Arguments:
     - `app_slug`: Identifier of the Bitrise app

5. `delete_app`
   - Delete an app from Bitrise
   - Arguments:
     - `app_slug`: Identifier of the Bitrise app

6. `update_app`
   - Update an app
   - Arguments:
     - `app_slug`: Identifier of the Bitrise app
     - `is_public`: Whether the app's builds visibility is "public"
     - `project_type`: Type of project
     - `provider`: Repository provider
     - `repo_url`: Repository URL

7. `get_bitrise_yml`
   - Get the current Bitrise YML config file of a specified Bitrise app
   - Arguments:
     - `app_slug`: Identifier of the Bitrise app

8. `update_bitrise_yml`
   - Update the Bitrise YML config file of a specified Bitrise app
   - Arguments:
     - `app_slug`: Identifier of the Bitrise app
     - `bitrise_yml_as_json`: The new Bitrise YML config file content

9. `list_branches`
   - List the branches with existing builds of an app's repository
   - Arguments:
     - `app_slug`: Identifier of the Bitrise app

10. `register_ssh_key`
    - Add an SSH-key to a specific app
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app
      - `auth_ssh_private_key`: Private SSH key
      - `auth_ssh_public_key`: Public SSH key
      - `is_register_key_into_provider_service`: Register the key in the provider service

11. `register_webhook`
    - Register an incoming webhook for a specific application
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app

### Builds

12. `list_builds`
    - List all the builds of a specified Bitrise app or all accessible builds
    - Arguments:
      - `app_slug` (optional): Identifier of the Bitrise app
      - `sort_by` (optional): Order of builds: created_at (default), running_first
      - `branch` (optional): Filter builds by branch
      - `workflow` (optional): Filter builds by workflow
      - `status` (optional): Filter builds by status (0: not finished, 1: successful, 2: failed, 3: aborted, 4: in-progress)
      - `next` (optional): Slug of the first build in the response
      - `limit` (optional): Max number of elements per page (default: 50)

13. `trigger_bitrise_build`
    - Trigger a new build/pipeline for a specified Bitrise app
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app
      - `branch` (optional): The branch to build (default: main)
      - `pipeline_id` (optional): The pipeline to build
      - `workflow_id` (optional): The workflow to build
      - `pipeline_id` (optional): The pipeline to build
      - `commit_message` (optional): The commit message for the build
      - `commit_hash` (optional): The commit hash for the build
      - `stack` (optional): Stack to run the build on, overriding the workflow's `meta.bitrise.io.stack` for this build only (e.g. "osx-xcode-16.0.x")
      - `environments` (optional): Custom environment variables for the build (array of objects with `mapped_to`, `value`, and optional `is_expand` properties)

14. `get_build`
    - Get a specific build of a given app
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app
      - `build_slug`: Identifier of the build

15. `abort_build`
    - Abort a specific build
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app
      - `build_slug`: Identifier of the build
      - `reason` (optional): Reason for aborting the build

16. `get_build_log`
    - Get the build log of a specified build of a Bitrise app
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app
      - `build_slug`: Identifier of the Bitrise build

17. `get_build_bitrise_yml`
    - Get the bitrise.yml of a build
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app
      - `build_slug`: Identifier of the build

18. `list_build_workflows`
    - List the workflows of an app
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app

19. `get_build_steps`
    - Get step statuses of a specific build of a given app
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app
      - `build_slug`: Identifier of the build

### Artifacts

20. `list_artifacts`
    - Get a list of all build artifacts
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app
      - `build_slug`: Identifier of the build
      - `next` (optional): Slug of the first artifact in the response
      - `limit` (optional): Max number of elements per page (default: 50)

20. `get_artifact`
    - Get a specific build artifact
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app
      - `build_slug`: Identifier of the build
      - `artifact_slug`: Identifier of the artifact

21. `delete_artifact`
    - Delete a build artifact
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app
      - `build_slug`: Identifier of the build
      - `artifact_slug`: Identifier of the artifact

22. `update_artifact`
    - Update a build artifact
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app
      - `build_slug`: Identifier of the build
      - `artifact_slug`: Identifier of the artifact
      - `is_public_page_enabled`: Enable public page for the artifact

### Outgoing Webhooks

24. `list_outgoing_webhooks`
    - List the outgoing webhooks of an app
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app

25. `delete_outgoing_webhook`
    - Delete the outgoing webhook of an app
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app
      - `webhook_slug`: Identifier of the webhook

26. `update_outgoing_webhook`
    - Update an outgoing webhook for an app
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app
      - `webhook_slug`: Identifier of the webhook
      - `events`: List of events to trigger the webhook
      - `url`: URL of the webhook
      - `headers` (optional): Headers to be sent with the webhook

27. `create_outgoing_webhook`
    - Create an outgoing webhook for an app
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app
      - `events`: List of events to trigger the webhook
      - `url`: URL of the webhook
      - `headers` (optional): Headers to be sent with the webhook

### Cache Items

28. `list_cache_items`
    - List the key-value cache items belonging to an app
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app

29. `delete_all_cache_items`
    - Delete all key-value cache items belonging to an app
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app

30. `delete_cache_item`
    - Delete a key-value cache item
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app
      - `cache_item_id`: Identifier of the cache item

31. `get_cache_item_download_url`
    - Get the download URL of a key-value cache item
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app
      - `cache_item_id`: Identifier of the cache item

### Pipelines

32. `list_pipelines`
    - List all pipelines and standalone builds of an app
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app

33. `get_pipeline`
    - Get a pipeline of a given app
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app
      - `pipeline_id`: Identifier of the pipeline

34. `abort_pipeline`
    - Abort a pipeline
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app
      - `pipeline_id`: Identifier of the pipeline
      - `reason` (optional): Reason for aborting the pipeline

35. `rebuild_pipeline`
    - Rebuild a pipeline
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app
      - `pipeline_id`: Identifier of the pipeline

### Group Roles

36. `list_group_roles`
    - List group roles for an app
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app
      - `role_name`: Name of the role

37. `replace_group_roles`
    - Replace group roles for an app
    - Arguments:
      - `app_slug`: Identifier of the Bitrise app
      - `role_name`: Name of the role
      - `group_slugs`: List of group slugs

### Workspaces

38. `list_workspaces`
    - List the workspaces the user has access to

39. `get_workspace`
    - Get details for one workspace
    - Arguments:
      - `workspace_slug`: Slug of the Bitrise workspace

40. `get_workspace_groups`
    - Get the groups in a workspace
    - Arguments:
      - `workspace_slug`: Slug of the Bitrise workspace

41. `create_workspace_group`
    - Create a group in a workspace
    - Arguments:
      - `workspace_slug`: Slug of the Bitrise workspace
      - `group_name`: Name of the group

42. `get_workspace_members`
    - Get the members in a workspace
    - Arguments:
      - `workspace_slug`: Slug of the Bitrise workspace

43. `invite_member_to_workspace`
    - Invite a member to a workspace
    - Arguments:
      - `workspace_slug`: Slug of the Bitrise workspace
      - `email`: Email address of the user

44. `add_member_to_group`
    - Add a member to a group
    - Arguments:
      - `group_slug`: Slug of the group
      - `user_slug`: Slug of the user

### Account

45. `me`
    - Get info from the currently authenticated user account

### Release Management

46. `create_connected_app`
   - Add a new Release Management connected app to Bitrise.
   - Arguments:
     - `platform`: The mobile platform for the connected app (ios/android).
     - `store_app_id`: The app store identifier for the connected app.
     - `workspace_slug`: Identifier of the Bitrise workspace.
     - `id`: (Optional) An uuidV4 identifier for your new connected app.
     - `manual_connection`: (Optional) Indicates a manual connection.
     - `project_id`: (Optional) Specifies which Bitrise Project to associate with.
     - `project_title`: (Optional) Title for the project created when no `project_id` is given. Rejected together with `project_id`.
     - `store_app_name`: (Optional) App name for manual connections.
     - `store_credential_id`: (Optional) Selection of credentials added on Bitrise.
     - `framework`: (Optional) The framework used to build the app (flutter/react_native/kotlin_multiplatform/other/native_ios/native_android), defaults to 'other'.
     - `tester_groups`: (Optional) Tester groups created together with the app (array of objects with `name`, and optional `type`, `auto_notify`, `user_slugs` and `emails` properties). Created one by one, so check the response's `errors` array.
     - `code_push_deployments`: (Optional) CodePush deployments created together with the app (array of objects with `name` and an optional `key` property). Created one by one, so check the response's `errors` array.

47. `list_connected_apps`
   - List Release Management connected apps available for the authenticated account within a workspace.
   - Arguments:
     - `workspace_slug`: Identifier of the Bitrise workspace.
     - `items_per_page`: (Optional) Maximum number of connected apps per page.
     - `page`: (Optional) Page number to return.
     - `platform`: (Optional) Filter for a specific mobile platform.
     - `project_id`: (Optional) Filter for a specific Bitrise Project.
     - `search`: (Optional) Search by bundle ID, package name, or app title.

48. `get_connected_app`
   - Gives back a Release Management connected app for the authenticated account.
   - Arguments:
     - `id`: Identifier of the Release Management connected app.

49. `update_connected_app`
   - Updates a connected app.
   - Arguments:
     - `connected_app_id`: The uuidV4 identifier for your connected app.
     - `store_app_id`: The store identifier for your app.
     - `connect_to_store`: (Optional) Check validity against the App Store or Google Play.
     - `store_credential_id`: (Optional) Selection of credentials added on Bitrise.

50. `list_installable_artifacts`
   - List Release Management installable artifacts of a connected app.
   - Arguments:
     - `connected_app_id`: Identifier of the Release Management connected app.
     - `after_date`: (Optional) Start of the interval for artifact creation/upload.
     - `artifact_type`: (Optional) Filter for a specific artifact type.
     - `before_date`: (Optional) End of the interval for artifact creation/upload.
     - `branch`: (Optional) Filter for the Bitrise CI branch.
     - `distribution_ready`: (Optional) Filter for distribution ready artifacts.
     - `items_per_page`: (Optional) Maximum number of artifacts per page.
     - `page`: (Optional) Page number to return.
     - `platform`: (Optional) Filter for a specific mobile platform.
     - `search`: (Optional) Search by version, filename or build number.
     - `source`: (Optional) Filter for the source of installable artifacts.
     - `store_signed`: (Optional) Filter for store ready installable artifacts.
     - `version`: (Optional) Filter for a specific version.
     - `workflow`: (Optional) Filter for a specific Bitrise CI workflow.

51. `generate_installable_artifact_upload_url`
   - Generates a signed upload URL for an installable artifact to be uploaded to Bitrise.
   - Arguments:
     - `connected_app_id`: Identifier of the Release Management connected app.
     - `installable_artifact_id`: An uuidv4 identifier for the installable artifact.
     - `file_name`: The name of the installable artifact file.
     - `file_size_bytes`: The byte size of the installable artifact file.
     - `branch`: (Optional) Name of the CI branch.
     - `with_public_page`: (Optional) Enable public install page.
     - `workflow`: (Optional) Name of the CI workflow.

52. `get_installable_artifact_upload_and_proc_status`
   - Gets the processing and upload status of an installable artifact.
   - Arguments:
     - `connected_app_id`: Identifier of the Release Management connected app.
     - `installable_artifact_id`: The uuidv4 identifier for the installable artifact.

53. `set_installable_artifact_public_install_page`
   - Changes whether public install page should be available for the installable artifact.
   - Arguments:
     - `connected_app_id`: Identifier of the Release Management connected app.
     - `installable_artifact_id`: The uuidv4 identifier for the installable artifact.
     - `with_public_page`: Boolean flag for enabling/disabling public install page.

54. `list_build_distribution_versions`
   - Lists Build Distribution versions available for testers.
   - Arguments:
     - `connected_app_id`: The uuidV4 identifier of the connected app.
     - `items_per_page`: (Optional) Maximum number of versions per page.
     - `page`: (Optional) Page number to return.

55. `list_build_distribution_version_test_builds`
   - Gives back a list of test builds for the given build distribution version.
   - Arguments:
     - `connected_app_id`: The uuidV4 identifier of the connected app.
     - `version`: The version of the build distribution.
     - `items_per_page`: (Optional) Maximum number of test builds per page.
     - `page`: (Optional) Page number to return.

56. `create_tester_group`
   - Creates a tester group for a Release Management connected app.
   - Arguments:
     - `connected_app_id`: The uuidV4 identifier of the connected app.
     - `name`: The name for the new tester group.
     - `auto_notify`: (Optional) Indicates automatic notifications for the group. Ignored for external tester groups.
     - `user_slugs`: (Optional) Slugs of workspace members to add as internal testers. Ignored for external tester groups.
     - `emails`: (Optional) Email addresses to add as testers for an external tester group (at most 1000).
     - `type`: (Optional) The type of the tester group, 'internal' or 'external' (default: internal).

57. `notify_tester_group`
   - Notifies an internal tester group about a new test build by email.
   - Arguments:
     - `connected_app_id`: The uuidV4 identifier of the connected app.
     - `id`: The uuidV4 identifier of the tester group.
     - `test_build_id`: The unique identifier of the test build.

58. `add_testers_to_tester_group`
   - Adds testers to a tester group of a connected app.
   - Arguments:
     - `connected_app_id`: The uuidV4 identifier of the connected app.
     - `id`: The uuidV4 identifier of the tester group.
     - `user_slugs`: (Optional) User slugs to add as internal testers. Required for internal tester groups; ignored for external tester groups.
     - `emails`: (Optional) Email addresses to add as external testers. Required for external tester groups (at most 1000 entries); ignored for internal tester groups.

59. `update_tester_group`
   - Updates the given tester group settings.
   - Arguments:
     - `connected_app_id`: The uuidV4 identifier of the connected app.
     - `id`: The uuidV4 identifier of the tester group.
     - `auto_notify`: (Optional) Setting for automatic email notifications.
     - `name`: (Optional) The new name for the tester group.

60. `list_tester_groups`
   - Gives back a list of tester groups related to a specific connected app.
   - Arguments:
     - `connected_app_id`: The uuidV4 identifier of the connected app.
     - `items_per_page`: (Optional) Maximum number of tester groups per page.
     - `page`: (Optional) Page number to return.
     - `type`: (Optional) Filters for a specific tester group type, 'internal' or 'external' (default: internal).
     - `installable_artifact_id`: (Optional) If given, decorates the response with notification details about that artifact for each tester group.

61. `get_tester_group`
   - Gives back the details of the selected tester group.
   - Arguments:
     - `connected_app_id`: The uuidV4 identifier of the connected app.
     - `id`: The uuidV4 identifier of the tester group.

62. `get_potential_testers`
   - Gets a list of potential testers who can be added to a tester group, or workspace members before one exists.
   - Arguments:
     - `workspace_slug`: (Optional) The slug of the workspace whose members are listed. Mutually exclusive with `connected_app_id`/`id`.
     - `project_id`: (Optional) Used with `workspace_slug`; also lists the project's outside contributors.
     - `connected_app_id`: (Optional) The uuidV4 identifier of the connected app. Required together with `id`.
     - `id`: (Optional) The uuidV4 identifier of the tester group. Required together with `connected_app_id`.
     - `items_per_page`: (Optional) Maximum number of potential testers per page, 1-1000 (default: 25).
     - `page`: (Optional) Page number to return (default: 1).
     - `search`: (Optional) Search for testers by email or username.

63. `get_testers`
   - Gets a list of testers that have been associated with a tester group related to a specific connected app.
   - Arguments:
     - `connected_app_id`: The uuidV4 identifier of the connected app.
     - `tester_group_id`: (Optional) The uuidV4 identifier of a tester group. If given, only testers within this specific tester group will be returned.
     - `items_per_page`: (Optional) Maximum number of testers per page (default: 10).
     - `page`: (Optional) Page number to return (default: 1).
     - `type`: (Optional) Filters for testers belonging to a specific tester group type, 'internal' or 'external' (default: internal).

### Configuration

64. `validate_bitrise_yml`
    - Validate a Bitrise YML config file. This endpoint checks if the provided bitrise.yml is valid.
    - Arguments:
      - `bitrise_yml`: The Bitrise YML config file content to be validated. It must be a string.
      - `app_slug` (optional): Slug of a Bitrise app. Specifying this value allows for validating the YML against workspace-specific settings like available stacks, machine types, license pools etc.

65. `step_search`
    - Find steps for building workflows or step bundles in a Bitrise YML config file. Finds steps based on name, description, tags or maintainers.
    - Arguments:
      - `query`: The phrase to search steps for like `clone`, `npm`, `deploy` etc.
      - `categories` (optional): Categories to filter steps. Available values: `build`, `code-sign`, `test`, `deploy`, `notification`, `access-control`, `artifact-info`, `installer`, `dependency`, `utility`
      - `maintainers` (optional): Filter steps by maintainers. Available values: `bitrise`, `verified`, `community`

66. `step_inputs`
    - List inputs of a step with their defaults, allowed values etc.
    - Arguments:
      - `step_ref`: Step reference formatted as `step_lib_source::step_id@version`. `step_id` and an exact `version` are required, `step_lib_source` is only necessary for custom step sources.

67. `list_available_stacks`
    - List available stacks with their machine configurations and version information. When a workspace_slug is provided, returns stacks available for that workspace including any custom stacks. When omitted, returns globally available stacks.
    - Arguments:
      - `workspace_slug` (optional): Slug of the Bitrise workspace. When provided, lists stacks available for that workspace (including custom stacks). When omitted, lists globally available stacks.

### CodePush

68. `codepush_list_deployments`
   - List CodePush deployments for a Bitrise app.
   - Arguments:
     - `app_id`: Identifier of the Bitrise app.
     - `search`: (Optional) Search deployments by name. The filter is case-sensitive.
     - `items_per_page`: (Optional) Maximum number of deployments per page (default: 10).
     - `page`: (Optional) Page number to return (default: 1).

69. `codepush_get_deployment`
   - Get a specific CodePush deployment by its ID.
   - Arguments:
     - `id`: Identifier (UUID) of the CodePush deployment.

70. `codepush_create_deployment`
   - Create a new CodePush deployment for a Bitrise app.
   - Arguments:
     - `name`: Name for the new deployment.
     - `app_id`: Identifier of the Bitrise app.
     - `key`: (Optional) Deployment key. Auto-generated if not provided.

71. `codepush_update_deployment`
   - Update the name of an existing CodePush deployment.
   - Arguments:
     - `id`: Identifier (UUID) of the CodePush deployment.
     - `name`: New name for the deployment.

72. `codepush_delete_deployment`
   - Delete a CodePush deployment. This action is irreversible.
   - Arguments:
     - `id`: Identifier (UUID) of the CodePush deployment to delete.

73. `codepush_promote_deployment`
   - Promote a package from a source deployment to a target deployment. The most recent package in the source deployment is promoted unless package_id is specified.
   - Arguments:
     - `id`: Identifier (UUID) of the source deployment.
     - `target_deployment_id`: Identifier (UUID) of the target deployment.
     - `package_id`: (Optional) UUID of a specific package to promote. Defaults to most recent.
     - `app_version`: (Optional) Semver app version constraint for the promoted package.
     - `description`: (Optional) Description for the promoted package.
     - `disabled`: (Optional) If true, clients will not download this update.
     - `mandatory`: (Optional) If true, clients must install immediately.
     - `rollout`: (Optional) Percentage (0-100) of users who receive this update.

74. `codepush_rollback_deployment`
   - Rollback a CodePush deployment to its previous version.
   - Arguments:
     - `id`: Identifier (UUID) of the CodePush deployment to rollback.
     - `package_id`: (Optional) UUID of a specific package to rollback to. Defaults to the previous package.

75. `codepush_list_updates`
   - List CodePush updates for a specific deployment.
   - Arguments:
     - `deployment_id`: Identifier (UUID) of the CodePush deployment.
     - `search`: (Optional) Search updates by label or description. The filter is case-sensitive.
     - `items_per_page`: (Optional) Maximum number of updates per page (default: 10).
     - `page`: (Optional) Page number to return (default: 1).

76. `codepush_get_update`
   - Get a specific CodePush update by its ID.
   - Arguments:
     - `id`: Identifier (UUID) of the CodePush update.

77. `codepush_patch_update`
   - Patch a CodePush update to change its disabled state, mandatory flag, or rollout percentage. Only include fields you want to change — omitted fields are left unchanged.
   - Arguments:
     - `id`: Identifier (UUID) of the CodePush update.
     - `disabled`: (Optional) Set to 'true' to disable or 'false' to re-enable.
     - `mandatory`: (Optional) Set to 'true' to make mandatory or 'false' to make optional.
     - `rollout`: (Optional) Percentage (0-100) of users who receive this update.

78. `codepush_delete_update`
   - Delete a CodePush update. This action is irreversible.
   - Arguments:
     - `id`: Identifier (UUID) of the CodePush update to delete.

79. `codepush_get_update_status`
   - Get the processing status of a CodePush update (e.g. pending, ready, failed).
   - Arguments:
     - `id`: Identifier (UUID) of the CodePush update.

80. `codepush_generate_update_upload_url`
   - Generate a signed upload URL (valid 1 hour) for uploading a CodePush update bundle. The response contains the URL, HTTP method, and headers needed for a direct upload. After uploading, check status with `codepush_get_update_status`.
   - Arguments:
     - `id`: Client-generated UUID for the new update.
     - `deployment_id`: Identifier (UUID) of the deployment this update belongs to.
     - `app_version`: Semver version of the app this update targets (e.g. '1.2.3').
     - `file_name`: File name of the update bundle to be uploaded (with extension).
     - `file_size_bytes`: Byte size of the update bundle file as a string.
     - `description`: (Optional) Description for this update.
     - `disabled`: (Optional) If true, clients will not download this update after upload.
     - `mandatory`: (Optional) If true, clients must install this update immediately.
     - `rollout`: (Optional) Percentage (0-100) of users who receive this update. Defaults to 100.

81. `codepush_get_metrics`
   - Get workspace-level CodePush usage metrics including data transfer, storage, and monthly active users, along with their limits and billing cycle information.
   - Arguments:
     - `workspace_slug`: Slug of the Bitrise workspace.

### Store Releases

Store release tools take an app version (what the Release Management UI calls a release) of a connected app through the store: selecting and uploading the release candidate, beta testing, approvals, store review and the production rollout. They call the Store Releases API (`https://api.bitrise.io/release-management/v2/store-releases/v1`), except the three App Store version tools at the end, which call the Apps API. `app_version_id` is the `id` of an app version as returned by `list_app_versions`.

Tools marked iOS only or Android only fail with HTTP 422 `ERR_INVALID_PLATFORM` on the other platform. Most store operations need a release candidate to be chosen first (HTTP 400 otherwise) and a Standard license on the connected app (HTTP 412 otherwise). `release_to_app_store`, `complete_app_store_phased_release`, `release_to_google_play` and the staged rollout schedule tools publish to end users and cannot be undone, so an agent should confirm with the user before calling them. The delete, stop, cancel and status-closing tools are irreversible too, and every one of them is marked with `destructiveHint: true`.

82. `list_app_versions`
   - Lists the app versions of a connected app, newest first, with their stages, release candidate selection and status.
   - Arguments:
     - `connected_app_id`: The uuidV4 identifier of the connected app.
     - `items_per_page`: (Optional) Maximum number of app versions per page (default: 10).
     - `page`: (Optional) Page number to return (default: 1).

83. `get_app_version`
   - Gives back an app version with its stages, release candidate selection, notification settings, automations and status.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.

84. `create_app_version`
   - Creates an app version under a connected app. Nothing is uploaded or published; the release opens with its release candidate stage in progress. Values given here override the ones from `presets_id`.
   - Arguments:
     - `connected_app_id`: The uuidV4 identifier of the connected app.
     - `name`: The name of the app version; for iOS apps the App Store Connect version in X.Y.Z format.
     - `artifact_source`: (Optional) `ci` for Bitrise CI builds (required with `release_branch`/`workflow`) or `api` for uploaded artifacts.
     - `description`: (Optional) The description of the app version.
     - `release_branch`: (Optional) Release branch of the release candidate build configuration.
     - `workflow`: (Optional) Workflow of the release candidate build configuration.
     - `automatic_store_upload`: (Optional) Upload every successful build of the branch and workflow to the store.
     - `release_candidate_locked`: (Optional) Do not select the latest matching artifact automatically.
     - `slack_webhook_url`, `slack_notification_integration_id`, `teams_webhook_url`: (Optional) Release update notification targets.
     - `approvals`: (Optional) Approval tasks (array of objects with `summary`, and optional `description` and `due_date`).
     - `automation`: (Optional) Workflow or pipeline runs on release events (array of objects with `event_name` and `workflow_name` or `pipeline_name`).
     - `presets_id`: (Optional) Identifier of a preset template to fill the app version from.

85. `update_app_version`
   - Updates an app version; omitted arguments are left unchanged. Selecting a `release_candidate_id` requires `release_candidate_locked` set to true (409 otherwise). Turning on `automatic_store_upload` also starts uploading the currently selected release candidate right away. An empty string clears `description`, `slack_webhook_url` or `teams_webhook_url`. `approvals` and `automation` replace the whole existing list. Setting `status` moves the app version to a final status, after which it cannot be modified (409).
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.
     - `artifact_source`, `description`, `release_branch`, `workflow`, `automatic_store_upload`, `slack_webhook_url`, `slack_notification_integration_id`, `teams_webhook_url`, `approvals`, `automation`: (Optional) As in `create_app_version`.
     - `release_candidate_id`: (Optional) The uuidV4 identifier of the installable artifact to select as the release candidate.
     - `release_candidate_locked`: (Optional) Lock the selected release candidate.
     - `status`: (Optional) `abandoned` or `completed_externally`. Closes the release for good; it cannot be modified afterwards.

86. `delete_app_version`
   - Deletes an app version with its stages, approval tasks, automations and event history. Cannot be undone; nothing already uploaded to or published on the store is affected.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.

87. `get_release_candidate`
   - Gives back the release candidate of an app version: the selected installable artifact and the Bitrise CI build it came from.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.

88. `upload_release_candidate`
   - Starts uploading the release candidate to TestFlight (iOS) or the Google Play Console (Android) in an asynchronous Bitrise build, locking the release candidate unless `automatic_store_upload` is on. 422 without a suitable artifact, 409 when an upload cannot be started.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.

89. `get_release_candidate_upload_status`
   - Gives back the store upload status of the release candidate: `upload_state`, the uploading build's URL and, for iOS, the App Store Connect `processing_state`.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.

90. `submit_release_candidate_for_beta_review`
   - Submits the uploaded release candidate for TestFlight beta app review, which external testing groups require. iOS only.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.
     - `auto_notify_enabled`: (Optional) Let TestFlight notify testers once the build is approved (default: false).

91. `list_beta_testing_groups`
   - Lists the TestFlight testing groups (iOS, each with an `id`) or the Google Play testing tracks (Android, grouped as internal, closed and open, each with a `name`) of an app version.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.

92. `start_beta_testing`
   - Distributes the uploaded release candidate to a TestFlight testing group or a Google Play testing track.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.
     - `group_id`: The TestFlight testing group id (iOS) or the Google Play track name (Android).

93. `stop_beta_testing`
   - Stops TestFlight beta testing of the release candidate on a testing group, removing the build from the group. iOS only.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.
     - `group_id`: The TestFlight testing group id.

94. `list_what_to_test_descriptions`
   - Lists the TestFlight "What to Test" descriptions of the uploaded release candidate per locale, with the app's primary locale. iOS only.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.

95. `create_what_to_test_description`
   - Adds a TestFlight "What to Test" description in a new locale. iOS only.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.
     - `locale`: The App Store Connect locale shortcode, e.g. `en-US`.
     - `whats_new`: The text shown to testers.

96. `update_what_to_test_description`
   - Replaces the text of an existing TestFlight "What to Test" description. iOS only.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.
     - `what_to_test_id`: The description id from `list_what_to_test_descriptions`.
     - `whats_new`: The new text shown to testers.

97. `delete_what_to_test_description`
   - Removes a TestFlight "What to Test" description. iOS only.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.
     - `what_to_test_id`: The description id from `list_what_to_test_descriptions`.

98. `list_approval_tasks`
   - Lists the approval tasks of an app version. The approval stage completes once every task is completed.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.

99. `get_approval_task`
   - Gives back one approval task.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.
     - `task_id`: The identifier of the approval task.

100. `create_approval_task`
   - Adds an approval task to the approval stage of an app version.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.
     - `summary`: The name of the task.
     - `description`: (Optional) Detailed explanation of the task.
     - `assigned_user_slug`: (Optional) Slug of the user to assign the task to.
     - `due_date`: (Optional) A date in the future.

101. `update_approval_task`
   - Updates an approval task, including completing or reopening it. On an assigned task only the assignee or a project admin may change `completed`, and only the creator or a project admin the other fields (403 otherwise).
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.
     - `task_id`: The identifier of the approval task.
     - `summary`, `description`, `assigned_user_slug`: (Optional) New values. An empty `description` clears it.
     - `completed`: (Optional) `true` to complete (approve), `false` to reopen.

102. `delete_approval_task`
   - Deletes an approval task; removing the last open task completes the approval stage. An assigned task can only be deleted by its creator or a project admin (403 otherwise).
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.
     - `task_id`: The identifier of the approval task.

103. `submit_for_app_store_review`
   - Submits the uploaded release candidate to App Store review, optionally setting App Store metadata per locale first. With release type `AFTER_APPROVAL` Apple publishes the version as soon as the review passes. iOS only.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.
     - `copy_primary_whats_new`: (Optional) Copy the primary localization's `whats_new` into empty ones (default: false).
     - `localizations`: (Optional) Array of objects with `locale` and optional `whats_new`, `description`, `keywords`, `promotional_text`, `marketing_url`, `support_url`.

104. `get_app_store_review_status`
   - Gives back the App Store review status as tracked by Release Management and as reported by App Store Connect. iOS only; 422 `ERR_REVIEW_NOT_REQUESTED` when never submitted.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.

105. `cancel_app_store_review`
   - Cancels the in-flight App Store review submission. The version loses its place in Apple's review queue and must be resubmitted. iOS only; 422 when the version is not under review.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.

106. `release_to_app_store`
   - Releases an approved `MANUAL` release-type app version to the App Store. Publishes to production and cannot be undone. iOS only.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.

107. `pause_app_store_phased_release`
   - Pauses the in-progress App Store phased release; users who already have the update keep it. iOS only.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.

108. `continue_app_store_phased_release`
   - Continues a paused App Store phased release on Apple's 7-day schedule. iOS only.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.

109. `complete_app_store_phased_release`
   - Completes the phased release, making the update available to 100% of users immediately. Cannot be undone. iOS only.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.

110. `get_app_store_release_status`
   - Gives back the App Store `app_release_status` and `phased_release_status` of an app version. iOS only; 422 `ERR_APP_STORE_VERSION_NOT_FOUND` when the version is missing from App Store Connect.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.

111. `get_app_store_release_settings`
   - Gives back the App Store Connect release settings: `release_type` (`MANUAL`, `AFTER_APPROVAL` or `SCHEDULED`), `phased_release` and `earliest_release_date`. iOS only.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.

112. `update_app_store_release_settings`
   - Updates the App Store Connect release settings; both `release_type` and `phased_release` must be given. iOS only.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.
     - `release_type`: `MANUAL`, `AFTER_APPROVAL` or `SCHEDULED` (case-insensitive).
     - `phased_release`: Roll out over 7 days (`true`) or to everyone at once (`false`).
     - `earliest_release_date`: (Optional) Earliest publish date-time; only with `SCHEDULED`.

113. `update_google_play_release_notes`
   - Sets the Google Play release notes of the uploaded release candidate per language. Android only.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.
     - `release_notes`: Array of objects with `language` (a Google Play language code) and `text` (max 500 characters).

114. `get_google_play_release`
   - Gives back the current Google Play production rollout as `user_fraction` (0 to 1, `1` is a full release, `null` when not on the production track). Android only.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.

115. `release_to_google_play`
   - Releases the uploaded release candidate on the Google Play production track to a fraction of users, or raises an in-progress rollout. Publishes to production and cannot be undone. Android only.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.
     - `user_fraction`: Greater than 0 and at most 1, where 1 is a full release.

116. `get_google_play_staged_rollout_schedule`
   - Gives back the staged rollout schedule: its steps (each with `id`, `rollout_percentage`, `when`, `has_run`, `status`), time zone, paused state and failure reason. Android only.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.

117. `create_google_play_staged_rollout_schedule`
   - Creates a staged rollout schedule: dates at which the production rollout is raised to the given percentage. Steps that have run cannot be undone. Android only.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.
     - `location`: Time zone of the dates: a TZ database name or `UTC`.
     - `schedule`: Array of objects with `percentage` (0 to 100, `100` is a full release) and `when` (ISO 8601).

118. `pause_google_play_staged_rollout_schedule`
   - Pauses the running staged rollout schedule. 412 when there is no schedule, it has not started, or it has finished. Android only.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.

119. `resume_google_play_staged_rollout_schedule`
   - Resumes a paused staged rollout schedule with the remaining steps re-timed. 412 when the schedule is not paused or has finished. Android only.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.
     - `location`: Time zone of the dates: a TZ database name or `UTC`.
     - `schedule`: Array of objects with the step `id` from `get_google_play_staged_rollout_schedule`, `percentage` and the new `when`.

120. `delete_google_play_staged_rollout_schedule`
   - Deletes the staged rollout schedule so no further steps run; steps that already ran stay in effect. 412 when there is no schedule or it has finished. Android only.
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.

121. `list_release_events`
   - Lists the event history of an app version, newest first, paginated with a cursor: pass `next_before` back as `before` until it is `null`. `total_count` and `available_actors` come with the first page only. Requires a Standard license (412 otherwise).
   - Arguments:
     - `app_version_id`: The uuidV4 identifier of the app version.
     - `before`: (Optional) Cursor: the `next_before` of the previous response, verbatim.
     - `limit`: (Optional) Maximum number of events per request (default: 10).
     - `search`: (Optional) Case-insensitive filter on the event title.
     - `from`, `to`: (Optional) Inclusive ISO 8601 time range; must be given together.
     - `triggered_by`: (Optional) User slugs or machine actor types (`store`, `automation`, `preset`) to filter by.

122. `create_app_store_version`
   - Creates the App Store Connect version record needed before an iOS app version can be submitted for review. An app version named `version_string` must already exist under the connected app (404 otherwise). iOS only.
   - Arguments:
     - `connected_app_id`: The uuidV4 identifier of the connected app.
     - `version_string`: The version to create in App Store Connect, e.g. `1.2.3`.

123. `get_app_store_draft_version`
   - Gives back the App Store Connect draft version of a connected app: the latest unreleased version, with its state, release type and the matching `app_version_id`. 404 when there is none. iOS only.
   - Arguments:
     - `connected_app_id`: The uuidV4 identifier of the connected app.

124. `update_app_store_draft_version`
   - Renames the App Store Connect draft version to match an existing app version. iOS only.
   - Arguments:
     - `connected_app_id`: The uuidV4 identifier of the connected app.
     - `version_string`: The new version string, matching an existing app version's name.

### Insights

Build and test metrics from Bitrise Insights (the public Insights API, `https://api.bitrise.io/insights/v1`). Requires a paid Insights plan on the workspace; without one every call fails with HTTP 402. Every tool reports on exactly one project, identified by its Insights project ID. Time windows are RFC3339 in UTC, `start` inclusive and `end` exclusive. Durations are seconds, rates are percentages in 0-100. The `branch` filter accepts `*` as a wildcard (`release/*`).

Common arguments of every Insights tool:
- `workspace_slug`: Slug of the workspace the project belongs to
- `project`: ID of the Insights project to report on
- `start`: Start of the time window, inclusive, RFC3339
- `end`: End of the time window, exclusive, RFC3339

1. `insights_get_build_totals`
   - Build count, failure rate, p50/p90 duration and total duration over the whole window, without grouping. The window may span at most 2 years.
   - Arguments:
     - `branch` (optional): Keep only builds on these branches; `*` is a wildcard
     - `workflow` (optional): Keep only builds of these workflows
     - `pipeline` (optional): Keep only builds of these pipelines

2. `insights_get_build_series`
   - The same build metrics as a time series: one row per time bucket, oldest first. With `group_by`, one row per bucket per group, each row naming its group. Window caps: 7 days hourly, 90 days daily, 1 year weekly, 2 years monthly.
   - Arguments:
     - `branch`, `workflow`, `pipeline` (optional): as above
     - `granularity` (optional): hourly, daily (default), weekly or monthly
     - `group_by` (optional): workflow, pipeline, stage or step
     - `limit` (optional): Max number of groups, 1-100 (default: 20); ignored without `group_by`

3. `insights_get_test_totals`
   - Test run count, failure rate, p50/p90 duration, total duration and flaky run count over the whole window, without grouping. Counts are of test case runs, not distinct test cases. The window may span at most 2 years.
   - Arguments:
     - `branch`, `workflow`, `pipeline` (optional): as above, on the builds the tests ran in
     - `test_suite` (optional): Keep only runs of these test suites

4. `insights_get_test_series`
   - The same test metrics as a time series, with the same window caps as the build series.
   - Arguments:
     - `branch`, `workflow`, `pipeline`, `test_suite` (optional): as above
     - `granularity` (optional): hourly, daily (default), weekly or monthly
     - `group_by` (optional): workflow, pipeline, stage, branch or test_suite
     - `limit` (optional): Max number of groups, 1-100 (default: 20)

5. `insights_list_flaky_tests`
   - Test cases that were rerun after a flaky result in the window, one page at a time, ranked. The window may span at most 2 years.
   - Arguments:
     - `branch`, `workflow`, `test_suite` (optional): as above (no `pipeline` filter on this tool)
     - `test_case` (optional): Keep only these test cases
     - `module` (optional): Keep only test cases of these modules
     - `order_by` (optional): flaky_rerun_count (default) or flaky_rate
     - `order` (optional): desc (default) or asc
     - `page` (optional): 1-based page (default: 1)
     - `per_page` (optional): Test cases per page, 1-100 (default: 20)

## API Groups

The Bitrise MCP server organizes tools into API groups that can be enabled or disabled via command-line arguments. The table below shows which API groups each tool belongs to:

| Tool | apps | builds | workspaces | outgoing-webhooks | artifacts | group-roles | cache-items | pipelines | account | read-only | release-management | configuration | release-management-code-push | release-management-store-releases | insights |
|------|------|--------|------------|-------------------|-----------|-------------|-------------|-----------|---------|-----------|--------------------|--------------|------------------------------|-----------------------------------|---|
| list_apps | ✅ | | | | | | | | | ✅ | | | | | |
| register_app | ✅ | | | | | | | | | | | | | | |
| finish_bitrise_app | ✅ | | | | | | | | | | | | | | |
| get_app | ✅ | | | | | | | | | ✅ | | | | | |
| delete_app | ✅ | | | | | | | | | | | | | | |
| update_app | ✅ | | | | | | | | | | | | | | |
| get_bitrise_yml | ✅ | | | | | | | | | ✅ | | | | | |
| update_bitrise_yml | ✅ | | | | | | | | | | | | | | |
| list_branches | ✅ | | | | | | | | | ✅ | | | | | |
| register_ssh_key | ✅ | | | | | | | | | | | | | | |
| register_webhook | ✅ | | | | | | | | | | | | | | |
| list_builds | | ✅ | | | | | | | | ✅ | | | | | |
| trigger_bitrise_build | | ✅ | | | | | | | | | | | | | |
| get_build | | ✅ | | | | | | | | ✅ | | | | | |
| abort_build | | ✅ | | | | | | | | | | | | | |
| get_build_log | | ✅ | | | | | | | | ✅ | | | | | |
| get_build_bitrise_yml | | ✅ | | | | | | | | ✅ | | | | | |
| list_build_workflows | | ✅ | | | | | | | | ✅ | | | | | |
| get_build_steps | | ✅ | | | | | | | | ✅ | | | | | |
| list_artifacts | | | | | ✅ | | | | | ✅ | | | | | |
| get_artifact | | | | | ✅ | | | | | ✅ | | | | | |
| delete_artifact | | | | | ✅ | | | | | | | | | | |
| update_artifact | | | | | ✅ | | | | | | | | | | |
| list_outgoing_webhooks | | | | ✅ | | | | | | ✅ | | | | | |
| delete_outgoing_webhook | | | | ✅ | | | | | | | | | | | |
| update_outgoing_webhook | | | | ✅ | | | | | | | | | | | |
| create_outgoing_webhook | | | | ✅ | | | | | | | | | | | |
| list_cache_items | | | | | | | ✅ | | | ✅ | | | | | |
| delete_all_cache_items | | | | | | | ✅ | | | | | | | | |
| delete_cache_item | | | | | | | ✅ | | | | | | | | |
| get_cache_item_download_url | | | | | | | ✅ | | | ✅ | | | | | |
| list_pipelines | | | | | | | | ✅ | | ✅ | | | | | |
| get_pipeline | | | | | | | | ✅ | | ✅ | | | | | |
| abort_pipeline | | | | | | | | ✅ | | | | | | | |
| rebuild_pipeline | | | | | | | | ✅ | | | | | | | |
| list_group_roles | | | | | | ✅ | | | | ✅ | | | | | |
| replace_group_roles | | | | | | ✅ | | | | | | | | | |
| list_workspaces | | | ✅ | | | | | | | ✅ | | | | | |
| get_workspace | | | ✅ | | | | | | | ✅ | | | | | |
| get_workspace_groups | | | ✅ | | | | | | | ✅ | | | | | |
| create_workspace_group | | | ✅ | | | | | | | | | | | | |
| get_workspace_members | | | ✅ | | | | | | | ✅ | | | | | |
| invite_member_to_workspace | | | ✅ | | | | | | | | | | | | |
| add_member_to_group | | | ✅ | | | | | | | | | | | | |
| me | | | | | | | | | ✅ | ✅ | | | | | |
| create_connected_app | | | | | | | | | | | ✅ | | | | |
| list_connected_apps | | | | | | | | | | ✅ | ✅ | | | | |
| get_connected_app | | | | | | | | | | ✅ | ✅ | | | | |
| update_connected_app | | | | | | | | | | | ✅ | | | | |
| list_installable_artifacts | | | | | | | | | | ✅ | ✅ | | | | |
| generate_installable_artifact_upload_url | | | | | | | | | | | ✅ | | | | |
| get_installable_artifact_upload_and_proc_status | | | | | | | | | | ✅ | ✅ | | | | |
| set_installable_artifact_public_install_page | | | | | | | | | | | ✅ | | | | |
| list_build_distribution_versions | | | | | | | | | | ✅ | ✅ | | | | |
| list_build_distribution_version_test_builds | | | | | | | | | | ✅ | ✅ | | | | |
| create_tester_group | | | | | | | | | | | ✅ | | | | |
| notify_tester_group | | | | | | | | | | | ✅ | | | | |
| add_testers_to_tester_group | | | | | | | | | | | ✅ | | | | |
| update_tester_group | | | | | | | | | | | ✅ | | | | |
| list_tester_groups | | | | | | | | | | ✅ | ✅ | | | | |
| get_tester_group | | | | | | | | | | ✅ | ✅ | | | | |
| get_potential_testers | | | | | | | | | | ✅ | ✅ | | | | |
| get_testers | | | | | | | | | | ✅ | ✅ | | | | |
| validate_bitrise_yml | | | | | | | | | | ✅ | | ✅ | | | |
| step_search | | | | | | | | | | ✅ | | ✅ | | | |
| step_inputs | | | | | | | | | | ✅ | | ✅ | | | |
| list_available_stacks | | | | | | | | | | ✅ | | ✅ | | | |
| codepush_list_deployments | | | | | | | | | | ✅ | ✅ | | ✅ | | |
| codepush_get_deployment | | | | | | | | | | ✅ | ✅ | | ✅ | | |
| codepush_create_deployment | | | | | | | | | | | ✅ | | ✅ | | |
| codepush_update_deployment | | | | | | | | | | | ✅ | | ✅ | | |
| codepush_delete_deployment | | | | | | | | | | | ✅ | | ✅ | | |
| codepush_promote_deployment | | | | | | | | | | | ✅ | | ✅ | | |
| codepush_rollback_deployment | | | | | | | | | | | ✅ | | ✅ | | |
| codepush_list_updates | | | | | | | | | | ✅ | ✅ | | ✅ | | |
| codepush_get_update | | | | | | | | | | ✅ | ✅ | | ✅ | | |
| codepush_patch_update | | | | | | | | | | | ✅ | | ✅ | | |
| codepush_delete_update | | | | | | | | | | | ✅ | | ✅ | | |
| codepush_get_update_status | | | | | | | | | | ✅ | ✅ | | ✅ | | |
| codepush_generate_update_upload_url | | | | | | | | | | | ✅ | | ✅ | | |
| codepush_get_metrics | | | | | | | | | | ✅ | ✅ | | ✅ | | |
| list_app_versions |  |  |  |  |  |  |  |  |  | ✅ | ✅ |  |  | ✅ |  |
| get_app_version |  |  |  |  |  |  |  |  |  | ✅ | ✅ |  |  | ✅ |  |
| create_app_version |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| update_app_version |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| delete_app_version |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| get_release_candidate |  |  |  |  |  |  |  |  |  | ✅ | ✅ |  |  | ✅ |  |
| upload_release_candidate |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| get_release_candidate_upload_status |  |  |  |  |  |  |  |  |  | ✅ | ✅ |  |  | ✅ |  |
| submit_release_candidate_for_beta_review |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| list_beta_testing_groups |  |  |  |  |  |  |  |  |  | ✅ | ✅ |  |  | ✅ |  |
| start_beta_testing |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| stop_beta_testing |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| list_what_to_test_descriptions |  |  |  |  |  |  |  |  |  | ✅ | ✅ |  |  | ✅ |  |
| create_what_to_test_description |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| update_what_to_test_description |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| delete_what_to_test_description |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| list_approval_tasks |  |  |  |  |  |  |  |  |  | ✅ | ✅ |  |  | ✅ |  |
| get_approval_task |  |  |  |  |  |  |  |  |  | ✅ | ✅ |  |  | ✅ |  |
| create_approval_task |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| update_approval_task |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| delete_approval_task |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| submit_for_app_store_review |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| get_app_store_review_status |  |  |  |  |  |  |  |  |  | ✅ | ✅ |  |  | ✅ |  |
| cancel_app_store_review |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| release_to_app_store |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| pause_app_store_phased_release |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| continue_app_store_phased_release |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| complete_app_store_phased_release |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| get_app_store_release_status |  |  |  |  |  |  |  |  |  | ✅ | ✅ |  |  | ✅ |  |
| get_app_store_release_settings |  |  |  |  |  |  |  |  |  | ✅ | ✅ |  |  | ✅ |  |
| update_app_store_release_settings |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| update_google_play_release_notes |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| get_google_play_release |  |  |  |  |  |  |  |  |  | ✅ | ✅ |  |  | ✅ |  |
| release_to_google_play |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| get_google_play_staged_rollout_schedule |  |  |  |  |  |  |  |  |  | ✅ | ✅ |  |  | ✅ |  |
| create_google_play_staged_rollout_schedule |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| pause_google_play_staged_rollout_schedule |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| resume_google_play_staged_rollout_schedule |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| delete_google_play_staged_rollout_schedule |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| list_release_events |  |  |  |  |  |  |  |  |  | ✅ | ✅ |  |  | ✅ |  |
| create_app_store_version |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| get_app_store_draft_version |  |  |  |  |  |  |  |  |  | ✅ | ✅ |  |  | ✅ |  |
| update_app_store_draft_version |  |  |  |  |  |  |  |  |  |  | ✅ |  |  | ✅ |  |
| insights_get_build_totals | | | | | | | | | | ✅ | | | | | ✅ |
| insights_get_build_series | | | | | | | | | | ✅ | | | | | ✅ |
| insights_get_test_totals | | | | | | | | | | ✅ | | | | | ✅ |
| insights_get_test_series | | | | | | | | | | ✅ | | | | | ✅ |
| insights_list_flaky_tests | | | | | | | | | | ✅ | | | | | ✅ |
