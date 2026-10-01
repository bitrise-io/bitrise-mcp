package devenvironments

import (
	"context"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/devenv"
	"github.com/mark3labs/mcp-go/mcp"
)

// GetWorkspaceUsage reports the workspace's active-session resource usage.
var GetWorkspaceUsage = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_get_workspace_usage",
		mcp.WithTitleAnnotation("Get Dev Environments Usage"),
		mcp.WithDescription(`Dev Environments (RDE) only, not billing and not CI (CI build aggregates: insights_get_build_totals; CodePush quotas: codepush_get_metrics; workspace profile: get_workspace). A point-in-time snapshot of the workspace's ACTIVE sessions: session counts and vCPU/memory totals split by OS, workspace-wide and per user.

This reports sessions currently consuming resources (starting, running, terminating, or draining). It is NOT a historical or billing-period report — poll it over time if you need trends.

Requires the workspace's billing-view permission (workspace owners and billing-managing custom roles); other members get a permission-denied error.

Response shape (zero-valued fields and empty objects may be omitted from the JSON — treat a missing field as 0/false/empty):
- totals: workspace-wide usage, split into linux, macos, and unknown (OS could not be determined) buckets. Each bucket has:
  - sessionCount: number of active sessions.
  - vcpu: total vCPUs across those sessions.
  - memoryGb: total memory in GB across those sessions.
- users: per-user breakdown, sorted by email. Each row has:
  - userSlug: the user's Bitrise slug, usable with other Bitrise APIs. Best-effort: may be empty even on user rows, so use isWorkspace (not an empty userSlug) to detect the workspace bucket.
  - email / username: the user's identity (username is best-effort and may be empty).
  - userId: internal identifier; prefer userSlug for cross-referencing.
  - isWorkspace: true for the single aggregate row of sessions owned by the workspace itself rather than by an individual user (for example sessions spawned by a preview link); that row has no userSlug/email.
  - totals: this row's usage, same linux/macos/unknown bucket shape as the workspace-wide totals.
- unknownMachineTypeCount: number of active sessions whose machine type had no resolvable vCPU/RAM spec. Those sessions are counted in sessionCount but contribute 0 to the vcpu/memory sums, so totals undercount when this is non-zero — mention that caveat when presenting the numbers.`),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		res, err := devenv.CallAPI(ctx, devenv.CallAPIParams{
			Method: http.MethodGet,
			Path:   devenv.WsPath(ctx, "/usage"),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("get workspace usage", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
