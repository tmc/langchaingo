package googleai

import (
	"context"
	"github.com/google/generative-ai-go/genai"
	"github.com/tmc/langchaingo/llms"
)

type PreSendingHook func(
	ctx context.Context,
	model *genai.GenerativeModel,
	meta PreSendingHookMetadata,
)

// PreSendingHookMetadata contains more metadata for the pre-sending hook.
type PreSendingHookMetadata struct {
	// Options contains the options used for the call.
	Options llms.CallOptions

	// History contains the history of the conversation.
	// It is empty if this is the first call of the conversation.
	History []*genai.Content

	// Parts contains the parts of the content that are being sent.
	Parts []genai.Part
}
