package tool

import (
	"context"
	"slices"
	"strings"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

// AllToolsets enables every toolset when it appears in a toolset list.
const AllToolsets = "all"

// Instructions are sent to clients in the initialize response, so the model
// knows that disabled toolsets exist and how to find them.
const Instructions = "Bitrise MCP tools are organized into toolsets, and the user's MCP client configuration decides which toolsets are enabled. " +
	"If a task needs a capability that none of the available tools provide, call list_toolsets: it lists every toolset, whether it is enabled, and the tools it would add. " +
	"Tools of a disabled toolset cannot be called. To get them, the user has to enable the toolset, with the toolset query parameter of the server URL or the x-bitrise-enabled-toolsets header for the remote server, or the ENABLED_TOOLSETS environment variable for a local one, and reconnect the client."

const howToEnableToolsets = "Tools of a disabled toolset are unavailable until the user enables the toolset in their MCP client configuration: " +
	"set the toolset query parameter of the server URL (for example https://mcp.bitrise.io/?toolset=apps,builds) or the x-bitrise-enabled-toolsets HTTP header (remote server), " +
	"or the ENABLED_TOOLSETS environment variable (local server), to a comma-separated list of toolsets or to all, then reconnect the client. " +
	"Enabling a toolset adds its unavailable_tools."

var toolsetDescriptions = map[string]string{
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
	"read-only":                             "Every tool that only reads data, across all toolsets.",
	"release-management":                    "Every Release Management tool.",
	"release-management-app":                "Release Management connected apps and installable artifacts.",
	"release-management-build-distribution": "Build distribution: versions, test builds, tester groups and testers.",
	"release-management-code-push":          "CodePush deployments and updates.",
	"release-management-store-releases":     "Store releases: app versions from release candidate and beta testing through approvals and store review to the rollout.",
}

type toolsetsResult struct {
	EnabledToolsets []string  `json:"enabled_toolsets" jsonschema:"The toolsets enabled for this connection."`
	HowToEnable     string    `json:"how_to_enable" jsonschema:"How the user enables more toolsets."`
	Toolsets        []toolset `json:"toolsets" jsonschema:"Every toolset of the server."`
}

type toolset struct {
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Enabled          bool     `json:"enabled"`
	UnavailableTools []string `json:"unavailable_tools,omitempty" jsonschema:"Tools of the toolset that no enabled toolset provides; enabling this toolset adds them. Omitted when there are none."`
}

func listToolsets(b *Belt) bitrise.Tool {
	return bitrise.Tool{
		AlwaysEnabled: true,
		Definition: mcp.NewTool("list_toolsets",
			mcp.WithDescription("Lists the toolsets of the Bitrise MCP server, whether each is enabled for this connection, and the tools each disabled toolset would add. "+
				"Call it when a task needs a capability that none of the available tools provide: unavailable_tools shows which disabled toolset would add the missing tool. "+
				"The user has to enable a toolset in their MCP client configuration; it cannot be enabled from a tool call."),
			mcp.WithOutputSchema[toolsetsResult](),
			mcp.WithTitleAnnotation("List Toolsets"),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithOpenWorldHintAnnotation(false),
			mcp.WithIdempotentHintAnnotation(true),
		),
		Handler: func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			enabled, err := bitrise.EnabledGroupsFromCtx(ctx)
			if err != nil {
				enabled = b.toolsetNames()
			}
			return mcp.NewToolResultStructuredOnly(b.toolsets(enabled)), nil
		},
	}
}

// EnabledToolsets splits a comma-separated toolset list. A list that names
// AllToolsets enables every toolset.
func (b *Belt) EnabledToolsets(list string) []string {
	names := strings.Split(list, ",")
	if slices.Contains(names, AllToolsets) {
		return b.toolsetNames()
	}
	return names
}

func (b *Belt) toolsetNames() []string {
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

func (b *Belt) toolsets(enabled []string) toolsetsResult {
	result := toolsetsResult{EnabledToolsets: enabled, HowToEnable: howToEnableToolsets}
	for _, name := range b.toolsetNames() {
		set := toolset{Name: name, Description: toolsetDescriptions[name], Enabled: slices.Contains(enabled, name)}
		for toolName, t := range b.tools {
			if slices.Contains(t.APIGroups, name) && !b.ToolEnabled(toolName, enabled) {
				set.UnavailableTools = append(set.UnavailableTools, toolName)
			}
		}
		slices.Sort(set.UnavailableTools)
		result.Toolsets = append(result.Toolsets, set)
	}
	return result
}
