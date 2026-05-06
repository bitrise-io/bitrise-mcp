package codesigning

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var UpdateProvisioningProfile = bitrise.Tool{
	APIGroups: []string{"apps"},
	Definition: mcp.NewTool("update_provisioning_profile",
		mcp.WithDescription("Update metadata for an iOS provisioning profile. Note: once is_protected is set to true it cannot be changed back."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("profile_slug",
			mcp.Description("Identifier of the provisioning profile"),
			mcp.Required(),
		),
		mcp.WithBoolean("is_protected",
			mcp.Description("Mark the profile as protected (irreversible once set to true)"),
		),
		mcp.WithBoolean("is_expose",
			mcp.Description("Whether to expose the profile to pull request builds"),
		),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		profileSlug, err := request.RequireString("profile_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{}
		args := request.GetArguments()
		if _, ok := args["is_protected"]; ok {
			body["is_protected"] = request.GetBool("is_protected", false)
		}
		if _, ok := args["is_expose"]; ok {
			body["is_expose"] = request.GetBool("is_expose", false)
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPatch,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/provisioning-profiles/%s", appSlug, profileSlug),
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
