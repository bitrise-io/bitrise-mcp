package devenvironments

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/devenv"
	"github.com/mark3labs/mcp-go/mcp"
)

// ListSavedInputs lists all saved inputs for the current user.
var ListSavedInputs = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_list_saved_inputs",
		mcp.WithTitleAnnotation("List saved inputs"),
		mcp.WithDescription("List the current user's Dev Environments saved inputs: personal credentials/values reused as template session inputs when creating sessions. Not Bitrise CI step inputs (step_inputs), and not CI app secrets or env vars, which this server does not expose. By default, secret values are redacted from the response; set include_secrets=true to receive plaintext values."),
		mcp.WithBoolean("include_secrets",
			mcp.Description("When true, secret values are returned in plaintext. Defaults to false (secret values are redacted)."),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params := map[string]string{}
		if request.GetBool("include_secrets", false) {
			params["include_secrets"] = "true"
		}
		res, err := devenv.CallAPI(ctx, devenv.CallAPIParams{
			Method: http.MethodGet,
			Path:   "/v1/saved-inputs",
			Params: params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("list saved inputs", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

// GetSavedInput retrieves a single saved input.
var GetSavedInput = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_get_saved_input",
		mcp.WithTitleAnnotation("Get saved input"),
		mcp.WithDescription("Get details of a specific saved input. By default, the secret value is redacted from the response; set include_secrets=true to receive the plaintext value."),
		mcp.WithString("saved_input_id", mcp.Description("The unique identifier of the saved input"), mcp.Required()),
		mcp.WithBoolean("include_secrets",
			mcp.Description("When true, the secret value is returned in plaintext. Defaults to false (secret value is redacted)."),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := requireUUID(request, "saved_input_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		params := map[string]string{}
		if request.GetBool("include_secrets", false) {
			params["include_secrets"] = "true"
		}
		res, err := devenv.CallAPI(ctx, devenv.CallAPIParams{
			Method: http.MethodGet,
			Path:   fmt.Sprintf("/v1/saved-inputs/%s", id),
			Params: params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("get saved input", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

// CreateSavedInput creates a new saved input.
var CreateSavedInput = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_create_saved_input",
		mcp.WithTitleAnnotation("Create saved input"),
		mcp.WithDescription("Create a new saved input (credential/value). The key should match a template's session input key for automatic pre-fill when creating sessions."),
		mcp.WithString("key", mcp.Description("Key/name of the input"), mcp.Required()),
		mcp.WithString("value", mcp.Description("Value of the input"), mcp.Required()),
		mcp.WithBoolean("is_secret", mcp.Description("Whether this is a secret value (will be encrypted at rest)")),
		mcp.WithDestructiveHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		body := map[string]any{
			"key":   request.GetString("key", ""),
			"value": request.GetString("value", ""),
		}
		if isSecret, ok := request.GetArguments()["is_secret"]; ok {
			body["is_secret"] = isSecret
		}

		res, err := devenv.CallAPI(ctx, devenv.CallAPIParams{
			Method: http.MethodPost,
			Path:   "/v1/saved-inputs",
			Body:   body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("create saved input", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

// UpdateSavedInput updates an existing saved input.
var UpdateSavedInput = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_update_saved_input",
		mcp.WithTitleAnnotation("Update saved input"),
		mcp.WithDescription("Update an existing saved input value."),
		mcp.WithString("saved_input_id", mcp.Description("The unique identifier of the saved input to update"), mcp.Required()),
		mcp.WithString("value", mcp.Description("Updated value"), mcp.Required()),
		mcp.WithBoolean("is_secret", mcp.Description("Updated secret flag")),
		mcp.WithDestructiveHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := requireUUID(request, "saved_input_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{
			"value": request.GetString("value", ""),
		}
		if isSecret, ok := request.GetArguments()["is_secret"]; ok {
			body["is_secret"] = isSecret
		}

		res, err := devenv.CallAPI(ctx, devenv.CallAPIParams{
			Method: http.MethodPatch,
			Path:   fmt.Sprintf("/v1/saved-inputs/%s", id),
			Body:   body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("update saved input", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

// DeleteSavedInput deletes a saved input.
var DeleteSavedInput = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_delete_saved_input",
		mcp.WithTitleAnnotation("Delete saved input"),
		mcp.WithDescription("Delete a saved input. Sessions that used this input are not affected (values are snapshotted at creation time)."),
		mcp.WithString("saved_input_id", mcp.Description("The unique identifier of the saved input to delete"), mcp.Required()),
		mcp.WithDestructiveHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := requireUUID(request, "saved_input_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		res, err := devenv.CallAPI(ctx, devenv.CallAPIParams{
			Method: http.MethodDelete,
			Path:   fmt.Sprintf("/v1/saved-inputs/%s", id),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("delete saved input", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
