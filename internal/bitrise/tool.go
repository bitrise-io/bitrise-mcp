package bitrise

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

type Tool struct {
	APIGroups  []string
	Definition mcp.Tool
	Handler    func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
}
