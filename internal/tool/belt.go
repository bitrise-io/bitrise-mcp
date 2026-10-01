package tool

import (
	"context"
	"slices"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/apps"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/artifacts"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/builds"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/cache"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/configuration"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/devenvironments"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/grouproles"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/insights"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/outputschema"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/pipelines"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/releasemanagement"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/releasemanagement/codepush"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/user"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/webhooks"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/workspaces"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Belt struct {
	tools map[string]bitrise.Tool
	// devenv is the Dev Environments sub-belt: it owns the bitrise_devenv_*
	// tools, their MCP resources, and the per-request workspace resolution
	// those tools need.
	devenv *devenvironments.Belt
}

func NewBelt() *Belt {
	var toolList = []bitrise.Tool{
		// User
		user.Me,

		// Apps
		apps.List,
		apps.Register,
		apps.Finish,
		apps.Get,
		apps.Delete,
		apps.Update,
		apps.GetBitriseYML,
		apps.UpdateBitriseYML,
		apps.ListBranches,
		apps.RegisterSSHKey,
		apps.RegisterWebhook,

		// Builds
		builds.Trigger,
		builds.List,
		builds.Get,
		builds.GetSteps,
		builds.Abort,
		builds.GetBuildLog,
		builds.GetBuildBitriseYML,
		builds.ListBuildWorkflows,

		// Artifacts
		artifacts.List,
		artifacts.Get,
		artifacts.Delete,
		artifacts.Update,

		// Workspaces
		workspaces.List,
		workspaces.Get,
		workspaces.GetWorkspaceGroups,
		workspaces.CreateWorkspaceGroup,
		workspaces.GetWorkspaceMembers,
		workspaces.InviteMemberToWorkspace,
		workspaces.AddMemberToGroup,

		// Webhooks
		webhooks.ListOutgoing,
		webhooks.DeleteOutgoing,
		webhooks.CreateOutgoing,
		webhooks.UpdateOutgoing,

		// Cache
		cache.ListItems,
		cache.DeleteAllItems,
		cache.DeleteItem,
		cache.GetItemDownloadURL,

		// Pipelines
		pipelines.List,
		pipelines.Get,
		pipelines.Abort,
		pipelines.Rebuild,

		// Group Roles
		grouproles.List,
		grouproles.Replace,

		// Release Management
		releasemanagement.CreateConnectedApp,
		releasemanagement.UpdateConnectedApp,
		releasemanagement.ListConnectedApps,
		releasemanagement.GetConnectedApp,
		releasemanagement.ListInstallableArtifacts,
		releasemanagement.GenerateInstallableArtifactUploadURL,
		releasemanagement.GetInstallableArtifactUploadAndProcessingStatus,
		releasemanagement.SetInstallableArtifactPublicInstallPage,
		releasemanagement.ListBuildDistributionVersions,
		releasemanagement.ListBuildDistributionVersionTestBuilds,
		releasemanagement.CreateTesterGroup,
		releasemanagement.NotifyTesterGroup,
		releasemanagement.AddTestersToTesterGroup,
		releasemanagement.UpdateTesterGroup,
		releasemanagement.ListTesterGroups,
		releasemanagement.GetTesterGroup,
		releasemanagement.GetPotentialTesters,
		releasemanagement.GetTesters,

		// Configuration
		configuration.ValidateBitriseYML,
		configuration.StepSearch,
		configuration.StepInputs,
		configuration.ListAvailableStacks,

		// CodePush
		codepush.ListDeployments,
		codepush.GetDeployment,
		codepush.CreateDeployment,
		codepush.UpdateDeployment,
		codepush.DeleteDeployment,
		codepush.PromoteDeployment,
		codepush.RollbackDeployment,
		codepush.ListUpdates,
		codepush.GetUpdate,
		codepush.PatchUpdate,
		codepush.DeleteUpdate,
		codepush.GetUpdateStatus,
		codepush.GenerateUpdateUploadURL,
		codepush.GetMetrics,

		// Insights
		insights.GetBuildTotals,
		insights.GetBuildSeries,
		insights.GetTestTotals,
		insights.GetTestSeries,
		insights.ListFlakyTests,
	}

	// Dev Environments (RDE)
	devenvBelt := devenvironments.NewBelt()
	toolList = append(toolList, devenvBelt.Tools()...)

	belt := &Belt{tools: make(map[string]bitrise.Tool, len(toolList)), devenv: devenvBelt}
	for _, tool := range toolList {
		if _, dup := belt.tools[tool.Definition.Name]; dup {
			panic("duplicate tool name: " + tool.Definition.Name)
		}
		// Every tool advertises the schema of its result (declared in code or
		// generated from the API document) and, to honour it, returns the
		// result as structuredContent too. Tools without a schema (image
		// results) are served as they are.
		if outputschema.Apply(&tool.Definition) {
			tool.Handler = outputschema.WithStructuredContent(tool.Handler)
		}
		// Clients prefer the spec-level title over annotations.title.
		if tool.Definition.Title == "" {
			tool.Definition.Title = tool.Definition.Annotations.Title
		}
		belt.tools[tool.Definition.Name] = tool
	}
	return belt
}

func (b *Belt) RegisterAll(server *server.MCPServer) {
	for _, tool := range b.tools {
		server.AddTool(tool.Definition, tool.Handler)
	}
	b.devenv.RegisterResources(server)
}

// Tools returns every registered tool, in no particular order.
func (b *Belt) Tools() []bitrise.Tool {
	out := make([]bitrise.Tool, 0, len(b.tools))
	for _, t := range b.tools {
		out = append(out, t)
	}
	return out
}

// ToolFilter returns the server.WithToolFilter function: it lists only the
// tools whose API group is enabled (from the x-bitrise-enabled-api-groups
// header on the HTTP transport, else defaultGroups) and hides the
// local-only Dev Environments tools on the hosted transport.
func (b *Belt) ToolFilter(defaultGroups []string) server.ToolFilterFunc {
	return func(ctx context.Context, tools []mcp.Tool) []mcp.Tool {
		enabledGroups, err := bitrise.EnabledGroupsFromCtx(ctx) // http transport only
		if err != nil {
			// stdio transport, or no per-request override on the http transport
			enabledGroups = defaultGroups
		}
		filtered := make([]mcp.Tool, 0, len(tools))
		for _, tool := range tools {
			if b.ToolEnabled(tool.Name, enabledGroups) {
				filtered = append(filtered, tool)
			}
		}
		return b.devenv.FilterTools(ctx, filtered)
	}
}

// GateAndResolveWorkspace is the per-call middleware step for the Dev
// Environments tools: it rejects local-only tools on the hosted transport and
// resolves the workspace the call operates in. Other tools pass through
// untouched. The PAT (and any per-connection default workspace) must already
// be in ctx.
func (b *Belt) GateAndResolveWorkspace(ctx context.Context, request mcp.CallToolRequest) (context.Context, *mcp.CallToolResult) {
	return b.devenv.GateAndResolveWorkspace(ctx, request)
}

func (b *Belt) ToolEnabled(name string, enabledGroups []string) bool {
	tool, ok := b.tools[name]
	if !ok {
		return false
	}
	for _, enabledGroup := range enabledGroups {
		if slices.Contains(tool.APIGroups, enabledGroup) {
			return true
		}
	}
	return false
}
