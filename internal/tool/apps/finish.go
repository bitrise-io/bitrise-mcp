package apps

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var Finish = bitrise.Tool{
	APIGroups: []string{"apps"},
	Definition: mcp.NewTool("finish_bitrise_app",
		mcp.WithDescription("Finish the setup of a Bitrise app. If this is successful, a build can be triggered via trigger_bitrise_build. If you have access to the repository, decide the project type, the stack ID, and the config to use, based on https://stacks.bitrise.io/, and the config should be also based on the project type."),
		mcp.WithString("app_slug",
			mcp.Description("The slug of the Bitrise app to finish setup for."),
			mcp.Required(),
		),
		mcp.WithString("project_type",
			mcp.Description("The type of project (e.g., android, ios, flutter, etc.)."),
			mcp.DefaultString("other"),
		),
		mcp.WithString("stack_id",
			mcp.Description("The stack ID to use for the app."),
			mcp.DefaultString("linux-docker-android-22.04"),
		),
		mcp.WithString("mode",
			mcp.Description("The mode of setup."),
			mcp.DefaultString("manual"),
		),
		mcp.WithString("config",
			mcp.Description("The configuration to use for the app (default is \"other-config\", other valid values are \"default-android-config\", \"default-ios-config\", \"default-macos-config\", etc)."),
			mcp.DefaultString("other-config"),
		),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/finish", appSlug),
			Body: map[string]any{
				"project_type": request.GetString("project_type", "other"),
				"stack_id":     request.GetString("stack_id", "linux-docker-android-22.04"),
				"mode":         request.GetString("mode", "manual"),
				"config":       request.GetString("config", "other-config"),
			},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
