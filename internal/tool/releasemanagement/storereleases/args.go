package storereleases

import (
	"fmt"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
)

const appVersionIDDescription = "The uuidV4 identifier of the app version (a release in the Release Management UI)."

// optionalBool tells an omitted argument apart from an explicit false, so a
// PATCH never overwrites a setting the caller did not mention. It coerces the
// same shapes as request.RequireBool and rejects anything else, because a
// silently dropped argument makes the caller believe a change it never sent.
func optionalBool(request mcp.CallToolRequest, key string) (value, ok bool, err error) {
	v, present := request.GetArguments()[key]
	if !present || v == nil {
		return false, false, nil
	}
	switch b := v.(type) {
	case bool:
		return b, true, nil
	case string:
		parsed, parseErr := strconv.ParseBool(b)
		if parseErr != nil {
			return false, false, fmt.Errorf("%s must be a boolean, got %q", key, b)
		}
		return parsed, true, nil
	case int:
		return b != 0, true, nil
	case float64:
		return b != 0, true, nil
	default:
		return false, false, fmt.Errorf("%s must be a boolean", key)
	}
}

// optionalArray rejects a present non-array value instead of dropping it, so a
// caller sending a bare object instead of a one-element array is told so.
func optionalArray(request mcp.CallToolRequest, key string) (value any, ok bool, err error) {
	v, present := request.GetArguments()[key]
	if !present || v == nil {
		return nil, false, nil
	}
	if _, isArray := v.([]any); !isArray {
		return nil, false, fmt.Errorf("%s must be an array", key)
	}
	return v, true, nil
}

func requireArray(request mcp.CallToolRequest, key string) (any, error) {
	v, ok, err := optionalArray(request, key)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("required argument %q not found", key)
	}
	if len(v.([]any)) == 0 {
		return nil, fmt.Errorf("%s must have at least one element", key)
	}
	return v, nil
}
