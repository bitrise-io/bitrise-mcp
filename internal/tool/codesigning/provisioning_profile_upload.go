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

var UploadProvisioningProfile = bitrise.Tool{
	APIGroups: []string{"apps"},
	Definition: mcp.NewTool("upload_provisioning_profile",
		mcp.WithDescription("Upload an iOS provisioning profile (.mobileprovision file) to a Bitrise app for code signing. Reads the file from a local path and uploads it to Bitrise in a single operation."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("file_path",
			mcp.Description("Local filesystem path to the provisioning profile file (e.g. /Users/me/MyApp.mobileprovision)"),
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

		content, err := os.ReadFile(filePath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("read file: %s", err)), nil
		}
		fileName := filepath.Base(filePath)

		createRes, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/provisioning-profiles", appSlug),
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
			Path:    fmt.Sprintf("/apps/%s/provisioning-profiles/%s/uploaded", appSlug, parsed.Data.Slug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("confirm upload", err), nil
		}
		return mcp.NewToolResultText(confirmRes), nil
	},
}
