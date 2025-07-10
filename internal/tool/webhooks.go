package tool

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
)

var listOutgoingWebhooks = Tool{
	APIGroups: []string{"webhooks", "read-only"},
	Definition: mcp.NewTool("list_outgoing_webhooks",
		mcp.WithDescription("List the outgoing webhooks of an app."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodGet,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/apps/%s/outgoing-webhooks", appSlug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var deleteOutgoingWebhook = Tool{
	APIGroups: []string{"webhooks"},
	Definition: mcp.NewTool("delete_outgoing_webhook",
		mcp.WithDescription("Delete the outgoing webhook of an app."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("webhook_slug",
			mcp.Description("Identifier of the webhook"),
			mcp.Required(),
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

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodDelete,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/apps/%s/outgoing-webhooks/%s", appSlug, webhookSlug),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var createOutgoingWebhook = Tool{
	APIGroups: []string{"webhooks"},
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

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodPost,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/apps/%s/outgoing-webhooks", appSlug),
			body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

var updateOutgoingWebhook = Tool{
	APIGroups: []string{"webhooks"},
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

		res, err := callAPI(ctx, callAPIParams{
			method:  http.MethodPatch,
			baseURL: apiBaseURL,
			path:    fmt.Sprintf("/apps/%s/outgoing-webhooks/%s", appSlug, webhookSlug),
			body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
