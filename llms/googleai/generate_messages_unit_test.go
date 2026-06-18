package googleai

import (
	"context"
	"errors"
	"testing"

	"github.com/google/generative-ai-go/genai"
	"github.com/tmc/langchaingo/llms"
)

// TestGenerateFromMessagesSystemOnly verifies that a message list containing
// only system messages returns ErrNoRequestMessage instead of panicking with
// an index-out-of-range when selecting the request message from the history.
func TestGenerateFromMessagesSystemOnly(t *testing.T) {
	t.Parallel()

	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart("you are a helpful assistant")},
		},
	}

	model := &genai.GenerativeModel{}
	_, err := generateFromMessages(context.Background(), model, messages, &llms.CallOptions{})

	if !errors.Is(err, ErrNoRequestMessage) {
		t.Fatalf("expected ErrNoRequestMessage, got %v", err)
	}
}
