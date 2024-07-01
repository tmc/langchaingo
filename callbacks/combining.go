package callbacks

import (
	"context"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// CombiningHandler is a callback handler that combine multi callbacks.
type CombiningHandler struct {
	Callbacks []Handler
}

var _ Handler = CombiningHandler{}

func (l CombiningHandler) HandleText(ctx context.Context, text string) {
	for _, handle := range l.Callbacks {
		handle.HandleText(ctx, text)
	}
}

func (l CombiningHandler) HandleLLM(ctx context.Context, messages []llms.MessageContent, info llms.CallOptions, next func(ctx context.Context) (*llms.ContentResponse, error)) (*llms.ContentResponse, error) {
	type nextFn func(ctx context.Context) (*llms.ContentResponse, error)
	var nextFunc nextFn = next
	for i := len(l.Callbacks) - 1; i >= 0; i-- {
		// wrapper is a closure that wraps the next function with the current handler.
		wrapper := func(nextFunc nextFn, h Handler) nextFn {
			return func(ctx context.Context) (*llms.ContentResponse, error) {
				return h.HandleLLM(ctx, messages, info, nextFunc)
			}
		}
		nextFunc = wrapper(nextFunc, l.Callbacks[i])
	}
	return nextFunc(ctx)
}

func (l CombiningHandler) HandleChain(ctx context.Context, inputs map[string]any, info schema.ChainInfo, next func(ctx context.Context) (map[string]any, error)) (map[string]any, error) {
	type nextFn func(ctx context.Context) (map[string]any, error)
	var nextFunc nextFn = next
	for i := len(l.Callbacks) - 1; i >= 0; i-- {
		// wrapper is a closure that wraps the next function with the current handler.
		wrapper := func(nextFunc nextFn, h Handler) nextFn {
			return func(ctx context.Context) (map[string]any, error) {
				return h.HandleChain(ctx, inputs, info, nextFunc)
			}
		}
		nextFunc = wrapper(nextFunc, l.Callbacks[i])
	}
	return nextFunc(ctx)
}

func (l CombiningHandler) HandleLLMStart(ctx context.Context, prompts []string) {
	for _, handle := range l.Callbacks {
		handle.HandleLLMStart(ctx, prompts)
	}
}

func (l CombiningHandler) HandleLLMGenerateContentStart(ctx context.Context, ms []llms.MessageContent) {
	for _, handle := range l.Callbacks {
		handle.HandleLLMGenerateContentStart(ctx, ms)
	}
}

func (l CombiningHandler) HandleLLMGenerateContentEnd(ctx context.Context, res *llms.ContentResponse) {
	for _, handle := range l.Callbacks {
		handle.HandleLLMGenerateContentEnd(ctx, res)
	}
}

func (l CombiningHandler) HandleChainStart(ctx context.Context, inputs map[string]any) {
	for _, handle := range l.Callbacks {
		handle.HandleChainStart(ctx, inputs)
	}
}

func (l CombiningHandler) HandleChainEnd(ctx context.Context, outputs map[string]any) {
	for _, handle := range l.Callbacks {
		handle.HandleChainEnd(ctx, outputs)
	}
}

func (l CombiningHandler) HandleToolStart(ctx context.Context, input string) {
	for _, handle := range l.Callbacks {
		handle.HandleToolStart(ctx, input)
	}
}

func (l CombiningHandler) HandleToolEnd(ctx context.Context, output string) {
	for _, handle := range l.Callbacks {
		handle.HandleToolEnd(ctx, output)
	}
}

func (l CombiningHandler) HandleAgentAction(ctx context.Context, action schema.AgentAction) {
	for _, handle := range l.Callbacks {
		handle.HandleAgentAction(ctx, action)
	}
}

func (l CombiningHandler) HandleAgentFinish(ctx context.Context, finish schema.AgentFinish) {
	for _, handle := range l.Callbacks {
		handle.HandleAgentFinish(ctx, finish)
	}
}

func (l CombiningHandler) HandleRetrieverStart(ctx context.Context, query string) {
	for _, handle := range l.Callbacks {
		handle.HandleRetrieverStart(ctx, query)
	}
}

func (l CombiningHandler) HandleRetrieverEnd(ctx context.Context, query string, documents []schema.Document) {
	for _, handle := range l.Callbacks {
		handle.HandleRetrieverEnd(ctx, query, documents)
	}
}

func (l CombiningHandler) HandleStreamingFunc(ctx context.Context, chunk []byte) {
	for _, handle := range l.Callbacks {
		handle.HandleStreamingFunc(ctx, chunk)
	}
}

func (l CombiningHandler) HandleChainError(ctx context.Context, err error) {
	for _, handle := range l.Callbacks {
		handle.HandleChainError(ctx, err)
	}
}

func (l CombiningHandler) HandleLLMError(ctx context.Context, err error) {
	for _, handle := range l.Callbacks {
		handle.HandleLLMError(ctx, err)
	}
}

func (l CombiningHandler) HandleToolError(ctx context.Context, err error) {
	for _, handle := range l.Callbacks {
		handle.HandleToolError(ctx, err)
	}
}
