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

var UploadAndroidKeystoreFile = bitrise.Tool{
	APIGroups: []string{"apps"},
	Definition: mcp.NewTool("upload_android_keystore_file",
		mcp.WithDescription("Upload an Android keystore file to a Bitrise app for code signing. Reads the file from a local path, uploads it to Bitrise, and confirms the upload in a single operation."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("file_path",
			mcp.Description("Local filesystem path to the keystore file (e.g. /Users/me/release.jks)"),
			mcp.Required(),
		),
		mcp.WithString("alias",
			mcp.Description("Keystore alias"),
			mcp.Required(),
		),
		mcp.WithString("keystore_password",
			mcp.Description("Password for the keystore"),
			mcp.Required(),
		),
		mcp.WithString("private_key_password",
			mcp.Description("Password for the private key within the keystore"),
			mcp.Required(),
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
		alias, err := request.RequireString("alias")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		keystorePassword, err := request.RequireString("keystore_password")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		privateKeyPassword, err := request.RequireString("private_key_password")
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
			Path:    fmt.Sprintf("/apps/%s/android-keystore-files", appSlug),
			Body: map[string]any{
				"upload_file_name":     fileName,
				"upload_file_size":     len(content),
				"alias":                alias,
				"password":             keystorePassword,
				"private_key_password": privateKeyPassword,
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
			Path:    fmt.Sprintf("/apps/%s/android-keystore-files/%s/uploaded", appSlug, parsed.Data.Slug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("confirm upload", err), nil
		}
		return mcp.NewToolResultText(confirmRes), nil
	},
}
