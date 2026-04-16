package apps

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var ListBranches = bitrise.Tool{
	APIGroups: []string{"apps", "read-only"},
	Definition: mcp.NewTool("list_branches",
		mcp.WithDescription("List the branches with existing builds of an app's repository."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithNumber("limit",
			mcp.Description("Max number of branches to return (default: 50). The API returns all branches sorted by last build date, so lower limits return the most recently active branches first."),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		limit := request.GetInt("limit", 50)

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/branches", appSlug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}

		// The API returns all branches in one shot with no server-side
		// pagination. Active repos can have thousands, so we truncate
		// client-side. Branches are already sorted by most-recently-built
		// first, so this naturally returns the most relevant ones.
		var response map[string]any
		if err := json.Unmarshal([]byte(res), &response); err != nil {
			return mcp.NewToolResultText(res), nil
		}

		if branches, ok := response["data"].([]any); ok {
			total := len(branches)
			if limit > 0 && len(branches) > limit {
				branches = branches[:limit]
			}
			response["data"] = branches
			response["meta"] = map[string]any{
				"returned": len(branches),
				"total":    total,
			}
		}

		return mcp.NewToolResultStructuredOnly(response), nil
	},
}

