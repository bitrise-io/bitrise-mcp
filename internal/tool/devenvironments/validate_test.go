package devenvironments

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

func TestGetOptionalIntRejectsFractions(t *testing.T) {
	req := func(v any) mcp.CallToolRequest {
		var r mcp.CallToolRequest
		r.Params.Arguments = map[string]any{"n": v}
		return r
	}
	n, ok, err := getOptionalInt(req(3.0), "n")
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, 3, n)

	_, _, err = getOptionalInt(req(1.9), "n")
	assert.Error(t, err)

	_, _, err = getOptionalInt(req(-1.0), "n")
	assert.Error(t, err)

	_, ok, err = getOptionalInt(mcp.CallToolRequest{}, "n")
	assert.NoError(t, err)
	assert.False(t, ok)
}
