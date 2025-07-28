package llms_test

import (
	"testing"

	"github.com/0xDezzy/langchaingo/llms"
)

func TestGetBufferString(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name        string
		messages    []llms.ChatMessage
		humanPrefix string
		aiPrefix    string
		expected    string
		expectError bool
	}{
		{
			name:        "No messages",
			messages:    []llms.ChatMessage{},
			humanPrefix: "Human",
			aiPrefix:    "AI",
			expected:    "",
			expectError: false,
		},
		{
			name: "Mixed messages",
			messages: []llms.ChatMessage{
				llms.SystemChatMessage{Content: "Please be polite."},
				llms.HumanChatMessage{Content: "Hello, how are you?"},
				llms.AIChatMessage{Content: "I'm doing great!"},
				llms.GenericChatMessage{Role: "Moderator", Content: "Keep the conversation on topic."},
			},
			humanPrefix: "Human",
			aiPrefix:    "AI",
			expected:    "system: Please be polite.\nHuman: Hello, how are you?\nAI: I'm doing great!\nModerator: Keep the conversation on topic.", //nolint:lll
			expectError: false,
		},
		{
			name: "Unsupported message type",
			messages: []llms.ChatMessage{
				unsupportedChatMessage{},
			},
			humanPrefix: "Human",
			aiPrefix:    "AI",
			expected:    "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result, err := llms.GetBufferString(tc.messages, tc.humanPrefix, tc.aiPrefix)
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

func (m unsupportedChatMessage) GetType() llms.ChatMessageType { return "unsupported" }
func (m unsupportedChatMessage) GetContent() string            { return "Unsupported message" }
