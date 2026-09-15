package releasemanagement

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var AddTestersToTesterGroup = bitrise.Tool{
	APIGroups: []string{"release-management"},
	Definition: mcp.NewTool("add_testers_to_tester_group",
		mcp.WithDescription("Adds testers to a tester group of a connected app."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier of the related Release Management connected app."),
			mcp.Required(),
		),
		mcp.WithString("id",
			mcp.Description("The uuidV4 identifier of the tester group to which testers will be added."),
			mcp.Required(),
		),
		mcp.WithArray("user_slugs",
			mcp.Description("User slugs to add as internal testers. Required for internal tester groups; ignored for external tester groups."),
			mcp.WithStringItems(),
		),
		mcp.WithArray("emails",
			mcp.Description("Email addresses to add as external testers. Required for external tester groups (at most 1000 entries); ignored for internal tester groups."),
			mcp.WithStringItems(),
		),
		mcp.WithTitleAnnotation("Add Testers to Tester Group"),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if _, err := request.RequireString("connected_app_id"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		body := map[string]any{}
		if v := request.GetStringSlice("user_slugs", nil); len(v) > 0 {
			body["user_slugs"] = v
		}
		if v := request.GetStringSlice("emails", nil); len(v) > 0 {
			body["emails"] = v
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APIRMBuildDistributionsBaseURL,
			Path:    fmt.Sprintf("/tester-groups/%s/add-testers", id),
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
