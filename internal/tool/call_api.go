package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	apiBaseURL   = "https://api.bitrise.io/v0.1"
	apiRMBaseURL = "https://api.bitrise.io/release-management/v1"
	userAgent    = "bitrise-mcp/1.0"
)

type callAPIParams struct {
	method  string
	baseURL string
	path    string
	params  map[string]string
	body    any
}

func callAPI(ctx context.Context, p callAPIParams) (string, error) {
	apiKey, err := patFromCtx(ctx)
	if err != nil {
		return "", errors.New("set authorization header to your bitrise pat")
	}

	var reqBody io.Reader
	if p.body != nil {
		a, err := json.Marshal(p.body)
		if err != nil {
			return "", fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(a)
	}

	fullURL := p.baseURL
	if !strings.HasPrefix(p.path, "/") {
		fullURL += "/"
	}
	fullURL += p.path

	req, err := http.NewRequest(p.method, fullURL, reqBody)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	if p.params != nil {
		q := req.URL.Query()
		for key, value := range p.params {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", apiKey)

	client := http.Client{Timeout: 30 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute request: %w", err)
	}
	if res.StatusCode >= 400 {
		return "", fmt.Errorf("unexpected status code %d", res.StatusCode)
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}
	return string(resBody), nil
}
