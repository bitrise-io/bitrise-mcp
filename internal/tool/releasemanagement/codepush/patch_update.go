package codepush

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var PatchUpdate = bitrise.Tool{
	APIGroups: []string{"release-management-code-push", "release-management"},
	Definition: mcp.NewTool("codepush_patch_update",
		mcp.WithDescription("Patch a CodePush update to change its disabled state, mandatory flag, or rollout percentage. Only include fields you want to change — omitted fields are left unchanged."),
		mcp.WithString("id",
			mcp.Description("Identifier (UUID) of the CodePush update"),
			mcp.Required(),
		),
		mcp.WithString("disabled",
			mcp.Description("Set to 'true' to disable (clients won't download) or 'false' to re-enable. Omit to leave unchanged."),
		),
		mcp.WithString("mandatory",
			mcp.Description("Set to 'true' to make mandatory (clients must install immediately) or 'false' to make optional. Omit to leave unchanged."),
		),
		mcp.WithNumber("rollout",
			mcp.Description("Percentage (0-100) of users who will receive this update. Omit to leave unchanged."),
		),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{}
		// Use string params for booleans so callers can explicitly set false.
		// GetBool can't distinguish "not provided" from false, so we use strings.
		if v := request.GetString("disabled", ""); v != "" {
			body["disabled"] = v == "true"
		}
		if v := request.GetString("mandatory", ""); v != "" {
			body["mandatory"] = v == "true"
		}
		// rollout=0 is a valid value (0% rollout), use -1 as sentinel
		if rollout := request.GetInt("rollout", -1); rollout >= 0 {
			body["rollout"] = rollout
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPatch,
			BaseURL: bitrise.APICodePushBaseURL,
			Path:    fmt.Sprintf("/updates/%s", id),
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
