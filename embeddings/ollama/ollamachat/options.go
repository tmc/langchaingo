package ollamachat

import (
	"github.com/tmc/langchaingo/llms/ollama"
)

const (
	_defaultBatchSize     = 512
	_defaultStripNewLines = true
)

type ChatOption func(p *ChatOllama)

// WithClient is an option for providing the LLM client.
func WithClient(client *ollama.Chat) ChatOption {
	return func(p *ChatOllama) {
		p.client = client
	}
}

// WithStripNewLines is an option for specifying the should it strip new lines.
func WithStripNewLines(stripNewLines bool) ChatOption {
	return func(p *ChatOllama) {
		p.StripNewLines = stripNewLines
	}
}

// WithBatchSize is an option for specifying the batch size.
func WithBatchSize(batchSize int) ChatOption {
	return func(p *ChatOllama) {
		p.BatchSize = batchSize
	}
}

func applyChatClientOptions(opts ...ChatOption) (ChatOllama, error) {
	o := &ChatOllama{
		StripNewLines: _defaultStripNewLines,
		BatchSize:     _defaultBatchSize,
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.client == nil {
		client, err := ollama.NewChat()
		if err != nil {
			return ChatOllama{}, err
		}
		o.client = client
	}

	return *o, nil
}
