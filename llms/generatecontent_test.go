package llms

import (
	"reflect"
	"testing"
)

func TestTextParts(t *testing.T) {
	t.Parallel()
	type args struct {
		role  ChatMessageType
		parts []string
	}
	tests := []struct {
		name string
		args args
		want MessageContent
	}{
		{"basics", args{ChatMessageTypeHuman, []string{"a", "b", "c"}}, MessageContent{
			Role: ChatMessageTypeHuman,
			Parts: []ContentPart{
				TextContent{Text: "a"},
				TextContent{Text: "b"},
				TextContent{Text: "c"},
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := TextParts(tt.args.role, false, tt.args.parts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TextParts() = %v, want %v", got, tt.want)
			}
		})
	}
}
