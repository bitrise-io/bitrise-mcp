package storereleases

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var ResumeGooglePlayStagedRolloutSchedule = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management"},
	Definition: mcp.NewTool("resume_google_play_staged_rollout_schedule",
		mcp.WithDescription("Resumes a paused staged rollout schedule of an Android app version with the remaining steps re-timed. Send the steps that are still to run, each identified by the 'id' from 'get_google_play_staged_rollout_schedule'. "+
			"Resuming re-arms production rollouts, so confirm the new dates with the user before calling. Fails with 412 when the schedule is not paused or has finished. Android only."),
		mcp.WithString("app_version_id",
			mcp.Description(appVersionIDDescription),
			mcp.Required(),
		),
		mcp.WithString("location",
			mcp.Description("The time zone the schedule dates are given in: any TZ database name (e.g. 'Europe/Budapest') or 'UTC'."),
			mcp.Required(),
		),
		mcp.WithArray("schedule",
			mcp.Description("The remaining rollout steps, each with the 'id' of the existing step, its 'percentage' (greater than 0 and at most 100) and the new 'when' ISO 8601 date-time."),
			mcp.Required(),
			mcp.Items(map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id":         map[string]any{"type": "string"},
					"percentage": map[string]any{"type": "number"},
					"when":       map[string]any{"type": "string"},
				},
				"required": []string{"id", "percentage", "when"},
			}),
		),
		mcp.WithTitleAnnotation("Resume Google Play Staged Rollout Schedule"),
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
			Path:    fmt.Sprintf("/app-versions/%s/google-play-store/staged-rollout-schedule/resume", appVersionID),
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
