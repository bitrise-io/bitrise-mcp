package devenvironments

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/devenv"
	"github.com/mark3labs/mcp-go/mcp"
)

// executeTimeoutLocal caps one execute call on a locally-run (stdio) server.
const executeTimeoutLocal = 2 * time.Minute

// executeTimeoutHosted caps one execute call on the hosted HTTP server. The
// proxy chain in front of the hosted server answers 504 at roughly 40 s
// (measured on staging: a `sleep 100` came back as a Cloudflare 504 after
// 39.5 s, well under Cloudflare's own 100 s default — the cut sits closer to
// the origin). A longer server-side deadline only produces a gateway error the
// client cannot interpret (the command may have finished) instead of this
// tool's own error, so the cap stays under the cut with margin. Raise it only
// together with that ingress timeout.
const executeTimeoutHosted = 30 * time.Second

// executeTimeoutFor returns the per-call deadline for the transport in ctx.
func executeTimeoutFor(ctx context.Context) time.Duration {
	if devenv.HostedModeFromCtx(ctx) {
		return executeTimeoutHosted
	}
	return executeTimeoutLocal
}

// getSessionResponse mirrors the minimal shape of GetSessionResponse from the
// codespaces backend: the session object is wrapped under a top-level
// "session" key (see proto/codespaces/v1/codespaces.proto). Only the fields
// needed to open an SSH connection are deserialized.
type getSessionResponse struct {
	Session sessionSSHFields `json:"session"`
}

// Field names are camelCase because grpc-gateway's default protojson marshaler
// emits proto3 JSON spec names (lowerCamelCase), not the snake_case proto
// field names. The backend doesn't set UseProtoNames: true.
type sessionSSHFields struct {
	Status            string `json:"status"`
	SSHAddress        string `json:"sshAddress"`
	SSHPassword       string `json:"sshPassword"`
	SSHConnectionOpen bool   `json:"sshConnectionOpen"`
}

// Execute runs a bash command on a session's machine over a direct SSH
// connection from the MCP server to the session VM.
// ExecuteResult is the result of bitrise_devenv_execute.
type ExecuteResult struct {
	ExitCode int    `json:"exit_code" jsonschema_description:"The command's exit status; 0 means success."`
	Stdout   string `json:"stdout" jsonschema_description:"Standard output of the command."`
	Stderr   string `json:"stderr" jsonschema_description:"Standard error of the command, including bash's harmless 'no job control' startup lines."`
}

var Execute = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_execute",
		mcp.WithOutputSchema[ExecuteResult](),
		mcp.WithTitleAnnotation("Run command in session"),
		mcp.WithDescription(`Run a shell command inside an EXISTING Dev Environments session (a remote VM, identified by session_id). This is not Bitrise CI: to run a CI workflow or pipeline use trigger_bitrise_build.

The command runs over a direct SSH connection from the MCP server to the session VM,
inside a forced-interactive login bash shell (bash -i -l -c). This ensures the full
shell environment is available — PATH, brew-installed binaries, git-lfs, language
version managers (nvm, pyenv, rbenv, asdf), and anything the template's warmup
script writes to ~/.bashrc, ~/.bash_profile, ~/.profile, or /etc/profile are all
loaded.

Output is returned as a JSON object with three fields:
- exit_code: the command's exit status (0 = success)
- stdout:    captured stdout as a string
- stderr:    captured stderr as a string

When the MCP server runs locally (stdio) and the host has an SSH agent
(SSH_AUTH_SOCK), that agent is forwarded into the session, so remote commands
that authenticate over SSH ("git push", "git clone git@github.com:...") can use
the caller's keys. The hosted server forwards nothing: set up credentials in
the session yourself there.

NOTE: Because the remote shell is forced-interactive without a TTY, bash emits two
harmless startup diagnostic lines to stderr on every invocation:
  "bash: cannot set terminal process group (-1): Inappropriate ioctl for device"
  "bash: no job control in this shell"
These are not errors from the user's command — ignore them. The exit_code field is
the source of truth for success/failure, not the presence of stderr output.

IMPORTANT:
- The session must be in "running" status with SSH remote access available.
  SSH remote access is provisioned automatically; if credentials aren't populated
  yet, the session is likely still starting up — wait briefly and retry.
- Time cap per call: about 30 seconds on the hosted server, 2 minutes on a
  locally-run one. Past it the call fails and the command is cancelled. Chunk
  work to the cap; background anything longer and poll its log (see next
  point).
- A gateway timeout / 504 does NOT mean the command did not run — it may have
  completed or still be running on the VM. Before retrying anything that mutates
  state (installs, boots, file writes), verify with a read-only command.
- For background processes, redirect output: "nohup ./server &>/dev/null &"
  and poll: "tail -n 50 /tmp/x.log"
- For large outputs, pipe through head: "find / -name '*.log' | head -100"
- Commands run as the session user (vagrant on macOS, ubuntu on Linux).

macOS UI automation: most UI actions (System Settings panes, launching apps,
menus, keystrokes, defaults) are scriptable from this tool with "open",
"osascript" and "defaults", which is faster and more reliable than a
screenshot + click loop; wrap osascript in "timeout 15s" so an unexpected
permission prompt fails fast. Recipes: bitrise_devenv_device_guide with
guide="macos-automation" (also the resource bitrise-devenv://guides/macos-automation).`),
		mcp.WithString("session_id",
			mcp.Description("The unique identifier of the running session"),
			mcp.Required(),
		),
		mcp.WithString("command",
			mcp.Description("The bash command to execute"),
			mcp.Required(),
		),
		mcp.WithDestructiveHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sessionID, err := requireUUID(request, "session_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		command, err := request.RequireString("command")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		sessionJSON, err := devenv.CallAPI(ctx, devenv.CallAPIParams{
			Method: http.MethodGet,
			Path:   devenv.WsPath(ctx, fmt.Sprintf("/sessions/%s", sessionID)),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("fetch session for execute", err), nil
		}
		var resp getSessionResponse
		if err := json.Unmarshal([]byte(sessionJSON), &resp); err != nil {
			return mcp.NewToolResultErrorFromErr("parse session response", err), nil
		}
		s := resp.Session
		if s.Status != "SESSION_STATUS_RUNNING" {
			return mcp.NewToolResultError(fmt.Sprintf(
				"session is not running (status: %q); start the session before running commands",
				s.Status,
			)), nil
		}
		if !s.SSHConnectionOpen || s.SSHAddress == "" || s.SSHPassword == "" {
			return mcp.NewToolResultError(
				"session SSH is not ready yet: the session read does not report sshConnectionOpen=true with an sshAddress and sshPassword. Remote access opens automatically during provisioning, but it can take a few minutes after the session turns running — poll bitrise_devenv_get until sshConnectionOpen is true, then retry",
			), nil
		}

		target, err := parseSSHAddress(s.SSHAddress)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("parse session ssh address", err), nil
		}
		target.Password = s.SSHPassword

		execCtx, cancel := context.WithTimeout(ctx, executeTimeoutFor(ctx))
		defer cancel()

		client, err := dialSSH(execCtx, target)
		if err != nil {
			if hint := dnsFailureHint(err, runtime.GOOS); hint != "" {
				return mcp.NewToolResultError(fmt.Sprintf("ssh dial: %v\n\n%s", err, hint)), nil
			}
			return mcp.NewToolResultErrorFromErr("ssh dial", err), nil
		}
		defer client.Close()

		res, err := client.run(execCtx, command)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("execute command", err), nil
		}

		result := ExecuteResult{ExitCode: res.ExitCode, Stdout: string(res.Stdout), Stderr: string(res.Stderr)}
		payload, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("marshal execute result", err), nil
		}
		return mcp.NewToolResultStructured(result, string(payload)), nil
	},
}
