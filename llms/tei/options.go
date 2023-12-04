package tei

import (
	"errors"
	"os"
	"runtime"
	"time"

	client "github.com/gage-technologies/tei-go"
)

const (
	defaultTimeNanoSeconds = 60 * 1000000000
	defaultURLEnvVarName   = "TEI_API_URL"
)

var ErrMissingAPIBaseURL = errors.New("missing the API Base URL") //nolint:lll

type Option func(emb *TextEmbeddingsInference)

// WithPoolSize is an option for specifying the number of goroutines.
func WithPoolSize(poolSize int) Option {
	return func(p *TextEmbeddingsInference) {
		p.poolSize = poolSize
	}
}

// WithAPIBaseURL adds base url for api.
func WithAPIBaseURL(url string) Option {
	return func(emb *TextEmbeddingsInference) {
		emb.baseURL = url
	}
}

// WithHeaders add request headers.
func WithHeaders(headers map[string]string) Option {
	return func(emb *TextEmbeddingsInference) {
		if emb.headers == nil {
			emb.headers = make(map[string]string, len(headers))
		}
		for k, v := range headers {
			emb.headers[k] = v
		}
	}
}

// WithCookies add request cookies.
func WithCookies(cookies map[string]string) Option {
	return func(emb *TextEmbeddingsInference) {
		if emb.cookies == nil {
			emb.cookies = make(map[string]string, len(cookies))
		}
		for k, v := range cookies {
			emb.cookies[k] = v
		}
	}
}

// WithTimeout set the request timeout.
func WithTimeout(dur time.Duration) Option {
	return func(emb *TextEmbeddingsInference) {
		emb.timeout = dur
	}
}

// WithTruncate set the embedder to truncate input length.
func WithTruncate() Option {
	return func(emb *TextEmbeddingsInference) {
		emb.truncate = true
	}
}

func applyClientOptions(opts ...Option) (TextEmbeddingsInference, error) {
	emb := TextEmbeddingsInference{
		timeout:  time.Duration(defaultTimeNanoSeconds),
		poolSize: runtime.GOMAXPROCS(0),
		baseURL:  os.Getenv(defaultURLEnvVarName),
	}
	for _, opt := range opts {
		opt(&emb)
	}
	if emb.baseURL == "" {
		return emb, ErrMissingAPIBaseURL
	}
	if emb.client == nil {
		emb.client = client.NewClient(emb.baseURL, emb.headers, emb.cookies, emb.timeout)
	}
	return emb, nil
}
