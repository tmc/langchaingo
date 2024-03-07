package fstring

import (
	"strings"
	"testing"
)

func TestFormat(t *testing.T) {
	t.Parallel()

	type args struct {
		format string
		values map[string]any
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr string
	}{
		{"1", args{"{", map[string]any{}}, "", "single '{' is not allowed"},
		{"2", args{"{{", map[string]any{}}, "{", ""},
		{"3", args{"}", map[string]any{}}, "", "single '}' is not allowed"},
		{"4", args{"}}", map[string]any{}}, "}", ""},
		{"4", args{"{}", map[string]any{}}, "", "empty expression not allowed"},
		{"4", args{"{val}", map[string]any{}}, "", "args not defined"},
		{"4", args{"a={val}", map[string]any{"val": 1}}, "a=1", ""},
		{"4", args{"a= {val}", map[string]any{"val": 1}}, "a= 1", ""},
		{"4", args{"a= { val }", map[string]any{"val": 1}}, "a= 1", ""},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Format(tt.args.format, tt.args.values)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("Format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Format() got = %v, want %v", got, tt.want)
			}
		})
	}
}
