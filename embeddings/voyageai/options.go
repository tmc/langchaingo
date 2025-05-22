package voyageai

import (
	"errors"
	"net/http"
	"os"
)

const (
	_defaultBaseURL       = "https://api.voyageai.com/v1"
	_defaultBatchSize     = 512
	_defaultStripNewLines = true
	_defaultModel         = "voyage-2"
)

// Option is a function type that can be used to modify the client.
type Option func(v *VoyageAI)

// WithModel is an option for providing the model name to use.
func WithModel(model string) Option {
	return func(v *VoyageAI) {
		v.Model = model
	}
}

// WithClient is an option for providing a custom http client.
func WithClient(client http.Client) Option {
	return func(v *VoyageAI) {
		v.client = &client
	}
}

// WithHTTPClient is an option for providing a custom http client pointer.
func WithHTTPClient(client *http.Client) Option {
	return func(v *VoyageAI) {
		v.client = client
	}
}

// WithToken is an option for providing the VoyageAI token.
func WithToken(token string) Option {
	return func(v *VoyageAI) {
		v.token = token
	}
}

// WithStripNewLines is an option for specifying the should it strip new lines.
func WithStripNewLines(stripNewLines bool) Option {
	return func(v *VoyageAI) {
		v.StripNewLines = stripNewLines
	}
}

// WithBatchSize is an option for specifying the batch size.
func WithBatchSize(batchSize int) Option {
	return func(v *VoyageAI) {
		v.BatchSize = batchSize
	}
}

func applyOptions(opts ...Option) (*VoyageAI, error) {
	o := &VoyageAI{
		baseURL:       _defaultBaseURL,
		Model:         _defaultModel,
		StripNewLines: _defaultStripNewLines,
		BatchSize:     _defaultBatchSize,
	}
	for _, opt := range opts {
		opt(o)
	}
	if o.client == nil {
		o.client = http.DefaultClient
	}
	if o.token == "" {
		token := os.Getenv("VOYAGEAI_API_KEY")
		if token != "" {
			o.token = token
		} else {
			return nil, errors.New("missing the VoyageAI API key, set it as VOYAGEAI_API_KEY environment variable")
		}
	}
	return o, nil
}
