package llms_test

import (
	"strings"
	"testing"

	"github.com/vendasta/langchaingo/llms"
)

// mockPromptValue implements the PromptValue interface for testing
type mockPromptValue struct {
	stringValue string
	messages    []llms.ChatMessage
}

func (m mockPromptValue) String() string {
	return m.stringValue
}

func (m mockPromptValue) Messages() []llms.ChatMessage {
	return m.messages
}

func TestPromptValueInterface(t *testing.T) {
	// Test that our mock implements the interface
	var _ llms.PromptValue = mockPromptValue{}

	// Test String() method
	prompt := mockPromptValue{
		stringValue: "Hello, world!",
		messages: []llms.ChatMessage{
			llms.HumanChatMessage{Content: "Hello, world!"},
		},
	}

	if got := prompt.String(); got != "Hello, world!" {
		t.Errorf("String() = %v, want %v", got, "Hello, world!")
	}

	// Test Messages() method
	messages := prompt.Messages()
	if len(messages) != 1 {
		t.Errorf("Messages() returned %d messages, want 1", len(messages))
	}

	if humanMsg, ok := messages[0].(llms.HumanChatMessage); !ok {
		t.Error("Expected HumanChatMessage")
	} else if humanMsg.Content != "Hello, world!" {
		t.Errorf("Message content = %v, want %v", humanMsg.Content, "Hello, world!")
	}
}

// Test multiple prompt values with different implementations
func TestMultiplePromptValues(t *testing.T) {
	tests := []struct {
		name     string
		prompt   llms.PromptValue
		wantStr  string
		wantMsgs int
	}{
		{
			name: "simple prompt",
			prompt: mockPromptValue{
				stringValue: "Simple text",
				messages: []llms.ChatMessage{
					llms.HumanChatMessage{Content: "Simple text"},
				},
			},
			wantStr:  "Simple text",
			wantMsgs: 1,
		},
		{
			name: "conversation prompt",
			prompt: mockPromptValue{
				stringValue: "User: Hello\nAssistant: Hi there!\nUser: How are you?",
				messages: []llms.ChatMessage{
					llms.HumanChatMessage{Content: "Hello"},
					llms.AIChatMessage{Content: "Hi there!"},
					llms.HumanChatMessage{Content: "How are you?"},
				},
			},
			wantStr:  "User: Hello\nAssistant: Hi there!\nUser: How are you?",
			wantMsgs: 3,
		},
		{
			name: "empty prompt",
			prompt: mockPromptValue{
				stringValue: "",
				messages:    []llms.ChatMessage{},
			},
			wantStr:  "",
			wantMsgs: 0,
		},
		{
			name: "system message prompt",
			prompt: mockPromptValue{
				stringValue: "System: You are a helpful assistant.\nUser: Hello",
				messages: []llms.ChatMessage{
					llms.SystemChatMessage{Content: "You are a helpful assistant."},
					llms.HumanChatMessage{Content: "Hello"},
				},
			},
			wantStr:  "System: You are a helpful assistant.\nUser: Hello",
			wantMsgs: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.prompt.String(); got != tt.wantStr {
				t.Errorf("String() = %v, want %v", got, tt.wantStr)
			}

			msgs := tt.prompt.Messages()
			if len(msgs) != tt.wantMsgs {
				t.Errorf("Messages() returned %d messages, want %d", len(msgs), tt.wantMsgs)
			}
		})
	}
}

// complexPromptValue demonstrates a more complex implementation
type complexPromptValue struct {
	template string
	values   map[string]string
}

func (c complexPromptValue) String() string {
	result := c.template
	for k, v := range c.values {
		result = strings.ReplaceAll(result, "{"+k+"}", v)
	}
	return result
}

func (c complexPromptValue) Messages() []llms.ChatMessage {
	content := c.String()
	return []llms.ChatMessage{
		llms.HumanChatMessage{Content: content},
	}
}

func TestComplexPromptValue(t *testing.T) {
	prompt := complexPromptValue{
		template: "Hello {name}, the weather is {weather} today.",
		values: map[string]string{
			"name":    "Alice",
			"weather": "sunny",
		},
	}

	// Verify it implements the interface
	var _ llms.PromptValue = prompt

	expectedStr := "Hello Alice, the weather is sunny today."
	if got := prompt.String(); got != expectedStr {
		t.Errorf("String() = %v, want %v", got, expectedStr)
	}

	msgs := prompt.Messages()
	if len(msgs) != 1 {
		t.Fatalf("Messages() returned %d messages, want 1", len(msgs))
	}

	if humanMsg, ok := msgs[0].(llms.HumanChatMessage); !ok {
		t.Error("Expected HumanChatMessage")
	} else if humanMsg.Content != expectedStr {
		t.Errorf("Message content = %v, want %v", humanMsg.Content, expectedStr)
	}
}
