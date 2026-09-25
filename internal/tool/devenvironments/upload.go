package devenvironments

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/devenv"
	"github.com/mark3labs/mcp-go/mcp"
)

type startUploadResp struct {
	SignedURL string `json:"signedUrl"`
	UploadID  string `json:"uploadId"`
}

// Upload uploads a local file or directory to a session.
var Upload = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_upload",
		mcp.WithTitleAnnotation("Upload files to session"),
		mcp.WithDescription(`Upload a local file or directory from this machine into a running Dev Environments session (a remote VM). Not for app builds: to upload an installable app build to Release Management use generate_installable_artifact_upload_url.

The local path is compressed into a tar.gz archive, uploaded to cloud storage via a signed URL,
then extracted on the remote machine INTO destination_folder, which must be a directory
(created if missing). A directory source lands as its contents inside destination_folder;
a single file lands as destination_folder/<basename>. To replace one remote file, upload
it into its parent directory — a destination that is an existing file is rejected.
Extracted files are owned by the session user.

Example: Upload a local project directory to the VM:
  source_path: /Users/me/project
  destination_folder: /Users/vagrant/project`),
		mcp.WithString("session_id", mcp.Description("The unique identifier of the running session"), mcp.Required()),
		mcp.WithString("source_path", mcp.Description("Local file or directory path to upload"), mcp.Required()),
		mcp.WithString("destination_folder", mcp.Description("Absolute DIRECTORY path on the remote machine to extract into (created if missing; must not be an existing file)"), mcp.Required()),
		mcp.WithDestructiveHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sessionID, err := requireUUID(request, "session_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		sourcePath, err := request.RequireString("source_path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		destFolder, err := request.RequireString("destination_folder")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Step 1: Start upload to get signed PUT URL
		startRes, err := devenv.CallAPI(ctx, devenv.CallAPIParams{
			Method: http.MethodPost,
			Path:   devenv.WsPath(ctx, fmt.Sprintf("/sessions/%s/start-file-upload", sessionID)),
			Body:   map[string]any{"destination_folder": destFolder},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("start upload", err), nil
		}

		var startResp startUploadResp
		if err := json.Unmarshal([]byte(startRes), &startResp); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("parse start upload response: %v", err)), nil
		}

		// Step 2: Create tar.gz archive from local path
		archive, err := createTarGz(sourcePath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("create archive: %v", err)), nil
		}

		// Step 3: Upload to GCS via signed PUT URL
		if err := uploadToGCS(ctx, startResp.SignedURL, archive); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("upload to cloud storage: %v", err)), nil
		}

		// Step 4: Complete upload to trigger extraction on the VM
		completeRes, err := devenv.CallAPILongTimeout(ctx, devenv.CallAPIParams{
			Method: http.MethodPost,
			Path:   devenv.WsPath(ctx, fmt.Sprintf("/sessions/%s/complete-file-upload", sessionID)),
			Body: map[string]any{
				"upload_id":          startResp.UploadID,
				"destination_folder": destFolder,
			},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("complete upload", err), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Upload complete. Files extracted to %s on the remote machine.\n%s", destFolder, completeRes)), nil
	},
}

func createTarGz(sourcePath string) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	info, err := os.Stat(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("stat source: %w", err)
	}

	baseDir := filepath.Dir(sourcePath)
	if info.IsDir() {
		baseDir = sourcePath
	}

	walkFn := func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("file info header: %w", err)
		}
		clearTarOwner(header)

		relPath, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("write header: %w", err)
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open file: %w", err)
		}
		defer f.Close()

		if _, err := io.Copy(tw, f); err != nil {
			return fmt.Errorf("copy file: %w", err)
		}
		return nil
	}

	if info.IsDir() {
		if err := filepath.WalkDir(sourcePath, walkFn); err != nil {
			return nil, fmt.Errorf("walk directory: %w", err)
		}
	} else {
		// Single file
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return nil, fmt.Errorf("file info header: %w", err)
		}
		clearTarOwner(header)
		header.Name = filepath.Base(sourcePath)
		if err := tw.WriteHeader(header); err != nil {
			return nil, fmt.Errorf("write header: %w", err)
		}
		f, err := os.Open(sourcePath)
		if err != nil {
			return nil, fmt.Errorf("open file: %w", err)
		}
		defer f.Close()
		if _, err := io.Copy(tw, f); err != nil {
			return nil, fmt.Errorf("copy file: %w", err)
		}
	}

	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("close tar: %w", err)
	}
	if err := gw.Close(); err != nil {
		return nil, fmt.Errorf("close gzip: %w", err)
	}

	return buf.Bytes(), nil
}

func uploadToGCS(ctx context.Context, signedURL string, data []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, signedURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/gzip")

	client := http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed (status %d): %s", resp.StatusCode, string(body))
	}
	return nil
}

// clearTarOwner drops the local owner from an archive entry. The archive is
// extracted on the VM; the caller's numeric uid/gid means nothing there and,
// when honoured, leaves files the session user cannot write.
func clearTarOwner(h *tar.Header) {
	h.Uid, h.Gid = 0, 0
	h.Uname, h.Gname = "", ""
}
