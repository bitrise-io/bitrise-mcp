package storereleases

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var CreateGooglePlayStagedRolloutSchedule = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management"},
	Definition: mcp.NewTool("create_google_play_staged_rollout_schedule",
		mcp.WithDescription("Creates a staged rollout schedule for an Android app version: a list of dates at which Release Management raises the Google Play production rollout to the given percentage. "+
			"Once a step runs it has published to production and cannot be undone, so confirm the schedule with the user before calling. The schedule can be paused and resumed later. "+
			"Android only, and a release candidate must already be chosen (400 otherwise)."),
		mcp.WithString("app_version_id",
			mcp.Description(appVersionIDDescription),
			mcp.Required(),
		),
		mcp.WithString("location",
			mcp.Description("The time zone the schedule dates are given in: any TZ database name (e.g. 'Europe/Budapest') or 'UTC'."),
			mcp.Required(),
		),
		mcp.WithArray("schedule",
			mcp.Description("The rollout steps, each with a 'percentage' (greater than 0 and at most 100, where 100 is a full release) and a 'when' ISO 8601 date-time at which to apply it."),
			mcp.Required(),
			mcp.Items(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"percentage": map[string]any{"type": "number"},
					"when":       map[string]any{"type": "string"},
				},
				"required": []string{"percentage", "when"},
			}),
		),
		mcp.WithTitleAnnotation("Create Google Play Staged Rollout Schedule"),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appVersionID, err := request.RequireString("app_version_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		location, err := request.RequireString("location")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		schedule, err := requireArray(request, "schedule")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodPost,
			BaseURL: bitrise.APIRMStoreReleasesBaseURL,
			Path:    fmt.Sprintf("/app-versions/%s/google-play-store/staged-rollout-schedule", appVersionID),
			Body: map[string]any{
				"location": location,
				"schedule": schedule,
			},
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
