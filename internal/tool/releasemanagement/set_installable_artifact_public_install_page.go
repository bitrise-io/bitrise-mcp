package releasemanagement

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var SetInstallableArtifactPublicInstallPage = bitrise.Tool{
	APIGroups: []string{"release-management"},
	Definition: mcp.NewTool("set_installable_artifact_public_install_page",
		mcp.WithDescription("Enable or disable the Release Management public install page of an installable artifact (an over-the-air install link for real devices). Not a browser simulator preview: see bitrise_devenv_create_preview_link."),
		mcp.WithString("connected_app_id",
			mcp.Description("Identifier of the Release Management connected app for the installable artifact. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithString("installable_artifact_id",
			mcp.Description("The uuidv4 identifier for the installable artifact. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithBoolean("with_public_page",
			mcp.Description("Boolean flag for enabling/disabling public install page for the installable artifact. This field is mandatory."),
			mcp.Required(),
		),
		mcp.WithTitleAnnotation("Set Installable Artifact Public Install Page"),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if _, err := request.RequireString("connected_app_id"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		installableArtifactID, err := request.RequireString("installable_artifact_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		withPublicPage, err := request.RequireBool("with_public_page")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{
			"with_public_page": withPublicPage,
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPatch,
			BaseURL: bitrise.APIRMAppsBaseURL,
			Path:    fmt.Sprintf("/installable-artifacts/%s/public-install-page", installableArtifactID),
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
