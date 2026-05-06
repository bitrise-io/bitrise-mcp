package bitrise

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

// UploadFileToPresignedURL uploads file content to an AWS S3 presigned URL.
// Unlike CallAPI, this does not add Bitrise auth headers — presigned URLs are self-authenticating.
func UploadFileToPresignedURL(uploadURL string, content []byte) error {
	req, err := http.NewRequest(http.MethodPut, uploadURL, bytes.NewReader(content))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.ContentLength = int64(len(content))

	client := &http.Client{Timeout: 60 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		resBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("unexpected status code %d; response body: %s", res.StatusCode, resBody)
	}
	return nil
}
