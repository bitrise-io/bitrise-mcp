package releasemanagement

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var GetInstallableArtifactUploadAndProcessingStatus = bitrise.Tool{
	APIGroups: []string{"release-management", "read-only"},
	Definition: mcp.NewTool("get_installable_artifact_upload_and_proc_status",
		mcp.WithDescription("Gets the processing and upload status of an installable artifact. An artifact will need to be processed after upload to be usable. This endpoint helps understanding when an uploaded installable artifacts becomes usable for later purposes."),
		mcp.WithString("connected_app_id",
			mcp.Description("Identifier of the Release Management connected app for the installable artifact. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithString("installable_artifact_id",
			mcp.Description("The uuidv4 identifier for the installable artifact. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connectedAppID, err := request.RequireString("connected_app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		installableArtifactID, err := request.RequireString("installable_artifact_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIRMBaseURL,
			Path:    fmt.Sprintf("/connected-apps/%s/installable-artifacts/%s/status", connectedAppID, installableArtifactID),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
