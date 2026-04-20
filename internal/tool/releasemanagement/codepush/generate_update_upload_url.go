package codepush

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var GenerateUpdateUploadURL = bitrise.Tool{
	APIGroups: []string{"release-management-code-push", "release-management"},
	Definition: mcp.NewTool("codepush_generate_update_upload_url",
		mcp.WithDescription("Generate a signed upload URL (valid 1 hour) for uploading a CodePush update bundle. The response contains the URL, HTTP method, and headers needed for a direct upload. After uploading, check status with codepush_get_update_status."),
		mcp.WithString("id",
			mcp.Description("Client-generated UUID for the new update"),
			mcp.Required(),
		),
		mcp.WithString("deployment_id",
			mcp.Description("Identifier (UUID) of the deployment this update belongs to"),
			mcp.Required(),
		),
		mcp.WithString("app_version",
			mcp.Description("Semver version of the app this update targets (e.g. '1.2.3')"),
			mcp.Required(),
		),
		mcp.WithString("file_name",
			mcp.Description("File name of the update bundle to be uploaded (with extension)"),
			mcp.Required(),
		),
		mcp.WithString("file_size_bytes",
			mcp.Description("Byte size of the update bundle file as a string"),
			mcp.Required(),
		),
		mcp.WithString("description",
			mcp.Description("Optional description for this update."),
		),
		mcp.WithBoolean("disabled",
			mcp.Description("If true, clients will not download this update after upload."),
			mcp.DefaultBool(false),
		),
		mcp.WithBoolean("mandatory",
			mcp.Description("If true, clients must install this update immediately."),
			mcp.DefaultBool(false),
		),
		mcp.WithNumber("rollout",
			mcp.Description("Percentage (0-100) of users who will receive this update. Defaults to 100."),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		deploymentID, err := request.RequireString("deployment_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		appVersion, err := request.RequireString("app_version")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		fileName, err := request.RequireString("file_name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		fileSizeBytes, err := request.RequireString("file_size_bytes")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		params := map[string]any{
			"deployment_id":   deploymentID,
			"app_version":     appVersion,
			"file_name":       fileName,
			"file_size_bytes": fileSizeBytes,
		}
		if v := request.GetString("description", ""); v != "" {
			params["description"] = v
		}
		if v := request.GetBool("disabled", false); v {
			params["disabled"] = "true"
		}
		if v := request.GetBool("mandatory", false); v {
			params["mandatory"] = "true"
		}
		// rollout=0 is a valid value (0% rollout); use -1 as sentinel for "not provided"
		if v := request.GetInt("rollout", -1); v >= 0 {
			params["rollout"] = strconv.Itoa(v)
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APICodePushBaseURL,
			Path:    fmt.Sprintf("/updates/%s/upload-url", id),
			Params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
