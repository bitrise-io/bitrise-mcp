package codesigning

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var UpdateBuildCertificate = bitrise.Tool{
	APIGroups: []string{"apps"},
	Definition: mcp.NewTool("update_build_certificate",
		mcp.WithDescription("Update metadata for an iOS build certificate. Note: once is_protected is set to true it cannot be changed back."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("certificate_slug",
			mcp.Description("Identifier of the build certificate"),
			mcp.Required(),
		),
		mcp.WithString("certificate_password",
			mcp.Description("Password for the .p12 certificate file"),
		),
		mcp.WithBoolean("is_protected",
			mcp.Description("Mark the certificate as protected (irreversible once set to true)"),
		),
		mcp.WithBoolean("is_expose",
			mcp.Description("Whether to expose the certificate to pull request builds"),
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
		certificateSlug, err := request.RequireString("certificate_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{}
		args := request.GetArguments()
		if _, ok := args["certificate_password"]; ok {
			body["certificate_password"] = request.GetString("certificate_password", "")
		}
		if _, ok := args["is_protected"]; ok {
			body["is_protected"] = request.GetBool("is_protected", false)
		}
		if _, ok := args["is_expose"]; ok {
			body["is_expose"] = request.GetBool("is_expose", false)
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPatch,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/build-certificates/%s", appSlug, certificateSlug),
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
