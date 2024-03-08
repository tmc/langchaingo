package bedrock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

const (
	_defaultBatchSize     = 512
	_defaultStripNewLines = true
	_defaultModel         = ModelTitanEmbedG1
)

// Option is a function type that can be used to modify the client.
type Option func(p *Bedrock)

// WithStripNewLines is an option for specifying the should it strip new lines.
func WithStripNewLines(stripNewLines bool) Option {
	return func(p *Bedrock) {
		p.StripNewLines = stripNewLines
	}
}

// WithBatchSize is an option for specifying the batch size.
// Only applicable to Cohere provider.
func WithBatchSize(batchSize int) Option {
	return func(p *Bedrock) {
		p.BatchSize = batchSize
	}
}

// WithModel is an option for providing the model name to use.
func WithModel(model string) Option {
	return func(p *Bedrock) {
		p.ModelID = model
	}
}

// WithClient is an option for providing the Bedrock client.
func WithClient(client *bedrockruntime.Client) Option {
	return func(p *Bedrock) {
		p.client = client
	}
}

func applyOptions(opts ...Option) (*Bedrock, error) {
	o := &Bedrock{
		StripNewLines: _defaultStripNewLines,
		BatchSize:     _defaultBatchSize,
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.client == nil {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			return nil, err
		}
		o.client = bedrockruntime.NewFromConfig(cfg)
	}
	return o, nil
}
