package codepush

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var PromoteDeployment = bitrise.Tool{
	APIGroups: []string{"release-management-code-push", "release-management"},
	Definition: mcp.NewTool("codepush_promote_deployment",
		mcp.WithDescription("Promote a package from a source deployment to a target deployment. The most recent package in the source deployment is promoted unless package_id is specified."),
		mcp.WithString("id",
			mcp.Description("Identifier (UUID) of the source CodePush deployment"),
			mcp.Required(),
		),
		mcp.WithString("target_deployment_id",
			mcp.Description("Identifier (UUID) of the target deployment to promote the package to"),
			mcp.Required(),
		),
		mcp.WithString("package_id",
			mcp.Description("Optional UUID of a specific package to promote. Defaults to the most recent package."),
		),
		mcp.WithString("app_version",
			mcp.Description("Optional semver app version constraint for the promoted package."),
		),
		mcp.WithString("description",
			mcp.Description("Optional description for the promoted package."),
		),
		mcp.WithBoolean("disabled",
			mcp.Description("If true, the promoted package will not be downloaded by clients."),
			mcp.DefaultBool(false),
		),
		mcp.WithBoolean("mandatory",
			mcp.Description("If true, clients must install this update immediately."),
			mcp.DefaultBool(false),
		),
		mcp.WithNumber("rollout",
			mcp.Description("Percentage (0-100) of users who will receive this update. Defaults to 100."),
		),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		targetDeploymentID, err := request.RequireString("target_deployment_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{
			"target_deployment_id": targetDeploymentID,
		}
		if v := request.GetString("package_id", ""); v != "" {
			body["package_id"] = v
		}
		if v := request.GetString("app_version", ""); v != "" {
			body["app_version"] = v
		}
		if v := request.GetString("description", ""); v != "" {
			body["description"] = v
		}
		if v := request.GetBool("disabled", false); v {
			body["disabled"] = v
		}
		if v := request.GetBool("mandatory", false); v {
			body["mandatory"] = v
		}
		// rollout=0 is a valid value (0% rollout); use -1 as sentinel for "not provided"
		if v := request.GetInt("rollout", -1); v >= 0 {
			body["rollout"] = v
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APICodePushBaseURL,
			Path:    fmt.Sprintf("/deployments/%s/promote", id),
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
