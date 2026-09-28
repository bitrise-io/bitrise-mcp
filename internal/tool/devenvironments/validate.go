package devenvironments

import (
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
)

// requireUUID extracts a required string parameter and validates it as a UUID.
func requireUUID(request mcp.CallToolRequest, name string) (string, error) {
	value, err := request.RequireString(name)
	if err != nil {
		return "", err
	}
	if _, err := uuid.Parse(value); err != nil {
		return "", fmt.Errorf("invalid %s: not a valid UUID", name)
	}
	return value, nil
}

// optionalUUID extracts an optional string parameter and, when it is present
// and non-empty, validates it as a UUID. Returns "" when absent or empty.
func optionalUUID(request mcp.CallToolRequest, name string) (string, error) {
	value := request.GetString(name, "")
	if value == "" {
		return "", nil
	}
	if _, err := uuid.Parse(value); err != nil {
		return "", fmt.Errorf("invalid %s: not a valid UUID", name)
	}
	return value, nil
}

// requireInt extracts a required whole-number parameter (a screen coordinate or
// size); a missing, non-numeric or fractional value is an error, so a
// schema-valid call is never silently moved to a different pixel.
func requireInt(request mcp.CallToolRequest, name string) (int, error) {
	val, ok := request.GetArguments()[name]
	if !ok {
		return 0, fmt.Errorf("%s is required", name)
	}
	f, ok := val.(float64)
	if !ok {
		return 0, fmt.Errorf("%s must be a number", name)
	}
	if f != math.Trunc(f) {
		return 0, fmt.Errorf("%s must be a whole number", name)
	}
	return int(f), nil
}

// requireInts is requireInt for four parameters at once.
func requireInts(request mcp.CallToolRequest, a, b, c, d string) (int, int, int, int, error) {
	var out [4]int
	for i, name := range []string{a, b, c, d} {
		v, err := requireInt(request, name)
		if err != nil {
			return 0, 0, 0, 0, err
		}
		out[i] = v
	}
	return out[0], out[1], out[2], out[3], nil
}

// getOptionalInt extracts an optional numeric parameter as an int.
// Returns (value, true, nil) if present and valid, (0, false, nil) if absent,
// or (0, false, error) if the value is not a valid number or is negative.
func getOptionalInt(request mcp.CallToolRequest, name string) (int, bool, error) {
	val, ok := request.GetArguments()[name]
	if !ok {
		return 0, false, nil
	}
	f, ok := val.(float64)
	if !ok {
		return 0, false, fmt.Errorf("%s must be a number", name)
	}
	if f != math.Trunc(f) {
		return 0, false, fmt.Errorf("%s must be a whole number", name)
	}
	n := int(f)
	if n < 0 {
		return 0, false, fmt.Errorf("%s must be >= 0", name)
	}
	return n, true, nil
}
