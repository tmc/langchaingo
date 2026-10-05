package main

import "testing"

func TestRemoveBlankLineAfterBrace(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "blank after function brace",
			in:   "func f() {\n\n\tx := 1\n}\n",
			want: "func f() {\n\tx := 1\n}\n",
		},
		{
			name: "no blank line untouched",
			in:   "func f() {\n\tx := 1\n}\n",
			want: "func f() {\n\tx := 1\n}\n",
		},
		{
			name: "blank line not after brace preserved",
			in:   "func f() {\n\tx := 1\n\n\ty := 2\n}\n",
			want: "func f() {\n\tx := 1\n\n\ty := 2\n}\n",
		},
		{
			name: "only collapses first blank after brace",
			in:   "func f() {\n\n\n\tx := 1\n}\n",
			want: "func f() {\n\n\tx := 1\n}\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(removeBlankLineAfterBrace([]byte(tt.in)))
			if got != tt.want {
				t.Errorf("removeBlankLineAfterBrace(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
