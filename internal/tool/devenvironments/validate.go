package devenvironments

import (
	"fmt"

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
	n := int(f)
	if n < 0 {
		return 0, false, fmt.Errorf("%s must be >= 0", name)
	}
	return n, true, nil
}
