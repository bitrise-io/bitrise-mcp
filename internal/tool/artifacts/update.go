package artifacts

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var Update = bitrise.Tool{
	APIGroups: []string{"artifacts"},
	Definition: mcp.NewTool("update_artifact",
		mcp.WithDescription("Update a build artifact."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("build_slug",
			mcp.Description("Identifier of the build"),
			mcp.Required(),
		),
		mcp.WithString("artifact_slug",
			mcp.Description("Identifier of the artifact"),
			mcp.Required(),
		),
		mcp.WithBoolean("is_public_page_enabled",
			mcp.Description("Enable public page for the artifact"),
			mcp.Required(),
		),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		buildSlug, err := request.RequireString("build_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		artifactSlug, err := request.RequireString("artifact_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		isPublicPageEnabled, err := request.RequireBool("is_public_page_enabled")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPatch,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/builds/%s/artifacts/%s", appSlug, buildSlug, artifactSlug),
			Body: map[string]any{
				"is_public_page_enabled": isPublicPageEnabled,
			},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
