package tool

import (
	"context"
	"slices"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Tool struct {
	APIGroups  []string
	Definition mcp.Tool
	Handler    func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
}

type Belt struct {
	tools map[string]Tool
}

func NewBelt() *Belt {
	var toolList = []Tool{
		// User
		me,

		// Apps
		listApps,
		registerApp,
		finishBitriseApp,
		getApp,
		deleteApp,
		updateApp,
		getBitriseYML,
		updateBitriseYML,
		listBranches,
		registerSSHKey,
		registerWebhook,

		// Builds
		triggerBitriseBuild,
		listBuilds,
		getBuild,
		abortBuild,
		getBuildLog,
		getBuildBitriseYML,
		listBuildWorkflows,

		// Artifacts
		listArtifacts,
		getArtifact,
		deleteArtifact,
		updateArtifact,

		// Workspaces
		listWorkspaces,
		getWorkspace,
		getWorkspaceGroups,
		createWorkspaceGroup,
		getWorkspaceMembers,
		inviteMemberToWorkspace,
		addMemberToGroup,

		// Webhooks
		listOutgoingWebhooks,
		deleteOutgoingWebhook,
		createOutgoingWebhook,
		updateOutgoingWebhook,

		// Cache
		listCacheItems,
		deleteAllCacheItems,
		deleteCacheItem,
		getCacheItemDownloadURL,

		// Pipelines
		listPipelines,
		getPipeline,
		abortPipeline,
		rebuildPipeline,

		// Group Roles
		listGroupRoles,
		replaceGroupRoles,

		// Release Management
		createConnectedApp,
		updateConnectedApp,
		listConnectedApps,
		getConnectedApp,
		listInstallableArtifacts,
		generateInstallableArtifactUploadURL,
		getInstallableArtifactUploadAndProcessingStatus,
		setInstallableArtifactPublicInstallPage,
		listBuildDistributionVersions,
		listBuildDistributionVersionTestBuilds,
		createTesterGroup,
		notifyTesterGroup,
		addTestersToTesterGroup,
		updateTesterGroup,
		listTesterGroups,
		getTesterGroup,
		getPotentialTesters,
		getTesters,
	}
	belt := &Belt{tools: make(map[string]Tool)}
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
