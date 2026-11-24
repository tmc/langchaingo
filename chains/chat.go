package chains

import (
	"context"
	"errors"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/outputparser"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

const _chatChainDefaultOutputKey = "text"

type ChatChain struct {
	Prompt           prompts.MessageFormatter
	LLM              llms.Model
	Memory           schema.Memory
	CallbacksHandler callbacks.Handler
	OutputParser     schema.OutputParser[any]

	OutputKey string
}

var (
	_ Chain                  = &ChatChain{}
	_ callbacks.HandlerHaver = &ChatChain{}
)

// NewChatChain creates a new ChatChain with an LLM and a prompt.
func NewChatChain(llm llms.Model, prompt prompts.MessageFormatter, opts ...ChainCallOption) *ChatChain {
	opt := &chainCallOption{}
	for _, o := range opts {
		o(opt)
	}
	chain := &ChatChain{
		Prompt:           prompt,
		LLM:              llm,
		OutputParser:     outputparser.NewSimple(),
		Memory:           memory.NewSimple(),
		OutputKey:        _chatChainDefaultOutputKey,
		CallbacksHandler: opt.CallbackHandler,
	}

	return chain
}

// Call formats the messages with the input values, generates using the llm, and parses
// the output from the llm with the output parser. This function should not be called
// directly, use rather the Call or Run function if the prompt only requires one input
// value.
func (c ChatChain) Call(ctx context.Context, values map[string]any, options ...ChainCallOption) (map[string]any, error) {
	msgs, err := c.Prompt.FormatMessages(values)
	if err != nil {
		return nil, err
	}

	messageContent := make([]llms.MessageContent, len(msgs))
	for i, msg := range msgs {
		messageContent[i] = llms.TextParts(msg.GetType(), msg.GetContent())
	}

	resp, err := c.LLM.GenerateContent(ctx, messageContent, getLLMCallOptions(options...)...)
	if err != nil {
		return nil, err
	}

	choices := resp.Choices
	if len(choices) < 1 {
		return nil, errors.New("empty response from model")
	}
	c1 := choices[0]

	return map[string]any{c.OutputKey: c1.Content}, nil
}

// GetMemory returns the memory of the chain.
func (c ChatChain) GetMemory() schema.Memory {
	return c.Memory
}

func (c ChatChain) GetCallbackHandler() callbacks.Handler { //nolint:ireturn
	return c.CallbacksHandler
}

// GetInputKeys returns the expected input keys.
func (c ChatChain) GetInputKeys() []string {
	return append([]string{}, c.Prompt.GetInputVariables()...)
}

// GetOutputKeys returns the output keys the chain will return.
func (c ChatChain) GetOutputKeys() []string {
	return []string{c.OutputKey}
}
