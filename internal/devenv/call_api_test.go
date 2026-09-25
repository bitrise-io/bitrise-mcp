package devenv

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCallAPIQueryParams(t *testing.T) {
	// callWith performs a CallAPI against a stub server and returns the raw
	// query string the server received.
	callWith := func(t *testing.T, p CallAPIParams) string {
		t.Helper()
		var query string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			query = r.URL.RawQuery
			_, _ = w.Write([]byte(`{}`))
		}))
		defer srv.Close()

		old := BaseURL
		BaseURL = srv.URL
		defer func() { BaseURL = old }()

		p.Method = http.MethodGet
		p.Path = "/v1/x"
		_, err := CallAPI(ContextWithPAT(context.Background(), "pat"), p)
		assert.NoError(t, err)
		return query
	}

	t.Run("no params yields no query string", func(t *testing.T) {
		assert.Equal(t, "", callWith(t, CallAPIParams{}))
	})

	t.Run("single-value params", func(t *testing.T) {
		q := callWith(t, CallAPIParams{Params: map[string]string{"include_secrets": "true", "skipped": ""}})
		assert.Equal(t, "include_secrets=true", q)
	})

	t.Run("repeated params appear once per value", func(t *testing.T) {
		q := callWith(t, CallAPIParams{RepeatedParams: map[string][]string{
			"label_selectors": {"team=mobile", "branch=main", ""},
		}})
		assert.Equal(t, "label_selectors=team%3Dmobile&label_selectors=branch%3Dmain", q)
	})

	t.Run("single and repeated params combine", func(t *testing.T) {
		q := callWith(t, CallAPIParams{
			Params:         map[string]string{"a": "1"},
			RepeatedParams: map[string][]string{"b": {"2", "3"}},
		})
		assert.Equal(t, "a=1&b=2&b=3", q)
	})
}
