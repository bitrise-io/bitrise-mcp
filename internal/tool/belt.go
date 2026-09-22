package tool

import (
	"slices"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/apps"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/artifacts"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/builds"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/cache"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/configuration"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/grouproles"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/insights"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/pipelines"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/releasemanagement"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/releasemanagement/codepush"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/releasemanagement/storereleases"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/user"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/webhooks"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool/workspaces"
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

		// Store Releases
		storereleases.ListAppVersions,
		storereleases.GetAppVersion,
		storereleases.CreateAppVersion,
		storereleases.UpdateAppVersion,
		storereleases.DeleteAppVersion,
		storereleases.GetReleaseCandidate,
		storereleases.UploadReleaseCandidate,
		storereleases.GetReleaseCandidateUploadStatus,
		storereleases.SubmitReleaseCandidateForBetaReview,
		storereleases.ListBetaTestingGroups,
		storereleases.StartBetaTesting,
		storereleases.StopBetaTesting,
		storereleases.ListWhatToTestDescriptions,
		storereleases.CreateWhatToTestDescription,
		storereleases.UpdateWhatToTestDescription,
		storereleases.DeleteWhatToTestDescription,
		storereleases.ListApprovalTasks,
		storereleases.GetApprovalTask,
		storereleases.CreateApprovalTask,
		storereleases.UpdateApprovalTask,
		storereleases.DeleteApprovalTask,
		storereleases.SubmitForAppStoreReview,
		storereleases.GetAppStoreReviewStatus,
		storereleases.CancelAppStoreReview,
		storereleases.ReleaseToAppStore,
		storereleases.PauseAppStorePhasedRelease,
		storereleases.ContinueAppStorePhasedRelease,
		storereleases.CompleteAppStorePhasedRelease,
		storereleases.GetAppStoreReleaseStatus,
		storereleases.GetAppStoreReleaseSettings,
		storereleases.UpdateAppStoreReleaseSettings,
		storereleases.UpdateGooglePlayReleaseNotes,
		storereleases.GetGooglePlayRelease,
		storereleases.ReleaseToGooglePlay,
		storereleases.GetGooglePlayStagedRolloutSchedule,
		storereleases.CreateGooglePlayStagedRolloutSchedule,
		storereleases.PauseGooglePlayStagedRolloutSchedule,
		storereleases.ResumeGooglePlayStagedRolloutSchedule,
		storereleases.DeleteGooglePlayStagedRolloutSchedule,
		storereleases.ListReleaseEvents,
		storereleases.CreateAppStoreVersion,
		storereleases.GetAppStoreDraftVersion,
		storereleases.UpdateAppStoreDraftVersion,

		// Insights
		insights.GetBuildTotals,
		insights.GetBuildSeries,
		insights.GetTestTotals,
		insights.GetTestSeries,
		insights.ListFlakyTests,
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
