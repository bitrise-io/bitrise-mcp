package devenvironments

import (
	"context"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/devenv"
	"github.com/mark3labs/mcp-go/mcp"
)

// CreatePreviewLink mints a shareable device preview link for an app build.
var CreatePreviewLink = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_create_preview_link",
		mcp.WithTitleAnnotation("Create device preview link"),
		mcp.WithDescription(`Create a Dev Environments preview link: a shareable URL that opens an app build on a live iOS simulator or Android emulator in the recipient's browser — no Bitrise login, no local tooling, nothing for them to install. This is not Release Management build distribution (over-the-air install on real devices: set_installable_artifact_public_install_page, notify_tester_group, list_build_distribution_versions); use it when a human should try the app in a browser without installing anything.

WHEN TO USE THIS: you have built an app and want a HUMAN to try it — a reviewer on a pull request, a designer checking a screen, a PM signing off. Hand them the url from the response.

WHEN NOT TO USE THIS: you want to drive a device YOURSELF (install, tap, screenshot, run tests). That is bitrise_devenv_create with device_spec — a session you control over adb / xcrun simctl. A preview link gives YOU nothing to drive; it only produces a URL for someone else.

Nothing is stored: the response is the only artifact, a link nobody opens costs nothing, and every open spawns its own device, so one link serves several reviewers at once.

THE APP BUILD (artifact.url) must be an absolute http(s) URL that a plain anonymous GET can fetch — a presigned URL is fine, and it is never shown to viewers. It must be a SIMULATOR/EMULATOR build, not a device build: iOS wants a zipped .app carrying an arm64 simulator slice, Android an .apk that runs on x86_64. A device-signed build installs on neither, and the viewer finds out only after a device has booted. If all you have is a device build, rebuild for the simulator/emulator first.

THE LINK IS A BEARER CREDENTIAL. Anyone holding the URL can open a device on this workspace's bill, and there is no revocation list — a short ttl_seconds is the control. Share it the way you would share a password: a pull request or a team channel, never a public one.

FASTER OPENS FROM A WARM POOL: pass warm_pool_id to serve the link's opens from a warm pool (bitrise_devenv_list_warm_pools) — each open claims one of the pool's booted device sessions when available and cold-boots from the pool's configuration otherwise, so click-to-app is the app download rather than a VM boot. The pool must be WORKSPACE-owned (owner_type "workspace"; a personal pool's sessions carry personal credentials that must never reach an anonymous viewer) and its configuration must boot a device. The pool fixes the machine and the device: omit device_spec, stack_id and machine_type (device_spec may be given only if it equals the pool's device). Deleting the pool invalidates the link (later opens fail as expired); a pool that still exists but no longer boots the link's device degrades opens to the ordinary cold path instead, with a notice shown to the viewer.

LIMITS: ttl_seconds defaults to 24 hours, capped at 72. At most 5 devices alive per link and 20 per workspace; opens past the cap are refused. Each device auto-terminates after its idle window (session_auto_terminate_minutes, at least 10, default 60).

WHO OWNS THE DEVICES: every open spawns a session, and it is always owned by the workspace — also when you are logged in as yourself. You are recorded as its creator, but it is not one of your personal sessions: it is visible to and manageable by every workspace member, and you find it with bitrise_devenv_list scope="workspace", not in the default "mine" listing. A Workspace API Token (bitwat_…, e.g. from CI) mints the same kind of link.

Returns url (the shareable link — this is what you hand over), token (the same credential bare), jti (the link's id, recorded on every session it spawns) and expiresAt.

Device preview is enabled per workspace and per platform. PermissionDenied means the workspace does not have it yet — tell the user to ask Bitrise support, do not retry.`),
		mcp.WithObject("device_spec",
			deviceSpecSchema(`The virtual device each open boots. `+deviceSpecFieldsDoc+` platform is required. Stick to platform alone unless the user asked for a specific device or OS — the deployment's defaults are known-good, and a device_model or os_version the stack lacks is silently substituted. Omit it when warm_pool_id is set: the pool's configuration names the device.`, "platform")...,
		),
		mcp.WithString("warm_pool_id",
			mcp.Description("Serve the link's opens from this workspace-owned warm pool (UUID, from bitrise_devenv_list_warm_pools) whose configuration boots a device. When set, omit device_spec, stack_id and machine_type — the pool fixes them (stack_id / machine_type are rejected alongside it; device_spec only if it differs from the pool's). Sessions the link spawns are owned by the workspace."),
		),
		mcp.WithObject("artifact",
			mcp.Description(`The app build installed on every device this link opens (required — a preview link without an app has nothing to preview). url is an absolute http(s) URL fetched by anonymous GET; a presigned URL is fine and is never shown to viewers. iOS: a zipped simulator .app. Android: an .apk. app_name / build_number / commit_sha are display metadata shown on the viewer page — fill them in when you know them, so the reviewer can see which build they are looking at.`),
			mcp.Properties(map[string]any{
				"url":          map[string]any{"type": "string", "description": "absolute http(s) download URL of the simulator/emulator app build"},
				"app_name":     map[string]any{"type": "string"},
				"build_number": map[string]any{"type": "string"},
				"commit_sha":   map[string]any{"type": "string"},
			}),
			requiredProperties("url"),
		),
		mcp.WithNumber("ttl_seconds",
			mcp.Description("How long the link stays openable, in seconds. Omit for the default (24 hours); the maximum is 72 hours (259200). Prefer the shortest lifetime that covers the review — there is no way to revoke a link early."),
		),
		mcp.WithNumber("session_auto_terminate_minutes",
			mcp.Description("How long a device stays alive after its last viewer disconnects. Omit for the deployment default (60). At least 10 — a shorter window would reap devices mid-boot — and it can be tuned but never turned off."),
		),
		mcp.WithString("stack_id",
			mcp.Description("Stack the link's devices run on. Omit for the platform default, which is what you want in almost every case. If you pass it, it must fit the platform: an Android-flavored dockerless Linux stack for android, a macOS 26+ stack for ios. Rejected at mint with the reason if it does not."),
		),
		mcp.WithString("machine_type",
			mcp.Description("Machine type the link's devices run on. Omit for the platform default. If you pass it, it must match the platform's OS family and clear the device minimum (>= 4 vCPU / 6-8 GB) — a smaller machine never finishes booting the device. Discover values with bitrise_devenv_list_machine_types."),
		),
		mcp.WithDestructiveHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		deviceSpec, hasDevice := request.GetArguments()["device_spec"]
		artifact, hasArtifact := request.GetArguments()["artifact"]
		warmPoolID, err := optionalUUID(request, "warm_pool_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		stackID := request.GetString("stack_id", "")
		machineType := request.GetString("machine_type", "")

		// The nested "required" lists are advisory to the client, so check
		// what the backend is certain to reject before spending a round trip.
		// A warm pool names the device itself, so device_spec is required only
		// without one; with one, the machine must come from the pool.
		if warmPoolID != "" && (stackID != "" || machineType != "") {
			return mcp.NewToolResultError("warm_pool_id cannot be combined with stack_id or machine_type — the pool fixes the machine its devices run on"), nil
		}
		if !hasDevice && warmPoolID == "" {
			return mcp.NewToolResultError("device_spec is required — a preview link always names the device to boot (platform \"ios\" or \"android\"), unless warm_pool_id names a pool that does"), nil
		}
		if hasDevice {
			if _, ok := deviceSpec.(map[string]any); !ok {
				return mcp.NewToolResultError("device_spec must be an object with a \"platform\" field"), nil
			}
			if err := requireNonEmptyString(deviceSpec, "device_spec", "platform"); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
		}
		// Unlike a device session, which may boot empty for an agent to build
		// into, a link always installs an app — that is the whole point of it.
		if !hasArtifact {
			return mcp.NewToolResultError("artifact is required — a preview link exists to show an app build, so artifact.url must point at a simulator/emulator build (iOS: zipped .app, Android: .apk)"), nil
		}
		if err := requireNonEmptyString(artifact, "artifact", "url"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{
			"artifact": artifact,
		}
		if hasDevice {
			body["device_spec"] = deviceSpec
		}
		if warmPoolID != "" {
			body["warm_pool_id"] = warmPoolID
		}
		if ttl, ok, err := getOptionalInt(request, "ttl_seconds"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		} else if ok {
			body["ttl_seconds"] = ttl
		}
		if minutes, ok, err := getOptionalInt(request, "session_auto_terminate_minutes"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		} else if ok {
			body["session_auto_terminate_minutes"] = minutes
		}
		if stackID != "" {
			body["stack_id"] = stackID
		}
		if machineType != "" {
			body["machine_type"] = machineType
		}

		res, err := devenv.CallAPI(ctx, devenv.CallAPIParams{
			Method: http.MethodPost,
			Path:   devenv.WsPath(ctx, "/preview-links"),
			Body:   body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("create preview link", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
