package vertexaichat

import (
	"github.com/tmc/langchaingo/llms/vertexai"
)

const (
	_defaultBatchSize     = 512
	_defaultStripNewLines = true
)

type ChatOption func(p *ChatVertexAI)

// WithClient is an option for providing the LLM client.
func WithClient(client vertexai.Chat) ChatOption {
	return func(p *ChatVertexAI) {
		p.client = &client
	}
}

// WithStripNewLines is an option for specifying the should it strip new lines.
func WithStripNewLines(stripNewLines bool) ChatOption {
	return func(p *ChatVertexAI) {
		p.StripNewLines = stripNewLines
	}
}

// WithBatchSize is an option for specifying the batch size.
func WithBatchSize(batchSize int) ChatOption {
	return func(p *ChatVertexAI) {
		p.BatchSize = batchSize
	}
}

func applyChatClientOptions(opts ...ChatOption) (ChatVertexAI, error) {
	o := &ChatVertexAI{
		StripNewLines: _defaultStripNewLines,
		BatchSize:     _defaultBatchSize,
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.client == nil {
		client, err := vertexai.NewChat()
		if err != nil {
			return ChatVertexAI{}, err
		}
		o.client = client
	}

	return *o, nil
}
