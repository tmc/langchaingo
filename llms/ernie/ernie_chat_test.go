package ernie

import (
	"reflect"
	"testing"

	"github.com/tmc/langchaingo/llms/ernie/internal/ernieclient"

	"github.com/tmc/langchaingo/schema"
)

func TestNewChat(t *testing.T) {
	t.Parallel()
	type args struct {
		opts []Option
	}
	tests := []struct {
		name    string
		args    args
		want    *Chat
		wantErr bool
	}{
		{name: "", args: args{opts: []Option{
			WithModelName(ModelNameERNIEBot),
			WithAKSK("ak", "sk"),
		}}, want: nil, wantErr: true},
		{name: "", args: args{opts: []Option{
			WithModelName(ModelNameERNIEBot),
			WithAKSK("ak", "sk"),
			WithAccessToken("xxx"),
		}}, want: nil, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewChat(tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewChat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			expectedType := reflect.TypeOf(tt.want)
			if reflect.TypeOf(got) == expectedType {
				t.Errorf("NewChat() got = %T, want %T", got, tt.want)
			}
		})
	}
}

func TestGetSystem(t *testing.T) {
	t.Parallel()
	type args struct {
		messages []schema.ChatMessage
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "system message exists",
			args: args{
				messages: []schema.ChatMessage{
					schema.SystemChatMessage{Content: "you are a robot."},
					schema.HumanChatMessage{Content: "who are you?"},
				},
			},
			want: "you are a robot.",
		},
		{
			name: "no system message",
			args: args{
				messages: []schema.ChatMessage{
					schema.HumanChatMessage{Content: "who are you?"},
				},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSystem(tt.args.messages); got != tt.want {
				t.Errorf("getSystem() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessagesToClientMessages(t *testing.T) {
	t.Parallel()
	type args struct {
		messages []schema.ChatMessage
	}
	tests := []struct {
		name string
		args args
		want []*ernieclient.ChatMessage
	}{
		{
			name: "Test_MessagesToClientMessages_OK",
			args: args{messages: []schema.ChatMessage{
				schema.AIChatMessage{Content: "assistant"},
				schema.HumanChatMessage{Content: "user"},
				schema.SystemChatMessage{Content: ""},
				schema.FunctionChatMessage{Content: "function"},
				schema.GenericChatMessage{Content: "user"},
			}},
			want: []*ernieclient.ChatMessage{
				{Content: "assistant", Role: "assistant"},
				{Content: "user", Role: "user"},
				{Content: "function", Role: "function"},
				{Content: "user", Role: "user"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := messagesToClientMessages(tt.args.messages)
			for i, v := range got {
				if !reflect.DeepEqual(v.Content, tt.want[i].Content) {
					t.Errorf("messagesToClientMessages() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
