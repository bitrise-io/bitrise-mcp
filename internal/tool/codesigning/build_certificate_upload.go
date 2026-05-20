package codesigning

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var UploadBuildCertificate = bitrise.Tool{
	APIGroups: []string{"apps"},
	Definition: mcp.NewTool("upload_build_certificate",
		mcp.WithDescription("Upload an iOS build certificate (.p12 file) to a Bitrise app for code signing. Reads the file from a local path, uploads it to Bitrise, and sets the certificate password in a single operation."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("file_path",
			mcp.Description("Local filesystem path to the .p12 certificate file (e.g. /Users/me/Certificates.p12)"),
			mcp.Required(),
		),
		mcp.WithString("certificate_password",
			mcp.Description("Password for the .p12 certificate file (leave empty if the certificate has no password)"),
		),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		filePath, err := request.RequireString("file_path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("read file: %s", err)), nil
		}
		fileName := filepath.Base(filePath)

		createRes, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/build-certificates", appSlug),
			Body: map[string]any{
				"upload_file_name": fileName,
				"upload_file_size": len(content),
			},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("create upload entry", err), nil
		}

		var parsed struct {
			Data struct {
				UploadURL string `json:"upload_url"`
				Slug      string `json:"slug"`
			} `json:"data"`
		}
		if err := json.Unmarshal([]byte(createRes), &parsed); err != nil {
			return mcp.NewToolResultErrorFromErr("parse create response", err), nil
		}
		if parsed.Data.UploadURL == "" {
			return mcp.NewToolResultError("API did not return an upload URL"), nil
		}
		if parsed.Data.Slug == "" {
			return mcp.NewToolResultError("API did not return a file slug"), nil
		}

		if err := bitrise.UploadFileToPresignedURL(parsed.Data.UploadURL, content); err != nil {
			return mcp.NewToolResultErrorFromErr("upload file", err), nil
		}

		confirmRes, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/build-certificates/%s/uploaded", appSlug, parsed.Data.Slug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("confirm upload", err), nil
		}

		if password := request.GetString("certificate_password", ""); password != "" {
			patchRes, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
				Method:  http.MethodPatch,
				BaseURL: bitrise.APIBaseURL,
				Path:    fmt.Sprintf("/apps/%s/build-certificates/%s", appSlug, parsed.Data.Slug),
				Body:    map[string]any{"certificate_password": password},
			})
			if err != nil {
				return mcp.NewToolResultErrorFromErr("set certificate password", err), nil
			}
			return mcp.NewToolResultText(patchRes), nil
		}

		return mcp.NewToolResultText(confirmRes), nil
	},
}
