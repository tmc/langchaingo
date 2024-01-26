package llms

import (
	"reflect"
	"testing"

	"github.com/tmc/langchaingo/schema"
)

func TestTextParts(t *testing.T) {
	t.Parallel()
	type args struct {
		role  schema.ChatMessageType
		parts []string
	}
	tests := []struct {
		name string
		args args
		want MessageContent
	}{
		{"basics", args{schema.ChatMessageTypeHuman, []string{"a", "b", "c"}}, MessageContent{
			Role: schema.ChatMessageTypeHuman,
			Parts: []ContentPart{
				TextContent{Text: "a"},
				TextContent{Text: "b"},
				TextContent{Text: "c"},
			},
		}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := TextParts(tt.args.role, tt.args.parts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TextParts() = %v, want %v", got, tt.want)
			}
		})
	}
}
