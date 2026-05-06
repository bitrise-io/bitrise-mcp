package codesigning

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var ListBuildCertificates = bitrise.Tool{
	APIGroups: []string{"apps", "read-only"},
	Definition: mcp.NewTool("list_build_certificates",
		mcp.WithDescription("List all iOS build certificates (.p12 files) for a Bitrise app."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("next",
			mcp.Description("Slug of the first certificate to return (for pagination)"),
		),
		mcp.WithString("limit",
			mcp.Description("Maximum number of results to return (default: 50)"),
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

		params := map[string]any{}
		if v := request.GetString("next", ""); v != "" {
			params["next"] = v
		}
		if v := request.GetString("limit", ""); v != "" {
			params["limit"] = v
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/build-certificates", appSlug),
			Params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
