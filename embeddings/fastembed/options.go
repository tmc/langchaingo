package fastembed

import (
	"net/http"

	"github.com/0xDezzy/langchaingo/httputil"
	fastembed "github.com/anush008/fastembed-go"
)

const (
	defaultModel         = fastembed.BGESmallENV15
	defaultMaxLength     = 512
	defaultCacheDir      = "local_cache"
	defaultBatchSize     = 256
	defaultStripNewLines = true
	defaultDocEmbedType  = "default"
	defaultParallel      = 0
)

type Option func(*FastEmbed)

func WithModel(model fastembed.EmbeddingModel) Option {
	return func(f *FastEmbed) {
		f.Model = model
	}
}

func WithMaxLength(maxLength int) Option {
	return func(f *FastEmbed) {
		f.MaxLength = maxLength
	}
}

func WithCacheDir(cacheDir string) Option {
	return func(f *FastEmbed) {
		f.CacheDir = cacheDir
	}
}

func WithBatchSize(batchSize int) Option {
	return func(f *FastEmbed) {
		f.BatchSize = batchSize
	}
}

func WithStripNewLines(stripNewLines bool) Option {
	return func(f *FastEmbed) {
		f.StripNewLines = stripNewLines
	}
}

func WithDocEmbedType(docEmbedType string) Option {
	return func(f *FastEmbed) {
		f.DocEmbedType = docEmbedType
	}
}

func WithExecutionProviders(providers []string) Option {
	return func(f *FastEmbed) {
		f.ExecutionProviders = providers
	}
}

func WithParallel(parallel int) Option {
	return func(f *FastEmbed) {
		f.Parallel = parallel
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(f *FastEmbed) {
		f.client = client
	}
}

func WithShowDownloadProgress(show bool) Option {
	return func(f *FastEmbed) {
		f.ShowDownloadProgress = show
	}
}

func applyOptions(opts ...Option) (*FastEmbed, error) {
	f := &FastEmbed{
		Model:                defaultModel,
		MaxLength:            defaultMaxLength,
		CacheDir:             defaultCacheDir,
		BatchSize:            defaultBatchSize,
		StripNewLines:        defaultStripNewLines,
		DocEmbedType:         defaultDocEmbedType,
		Parallel:             defaultParallel,
		ShowDownloadProgress: true,
		client:               httputil.DefaultClient,
	}

	for _, opt := range opts {
		opt(f)
	}

	return f, nil
}
