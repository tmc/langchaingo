package schema_test

import (
	"testing"

	"github.com/tmc/langchaingo/schema"
)

func TestGetBufferString(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name        string
		messages    []schema.ChatMessage
		humanPrefix string
		aiPrefix    string
		expected    string
		expectError bool
	}{
		{
			name:        "No messages",
			messages:    []schema.ChatMessage{},
			humanPrefix: "Human",
			aiPrefix:    "AI",
			expected:    "",
			expectError: false,
		},
		{
			name: "Mixed messages",
			messages: []schema.ChatMessage{
				schema.HumanChatMessage{Text: "Hello, how are you?"},
				schema.AIChatMessage{Text: "I'm doing great!"},
				schema.SystemChatMessage{Text: "Please be polite."},
				schema.GenericChatMessage{Role: "Moderator", Text: "Keep the conversation on topic."},
			},
			humanPrefix: "Human",
			aiPrefix:    "AI",
			expected: "Human: Hello, how are you?\nAI: I'm doing great!\n" +
				"System: Please be polite.\nModerator: Keep the conversation on topic.",
			expectError: false,
		},
		{
			name: "Unsupported message type",
			messages: []schema.ChatMessage{
				unsupportedChatMessage{},
			},
			humanPrefix: "Human",
			aiPrefix:    "AI",
			expected:    "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result, err := schema.GetBufferString(tc.messages, tc.humanPrefix, tc.aiPrefix)
			if (err != nil) != tc.expectError {
				t.Fatalf("expected error: %v, got: %v", tc.expectError, err)
			}

			if result != tc.expected {
				t.Errorf("expected: %q, got: %q", tc.expected, result)
			}
		})
	}
}

type unsupportedChatMessage struct{}

func (m unsupportedChatMessage) GetType() schema.ChatMessageType { return "unsupported" }
func (m unsupportedChatMessage) GetText() string                 { return "Unsupported message" }
