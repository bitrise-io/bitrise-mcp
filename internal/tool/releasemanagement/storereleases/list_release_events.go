package storereleases

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

var ListReleaseEvents = bitrise.Tool{
	APIGroups: []string{"release-management-store-releases", "release-management", "read-only"},
	Definition: mcp.NewTool("list_release_events",
		mcp.WithDescription("Lists the recorded event history of an app version, newest first: who did what and when across the release stages. "+
			"Paginate with a cursor: pass a response's 'next_before' back as 'before' to read the next page, and stop once it comes back null. 'total_count' and 'available_actors' are present on the first page only. "+
			"Requires a Standard license for the connected app (412 otherwise)."),
		mcp.WithString("app_version_id",
			mcp.Description(appVersionIDDescription),
			mcp.Required(),
		),
		mcp.WithString("before",
			mcp.Description("Returns the events recorded strictly before this timestamp. Echo back the 'next_before' of the previous response verbatim, keeping its microseconds."),
		),
		mcp.WithNumber("limit",
			mcp.Description("Specifies the maximum number of events returned per request. Default value is 10."),
			mcp.DefaultNumber(10),
		),
		mcp.WithString("search",
			mcp.Description("Returns only the events whose title contains this text, case-insensitively."),
		),
		mcp.WithString("from",
			mcp.Description("Start of the inclusive ISO 8601 time range to return events from. Must be given together with 'to'."),
		),
		mcp.WithString("to",
			mcp.Description("End of the inclusive ISO 8601 time range to return events from. Must be given together with 'from'."),
		),
		mcp.WithArray("triggered_by",
			mcp.Description("Returns only the events triggered by any of these actors: a user slug, or a machine actor type ('store', 'automation' or 'preset'). Actor types are listed in 'available_actors' of the first page."),
			mcp.WithStringItems(),
		),
		mcp.WithTitleAnnotation("List Release Events"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appVersionID, err := request.RequireString("app_version_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		params := map[string]any{}
		for _, key := range []string{"before", "search", "from", "to"} {
			if v := request.GetString(key, ""); v != "" {
				params[key] = v
			}
		}
		if v := request.GetInt("limit", 10); v != 10 {
			params["limit"] = strconv.Itoa(v)
		}
		if v := request.GetStringSlice("triggered_by", nil); len(v) > 0 {
			params["triggered_by[]"] = v
		}

		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIRMStoreReleasesBaseURL,
			Path:    fmt.Sprintf("/app-versions/%s/release-events", appVersionID),
			Params:  params,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}
		return mcp.NewToolResultText(res), nil
	},
}
