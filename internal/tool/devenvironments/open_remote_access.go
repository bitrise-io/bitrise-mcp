package devenvironments

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/devenv"
	"github.com/mark3labs/mcp-go/mcp"
)

// OpenRemoteAccess opens remote access (SSH/VNC) for a session.
var OpenRemoteAccess = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_open_remote_access",
		mcp.WithTitleAnnotation("Open remote access"),
		mcp.WithDescription(`Open SSH/VNC remote access on a running Dev Environments session and return its credentials (unrelated to Bitrise CI SSH keys, register_ssh_key).

This establishes the remote access tunnel and returns connection details.
The session must be in "running" status.

On macOS sessions, this returns SSH and VNC connection details (address, username, password).
On Linux sessions, this returns SSH connection details only.

IMPORTANT: Most running sessions already have remote access connection details available
from bitrise_devenv_get. Only use this tool if the session does NOT already have these
details populated.

The VNC details form a vnc://username:password@host:port URL a VNC client can open.`),
		mcp.WithString("session_id",
			mcp.Description("The unique identifier of the running session"),
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
			Path:   devenv.WsPath(ctx, fmt.Sprintf("/sessions/%s/open-remote-access", sessionID)),
			Body:   map[string]any{},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("open remote access", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
