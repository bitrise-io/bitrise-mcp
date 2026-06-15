package registration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

// newRequest builds a CallToolRequest with the given arguments, mirroring how
// the MCP server passes decoded JSON arguments to a handler.
func newRequest(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: args}}
}

// resultText extracts the single text content from a tool result.
func resultText(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if res == nil {
		t.Fatal("result is nil")
	}
	if len(res.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(res.Content))
	}
	tc, ok := mcp.AsTextContent(res.Content[0])
	if !ok {
		t.Fatalf("content is not text: %#v", res.Content[0])
	}
	return tc.Text
}

// decodeJSON unmarshals a JSON string into a map, failing the test on error.
func decodeJSON(t *testing.T, s string) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		t.Fatalf("decode JSON %q: %v", s, err)
	}
	return out
}

// withTestServer spins up an httptest server and points the registration tools'
// API base URL at it for the duration of the test. The given handler captures
// and responds to the request the tool makes.
func withTestServer(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	prev := bitrise.APIBaseURL
	bitrise.APIBaseURL = srv.URL
	t.Cleanup(func() { bitrise.APIBaseURL = prev })
}

func TestRegisterHandler(t *testing.T) {
	t.Run("missing email returns an error result before any request", func(t *testing.T) {
		called := false
		withTestServer(t, func(http.ResponseWriter, *http.Request) { called = true })

		res, err := Register.Handler(context.Background(), newRequest(map[string]any{}))

		assert.NoError(t, err)
		assert.True(t, res.IsError)
		assert.Contains(t, resultText(t, res), `required argument "email"`)
		assert.False(t, called, "no HTTP call should be made when email is missing")
	})

	t.Run("happy path forwards the email and returns the API body", func(t *testing.T) {
		var gotMethod, gotPath string
		var gotBody map[string]any
		withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotPath = r.URL.Path
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &gotBody)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"pending_signup_id":"ps_123","expires_at":"2026-06-12T10:00:00Z"}`))
		})

		res, err := Register.Handler(context.Background(), newRequest(map[string]any{"email": "dev@bitrise.io"}))

		assert.NoError(t, err)
		assert.False(t, res.IsError)
		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/agent-signup/start", gotPath)
		assert.Equal(t, map[string]any{"email": "dev@bitrise.io"}, gotBody)
		assert.JSONEq(t, `{"pending_signup_id":"ps_123","expires_at":"2026-06-12T10:00:00Z"}`, resultText(t, res))
	})

	t.Run("does not send an Authorization header", func(t *testing.T) {
		var hadAuth bool
		withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			_, hadAuth = r.Header["Authorization"]
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		})

		_, err := Register.Handler(context.Background(), newRequest(map[string]any{"email": "dev@bitrise.io"}))

		assert.NoError(t, err)
		assert.False(t, hadAuth, "agent signup must be unauthenticated")
	})

	t.Run("feature flag disabled surfaces the 404 with status", func(t *testing.T) {
		withTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error_msg":"Not Found"}`))
		})

		res, err := Register.Handler(context.Background(), newRequest(map[string]any{"email": "dev@bitrise.io"}))

		assert.NoError(t, err)
		assert.True(t, res.IsError)

		payload := decodeJSON(t, resultText(t, res))
		assert.Equal(t, float64(http.StatusNotFound), payload["status"])
		assert.Equal(t, "Not Found", payload["error_msg"])
	})

	t.Run("validation error surfaces status and error_msg", func(t *testing.T) {
		withTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_, _ = w.Write([]byte(`{"error_msg":"Invalid email format"}`))
		})

		res, err := Register.Handler(context.Background(), newRequest(map[string]any{"email": "nope"}))

		assert.NoError(t, err)
		assert.True(t, res.IsError)

		payload := decodeJSON(t, resultText(t, res))
		assert.Equal(t, float64(http.StatusUnprocessableEntity), payload["status"])
		assert.Equal(t, "Invalid email format", payload["error_msg"])
	})
}

func TestVerifyRegistrationHandler(t *testing.T) {
	t.Run("missing pending_signup_id returns an error result", func(t *testing.T) {
		res, err := VerifyRegistration.Handler(context.Background(), newRequest(map[string]any{"otp": "123456"}))

		assert.NoError(t, err)
		assert.True(t, res.IsError)
		assert.Contains(t, resultText(t, res), `required argument "pending_signup_id"`)
	})

	t.Run("missing otp returns an error result", func(t *testing.T) {
		res, err := VerifyRegistration.Handler(context.Background(), newRequest(map[string]any{"pending_signup_id": "ps_123"}))

		assert.NoError(t, err)
		assert.True(t, res.IsError)
		assert.Contains(t, resultText(t, res), `required argument "otp"`)
	})

	t.Run("happy path forwards both fields and returns the token body", func(t *testing.T) {
		var gotPath string
		var gotBody map[string]any
		body := `{"user_slug":"u_1","api_token":"pat_abc","token_expires_at":"2026-06-13T10:00:00Z","workspace_slug":"ws_1"}`
		withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &gotBody)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(body))
		})

		res, err := VerifyRegistration.Handler(context.Background(), newRequest(map[string]any{
			"pending_signup_id": "ps_123",
			"otp":               "123456",
		}))

		assert.NoError(t, err)
		assert.False(t, res.IsError)
		assert.Equal(t, "/agent-signup/confirm", gotPath)
		assert.Equal(t, map[string]any{"pending_signup_id": "ps_123", "otp": "123456"}, gotBody)
		assert.JSONEq(t, body, resultText(t, res))
	})

	t.Run("invalid otp surfaces status and code", func(t *testing.T) {
		withTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error_msg":"Invalid verification code","code":"invalid_otp"}`))
		})

		res, err := VerifyRegistration.Handler(context.Background(), newRequest(map[string]any{
			"pending_signup_id": "ps_123",
			"otp":               "000000",
		}))

		assert.NoError(t, err)
		assert.True(t, res.IsError)

		payload := decodeJSON(t, resultText(t, res))
		assert.Equal(t, float64(http.StatusUnauthorized), payload["status"])
		assert.Equal(t, "invalid_otp", payload["code"])
	})
}

func TestApiErrorResult(t *testing.T) {
	t.Run("non-API error falls back to the wrapped message", func(t *testing.T) {
		res := apiErrorResult(errStub{})

		assert.True(t, res.IsError)
		assert.Contains(t, resultText(t, res), "call api")
		assert.Contains(t, resultText(t, res), "boom")
	})

	t.Run("API error with JSON body merges fields alongside status", func(t *testing.T) {
		res := apiErrorResult(&bitrise.APIError{
			StatusCode: http.StatusConflict,
			Body:       `{"error_msg":"Email already registered","code":"email_already_registered"}`,
		})

		assert.True(t, res.IsError)
		payload := decodeJSON(t, resultText(t, res))
		assert.Equal(t, float64(http.StatusConflict), payload["status"])
		assert.Equal(t, "email_already_registered", payload["code"])
		assert.Equal(t, "Email already registered", payload["error_msg"])
	})

	t.Run("API error with non-JSON body is preserved under body", func(t *testing.T) {
		res := apiErrorResult(&bitrise.APIError{
			StatusCode: http.StatusBadGateway,
			Body:       "upstream exploded",
		})

		assert.True(t, res.IsError)
		payload := decodeJSON(t, resultText(t, res))
		assert.Equal(t, float64(http.StatusBadGateway), payload["status"])
		assert.Equal(t, "upstream exploded", payload["body"])
	})
}

type errStub struct{}

func (errStub) Error() string { return "boom" }
