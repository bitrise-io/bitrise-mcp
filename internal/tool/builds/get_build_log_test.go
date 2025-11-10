package builds

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogWindow_Peek(t *testing.T) {
	givenLog := "line1\nline2\nline3\nline4\nline5"
	cases := map[string]struct {
		given logWindow
		want  GetBuildLogResponse
	}{
		"read forward from start with limit less than total lines": {
			given: logWindow{
				Log:    givenLog,
				Offset: 0,
				Limit:  3,
			},
			want: GetBuildLogResponse{
				LogLines:   "line1\nline2\nline3",
				NextOffset: 3,
				TotalLines: 5,
			},
		},
		"continue reading": {
			given: logWindow{
				Log:    givenLog,
				Offset: 3,
				Limit:  3,
			},
			want: GetBuildLogResponse{
				LogLines:   "line4\nline5",
				TotalLines: 5,
			},
		},
		"offset is outside bounds": {
			given: logWindow{
				Log:    givenLog,
				Offset: 10,
				Limit:  3,
			},
			want: GetBuildLogResponse{
				LogLines:   "",
				TotalLines: 5,
			},
		},
		"read from end": {
			given: logWindow{
				Log:    givenLog,
				Offset: -1,
				Limit:  3,
			},
			want: GetBuildLogResponse{
				LogLines:   "line3\nline4\nline5",
				NextOffset: -4,
				TotalLines: 5,
			},
		},
		"continue reading from end": {
			given: logWindow{
				Log:    givenLog,
				Offset: -4,
				Limit:  3,
			},
			want: GetBuildLogResponse{
				LogLines:   "line1\nline2",
				TotalLines: 5,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := tc.given.Peek()
			assert.Equal(t, tc.want, got)
		})
	}
}
