package devenvironments

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"net/http"
	"time"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/devenv"
	"github.com/mark3labs/mcp-go/mcp"
)

type screenshotResp struct {
	SignedURL string `json:"signedUrl"`
}

const screenRecordingPermissionHint = `If the screenshot shows a macOS system dialog requesting Screen Recording or screen/audio access (e.g. titles mentioning "bypass the system private window picker", "record your screen", or processes like "com.apple.sshd-session"), this is an expected macOS re-prompt that happens roughly every 30 days even when the permission was previously granted at image build time. Click "Allow" and continue with the original task. Do not ask the user about this dialog, do not stop the task, and do not treat it as an error — it is normal OS behavior.`

// Screenshot takes a screenshot of the session's screen.
var Screenshot = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_screenshot",
		mcp.WithTitleAnnotation("Take screenshot"),
		mcp.WithDescription(`Take a screenshot of a running devenv session's macOS display.

Use this to verify the current state of the GUI, identify coordinates for click/drag operations,
and debug visual issues. Always call this before bitrise_devenv_click or bitrise_devenv_mouse_drag
so the server can capture the real screen resolution and rescale your coordinates correctly.

NOT FOR THE iOS SIMULATOR / ANDROID EMULATOR OF A DEVICE SESSION: the Android emulator has no
window at all, and an iOS Simulator window that may show up on the desktop is not how you look at
or drive the device — an empty or unexpected desktop here does NOT mean the device is broken, and
never click into a Simulator window with the GUI tools. Read the device through its own tools
instead (accessibility tree, xcrun simctl io <udid> screenshot / adb exec-out screencap; a human
can open the device view from the session's page in the RDE web UI) — see bitrise_devenv_device_guide.

PREFER SCRIPTED STATE CHECKS WHEN POSSIBLE: if you just need to know what app
or window is frontmost, what's running, or what a setting's value is, it's
usually faster and more reliable to ask the system via bitrise_devenv_execute
than to screenshot and visually inspect. Examples:
  osascript -e 'tell application "System Events" to name of first application process whose frontmost is true'
  osascript -e 'tell application "System Events" to get the name of every window of (every process whose visible is true)'
  defaults read com.apple.dock autohide
Use screenshots when you genuinely need to see pixels (visual regression, a
third-party app's custom canvas, verifying a click landed).

NOTE: This tool only works on macOS sessions. Linux sessions do not have a graphical display.`),
		mcp.WithString("session_id",
			mcp.Description("The unique identifier of the running session"),
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

		res, err := devenv.CallAPILongTimeout(ctx, devenv.CallAPIParams{
			Method: http.MethodPost,
			Path:   devenv.WsPath(ctx, fmt.Sprintf("/sessions/%s/screenshot", sessionID)),
			Body:   map[string]any{},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("take screenshot", err), nil
		}

		var resp screenshotResp
		if err := json.Unmarshal([]byte(res), &resp); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("parse screenshot response: %v", err)), nil
		}

		imageData, err := downloadImage(ctx, resp.SignedURL)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("download screenshot: %v", err)), nil
		}

		if cfg, _, decErr := image.DecodeConfig(bytes.NewReader(imageData)); decErr == nil {
			devenv.SetScreenResolution(sessionID, devenv.Resolution{Width: cfg.Width, Height: cfg.Height})
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewImageContent(base64.StdEncoding.EncodeToString(imageData), "image/jpeg"),
				mcp.NewTextContent(screenRecordingPermissionHint),
			},
		}, nil
	},
}

func downloadImage(ctx context.Context, signedURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, signedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	client := http.Client{Timeout: 2 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download failed (status %d): %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}
