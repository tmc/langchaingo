package openai_chat

import (
	"github.com/tmc/langchaingo/llms/openai"
)

const (
	_defaultBatchSize     = 512
	_defaultStripNewLines = true
)

type ChatOption func(p *ChatOpenAI)

// WithClient is an option for providing the LLM client.
func WithClient(client openai.Chat) ChatOption {
	return func(p *ChatOpenAI) {
		p.client = &client
	}
}

// WithStripNewLines is an option for specifying the should it strip new lines.
func WithStripNewLines(stripNewLines bool) ChatOption {
	return func(p *ChatOpenAI) {
		p.StripNewLines = stripNewLines
	}
}

// WithBatchSize is an option for specifying the batch size.
func WithBatchSize(batchSize int) ChatOption {
	return func(p *ChatOpenAI) {
		p.BatchSize = batchSize
	}
}

func applyChatClientOptions(opts ...ChatOption) (ChatOpenAI, error) {
	o := &ChatOpenAI{
		StripNewLines: _defaultStripNewLines,
		BatchSize:     _defaultBatchSize,
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.client == nil {
		client, err := openai.NewChat()
		if err != nil {
			return ChatOpenAI{}, err
		}
		o.client = client
	}

	return *o, nil
}
