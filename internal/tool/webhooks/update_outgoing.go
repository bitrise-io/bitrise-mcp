package webhooks

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var UpdateOutgoing = bitrise.Tool{
	APIGroups: []string{"outgoing-webhooks"},
	Definition: mcp.NewTool("update_outgoing_webhook",
		mcp.WithDescription("Update an outgoing webhook for an app. Even if you do not want to change one of the parameters, you still have to provide that parameter as well: simply use its existing value."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("webhook_slug",
			mcp.Description("Identifier of the webhook"),
			mcp.Required(),
		),
		mcp.WithArray("events",
			mcp.Description("List of events to trigger the webhook"),
			mcp.WithStringItems(),
		),
		mcp.WithString("url",
			mcp.Description("URL of the webhook"),
		),
		mcp.WithObject("headers",
			mcp.Description("Headers to be sent with the webhook"),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		webhookSlug, err := request.RequireString("webhook_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{}
		body["events"] = request.GetStringSlice("events", nil)
		body["url"] = request.GetString("url", "")
		body["headers"] = request.GetArguments()["headers"]

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPatch,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/outgoing-webhooks/%s", appSlug, webhookSlug),
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
