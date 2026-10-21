package storereleases

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func requestWith(args map[string]any) mcp.CallToolRequest {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args
	return req
}

func TestOptionalBool(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]any
		wantValue bool
		wantOK    bool
		wantErr   bool
	}{
		{name: "absent", args: map[string]any{}},
		{name: "null", args: map[string]any{"flag": nil}},
		{name: "true", args: map[string]any{"flag": true}, wantValue: true, wantOK: true},
		{name: "false is sent, not dropped", args: map[string]any{"flag": false}, wantValue: false, wantOK: true},
		{name: "string true", args: map[string]any{"flag": "true"}, wantValue: true, wantOK: true},
		{name: "string false", args: map[string]any{"flag": "false"}, wantValue: false, wantOK: true},
		{name: "number", args: map[string]any{"flag": 1.0}, wantValue: true, wantOK: true},
		{name: "zero", args: map[string]any{"flag": 0.0}, wantValue: false, wantOK: true},
		{name: "unparsable string", args: map[string]any{"flag": "yes please"}, wantErr: true},
		{name: "object", args: map[string]any{"flag": map[string]any{}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, ok, err := optionalBool(requestWith(tt.args), "flag")
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if ok != tt.wantOK || value != tt.wantValue {
				t.Fatalf("got (%v, %v), want (%v, %v)", value, ok, tt.wantValue, tt.wantOK)
			}
		})
	}
}

func TestOptionalArray(t *testing.T) {
	if _, ok, err := optionalArray(requestWith(map[string]any{}), "items"); ok || err != nil {
		t.Fatalf("absent: got ok=%v err=%v", ok, err)
	}
	if _, ok, err := optionalArray(requestWith(map[string]any{"items": nil}), "items"); ok || err != nil {
		t.Fatalf("null: got ok=%v err=%v", ok, err)
	}
	if _, _, err := optionalArray(requestWith(map[string]any{"items": map[string]any{"a": 1}}), "items"); err == nil {
		t.Fatal("bare object: want error, got nil")
	}
	v, ok, err := optionalArray(requestWith(map[string]any{"items": []any{"a"}}), "items")
	if err != nil || !ok || len(v.([]any)) != 1 {
		t.Fatalf("array: got v=%v ok=%v err=%v", v, ok, err)
	}
}

func TestRequireArray(t *testing.T) {
	if _, err := requireArray(requestWith(map[string]any{}), "items"); err == nil {
		t.Fatal("absent: want error")
	}
	if _, err := requireArray(requestWith(map[string]any{"items": []any{}}), "items"); err == nil {
		t.Fatal("empty: want error")
	}
	if _, err := requireArray(requestWith(map[string]any{"items": "a"}), "items"); err == nil {
		t.Fatal("wrong type: want error")
	}
	if v, err := requireArray(requestWith(map[string]any{"items": []any{"a", "b"}}), "items"); err != nil || len(v.([]any)) != 2 {
		t.Fatalf("array: got v=%v err=%v", v, err)
	}
}

func TestOptionalString(t *testing.T) {
	if _, ok, err := optionalString(requestWith(map[string]any{}), "field"); ok || err != nil {
		t.Fatalf("absent: got ok=%v err=%v", ok, err)
	}
	if _, ok, err := optionalString(requestWith(map[string]any{"field": nil}), "field"); ok || err != nil {
		t.Fatalf("null: got ok=%v err=%v", ok, err)
	}
	if v, ok, err := optionalString(requestWith(map[string]any{"field": ""}), "field"); !ok || err != nil || v != "" {
		t.Fatalf("empty string must be forwarded to clear the field: got v=%q ok=%v err=%v", v, ok, err)
	}
	if v, ok, err := optionalString(requestWith(map[string]any{"field": "x"}), "field"); !ok || err != nil || v != "x" {
		t.Fatalf("value: got v=%q ok=%v err=%v", v, ok, err)
	}
	if _, _, err := optionalString(requestWith(map[string]any{"field": 1.0}), "field"); err == nil {
		t.Fatal("number: want error")
	}
}
