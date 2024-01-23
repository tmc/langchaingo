package huggingface

const (
	tokenEnvVarName = "HUGGINGFACEHUB_API_TOKEN"
	defaultModel    = "gpt2"
	defaultURL      = "https://api-inference.huggingface.co"
)

type options struct {
	token string
	model string
	url   string
}

type Option func(*options)

// WithToken passes the HuggingFace API token to the client. If not set, the token
// is read from the HUGGINGFACEHUB_API_TOKEN environment variable.
func WithToken(token string) Option {
	return func(opts *options) {
		opts.token = token
	}
}

// WithModel passes the HuggingFace model to the client. If not set, then will be
// used default model.
func WithModel(model string) Option {
	return func(opts *options) {
		opts.model = model
	}
}

// WithUrl passes the HuggingFace url to the client. If not set, then will be
// used default url.
func WithUrl(url string) Option {
	return func(opts *options) {
		opts.url = url
	}
}
