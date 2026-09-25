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
	"strings"
	"time"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/devenv"
	"github.com/mark3labs/mcp-go/mcp"
)

type downloadResp struct {
	SignedURL string `json:"signedUrl"`
}

// Download downloads files from a session to the local machine.
var Download = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_download",
		mcp.WithTitleAnnotation("Download files from session"),
		mcp.WithDescription(`Download a file or directory from a running Dev Environments session's VM to this machine (locally-run server only). Not for Bitrise CI build artifacts (get_artifact returns their download URL) or cache items (get_cache_item_download_url).

The remote path is archived as tar.gz, uploaded to cloud storage, then downloaded and extracted locally.

If the server answers "File download is not available…", this deployment has no file store behind the
tool — do not retry. Copy the file with scp using the session's sshAddress / sshPassword from
bitrise_devenv_get instead (the device guide, bitrise_devenv_device_guide, has the password-feeding
recipe); do not base64 binaries through bitrise_devenv_execute output.

Example: Download build output from the VM:
  session_id: <uuid>
  source_path: /Users/vagrant/project/build/output
  local_destination: /Users/me/downloads/output`),
		mcp.WithString("session_id", mcp.Description("The unique identifier of the running session"), mcp.Required()),
		mcp.WithString("source_path", mcp.Description("Absolute path on the remote machine to download"), mcp.Required()),
		mcp.WithString("local_destination", mcp.Description("Local directory path where files will be extracted"), mcp.Required()),
		mcp.WithBoolean("only_contents", mcp.Description("If true and source is a directory, extract only its contents (not the directory itself)")),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
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
		localDest, err := request.RequireString("local_destination")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{
			"source_path": sourcePath,
		}
		if oc, ok := request.GetArguments()["only_contents"]; ok {
			body["only_contents_of_folder"] = oc
		}

		// Step 1: Request download archive from the VM
		res, err := devenv.CallAPILongTimeout(ctx, devenv.CallAPIParams{
			Method: http.MethodPost,
			Path:   devenv.WsPath(ctx, fmt.Sprintf("/sessions/%s/download-file", sessionID)),
			Body:   body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("request download", err), nil
		}

		var resp downloadResp
		if err := json.Unmarshal([]byte(res), &resp); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("parse download response: %v", err)), nil
		}

		// Step 2: Download the tar.gz from GCS
		archiveData, err := downloadArchive(ctx, resp.SignedURL)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("download archive: %v", err)), nil
		}

		// Step 3: Extract locally
		if err := extractTarGz(archiveData, localDest); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("extract archive: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Download complete. Files extracted to %s", localDest)), nil
	},
}

func downloadArchive(ctx context.Context, signedURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, signedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	client := http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download failed (status %d): %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxArchiveBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read archive: %w", err)
	}
	if int64(len(data)) > maxArchiveBytes {
		return nil, fmt.Errorf("archive larger than %d bytes; download a smaller folder", maxArchiveBytes)
	}
	return data, nil
}

// The archive comes from the session VM, which runs whatever the user's
// repository and scripts run there, so it is not trusted: entries must stay
// under the destination, may not carry setuid/setgid bits, and the archive
// and the extracted bytes are capped.
const (
	maxArchiveBytes   int64 = 2 << 30
	maxExtractedBytes int64 = 10 << 30
)

// withinDir reports whether path, with every symlink in it resolved, lies
// under realDir (itself already symlink-free). It guards against a
// pre-existing symlinked parent directory redirecting a write outside the
// destination, which the lexical check in safeExtractTarget cannot see.
func withinDir(realDir, path string) error {
	real, err := filepath.EvalSymlinks(path)
	if err != nil {
		return fmt.Errorf("resolve %q: %w", path, err)
	}
	if real != realDir && !strings.HasPrefix(real, realDir+string(filepath.Separator)) {
		return fmt.Errorf("%q resolves outside the destination folder", path)
	}
	return nil
}

// safeExtractTarget returns the extraction path of an archive entry, or an
// error when the entry would land outside destDir.
func safeExtractTarget(destDir, name string) (string, error) {
	if filepath.IsAbs(name) {
		return "", fmt.Errorf("archive entry %q has an absolute path", name)
	}
	target := filepath.Join(destDir, name)
	rel, err := filepath.Rel(destDir, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("archive entry %q escapes the destination folder", name)
	}
	return target, nil
}

func extractTarGz(data []byte, destDir string) error {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("create destination: %w", err)
	}
	destDir, err := filepath.Abs(destDir)
	if err != nil {
		return fmt.Errorf("resolve destination: %w", err)
	}
	realDest, err := filepath.EvalSymlinks(destDir)
	if err != nil {
		return fmt.Errorf("resolve destination: %w", err)
	}
	var extracted int64

	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("open gzip: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar: %w", err)
		}

		target, err := safeExtractTarget(destDir, header.Name)
		if err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("create dir: %w", err)
			}
			if err := withinDir(realDest, target); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return fmt.Errorf("create parent dir: %w", err)
			}
			if err := withinDir(realDest, filepath.Dir(target)); err != nil {
				return err
			}
			if fi, err := os.Lstat(target); err == nil && fi.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("refusing to write through symlink %q", target)
			}
			f, err := os.Create(target)
			if err != nil {
				return fmt.Errorf("create file: %w", err)
			}
			n, err := io.Copy(f, io.LimitReader(tr, maxExtractedBytes-extracted+1))
			if err != nil {
				_ = f.Close()
				return fmt.Errorf("write file: %w", err)
			}
			extracted += n
			if extracted > maxExtractedBytes {
				_ = f.Close()
				return fmt.Errorf("archive expands beyond %d bytes; download a smaller folder", maxExtractedBytes)
			}
			if err := f.Close(); err != nil {
				return fmt.Errorf("close file: %w", err)
			}
			if err := os.Chmod(target, os.FileMode(header.Mode)&0o777); err != nil {
				return fmt.Errorf("chmod: %w", err)
			}
		}
	}
	return nil
}
