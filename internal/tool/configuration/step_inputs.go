package configuration

import (
	"context"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var StepInputs = bitrise.Tool{
	APIGroups: []string{"configuration", "read-only"},
	Definition: mcp.NewTool("step_inputs",
		mcp.WithDescription("List inputs of a step with their defaults, allowed values etc."),
		mcp.WithString("cvs",
			// Step reference format: https://docs.bitrise.io/en/bitrise-ci/references/steps-reference/step-reference-id-format.html
			mcp.Description("Step reference formatted as `step_lib_source::step_id@version`. Only `step_id` is required, `version` should be added when known, `step_lib_source` is only necessary for custom step sources."),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := request.RequireString("cvs")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIBaseURL,
			Path:    "/step-inputs",
			Params:  map[string]any{"cvs": query},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
