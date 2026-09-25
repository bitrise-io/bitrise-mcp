package apps

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var RegisterSSHKey = bitrise.Tool{
	APIGroups: []string{"apps"},
	Definition: mcp.NewTool("register_ssh_key",
		mcp.WithDescription("Add an SSH key pair to a Bitrise CI app so CI can clone its repository. Not for Dev Environments sessions: their SSH access is provisioned automatically (bitrise_devenv_get / bitrise_devenv_open_remote_access) and a local SSH agent is forwarded by bitrise_devenv_execute."),
		mcp.WithString("app_slug",
			mcp.Description("Identifier of the Bitrise app"),
			mcp.Required(),
		),
		mcp.WithString("auth_ssh_private_key",
			mcp.Description("Private SSH key"),
			mcp.Required(),
		),
		mcp.WithString("auth_ssh_public_key",
			mcp.Description("Public SSH key"),
			mcp.Required(),
		),
		mcp.WithBoolean("is_register_key_into_provider_service",
			mcp.Description("Register the key in the provider service"),
		),
		mcp.WithTitleAnnotation("Register SSH Key"),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		privateKey, err := request.RequireString("auth_ssh_private_key")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		publicKey, err := request.RequireString("auth_ssh_public_key")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		body := map[string]any{
			"auth_ssh_private_key": privateKey,
			"auth_ssh_public_key":  publicKey,
		}
		regKey := "is_register_key_into_provider_service"
		if _, ok := request.GetArguments()[regKey]; ok {
			body[regKey] = request.GetBool(regKey, false)
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APIBaseURL,
			Path:    fmt.Sprintf("/apps/%s/register-ssh-key", appSlug),
			Body:    body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
