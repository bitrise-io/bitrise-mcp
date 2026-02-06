package webhooks

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var CreateOutgoing = bitrise.Tool{
	APIGroups: []string{"outgoing-webhooks"},
	Definition: mcp.NewTool("create_outgoing_webhook",
		mcp.WithDescription("Create an outgoing webhook for an app."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithArray("events",
			mcp.Description("List of events to trigger the webhook"),
			mcp.Required(),
			mcp.WithStringItems(),
		),
		mcp.WithString("url",
			mcp.Description("URL of the webhook"),
			mcp.Required(),
		),
		mcp.WithObject("headers",
			mcp.Description("Headers to be sent with the webhook"),
		),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		events, err := request.RequireStringSlice("events")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		url, err := request.RequireString("url")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{
			"events": events,
			"url":    url,
		}
		if v, ok := request.GetArguments()["headers"]; ok {
			body["headers"] = v
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/outgoing-webhooks", appSlug),
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
