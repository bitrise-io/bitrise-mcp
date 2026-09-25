package devenvironments

import (
	"context"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/devenv"
	"github.com/mark3labs/mcp-go/mcp"
)

// ListStacks lists available development environment stacks.
var ListStacks = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_list_stacks",
		mcp.WithTitleAnnotation("List Dev Environment Stacks"),
		mcp.WithDescription(`List the stacks a Dev Environments session or template can be provisioned on. This is a separate catalog from Bitrise CI's list_available_stacks (IDs differ and are not interchangeable); pick one by what the user is doing, do not call both.

Each stack describes a provisionable development environment:
- id: stable stack identifier (e.g. 'osx-xcode-16.0.x-edge'). This is the value to store and to pass as stack_id when creating a session or template.
- title: human-friendly label (e.g. 'Xcode 16.0'). Show this to the user; fall back to id when title is empty.
- description / descriptionLink: summary and a link to the stack's pre-installed tools / system report.
- os: 'macos' or 'linux'.
- os_version: numeric OS version (e.g. 26 for macOS, 24 for Ubuntu 24.04).
- status: 'edge', 'stable', or 'frozen'.
- xcodeVersion: Xcode version (e.g. '16.0'); empty for non-Xcode stacks. Informational.
- isDefault: when true, this is the deployment's default stack — preselect it when the user has expressed no preference.
- clusterNames: the clusters where the stack can be provisioned. A machine type is compatible with the stack when its clusterName is one of these (see bitrise_devenv_list_machine_types).`),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		res, err := devenv.CallAPI(ctx, devenv.CallAPIParams{
			Method: http.MethodGet,
			Path:   devenv.WsPath(ctx, "/stacks"),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("list stacks", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

// ListMachineTypes lists available machine types.
var ListMachineTypes = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_list_machine_types",
		mcp.WithTitleAnnotation("List Dev Environment Machine Types"),
		mcp.WithDescription(`List the machine types for Dev Environments sessions and templates. Bitrise CI has no machine-type tool: CI machine sizes are listed per stack by list_available_stacks.

Each machine type includes a name (use the name, not the ID, when creating or updating templates), a friendly title, cpu/ram specs, the os it runs, and the clusterName it belongs to.

To pick a machine type compatible with a stack, choose one whose clusterName is in that stack's clusterNames (from bitrise_devenv_list_stacks).`),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		res, err := devenv.CallAPI(ctx, devenv.CallAPIParams{
			Method: http.MethodGet,
			Path:   devenv.WsPath(ctx, "/machine-types"),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("list machine types", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
