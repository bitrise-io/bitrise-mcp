package tool

import (
	"context"
	"slices"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

// Instructions are sent to clients in the initialize response, so the model
// knows that disabled API groups exist and how to find them.
const Instructions = "Bitrise MCP tools are organized into API groups, and the user's MCP client configuration decides which groups are enabled. " +
	"If a task needs a capability that none of the available tools provide, call list_api_groups: it lists every group, whether it is enabled, and the tools it would add. " +
	"Tools of a disabled group cannot be called. To get them, the user has to enable the group, with the x-bitrise-enabled-api-groups header for the remote server or the ENABLED_API_GROUPS environment variable for a local one, and reconnect the client."

const howToEnableGroups = "Tools of a disabled group are unavailable until the user enables the group in their MCP client configuration: " +
	"set the x-bitrise-enabled-api-groups HTTP header (remote server) or the ENABLED_API_GROUPS environment variable (local server) " +
	"to a comma-separated list of groups, then reconnect the client. Enabling a group adds its unavailable_tools."

var apiGroupDescriptions = map[string]string{
	"account":                               "The authenticated user's account.",
	"user":                                  "Older name of account, kept for existing configurations.",
	"apps":                                  "Bitrise apps: register and configure them, read and update bitrise.yml, branches, SSH keys and webhooks.",
	"builds":                                "Builds: trigger, list and inspect them, abort them, and read their logs, steps and test results.",
	"artifacts":                             "Build artifacts.",
	"pipelines":                             "Pipelines: list and inspect them, abort and rebuild them.",
	"workspaces":                            "Workspaces, their members and groups.",
	"outgoing-webhooks":                     "Outgoing webhooks of an app.",
	"group-roles":                           "Roles of workspace groups on an app.",
	"cache-items":                           "Build cache items.",
	"configuration":                         "bitrise.yml validation, Step search and Step inputs, available stacks.",
	"insights":                              "Build and test metrics from Bitrise Insights.",
	"read-only":                             "Every tool that only reads data, across all groups.",
	"release-management":                    "Every Release Management tool.",
	"release-management-app":                "Release Management connected apps and installable artifacts.",
	"release-management-build-distribution": "Build distribution: versions, test builds, tester groups and testers.",
	"release-management-code-push":          "CodePush deployments and updates.",
	"release-management-store-releases":     "Store releases: app versions from release candidate and beta testing through approvals and store review to the rollout.",
}

type apiGroupsResult struct {
	EnabledGroups []string   `json:"enabled_groups" jsonschema:"The API groups enabled for this connection."`
	HowToEnable   string     `json:"how_to_enable" jsonschema:"How the user enables more groups."`
	Groups        []apiGroup `json:"groups" jsonschema:"Every API group of the server."`
}

type apiGroup struct {
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Enabled          bool     `json:"enabled"`
	Tools            []string `json:"tools" jsonschema:"Every tool in the group."`
	UnavailableTools []string `json:"unavailable_tools,omitempty" jsonschema:"Tools of the group that no enabled group provides; enabling this group adds them."`
}

func listAPIGroups(b *Belt) bitrise.Tool {
	return bitrise.Tool{
		AlwaysEnabled: true,
		Definition: mcp.NewTool("list_api_groups",
			mcp.WithDescription("Lists the API groups of the Bitrise MCP server, whether each is enabled for this connection, and the tools each group provides. "+
				"Call it when a task needs a capability that none of the available tools provide: unavailable_tools shows which disabled group would add the missing tool. "+
				"The user has to enable a group in their MCP client configuration; it cannot be enabled from a tool call."),
			mcp.WithOutputSchema[apiGroupsResult](),
			mcp.WithTitleAnnotation("List API Groups"),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithOpenWorldHintAnnotation(false),
			mcp.WithIdempotentHintAnnotation(true),
		),
		Handler: func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			enabled, err := bitrise.EnabledGroupsFromCtx(ctx)
			if err != nil {
				enabled = b.groupNames()
			}
			return mcp.NewToolResultStructuredOnly(b.apiGroups(enabled)), nil
		},
	}
}

func (b *Belt) groupNames() []string {
	var names []string
	for _, t := range b.tools {
		for _, g := range t.APIGroups {
			if !slices.Contains(names, g) {
				names = append(names, g)
			}
		}
	}
	slices.Sort(names)
	return names
}

func (b *Belt) apiGroups(enabled []string) apiGroupsResult {
	result := apiGroupsResult{EnabledGroups: enabled, HowToEnable: howToEnableGroups}
	for _, name := range b.groupNames() {
		group := apiGroup{Name: name, Description: apiGroupDescriptions[name], Enabled: slices.Contains(enabled, name)}
		for toolName, t := range b.tools {
			if !slices.Contains(t.APIGroups, name) {
				continue
			}
			group.Tools = append(group.Tools, toolName)
			if !b.ToolEnabled(toolName, enabled) {
				group.UnavailableTools = append(group.UnavailableTools, toolName)
			}
		}
		slices.Sort(group.Tools)
		slices.Sort(group.UnavailableTools)
		result.Groups = append(result.Groups, group)
	}
	return result
}
