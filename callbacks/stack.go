package callbacks

import (
	"context"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

type StackHandler struct {
	handlers []Handler
}

var _ Handler = (*StackHandler)(nil)

// NewStackHandler creates a new stack handler with the given handlers.
// The handlers will be called in the order they are provided.
//
// Example:
//
//	  h := NewStackHandler(
//		&SimpleHandler{},
//	    &LogHandler{},
//	    &MyCustomHandler{},
//	  )
//
//	  h.HandleText(ctx, "Hello, world!")
func NewStackHandler(handlers ...Handler) StackHandler {
	return StackHandler{handlers: handlers}
}

func (s *StackHandler) HandleText(ctx context.Context, text string) {
	for _, h := range s.handlers {
		h.HandleText(ctx, text)
	}
}

func (s *StackHandler) HandleLLMStart(ctx context.Context, prompts []string) {
	for _, h := range s.handlers {
		h.HandleLLMStart(ctx, prompts)
	}
}

func (s *StackHandler) HandleLLMGenerateContentStart(ctx context.Context, ms []llms.MessageContent) {
	for _, h := range s.handlers {
		h.HandleLLMGenerateContentStart(ctx, ms)
	}
}

func (s *StackHandler) HandleLLMGenerateContentEnd(ctx context.Context, res *llms.ContentResponse) {
	for _, h := range s.handlers {
		h.HandleLLMGenerateContentEnd(ctx, res)
	}
}

func (s *StackHandler) HandleLLMError(ctx context.Context, err error) {
	for _, h := range s.handlers {
		h.HandleLLMError(ctx, err)
	}
}

func (s *StackHandler) HandleChainStart(ctx context.Context, inputs map[string]any) {
	for _, h := range s.handlers {
		h.HandleChainStart(ctx, inputs)
	}
}

func (s *StackHandler) HandleChainEnd(ctx context.Context, outputs map[string]any) {
	for _, h := range s.handlers {
		h.HandleChainEnd(ctx, outputs)
	}
}

func (s *StackHandler) HandleChainError(ctx context.Context, err error) {
	for _, h := range s.handlers {
		h.HandleChainError(ctx, err)
	}
}

func (s *StackHandler) HandleToolStart(ctx context.Context, input string) {
	for _, h := range s.handlers {
		h.HandleToolStart(ctx, input)
	}
}

func (s *StackHandler) HandleToolEnd(ctx context.Context, output string) {
	for _, h := range s.handlers {
		h.HandleToolEnd(ctx, output)
	}
}

func (s *StackHandler) HandleToolError(ctx context.Context, err error) {
	for _, h := range s.handlers {
		h.HandleToolError(ctx, err)
	}
}

func (s *StackHandler) HandleAgentAction(ctx context.Context, action schema.AgentAction) {
	for _, h := range s.handlers {
		h.HandleAgentAction(ctx, action)
	}
}

func (s *StackHandler) HandleAgentFinish(ctx context.Context, finish schema.AgentFinish) {
	for _, h := range s.handlers {
		h.HandleAgentFinish(ctx, finish)
	}
}

func (s *StackHandler) HandleRetrieverStart(ctx context.Context, query string) {
	for _, h := range s.handlers {
		h.HandleRetrieverStart(ctx, query)
	}
}

func (s *StackHandler) HandleRetrieverEnd(ctx context.Context, query string, documents []schema.Document) {
	for _, h := range s.handlers {
		h.HandleRetrieverEnd(ctx, query, documents)
	}
}

func (s *StackHandler) HandleStreamingFunc(ctx context.Context, chunk []byte) {
	for _, h := range s.handlers {
		h.HandleStreamingFunc(ctx, chunk)
	}
}
