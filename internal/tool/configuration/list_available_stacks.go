package configuration

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var ListAvailableStacks = bitrise.Tool{
	APIGroups: []string{"configuration", "read-only"},
	Definition: mcp.NewTool("list_available_stacks",
		mcp.WithDescription("List available stacks with their machine configurations and version information. When a workspace_slug is provided, returns stacks available for that workspace including any custom stacks. When omitted, returns globally available stacks."),
		mcp.WithString("workspace_slug",
			mcp.Description("Slug of the Bitrise workspace. When provided, lists stacks available for that workspace (including custom stacks). When omitted, lists globally available stacks."),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path := "/available-stacks"
		if slug := request.GetString("workspace_slug", ""); slug != "" {
			path = fmt.Sprintf("/organizations/%s/available-stacks", slug)
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIBaseURL,
			Path:    path,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
