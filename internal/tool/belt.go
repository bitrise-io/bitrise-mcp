package tool

import (
	"slices"

	"github.com/bitrise-io/bitrise-mcp/internal/bitrise"
	"github.com/bitrise-io/bitrise-mcp/internal/tool/apps"
	"github.com/bitrise-io/bitrise-mcp/internal/tool/artifacts"
	"github.com/bitrise-io/bitrise-mcp/internal/tool/builds"
	"github.com/bitrise-io/bitrise-mcp/internal/tool/cache"
	"github.com/bitrise-io/bitrise-mcp/internal/tool/grouproles"
	"github.com/bitrise-io/bitrise-mcp/internal/tool/pipelines"
	"github.com/bitrise-io/bitrise-mcp/internal/tool/releasemanagement"
	"github.com/bitrise-io/bitrise-mcp/internal/tool/user"
	"github.com/bitrise-io/bitrise-mcp/internal/tool/webhooks"
	"github.com/bitrise-io/bitrise-mcp/internal/tool/workspaces"
	"github.com/mark3labs/mcp-go/server"
)

type Belt struct {
	tools map[string]bitrise.Tool
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
	}
	belt := &Belt{tools: make(map[string]bitrise.Tool)}
	for _, tool := range toolList {
		belt.tools[tool.Definition.Name] = tool
	}
	return belt
}

func (b *Belt) RegisterAll(server *server.MCPServer) {
	for _, tool := range b.tools {
		server.AddTool(tool.Definition, tool.Handler)
	}
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
