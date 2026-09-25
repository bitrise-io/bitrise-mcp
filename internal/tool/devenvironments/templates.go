package devenvironments

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/devenv"
	"github.com/mark3labs/mcp-go/mcp"
)

// ListTemplates lists all available templates.
var ListTemplates = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_list_templates",
		mcp.WithTitleAnnotation("List templates"),
		mcp.WithDescription("List all available devenv templates. Each template defines the stack, startup/warmup scripts, template variables, session inputs (required and optional), feature flags, workspace links, and optionally a device_spec — the iOS simulator / Android emulator every session created from it boots unless the create request overrides it with its own device_spec or no_device. By default, secret template variable values are omitted from the response; set include_secrets=true to include them."),
		mcp.WithBoolean("include_secrets",
			mcp.Description("When true, secret template variable values are included in the response. Defaults to false (secret values are omitted)."),
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
			Path:   devenv.WsPath(ctx, "/templates"),
			Params: params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("list templates", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

// GetTemplate retrieves a template by ID.
var GetTemplate = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_get_template",
		mcp.WithTitleAnnotation("Get template"),
		mcp.WithDescription("Get details of a specific template including startup/warmup scripts, stack, working directory, template variables, session inputs (with required/default_value/expose_as_env_var fields), feature flags, workspace links, and device_spec (the virtual device sessions created from it boot by default; absent when the template declares none — see bitrise_devenv_create for how a create request inherits, overrides or skips it). By default, secret template variable values are omitted from the response; set include_secrets=true to include them."),
		mcp.WithString("template_id",
			mcp.Description("The unique identifier (UUID) of the template"),
			mcp.Required(),
		),
		mcp.WithBoolean("include_secrets",
			mcp.Description("When true, secret template variable values are included in the response. Defaults to false (secret values are omitted)."),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		templateID, err := requireUUID(request, "template_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		params := map[string]string{}
		if request.GetBool("include_secrets", false) {
			params["include_secrets"] = "true"
		}
		res, err := devenv.CallAPI(ctx, devenv.CallAPIParams{
			Method: http.MethodGet,
			Path:   devenv.WsPath(ctx, fmt.Sprintf("/templates/%s", templateID)),
			Params: params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("get template", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

// CreateTemplate creates a new template.
var CreateTemplate = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_create_template",
		mcp.WithTitleAnnotation("Create template"),
		mcp.WithDescription(`Create a new devenv template. Use bitrise_devenv_list_stacks to find a valid stack_id and bitrise_devenv_list_machine_types to find a valid machine_type name. IMPORTANT: Provide the stack id (the 'id' field from bitrise_devenv_list_stacks) for stack_id, and the machine type name (not UUID) for machine_type. Optionally declare a device_spec so every session created from the template boots an iOS simulator / Android emulator; the stack and machine type must then fit the platform (iOS: a macOS stack; Android: a dockerless Android Linux stack such as ubuntu-resolute-26.04-bitrise-2026-android, not linux-docker-*; both: >= 4 vCPU / 6-8 GB) or the request is rejected with a device_spec.* field violation.`),
		mcp.WithString("name", mcp.Description("Template name"), mcp.Required()),
		mcp.WithString("description", mcp.Description("Template description")),
		mcp.WithString("startup_script", mcp.Description("Bash script that runs every time a session starts")),
		mcp.WithString("warmup_script", mcp.Description("Bash script that runs once during initial session creation")),
		mcp.WithString("stack_id", mcp.Description(`Stack ID (use bitrise_devenv_list_stacks to find the id, e.g. 'osx-xcode-16.0.x-edge').`), mcp.Required()),
		mcp.WithString("machine_type", mcp.Description("Machine type name (use bitrise_devenv_list_machine_types to find the name, e.g. 'g2.mac.m2pro.4c')"), mcp.Required()),
		mcp.WithString("working_directory", mcp.Description("Working directory for terminal sessions (absolute path)")),
		mcp.WithArray("template_variables",
			mcp.Description("Template variables baked into this template (available in boot scripts)"),
			mcp.Items(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"key":               map[string]any{"type": "string", "description": "Key/name of the variable"},
					"value":             map[string]any{"type": "string", "description": "Value of the variable"},
					"is_secret":         map[string]any{"type": "boolean", "description": "Whether this is a secret value (encrypted at rest)"},
					"expose_as_env_var": map[string]any{"type": "boolean", "description": "Whether to expose as environment variable in terminal sessions"},
				},
				"required": []string{"key", "value"},
			}),
		),
		mcp.WithArray("session_inputs",
			mcp.Description("Session inputs that users provide when creating sessions from this template"),
			mcp.Items(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"key":               map[string]any{"type": "string", "description": "Key/name of the input"},
					"description":       map[string]any{"type": "string", "description": "Description explaining what this input is for"},
					"required":          map[string]any{"type": "boolean", "description": "Whether this input is required (default: false)"},
					"default_value":     map[string]any{"type": "string", "description": "Default value for the input. REQUIRED when required is false (must be a non-empty string). Omit or leave empty when required is true."},
					"expose_as_env_var": map[string]any{"type": "boolean", "description": "Whether to expose as environment variable in terminal sessions"},
				},
				"required": []string{"key"},
			}),
		),
		mcp.WithArray("feature_flags",
			mcp.Description("Feature flags to toggle optional behaviors"),
			mcp.Items(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name":        map[string]any{"type": "string", "description": "Name of the feature flag"},
					"description": map[string]any{"type": "string", "description": "Description of what this flag enables"},
				},
				"required": []string{"name"},
			}),
		),
		mcp.WithArray("workspace_links",
			mcp.Description("IDE workspace folder links for quick access"),
			mcp.Items(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"label":             map[string]any{"type": "string", "description": "Display label for the button"},
					"folder_path":       map[string]any{"type": "string", "description": "Remote folder path to open"},
					"feature_flag_name": map[string]any{"type": "string", "description": "Optional: feature flag name that controls visibility"},
				},
				"required": []string{"label", "folder_path"},
			}),
		),
		mcp.WithObject("device_spec",
			deviceSpecSchema(`Optional virtual device declared on the template: sessions created from this template boot this device unless the create request overrides it (a device_spec without a platform tweaks it per field; one with a platform replaces it whole) or skips it (no_device=true). `+deviceSpecFieldsDoc+` platform is required here. The template's stack_id and machine_type must fit the platform.`, "platform")...,
		),
		mcp.WithDestructiveHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		body := map[string]any{
			"name":         request.GetString("name", ""),
			"stack_id":     request.GetString("stack_id", ""),
			"machine_type": request.GetString("machine_type", ""),
		}
		for _, key := range []string{"description", "startup_script", "warmup_script", "working_directory"} {
			if v := request.GetString(key, ""); v != "" {
				body[key] = v
			}
		}
		for _, key := range []string{"template_variables", "session_inputs", "feature_flags", "workspace_links"} {
			if v, ok := request.GetArguments()[key]; ok {
				body[key] = v
			}
		}
		if deviceSpec, ok := request.GetArguments()["device_spec"]; ok {
			if err := requireNonEmptyString(deviceSpec, "device_spec", "platform"); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			body["device_spec"] = deviceSpec
		}

		res, err := devenv.CallAPI(ctx, devenv.CallAPIParams{
			Method: http.MethodPost,
			Path:   devenv.WsPath(ctx, "/templates"),
			Body:   body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("create template", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

// UpdateTemplate updates an existing template.
var UpdateTemplate = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_update_template",
		mcp.WithTitleAnnotation("Update template"),
		mcp.WithDescription("Update an existing devenv template. Only provided fields are updated. For array fields (template_variables, session_inputs, feature_flags, workspace_links), providing a new array replaces ALL existing entries. Omit an array field to leave it unchanged. The template's device works the same way: pass device_spec to replace the declared device as a whole (there is no per-field merge on the template itself), pass clear_device_spec=true to remove it so sessions boot without a device, or omit both to leave it unchanged. Existing sessions keep the device they were created with."),
		mcp.WithString("template_id", mcp.Description("The unique identifier of the template to update"), mcp.Required()),
		mcp.WithString("name", mcp.Description("Updated template name")),
		mcp.WithString("description", mcp.Description("Updated description")),
		mcp.WithString("startup_script", mcp.Description("Updated startup script")),
		mcp.WithString("warmup_script", mcp.Description("Updated warmup script")),
		mcp.WithString("stack_id", mcp.Description(`Updated stack ID (use bitrise_devenv_list_stacks to find the id, e.g. 'osx-xcode-16.0.x-edge').`)),
		mcp.WithString("machine_type", mcp.Description("Updated machine type name (use bitrise_devenv_list_machine_types to find the name)")),
		mcp.WithString("working_directory", mcp.Description("Updated working directory")),
		mcp.WithArray("template_variables",
			mcp.Description("Replace ALL template variables with this list. Omit to leave unchanged. Pass empty array to clear all."),
			mcp.Items(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"key":               map[string]any{"type": "string", "description": "Key/name of the variable"},
					"value":             map[string]any{"type": "string", "description": "Value of the variable"},
					"is_secret":         map[string]any{"type": "boolean", "description": "Whether this is a secret value (encrypted at rest)"},
					"expose_as_env_var": map[string]any{"type": "boolean", "description": "Whether to expose as environment variable in terminal sessions"},
				},
				"required": []string{"key", "value"},
			}),
		),
		mcp.WithArray("session_inputs",
			mcp.Description("Replace ALL session inputs with this list. Omit to leave unchanged. Pass empty array to clear all."),
			mcp.Items(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"key":               map[string]any{"type": "string", "description": "Key/name of the input"},
					"description":       map[string]any{"type": "string", "description": "Description explaining what this input is for"},
					"required":          map[string]any{"type": "boolean", "description": "Whether this input is required (default: false)"},
					"default_value":     map[string]any{"type": "string", "description": "Default value for the input. REQUIRED when required is false (must be a non-empty string). Omit or leave empty when required is true."},
					"expose_as_env_var": map[string]any{"type": "boolean", "description": "Whether to expose as environment variable in terminal sessions"},
				},
				"required": []string{"key"},
			}),
		),
		mcp.WithArray("feature_flags",
			mcp.Description("Replace ALL feature flags with this list. Omit to leave unchanged. Pass empty array to clear all."),
			mcp.Items(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name":        map[string]any{"type": "string", "description": "Name of the feature flag"},
					"description": map[string]any{"type": "string", "description": "Description of what this flag enables"},
				},
				"required": []string{"name"},
			}),
		),
		mcp.WithArray("workspace_links",
			mcp.Description("Replace ALL workspace links with this list. Omit to leave unchanged. Pass empty array to clear all."),
			mcp.Items(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"label":             map[string]any{"type": "string", "description": "Display label for the button"},
					"folder_path":       map[string]any{"type": "string", "description": "Remote folder path to open"},
					"feature_flag_name": map[string]any{"type": "string", "description": "Optional: feature flag name that controls visibility"},
				},
				"required": []string{"label", "folder_path"},
			}),
		),
		mcp.WithObject("device_spec",
			deviceSpecSchema(`Replace the template's declared virtual device with this one (the whole object is replaced, so include every field you want kept). Omit to leave the device unchanged; use clear_device_spec to remove it. `+deviceSpecFieldsDoc+` platform is required here. The template's stack_id and machine_type (after this update) must fit the platform.`, "platform")...,
		),
		mcp.WithBoolean("clear_device_spec",
			mcp.Description("When true, remove the template's declared device so new sessions boot without one. Cannot be combined with device_spec."),
		),
		mcp.WithDestructiveHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		templateID, err := requireUUID(request, "template_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{}
		for _, key := range []string{"name", "description", "startup_script", "warmup_script", "stack_id", "machine_type", "working_directory"} {
			if v := request.GetString(key, ""); v != "" {
				body[key] = v
			}
		}
		// Array fields: auto-set the corresponding update_* flag when the array is provided.
		// Backend requires the flag to be true for array changes to take effect.
		arrayFields := map[string]string{
			"template_variables": "update_template_variables",
			"session_inputs":     "update_session_inputs",
			"feature_flags":      "update_feature_flags",
			"workspace_links":    "update_workspace_links",
		}
		for arrayKey, flagKey := range arrayFields {
			if v, ok := request.GetArguments()[arrayKey]; ok {
				body[arrayKey] = v
				body[flagKey] = true
			}
		}
		// The device follows the same convention: update_device_spec is set
		// whenever the caller asks for a change. With device_spec the template's
		// device becomes that object; with clear_device_spec the flag goes out
		// alone, which the backend reads as "no device".
		deviceSpec, hasDevice := request.GetArguments()["device_spec"]
		clearDevice := request.GetBool("clear_device_spec", false)
		if hasDevice && clearDevice {
			return mcp.NewToolResultError("clear_device_spec cannot be combined with device_spec — either replace the template's device or remove it"), nil
		}
		if hasDevice {
			if err := requireNonEmptyString(deviceSpec, "device_spec", "platform"); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			body["device_spec"] = deviceSpec
			body["update_device_spec"] = true
		}
		if clearDevice {
			body["update_device_spec"] = true
		}

		res, err := devenv.CallAPI(ctx, devenv.CallAPIParams{
			Method: http.MethodPatch,
			Path:   devenv.WsPath(ctx, fmt.Sprintf("/templates/%s", templateID)),
			Body:   body,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("update template", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}

// DeleteTemplate soft-deletes a template.
var DeleteTemplate = bitrise.Tool{
	Definition: mcp.NewTool("bitrise_devenv_delete_template",
		mcp.WithTitleAnnotation("Delete template"),
		mcp.WithDescription("Delete a devenv template. Existing sessions continue to work from their snapshotted template configuration, but will be marked with templateDeleted=true."),
		mcp.WithString("template_id", mcp.Description("The unique identifier of the template to delete"), mcp.Required()),
		mcp.WithDestructiveHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		templateID, err := requireUUID(request, "template_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		res, err := devenv.CallAPI(ctx, devenv.CallAPIParams{
			Method: http.MethodDelete,
			Path:   devenv.WsPath(ctx, fmt.Sprintf("/templates/%s", templateID)),
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("delete template", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
