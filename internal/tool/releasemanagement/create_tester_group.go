package releasemanagement

import (
	"context"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var CreateTesterGroup = bitrise.Tool{
	APIGroups: []string{"release-management"},
	Definition: mcp.NewTool("create_tester_group",
		mcp.WithDescription("Creates a tester group for a Release Management connected app, optionally with its members: 'user_slugs' for internal groups, 'emails' for external ones. Tester groups are used to distribute installable artifacts to testers, who are notified by email either automatically or manually. "+
			"Adding an internal tester grants the App Tester role on the app; removing them later does not revoke it. A group is notified once per build, so members added afterwards get no email for that build. "+
			"This endpoint has an elevated access level requirement. Only the owner of the related Bitrise Workspace, a workspace manager or the related project's admin can manage tester groups."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier of the related Release Management connected app."),
			mcp.Required(),
		),
		mcp.WithString("name",
			mcp.Description("The name for the new tester group. Must be unique in the scope of the connected app."),
			mcp.Required(),
		),
		mcp.WithBoolean("auto_notify",
			mcp.Description("If set to true it indicates that the tester group will receive notifications automatically. Ignored for external tester groups."),
			mcp.DefaultBool(false),
		),
		mcp.WithArray("user_slugs",
			mcp.Description("User slugs to add as internal testers. Ignored for external tester groups. Use get_potential_testers to look the slugs up."),
			mcp.WithStringItems(),
		),
		mcp.WithArray("emails",
			mcp.Description("Email addresses to add as external testers. Required for external tester groups (at least one, at most 1000 entries, each at most 255 characters); ignored for internal tester groups."),
			mcp.WithStringItems(mcp.MaxLength(255)),
			mcp.MaxItems(1000),
		),
		mcp.WithString("type",
			mcp.Description("The type of the tester group. Available values are 'internal' (Bitrise project team members) and 'external' (testers added by email). Defaults to 'internal'."),
			mcp.Enum("internal", "external"),
		),
		mcp.WithTitleAnnotation("Create Tester Group"),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connectedAppID, err := request.RequireString("connected_app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		name, err := request.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{
			"app_id": connectedAppID,
			"name":   name,
		}
		if v := request.GetBool("auto_notify", false); v {
			body["auto_notify"] = v
		}
		if v := request.GetStringSlice("user_slugs", nil); len(v) > 0 {
			body["user_slugs"] = v
		}
		if v := request.GetStringSlice("emails", nil); len(v) > 0 {
			body["emails"] = v
		}

		params := map[string]any{}
		if v := request.GetString("type", ""); v != "" {
			params["type"] = v
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APIRMBuildDistributionsBaseURL,
			Path:    "/tester-groups",
			Params:  params,
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
