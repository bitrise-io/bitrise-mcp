package devenvironments

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/devenv"
	"github.com/mark3labs/mcp-go/mcp"
)

// ListSessions lists all sessions for the current user.
var ListSessions = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_list",
		mcp.WithTitleAnnotation("List sessions"),
		mcp.WithDescription(`List Dev Environments sessions (remote VMs); not CI builds (see list_builds). By default (scope="mine") returns the currently authenticated user's own sessions; set scope="workspace" to list sessions owned by the workspace itself instead.

Returns a lightweight view of each session: ID, name, description, status, agentSessionStatus, labels, owner_type ("user" or "workspace"), ownerId (user UUID or workspace slug), template_id, templateDeleted flag, SSH/VNC connection details, AI config, and a templateSnapshot containing the templateName, stack_id, and machine_type.

agentSessionStatus reflects the current state of the AI agent running in the session (working, waiting_for_input, idle, or unspecified). It is reset whenever the session is stopped or started.

Sessions created without a template have an empty template_id and a templateSnapshot with stack_id and machine_type but no templateName.

Use label_selectors to filter sessions server-side by their labels. Each selector is a "key=value" exact-match equality; multiple selectors are ANDed, so a session must match all of them. For example, label_selectors=["team=mobile", "branch=main"] returns only sessions carrying both labels, instead of listing everything and filtering client-side.

To get the full template snapshot (session inputs, feature flags, workspace links, working directory, script flags), the servicePorts of a device session and its SSH credentials, use bitrise_devenv_get on a specific session — the list view omits them.
To check if a session's template has been updated, look at the templateOutdated field on bitrise_devenv_get and use bitrise_devenv_compare_template for details.`),
		mcp.WithArray("label_selectors",
			mcp.Description(`Optional label filters of the form "key=value" (exact-match equality on one label). Multiple selectors are ANDed: only sessions matching every selector are returned. At most 8 selectors; duplicate keys are rejected (they can never match under AND); bare keys without "=" are invalid. System-owned "bitrise.io/"-prefixed keys may be used in selectors.`),
			mcp.WithStringItems(),
		),
		mcp.WithString("scope",
			mcp.Description(`Ownership scope of the listing. "mine" (default) returns the calling user's own sessions. "workspace" returns sessions owned by the workspace itself — created with owner="workspace" (or by a Workspace API Token, which always creates workspace-owned sessions) or started from a preview link (bitrise_devenv_create_preview_link — its devices are always workspace-owned) — which are visible to every workspace member. With a Workspace API Token the default is "workspace" and "mine" is rejected (the token has no personal sessions).`),
			mcp.Enum("mine", "workspace"),
			mcp.DefaultString("mine"),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var repeatedParams map[string][]string
		if selectors := request.GetStringSlice("label_selectors", nil); len(selectors) > 0 {
			repeatedParams = map[string][]string{"label_selectors": selectors}
		}
		params := map[string]string{}
		if request.GetString("scope", "") == "workspace" {
			params["scope"] = "SESSION_LIST_SCOPE_WORKSPACE"
		}
		res, err := devenv.CallAPI(ctx, devenv.CallAPIParams{
			Method:         http.MethodGet,
			Path:           devenv.WsPath(ctx, "/sessions"),
			Params:         params,
			RepeatedParams: repeatedParams,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("list sessions", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

// GetSession retrieves a single session by ID.
var GetSession = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_get",
		mcp.WithTitleAnnotation("Get session"),
		mcp.WithDescription(`Get full details of a specific devenv session.

Returns status, SSH/VNC connection details, AI config, and the complete templateSnapshot which contains:
- templateName: name of the template at creation time
- stack_id: stack ID
- machine_type: machine type name
- session_inputs: input values (key, value, is_secret, expose_as_env_var) snapshotted at creation
- feature_flags: flag states (name, enabled) snapshotted at creation
- workspace_links: IDE folder links (label, folder_path) filtered by enabled flags
- working_directory: terminal working directory
- hasWarmupScript / hasStartupScript: whether scripts were configured

Also includes:
- templateDeleted: true if the template was deleted after session creation (session still works from its snapshot)
- templateOutdated: true if the template has been updated since this session was created (use bitrise_devenv_compare_template to see what changed)
- agentSessionStatus: current state of the AI agent running in the session (working, waiting_for_input, idle, or unspecified). Reset on terminate/restore.
- agentSessionStatusUpdatedAt: timestamp when agentSessionStatus was last changed
- labels: key/value metadata attached to the session (set at creation or via bitrise_devenv_update; filterable in bitrise_devenv_list via label_selectors)
- device (on sessions that boot a virtual device): the device and its readiness. device.state is VM-asserted — PREVIEW_DEVICE_STATE_BOOTING while the VM runs but the device is not proven (wait; touch nothing on the VM), PREVIEW_DEVICE_STATE_READY once the device is booted AND streaming (start working), PREVIEW_DEVICE_STATE_FAILED when this boot gave up — or only its stream did: device.deviceNotes says which, and a stream-only failure leaves the device fully drivable over adb / simctl. Always read device.state together with the session status: PREVIEW_DEVICE_STATE_UNSPECIFIED with status pending/starting means the VM is not up yet (wait); UNSPECIFIED with a terminal status (terminated, terminating, draining, drained, failed) means the device is gone with the VM — restore the session or create a new one, do not keep polling. device.spec echoes the REQUEST, not the result: an explicitly requested os_version / system_image the stack lacks is substituted and only device.deviceNotes says so ("requested … not installed; using …") — diff the notes against your request before trusting the OS version. device.installStatus / installReason track the optional app install (the installer also launches the app). sshAddress (a ready-made "ssh user@host -p port" command), sshPassword, sshConnectionOpen and templateSnapshot.servicePorts are on THIS call only (not on bitrise_devenv_list): device-web-view is the browser view on both platforms, forwarded to local 3200 (VM side: serve-sim 3200 on iOS, ws-scrcpy 8000 on Android); Android adds adb (VM 5555 → local 15555). device.pageUrl is the device's page in the RDE web UI (the session page's "Open device view"): when a human should watch or drive the device, open it in a browser where the user is logged in to Bitrise, or give it to the user. It is not a shareable link — opening it needs a Bitrise login with access to the session (for someone outside the workspace, mint a bitrise_devenv_create_preview_link); it is empty on a warm pool's unclaimed sessions. Full know-how: bitrise_devenv_device_guide (or the resource bitrise-devenv://guides/device-sessions — the same text).

For sessions created without a template, template_id is empty and the snapshot is minimal: only stack_id and machine_type are populated, hasWarmupScript/hasStartupScript are false, and there is no templateName, session_inputs, feature_flags, or workspace_links. templateOutdated is always false for such sessions.

By default, secret session input values are redacted from the snapshot; set include_secrets=true to receive plaintext values.`),
		mcp.WithString("session_id",
			mcp.Description("The unique identifier (UUID) of the session"),
			mcp.Required(),
		),
		mcp.WithBoolean("include_secrets",
			mcp.Description("When true, secret session input values are returned in plaintext. Defaults to false (secret values are redacted)."),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sessionID, err := requireUUID(request, "session_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		params := map[string]string{}
		if request.GetBool("include_secrets", false) {
			params["include_secrets"] = "true"
		}
		res, err := devenv.CallAPI(ctx, devenv.CallAPIParams{
			Method: http.MethodGet,
			Path:   devenv.WsPath(ctx, fmt.Sprintf("/sessions/%s", sessionID)),
			Params: params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("get session", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

// CreateSession creates a new session, either from a template or directly from
// a stack + machine type (template-less).
var CreateSession = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_create",
		mcp.WithTitleAnnotation("Create session"),
		mcp.WithDescription(`Create a Dev Environments session: a remote macOS or Linux VM you then drive with bitrise_devenv_execute. This is not a Bitrise CI build (that is trigger_bitrise_build): choose it when the user wants a machine to work in, not a workflow run.

DOES YOUR TASK TOUCH A MOBILE APP? If it involves an iOS or Android app, a phone/tablet, a simulator or emulator, screenshots, UI tests, screen sizes, orientations or OS versions: create the session WITH device_spec (platform "ios" or "android") and READ THE GUIDE FIRST — bitrise_devenv_device_guide with guide="device-sessions", then "ios" or "android" (clients that read MCP resources can read bitrise-devenv://guides/device-sessions instead; same text). The platform boots and manages the simulator/emulator for you, streams it, and reports when it is ready. A device session is also the right choice for UNATTENDED and batch work — nobody has to watch it; the device stays fully drivable through xcrun simctl / adb / serve-sim from bitrise_devenv_execute. Never build your own simulator/emulator lifecycle on a bare machine (2/3 below without device_spec): that path is not supported for device work and loses readiness reporting, streaming and the viewer.

Three ways to create a session:

1) With a virtual device (device_spec set, or a template that declares one) — the default for anything mobile:
Boots an iOS simulator (platform "ios") or Android emulator (platform "android") alongside the session and streams it — ready for adb / xcrun simctl / serve-sim; optionally a human can watch and drive it on its page in the RDE web UI — device.pageUrl in the response, for a browser where the user is logged in to Bitrise. Zero-config rule: OMIT BOTH stack_id and machine_type (and cluster) — the deployment's known-good per-platform defaults apply; giving exactly one of the two is rejected. Name them only when you must (the guide says which stacks fit); with a template, its stack and machine type are used and must fit. Optionally pass artifact to pre-install an app build. Delete the session when done.
"running" is NOT "device ready": poll bitrise_devenv_get until session.device.state is PREVIEW_DEVICE_STATE_READY (and installStatus is PREVIEW_INSTALL_STATUS_OK if you passed an artifact). While it is PREVIEW_DEVICE_STATE_BOOTING touch nothing on the VM — do not run recovery scripts, do not recreate. On PREVIEW_DEVICE_STATE_FAILED read device.deviceNotes: a stream-only failure leaves the device fully drivable over adb / simctl (guide §6).
Templates can declare a device (device_spec on bitrise_devenv_get_template). Creating from one: omit device_spec to boot it as declared; pass a device_spec WITHOUT a platform to tweak it per field (empty fields inherit the template's); WITH a platform it is the complete device to boot (the template's is ignored); no_device=true skips the device (not combinable with device_spec; ignored when the template declares none).

2) From a template (template_id set), no device unless the template declares one:
1. List templates with bitrise_devenv_list_templates to find available templates and their session inputs
2. Optionally list saved inputs with bitrise_devenv_list_saved_inputs to find saved credentials
3. Provide values for session inputs (either direct values or references to saved inputs), or set map_saved_to_session_inputs=true to auto-fill session inputs from the user's saved inputs by key match
The session inherits the template's stack, machine type, scripts, feature flags, and workspace links. You may optionally pass stack_id and/or machine_type to override the template's values for this session only.

3) Without a template (template_id omitted) and without a device — a bare build machine:
Supply stack_id and machine_type directly to get a base environment with no warmup/startup scripts and no template configuration (no session inputs, feature flags, or workspace links). Use bitrise_devenv_list_stacks and bitrise_devenv_list_machine_types to discover valid values. This is the quickest way to spin up an environment for a repo when no template and no device is needed.

4) From a warm pool (warm_pool_id set) — the fastest path when a pool exists:
A warm pool (bitrise_devenv_list_warm_pools) keeps sessions of one stored configuration booted and idle. Pass its id as warm_pool_id and the backend hands you one of them, renamed to your name, with no machine to boot (the response's warmState is "claimed", and a device pool's session is usually READY at once — device.pageUrl opens it in the browser); when none is available it creates a session from the pool's configuration instead (warmState "cold"), so the call always succeeds. The pool fixes the configuration: do NOT pass template_id, session_inputs, map_saved_to_session_inputs, enabled_feature_flag_names, stack_id, machine_type, cluster, device_spec, no_device or ai_prompt — they are rejected, not ignored. Only name, description, labels, auto_terminate_minutes and artifact (pools whose configuration boots a device) apply to the claimed session; owner, if given, must be the pool's owner_type. Check bitrise_devenv_list_warm_pools before creating a session from a template that a pool already covers.

Who owns the session (owner): "user" (default) is a personal session of the authenticated user. "workspace" creates a session owned by the workspace itself — visible to and manageable by every member, listed with bitrise_devenv_list scope="workspace". A workspace-owned session carries no personal state: give every template session input as a plain value in session_inputs (saved_input_id references and map_saved_to_session_inputs are rejected), and ai_prompt is not available. When the server is authenticated with a Workspace API Token (bitwat_…, e.g. from CI) every session it creates is workspace-owned — omit owner or set "workspace"; "user" is rejected.

The session will start provisioning immediately after creation.`),
		mcp.WithString("name",
			mcp.Description("Human-readable name for the session"),
			mcp.Required(),
		),
		mcp.WithString("description",
			mcp.Description("Description of the session"),
		),
		mcp.WithString("template_id",
			mcp.Description("ID of the template to use. Optional: omit to create a session without a template, in which case either device_spec (platform defaults pick the machine) or both stack_id and machine_type are required, and no warmup/startup scripts run. Not allowed with warm_pool_id (the pool fixes the template)."),
		),
		mcp.WithString("warm_pool_id",
			mcp.Description("Claim a warm session from this pool (UUID, from bitrise_devenv_list_warm_pools) instead of building one — see 4 above. When set, do not pass template_id, session_inputs, map_saved_to_session_inputs, enabled_feature_flag_names, stack_id, machine_type, cluster, device_spec, no_device or ai_prompt: the pool fixes them and the request is rejected otherwise. name, description, labels, auto_terminate_minutes and artifact still apply to the claimed session."),
		),
		mcp.WithString("stack_id",
			mcp.Description("Stack ID (e.g. 'osx-xcode-16.0.x-edge'). Required when template_id is omitted — unless device_spec is set, then omit it (and machine_type) for the platform defaults. When a template is given, optionally overrides the template's stack for this session. Use bitrise_devenv_list_stacks to find valid IDs."),
		),
		mcp.WithString("machine_type",
			mcp.Description("Machine type name (e.g. 'g2.mac.m2pro.4c'). Required when template_id is omitted — unless device_spec is set, then omit it (and stack_id) for the platform defaults. When a template is given, optionally overrides the template's machine type for this session. Use bitrise_devenv_list_machine_types to find valid names."),
		),
		mcp.WithArray("session_inputs",
			mcp.Description("Values for the template's session inputs. Required inputs must have a value (direct or saved_input_id). Optional inputs use their default_value when omitted."),
			mcp.Items(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"key":            map[string]any{"type": "string", "description": "Key name matching a session input on the template"},
					"value":          map[string]any{"type": "string", "description": "Direct value (ignored if saved_input_id is set)"},
					"is_secret":      map[string]any{"type": "boolean", "description": "Whether the value is secret (ignored if saved_input_id is set)"},
					"saved_input_id": map[string]any{"type": "string", "description": "Optional: ID of a saved input to use instead of a direct value"},
				},
				"required": []string{"key"},
			}),
		),
		mcp.WithBoolean("map_saved_to_session_inputs",
			mcp.Description(`When true, the backend fills unreferenced template session inputs from the current user's saved inputs by matching keys, before required-input validation runs.

Use this as a shortcut instead of calling bitrise_devenv_list_saved_inputs and constructing a session_inputs entry for every saved credential that happens to match a template key.

Rules:
- Entries in session_inputs always win; auto-mapping only fills keys not already supplied.
- Required inputs that match neither session_inputs nor any saved input still fail with "missing required input: <key>" — the flag is not a bypass of required-input validation.
- The response includes an autoMappedInputs array listing {sessionInputKey, saved_input_id} for every key that was auto-filled, so you can report back exactly what the flag resolved.`),
		),
		mcp.WithArray("enabled_feature_flag_names",
			mcp.Description("Names of feature flags to enable for this session"),
			mcp.WithStringItems(),
		),
		mcp.WithString("cluster",
			mcp.Description("Target cluster name. Not needed with device_spec (the backend picks one). Otherwise required when the chosen stack + machine type are available in multiple clusters — whether they come from the template or, for a template-less session, from the stack_id and machine_type supplied directly. The candidates are the stack's clusterNames (from bitrise_devenv_list_stacks) that also match the machine type's clusterName (from bitrise_devenv_list_machine_types). Omit when only one cluster matches."),
		),
		mcp.WithString("ai_prompt",
			mcp.Description("Optional AI prompt to pass to Claude Code when the session starts"),
		),
		mcp.WithNumber("auto_terminate_minutes",
			mcp.Description("Minutes before auto-termination. Default: 7200 (5 days). Set to 0 to disable."),
		),
		mcp.WithObject("device_spec",
			deviceSpecSchema(`Optional virtual device to boot with the session (see 1 above; read the device guide first). `+deviceSpecFieldsDoc+` platform is required unless the session is created from a template that declares a device — then omit it to tweak that device per field (only the fields you set change), or name one to replace it whole. Prefer omitting stack_id/machine_type; if you pass them they must fit the platform (OS family, >= 4 vCPU / 6-8 GB) or the request is rejected with the reason.`)...,
		),
		mcp.WithBoolean("no_device",
			mcp.Description("Create the session WITHOUT the device its template declares (see 1 above). Only meaningful with a template that has a device_spec; ignored otherwise. Cannot be combined with device_spec."),
		),
		mcp.WithObject("artifact",
			mcp.Description(`Optional app build to install on the device once it is READY (requires device_spec). url is an absolute http(s) URL the VM downloads directly (a signed URL is fine; it is never returned) — iOS: a zipped simulator .app, Android: an .apk. app_name / build_number / commit_sha are display metadata (shown in the viewer). Progress: session.device.installStatus; a FAILED install (installReason says why) leaves the device usable — install the app yourself.`),
			mcp.Properties(map[string]any{
				"url":          map[string]any{"type": "string", "description": "absolute http(s) download URL of the app build"},
				"app_name":     map[string]any{"type": "string"},
				"build_number": map[string]any{"type": "string"},
				"commit_sha":   map[string]any{"type": "string"},
			}),
			requiredProperties("url"),
		),
		mcp.WithString("owner",
			mcp.Description(`Who owns the session. "user" (default): a personal session of the authenticated user. "workspace": owned by the workspace itself — shared with every member, listed under bitrise_devenv_list scope="workspace"; session inputs must then be plain values (no saved_input_id, no map_saved_to_session_inputs) and ai_prompt is unavailable. Required to be omitted or "workspace" when the server runs with a Workspace API Token.`),
			mcp.Enum("user", "workspace"),
		),
		mcp.WithObject("labels",
			mcp.Description(`Optional key/value string labels to attach to the session, e.g. {"team": "mobile", "branch": "main"}. At most 32 labels; keys are 1-63 characters of [a-zA-Z0-9._/-] starting and ending alphanumeric; values are 1-255 bytes of [a-zA-Z0-9._/:+-] with no positional rules (timestamps with offsets, branch names, paths, and semver all fit; spaces, '@', '=', newlines, and non-ASCII are rejected). The "bitrise.io/" key prefix is reserved for system-owned labels and rejected. Labels are returned on session reads and filterable in bitrise_devenv_list via label_selectors.`),
			mcp.AdditionalProperties(map[string]any{"type": "string"}),
		),
		mcp.WithDestructiveHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		templateID := request.GetString("template_id", "")
		stackID := request.GetString("stack_id", "")
		machineType := request.GetString("machine_type", "")
		deviceSpec, hasDevice := request.GetArguments()["device_spec"]
		artifact, hasArtifact := request.GetArguments()["artifact"]
		noDevice := request.GetBool("no_device", false)
		warmPoolID, err := optionalUUID(request, "warm_pool_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// A claim takes its whole configuration from the pool: the backend
		// rejects (does not ignore) every configuration field sent alongside
		// warm_pool_id, so name the offenders before spending the round trip.
		// The per-session fields (name, description, labels, auto-terminate,
		// artifact, owner) pass through untouched.
		if warmPoolID != "" {
			var conflicting []string
			for _, key := range []string{"template_id", "session_inputs", "map_saved_to_session_inputs", "enabled_feature_flag_names", "stack_id", "machine_type", "cluster", "device_spec", "no_device", "ai_prompt"} {
				if v, present := request.GetArguments()[key]; present && !isZeroArgument(v) {
					conflicting = append(conflicting, key)
				}
			}
			if len(conflicting) > 0 {
				return mcp.NewToolResultError(fmt.Sprintf("warm_pool_id cannot be combined with %s — the pool fixes the session's configuration; pass only name, description, labels, auto_terminate_minutes, artifact and owner", strings.Join(conflicting, ", "))), nil
			}
			// Whatever configuration keys remain are blank, and a blank key
			// means "absent" on a claim — so treat the device knobs as unset
			// too, instead of validating an empty device_spec.
			hasDevice, noDevice = false, false
		}
		claim := warmPoolID != ""

		// Without a template the session is built directly from a stack and
		// machine type, so both must be supplied — unless a device_spec is
		// set, in which case the backend fills whatever is missing from the
		// deployment's per-platform device defaults, or the session is
		// claimed from a warm pool, which fixes the configuration itself.
		if warmPoolID == "" && templateID == "" && !hasDevice && (stackID == "" || machineType == "") {
			return mcp.NewToolResultError("either template_id, device_spec, or both stack_id and machine_type (to create a session without a template), must be provided"), nil
		}
		if hasDevice && noDevice {
			return mcp.NewToolResultError("no_device cannot be combined with device_spec — either boot a device or skip it"), nil
		}
		// An artifact needs a device to land on: either a device_spec on the
		// request, a template-declared device that no_device does not
		// suppress, or a warm pool whose configuration boots one (whether the
		// template or pool really declares one is the backend's call).
		if hasArtifact && !hasDevice && warmPoolID == "" && (templateID == "" || noDevice) {
			return mcp.NewToolResultError("artifact requires a device — pass device_spec, or create from a template that declares one without no_device"), nil
		}
		// The nested "required" lists above are advisory to the client; check
		// what the backend cannot default before spending a round trip on a
		// request that is certain to be rejected. device_spec.platform is only
		// certain to be needed without a template: on a template that declares
		// a device, a platform-less spec is the documented per-field override,
		// and whether the template declares one is the backend's call.
		if hasDevice {
			if _, ok := deviceSpec.(map[string]any); !ok {
				return mcp.NewToolResultError("device_spec must be an object with a \"platform\" field"), nil
			}
			if templateID == "" {
				if err := requireNonEmptyString(deviceSpec, "device_spec", "platform"); err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
			}
		}
		if hasArtifact {
			if err := requireNonEmptyString(artifact, "artifact", "url"); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
		}

		body := map[string]any{
			"name": request.GetString("name", ""),
		}
		if templateID != "" {
			body["template_id"] = templateID
		}
		if warmPoolID != "" {
			body["warm_pool_id"] = warmPoolID
		}
		if stackID != "" {
			body["stack_id"] = stackID
		}
		if machineType != "" {
			body["machine_type"] = machineType
		}
		if desc := request.GetString("description", ""); desc != "" {
			body["description"] = desc
		}
		// On a claim these keys are known to be blank (checked above) and
		// belong to the pool, so they stay off the wire entirely.
		if inputs, ok := request.GetArguments()["session_inputs"]; ok && !claim {
			body["session_inputs"] = inputs
		}
		if mapSaved, ok := request.GetArguments()["map_saved_to_session_inputs"]; ok && !claim {
			body["map_saved_to_session_inputs"] = mapSaved
		}
		if flags, ok := request.GetArguments()["enabled_feature_flag_names"]; ok && !claim {
			body["enabled_feature_flag_names"] = flags
		}
		if cluster := request.GetString("cluster", ""); cluster != "" {
			body["cluster"] = cluster
		}
		if aiPrompt := request.GetString("ai_prompt", ""); aiPrompt != "" {
			body["ai_prompt"] = aiPrompt
		}
		if minutes, ok, err := getOptionalInt(request, "auto_terminate_minutes"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		} else if ok {
			body["auto_terminate_minutes"] = minutes
		}
		if labels, ok := request.GetArguments()["labels"]; ok {
			body["labels"] = labels
		}
		// Sent only when set: an absent owner_type is the backend's default
		// (personal for a user, workspace-owned for a Workspace API Token).
		if owner := request.GetString("owner", ""); owner != "" {
			body["owner_type"] = owner
		}
		if hasDevice {
			body["device_spec"] = deviceSpec
		}
		if noDevice {
			body["no_device"] = true
		}
		if hasArtifact {
			body["artifact"] = artifact
		}

		res, err := devenv.CallAPI(ctx, devenv.CallAPIParams{
			Method: http.MethodPost,
			Path:   devenv.WsPath(ctx, "/sessions"),
			Body:   body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("create session", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

// RestoreSession restores a terminated (or restarts a failed) session.
var RestoreSession = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_restore",
		mcp.WithTitleAnnotation("Restore session"),
		mcp.WithDescription(`Re-provision the VM of a terminated, drained or failed Dev Environments session (not a CI re-run: see rebuild_pipeline / trigger_bitrise_build). The session will begin provisioning and transition to running. Resets agentSessionStatus.

Restorable statuses: SESSION_STATUS_TERMINATED (user terminated), SESSION_STATUS_DRAINED (node was reclaimed under the session), SESSION_STATUS_FAILED. All three are terminal-and-restorable — restoring recreates the VM.

A session in SESSION_STATUS_UNKNOWN (the backend can't currently determine the machine state, e.g. its node lost network connectivity) cannot be restored, terminated or deleted until the state settles — retry shortly.

Only sessions that were terminated (not deleted) can be restored: bitrise_devenv_delete discards the VM and its disk for good.`),
		mcp.WithString("session_id",
			mcp.Description("The unique identifier (UUID) of the session to restore"),
			mcp.Required(),
		),
		mcp.WithDestructiveHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sessionID, err := requireUUID(request, "session_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		res, err := devenv.CallAPI(ctx, devenv.CallAPIParams{
			Method: http.MethodPost,
			Path:   devenv.WsPath(ctx, fmt.Sprintf("/sessions/%s/restore", sessionID)),
			Body:   map[string]any{},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("restore session", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

// TerminateSession terminates a running session (stops the VM, keeping the
// session for later restore).
var TerminateSession = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_terminate",
		mcp.WithTitleAnnotation("Terminate session"),
		mcp.WithDescription(`Stop a running Dev Environments session's VM but KEEP it for a later bitrise_devenv_restore (not a CI build: aborting a build or pipeline is abort_build / abort_pipeline): the VM is stopped, its disk is preserved, and the session stays listed as SESSION_STATUS_TERMINATED until it is restored (bitrise_devenv_restore) or deleted. Resets agentSessionStatus.

Use this only when the user wants to come back to this exact session later (e.g. to keep uncommitted work or an expensive warm state). A terminated session keeps occupying disk until it is deleted, and forgotten terminated sessions are the main source of waste — so when the session is simply no longer needed, call bitrise_devenv_delete directly instead; it works on running sessions and does NOT require terminating first.

Asynchronous: returns while the session is still SESSION_STATUS_TERMINATING; poll bitrise_devenv_get if you need to observe it reach SESSION_STATUS_TERMINATED.`),
		mcp.WithString("session_id",
			mcp.Description("The unique identifier (UUID) of the session to terminate"),
			mcp.Required(),
		),
		mcp.WithDestructiveHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sessionID, err := requireUUID(request, "session_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		res, err := devenv.CallAPI(ctx, devenv.CallAPIParams{
			Method: http.MethodPost,
			Path:   devenv.WsPath(ctx, fmt.Sprintf("/sessions/%s/terminate", sessionID)),
			Body:   map[string]any{},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("terminate session", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

// DeleteSession permanently deletes a session in any state (RDE-54): a
// running VM is stopped and discarded by the backend, no terminate needed.
var DeleteSession = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_delete",
		mcp.WithTitleAnnotation("Delete session"),
		mcp.WithDescription(`Permanently delete a devenv session in ANY state — running, starting, terminating, terminated or failed. This is the preferred way to get rid of a session you are done with: it does not have to be terminated first.

The session disappears from the list immediately and cannot be restored. If its VM is still running, the backend stops it and then discards it together with its disk in the background — any unsaved work on the VM is lost, so make sure anything worth keeping (commits, pushes, uploads) is already off the machine.

Prefer this over bitrise_devenv_terminate unless the user explicitly wants to restore the session later. Fails with a precondition error while the machine state is SESSION_STATUS_UNKNOWN — retry shortly.`),
		mcp.WithString("session_id",
			mcp.Description("The unique identifier (UUID) of the session to delete"),
			mcp.Required(),
		),
		mcp.WithDestructiveHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sessionID, err := requireUUID(request, "session_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		res, err := devenv.CallAPI(ctx, devenv.CallAPIParams{
			Method: http.MethodDelete,
			Path:   devenv.WsPath(ctx, fmt.Sprintf("/sessions/%s", sessionID)),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("delete session", err), nil
		}
		devenv.ForgetScreenResolution(sessionID)
		if strings.TrimSpace(res) == "" || strings.TrimSpace(res) == "{}" {
			// The API returns an empty body on success; say what happened so
			// the model doesn't have to guess from a blank result.
			return mcp.NewToolResultText(fmt.Sprintf("Session %s deleted. If its VM was still running it is being stopped and discarded in the background; the session cannot be restored.", sessionID)), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

// UpdateSession updates a session's name, description, labels, or
// auto-terminate settings.
var UpdateSession = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_update",
		mcp.WithTitleAnnotation("Update session"),
		mcp.WithDescription("Update a session's name, description, labels, or auto-terminate settings. Only provided fields are updated."),
		mcp.WithString("session_id",
			mcp.Description("The unique identifier (UUID) of the session to update"),
			mcp.Required(),
		),
		mcp.WithString("name",
			mcp.Description("Updated session name"),
		),
		mcp.WithString("description",
			mcp.Description("Updated session description"),
		),
		mcp.WithNumber("auto_terminate_minutes",
			mcp.Description("Update auto-terminate duration in minutes. Resets the deadline to now + minutes. Set to 0 to disable."),
		),
		mcp.WithObject("labels",
			mcp.Description(`Labels to add or update on the session. Merged into the existing labels: listed keys are overwritten, unlisted keys are left untouched. Same constraints as in bitrise_devenv_create (at most 32 labels total; keys 1-63 chars of [a-zA-Z0-9._/-] starting and ending alphanumeric; values 1-255 bytes of [a-zA-Z0-9._/:+-]; "bitrise.io/" key prefix reserved). To delete keys use remove_labels; if a key appears in both, the removal wins.`),
			mcp.AdditionalProperties(map[string]any{"type": "string"}),
		),
		mcp.WithArray("remove_labels",
			mcp.Description("Label keys to remove from the session. Unknown keys are ignored."),
			mcp.WithStringItems(),
		),
		mcp.WithDestructiveHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sessionID, err := requireUUID(request, "session_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{}
		if _, ok := request.GetArguments()["name"]; ok {
			body["name"] = request.GetString("name", "")
		}
		if _, ok := request.GetArguments()["description"]; ok {
			body["description"] = request.GetString("description", "")
		}
		if minutes, ok, err := getOptionalInt(request, "auto_terminate_minutes"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		} else if ok {
			body["auto_terminate_minutes"] = minutes
		}
		if labels, ok := request.GetArguments()["labels"]; ok {
			body["labels"] = labels
		}
		if removeLabels := request.GetStringSlice("remove_labels", nil); len(removeLabels) > 0 {
			body["remove_labels"] = removeLabels
		}

		res, err := devenv.CallAPI(ctx, devenv.CallAPIParams{
			Method: http.MethodPatch,
			Path:   devenv.WsPath(ctx, fmt.Sprintf("/sessions/%s", sessionID)),
			Body:   body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("update session", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

// CompareSessionTemplate compares a session's template snapshot with the current template.
var CompareSessionTemplate = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_compare_template",
		mcp.WithTitleAnnotation("Compare session with template"),
		mcp.WithDescription(`Compare a session's template snapshot with the current template configuration.

Returns both the snapshot (template config at session creation time) and the current template config side-by-side, including:
- templateName, stack_id, machine_type, working_directory
- startup_script, warmup_script (full text)
- feature_flags (name, description, enabled)
- session_inputs (key, description, required, default_value)
- template_variables (key, is_secret — values never exposed)
- device_spec (the virtual device the template boots with its sessions; absent when it declares none)
- changedVariableKeys: list of variable keys whose values differ (computed server-side)

Use this when templateOutdated is true on a session to see exactly what changed.
If the current template was deleted, the current field will be null.
Sessions created without a template have nothing to compare against, so the current field is null for them.`),
		mcp.WithString("session_id",
			mcp.Description("The unique identifier (UUID) of the session to compare"),
			mcp.Required(),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sessionID, err := requireUUID(request, "session_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		res, err := devenv.CallAPI(ctx, devenv.CallAPIParams{
			Method: http.MethodGet,
			Path:   devenv.WsPath(ctx, fmt.Sprintf("/sessions/%s/template-diff", sessionID)),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("compare session template", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

// DeleteTerminatedSessions deletes all terminated sessions.
var DeleteTerminatedSessions = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_delete_terminated",
		mcp.WithTitleAnnotation("Delete terminated sessions"),
		mcp.WithDescription(`Delete all terminated devenv sessions in the given ownership scope. By default (scope="mine") deletes the current user's terminated sessions; set scope="workspace" to delete terminated workspace-owned sessions instead. With a Workspace API Token pass scope="workspace" (the token has no personal sessions, so "mine" is rejected). Returns the number of deleted sessions. Running sessions are left alone — use bitrise_devenv_delete to delete a specific session regardless of its state.`),
		mcp.WithString("scope",
			mcp.Description(`Ownership scope of the cleanup. "mine" (default) deletes the calling user's own terminated sessions. "workspace" deletes terminated sessions owned by the workspace itself — created with owner="workspace" or by a Workspace API Token, or started from a preview link. Use "workspace" when the server runs with a Workspace API Token.`),
			mcp.Enum("mine", "workspace"),
			mcp.DefaultString("mine"),
		),
		mcp.WithDestructiveHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		body := map[string]any{}
		if request.GetString("scope", "") == "workspace" {
			body["scope"] = "SESSION_LIST_SCOPE_WORKSPACE"
		}
		res, err := devenv.CallAPI(ctx, devenv.CallAPIParams{
			Method: http.MethodPost,
			Path:   devenv.WsPath(ctx, "/sessions:delete-terminated"),
			Body:   body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("delete terminated sessions", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

// requiredProperties marks the listed keys of an object-typed tool parameter
// as required in its nested JSON schema. mcp.Required() only marks the
// parameter itself as required on the top-level schema; nothing in mcp-go
// sets "required" inside a nested object built with mcp.Properties.
func requiredProperties(names ...string) mcp.PropertyOption {
	return func(schema map[string]any) {
		schema["required"] = names
	}
}

// isZeroArgument reports whether a tool argument carries no information — an
// empty string, false, an empty array or an empty object — so a client that
// sends every parameter it knows with a blank value is not mistaken for one
// that asked for something.
func isZeroArgument(v any) bool {
	switch x := v.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(x) == ""
	case bool:
		return !x
	case []any:
		return len(x) == 0
	case map[string]any:
		return len(x) == 0
	}
	return false
}

// requireNonEmptyString checks that an object-typed argument carries a
// non-empty string under key, returning a client-facing error that names the
// parameter when it does not.
func requireNonEmptyString(arg any, param, key string) error {
	obj, ok := arg.(map[string]any)
	if !ok {
		return fmt.Errorf("%s must be an object with a %q field", param, key)
	}
	val, ok := obj[key].(string)
	if !ok || strings.TrimSpace(val) == "" {
		return fmt.Errorf("%s.%s is required and must be a non-empty string", param, key)
	}
	return nil
}

// deviceSpecFieldsDoc explains the fields of a device_spec object. It is
// shared by every tool that accepts one so the field semantics are described
// identically on sessions and templates.
const deviceSpecFieldsDoc = `platform: "ios" (simulator; macOS stack) or "android" (emulator; Linux stack). Everything else is optional: device_model (simctl device type like "iPhone 16" / emulator device profile like "pixel_7" — a screen profile on the generic emulator/simulator software, not that phone's firmware; empty = platform default), os_version (iOS only: "18.2" or a simctl runtime id; empty = newest installed), system_image (Android only — the API-level knob: sdkmanager package like "system-images;android-34;google_apis;x86_64"; empty = the stack's default), ram_mb / cores (Android only: explicit emulator sizing, 0 = host-derived), cold_boot (Android only).`

// deviceSpecSchema returns the property options of a device_spec object
// parameter — the same shape wherever a virtual device is described
// (bitrise_devenv_create, bitrise_devenv_create_template,
// bitrise_devenv_update_template) — under the given tool-specific description.
// requiredKeys lists the nested fields the schema marks required: templates
// always need a platform (a declared device is a complete device), while a
// session's device_spec may omit it to tweak the template's device per field.
func deviceSpecSchema(description string, requiredKeys ...string) []mcp.PropertyOption {
	opts := []mcp.PropertyOption{
		mcp.Description(description),
		mcp.Properties(map[string]any{
			"platform":     map[string]any{"type": "string", "enum": []string{"ios", "android"}, "description": `"ios" or "android"`},
			"device_model": map[string]any{"type": "string", "description": "simctl device type (iOS) or emulator device profile (Android); empty = platform default"},
			"os_version":   map[string]any{"type": "string", "description": "iOS only: an iOS version such as \"18.2\" or a simctl runtime id — anything else (and any value on Android) is rejected; empty = newest installed"},
			"system_image": map[string]any{"type": "string", "description": "Android only: sdkmanager system image package; empty = platform default"},
			"ram_mb":       map[string]any{"type": "integer", "description": "Android only: emulator RAM in MB; 0 = host-derived"},
			"cores":        map[string]any{"type": "integer", "description": "Android only: emulator CPU cores; 0 = host-derived"},
			"cold_boot":    map[string]any{"type": "boolean", "description": "Android only: full cold boot every start (no quickboot)"},
		}),
	}
	if len(requiredKeys) > 0 {
		opts = append(opts, requiredProperties(requiredKeys...))
	}
	return opts
}
