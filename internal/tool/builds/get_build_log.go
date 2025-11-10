package builds

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/mark3labs/mcp-go/mcp"
)

type GetBuildLogResponse struct {
	LogLines   string `json:"log_lines" jsonschema_description:"The requested lines of the build log."`
	NextOffset int    `json:"next_offset,omitempty" jsonschema_description:"The offset to use to read the next portion of the log, if any."`
	TotalLines int    `json:"total_lines" jsonschema_description:"The total number of lines in the build log."`
}

var GetBuildLog = bitrise.Tool{
	APIGroups: []string{"builds", "read-only"},
	Definition: mcp.NewTool("get_build_log",
		mcp.WithDescription("Get the build log of a specified build of a Bitrise app."),
		mcp.WithString("app_slug",
			mcp.Description(`Identifier of the Bitrise app (e.g., "d8db74e2675d54c4" or "8eb495d0-f653-4eed-910b-8d6b56cc0ec7")`),
			mcp.Required(),
		),
		mcp.WithString("build_slug",
			mcp.Description("Identifier of the Bitrise build"),
			mcp.Required(),
		),
		mcp.WithString("step_uuid",
			mcp.Description("UUID of the step to get the log for. If not provided, the full build log is returned. Always provide this value whenever possible to avoid large log responses and running out of the LLM context window."),
		),
		mcp.WithNumber("offset",
			mcp.Description("The line number to start reading from. Defaults to 0. Set -1 to read from the end of the log. Failures are usually at the end of the log."),
			mcp.DefaultNumber(0),
		),
		mcp.WithNumber("limit",
			mcp.Description("The number of lines to read. Defaults to 2000. Set to a high value to read the entire log."),
			mcp.DefaultNumber(2000),
		),
		mcp.WithOutputSchema[GetBuildLogResponse](),
	),
	Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appSlug, err := request.RequireString("app_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		buildSlug, err := request.RequireString("build_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		stepUUID := request.GetString("step_uuid", "")
		offset := request.GetInt("offset", 0)
		limit := request.GetInt("limit", 2000)
		if limit <= 0 {
			return mcp.NewToolResultError("limit must be greater than 0"), nil
		}

		path := fmt.Sprintf("/apps/%s/builds/%s/log", appSlug, buildSlug)
		if stepUUID != "" {
			path += fmt.Sprintf("/steps/%s", stepUUID)
		}
		res, err := bitrise.CallAPI(ctx, bitrise.CallAPIParams{
			Method:  http.MethodGet,
			BaseURL: bitrise.APIBaseURL,
			Path:    path,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("call api", err), nil
		}

		logGetter := getFullLog
		if stepUUID != "" {
			logGetter = getStepLog
		}
		log, err := logGetter(res)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("get log", err), nil
		}
		logWindow := logWindow{Log: log, Offset: offset, Limit: limit}
		a, err := json.Marshal(logWindow.Peek())
		if err != nil {
			return mcp.NewToolResultErrorFromErr("marshal log window", err), nil
		}
		return mcp.NewToolResultText(string(a)), nil
	},
}

func getFullLog(resBitriseRaw string) (string, error) {
	var resBitrise struct {
		URL       string `json:"expiring_raw_log_url"`
		LogChunks []struct {
			Chunk    string `json:"chunk"`
			Position int    `json:"position"`
		} `json:"log_chunks"`
	}
	if err := json.Unmarshal([]byte(resBitriseRaw), &resBitrise); err != nil {
		return "", fmt.Errorf("unmarshal bitrise response: %w", err)
	}
	if resBitrise.URL == "" {
		var foundFirstChunk bool
		var log string
		for _, chunk := range resBitrise.LogChunks {
			if chunk.Position == 1 {
				foundFirstChunk = true
			}
			log += chunk.Chunk
		}
		if !foundFirstChunk {
			// We could keep scrolling back on the API in this case using
			// `before_timestamp` to collect the full log but I didn't want to
			// complicate things further now, this is a temporary edge case.
			log = "[incomplete log: processing is still ongoing]\n\n" + log
		}
		return log, nil
	}
	log, err := httpGet(resBitrise.URL)
	if err != nil {
		return "", fmt.Errorf("get raw log: %w", err)
	}
	return log, nil
}

func getStepLog(resBitriseRaw string) (string, error) {
	var resBitrise struct {
		URL string `json:"expiring_raw_log_url"`
	}
	if err := json.Unmarshal([]byte(resBitriseRaw), &resBitrise); err != nil {
		return "", fmt.Errorf("unmarshal bitrise response: %w", err)
	}

	rawLog, err := httpGet(resBitrise.URL)
	if err != nil {
		return "", fmt.Errorf("get step raw log: %w", err)
	}
	var logChunks []struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(rawLog), &logChunks); err != nil {
		return "", fmt.Errorf("unmarshal log chunks: %w", err)
	}
	var log string
	for _, chunk := range logChunks {
		log += chunk.Message
	}
	return log, nil
}

func httpGet(url string) (string, error) {
	httpClient := http.Client{Timeout: 15 * time.Second}
	resLog, err := httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("http get: %w", err)
	}
	defer resLog.Body.Close()
	if resLog.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http status code %d", resLog.StatusCode)
	}
	a, err := io.ReadAll(resLog.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	return string(a), nil
}

type logWindow struct {
	Offset int
	Limit  int
	Log    string
}

func (lw logWindow) Peek() GetBuildLogResponse {
	lines := strings.Split(lw.Log, "\n")
	// Read forward:
	if lw.Offset >= 0 {
		if lw.Offset >= len(lines) {
			return GetBuildLogResponse{TotalLines: len(lines)}
		}
		end := lw.Offset + lw.Limit
		if end > len(lines) {
			end = len(lines)
		}
		nextOffset := end
		if nextOffset >= len(lines) {
			nextOffset = 0
		}
		return GetBuildLogResponse{
			LogLines:   strings.Join(lines[lw.Offset:end], "\n"),
			NextOffset: nextOffset,
			TotalLines: len(lines),
		}
	}
	// Read from the end:
	end := len(lines) + lw.Offset + 1
	start := end - lw.Limit
	if start < 0 {
		start = 0
	}
	nextOffset := lw.Offset - lw.Limit
	if nextOffset < -len(lines) {
		nextOffset = 0
	}
	return GetBuildLogResponse{
		LogLines:   strings.Join(lines[start:end], "\n"),
		NextOffset: nextOffset,
		TotalLines: len(lines),
	}
}
