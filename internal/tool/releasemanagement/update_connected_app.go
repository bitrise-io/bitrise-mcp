package releasemanagement

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var UpdateConnectedApp = bitrise.Tool{
	APIGroups: []string{"release-management"},
	Definition: mcp.NewTool("update_connected_app",
		mcp.WithDescription("Updates a connected app."),
		mcp.WithString("connected_app_id",
			mcp.Description("The uuidV4 identifier for your connected app."),
			mcp.Required(),
		),
		mcp.WithBoolean("connect_to_store",
			mcp.Description("If true, will check connected app validity against the Apple App Store or Google Play Store (dependent on the platform of your connected app). This means, that the already set or just given store_app_id will be validated against the Store, using the already set or just given store credential id."),
			mcp.DefaultBool(false),
		),
		mcp.WithString("store_app_id",
			mcp.Description("The store identifier for your app. You can change the previously set store_app_id to match the one in the App Store or Google Play depending on the app platform. This is especially useful if you want to connect your app with the store as the system will validate the given store_app_id against the Store. In case of iOS platform it is the bundle id. In case of Android platform it is the package name."),
		),
		mcp.WithString("store_credential_id",
			mcp.Description("If you have credentials added on Bitrise, you can decide to select one for your app. In case of ios platform it will be an Apple API credential id. In case of android platform it will be a Google Service credential id."),
		),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connectedAppID, err := request.RequireString("connected_app_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{}
		if v := request.GetBool("connect_to_store", false); v {
			body["connect_to_store"] = v
		}
		if v := request.GetString("store_app_id", ""); v != "" {
			body["store_app_id"] = v
		}
		if v := request.GetString("store_credential_id", ""); v != "" {
			body["store_credential_id"] = v
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPatch,
			BaseURL: bitrise.APIRMBaseURL,
			Path:    fmt.Sprintf("/connected-apps/%s", connectedAppID),
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
