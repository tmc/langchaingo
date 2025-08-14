package fastembed

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/0xDezzy/langchaingo/embeddings"
	fastembed "github.com/anush008/fastembed-go"
)

var _ embeddings.Embedder = &FastEmbed{}

type FastEmbed struct {
	flagEmbedding *fastembed.FlagEmbedding
	client        *http.Client

	Model         fastembed.EmbeddingModel
	MaxLength     int
	CacheDir      string
	BatchSize     int
	StripNewLines bool
	DocEmbedType  string
	Parallel      int

	ExecutionProviders   []string
	ShowDownloadProgress bool
}

func NewFastEmbed(opts ...Option) (*FastEmbed, error) {
	f, err := applyOptions(opts...)
	if err != nil {
		return nil, err
	}

	if err := f.validateOptions(); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	initOpts := &fastembed.InitOptions{
		Model:                f.Model,
		MaxLength:            f.MaxLength,
		CacheDir:             f.CacheDir,
		ExecutionProviders:   f.ExecutionProviders,
		ShowDownloadProgress: &f.ShowDownloadProgress,
	}

	flagEmbedding, err := fastembed.NewFlagEmbedding(initOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize FastEmbed model: %w", err)
	}

	f.flagEmbedding = flagEmbedding
	return f, nil
}

func (f *FastEmbed) Close() error {
	if f.flagEmbedding != nil {
		return f.flagEmbedding.Destroy()
	}
	return nil
}

func (f *FastEmbed) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	processedTexts := embeddings.MaybeRemoveNewLines(texts, f.StripNewLines)

	var embeddings [][]float32
	var err error

	if f.DocEmbedType == "passage" {
		embeddings, err = f.flagEmbedding.PassageEmbed(processedTexts, f.BatchSize)
	} else {
		embeddings, err = f.flagEmbedding.Embed(processedTexts, f.BatchSize)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to embed documents: %w", err)
	}

	return embeddings, nil
}

func (f *FastEmbed) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	if f.StripNewLines {
		text = strings.ReplaceAll(text, "\n", " ")
	}

	embedding, err := f.flagEmbedding.QueryEmbed(text)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	return embedding, nil
}

func (f *FastEmbed) validateOptions() error {
	if f.DocEmbedType != "default" && f.DocEmbedType != "passage" {
		return fmt.Errorf("invalid DocEmbedType: %s, must be 'default' or 'passage'", f.DocEmbedType)
	}

	if f.BatchSize <= 0 {
		return fmt.Errorf("BatchSize must be greater than 0, got: %d", f.BatchSize)
	}

	if f.MaxLength <= 0 {
		return fmt.Errorf("MaxLength must be greater than 0, got: %d", f.MaxLength)
	}

	supportedModels := fastembed.ListSupportedModels()
	for _, modelInfo := range supportedModels {
		if modelInfo.Model == f.Model {
			return nil
		}
	}

	return fmt.Errorf("unsupported model: %s", f.Model)
}
