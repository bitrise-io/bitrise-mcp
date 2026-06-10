package bitrise

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

	httptrace "github.com/DataDog/dd-trace-go/contrib/net/http/v2"
)

// APIBaseURL, APIRMBaseURL, and APICodePushBaseURL are vars so main can override
// them via environment variables to point at non-production API instances.
var (
	APIBaseURL         = "https://api.bitrise.io/v0.1"                          //nolint:gochecknoglobals
	APIRMBaseURL       = "https://api.bitrise.io/release-management/v1"         //nolint:gochecknoglobals
	APICodePushBaseURL = "https://api.bitrise.io/release-management/v2/code-push/v1" //nolint:gochecknoglobals
)

const userAgent = "bitrise-mcp/1.0"

type CallAPIParams struct {
	Method   string
	BaseURL  string
	Path     string
	Params   map[string]any
	Body     any
	SkipAuth bool
}

type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("unexpected status code %d; response body: %s", e.StatusCode, e.Body)
}

func CallAPI(ctx context.Context, p CallAPIParams) (string, error) {
	var apiKey string
	if !p.SkipAuth {
		key, err := patFromCtx(ctx)
		if err != nil || strings.TrimSpace(key) == "" {
			return "", errors.New("missing Bitrise authentication: set BITRISE_TOKEN in stdio mode or send Authorization: Bearer <bitrise_pat> in HTTP mode")
		}
		apiKey = key
	}

	var reqBody io.Reader
	if p.Body != nil {
		a, err := json.Marshal(p.Body)
		if err != nil {
			return "", fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(a)
	}

	fullURL := p.BaseURL
	if !strings.HasPrefix(p.Path, "/") {
		fullURL += "/"
	}
	fullURL += p.Path

	req, err := http.NewRequestWithContext(ctx, p.Method, fullURL, reqBody)

	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	if p.Params != nil {
		q := req.URL.Query()
		for key, value := range p.Params {
			switch v := value.(type) {
			case string:
				q.Add(key, v)
			case []string:
				for _, item := range v {
					q.Add(key, item)
				}
			default:
				q.Add(key, fmt.Sprintf("%v", v))
			}
		}
		req.URL.RawQuery = q.Encode()
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", apiKey)
	}

	httpClient := http.Client{Timeout: 30 * time.Second}
	client := httptrace.WrapClient(&httpClient)
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute request: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		resBody, _ := io.ReadAll(res.Body)
		return "", &APIError{StatusCode: res.StatusCode, Body: string(resBody)}
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}
	return string(resBody), nil
}
