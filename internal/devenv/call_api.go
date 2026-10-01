package devenv

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	httptrace "github.com/DataDog/dd-trace-go/contrib/net/http/v2"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
)

// BaseURL is set at init from config.
var BaseURL string

// WsPath prepends the workspace-scoped /v1 base to a resource path, using the
// workspace resolved into the request context (see ContextWithWorkspace and the
// workspace-resolution middleware). "/sessions" → "/v1/workspaces/{id}/sessions"
func WsPath(ctx context.Context, path string) string {
	return "/v1/workspaces/" + url.PathEscape(WorkspaceFromCtx(ctx)) + path
}

// CallAPIParams configures an API call.
type CallAPIParams struct {
	Method string
	Path   string
	Params map[string]string
	// RepeatedParams holds query parameters that appear multiple times in the
	// query string (e.g. label_selectors=a&label_selectors=b).
	RepeatedParams map[string][]string
	Body           any
}

// The clients are wrapped for Datadog APM like the main API client, so Dev
// Environments calls show up as outbound spans when tracing is enabled.
var (
	httpClient     = httptrace.WrapClient(&http.Client{Timeout: 30 * time.Second})
	longHTTPClient = httptrace.WrapClient(&http.Client{Timeout: 10 * time.Minute})
)

// CallAPI makes an authenticated HTTP request to the Dev Environments backend.
func CallAPI(ctx context.Context, p CallAPIParams) (string, error) {
	return callAPI(ctx, httpClient, p)
}

// CallAPILongTimeout makes an API call with a longer timeout for operations like file transfers.
func CallAPILongTimeout(ctx context.Context, p CallAPIParams) (string, error) {
	return callAPI(ctx, longHTTPClient, p)
}

func callAPI(ctx context.Context, client *http.Client, p CallAPIParams) (string, error) {
	authHeader, authValue, err := AuthFromCtx(ctx)
	if err != nil {
		return "", err
	}

	fullURL := BaseURL + p.Path
	if len(p.Params) > 0 || len(p.RepeatedParams) > 0 {
		params := url.Values{}
		for k, v := range p.Params {
			if v != "" {
				params.Set(k, v)
			}
		}
		for k, vs := range p.RepeatedParams {
			for _, v := range vs {
				if v != "" {
					params.Add(k, v)
				}
			}
		}
		if encoded := params.Encode(); encoded != "" {
			fullURL += "?" + encoded
		}
	}

	var reqBody io.Reader
	if p.Body != nil {
		bodyBytes, err := json.Marshal(p.Body)
		if err != nil {
			return "", fmt.Errorf("marshal body: %w", err)
		}
		reqBody = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, p.Method, fullURL, reqBody)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set(authHeader, authValue)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", bitrise.UserAgent)
	req.Header.Set("X-Request-Source", "mcp")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return string(body), nil
}
