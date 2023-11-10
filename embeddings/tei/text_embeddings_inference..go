package tei

import (
	"context"
	"strings"
	"time"

	client "github.com/gage-technologies/tei-go"
	"github.com/sourcegraph/conc/pool"
	"github.com/tmc/langchaingo/embeddings"
)

type TextEmbeddingsInference struct {
	client        *client.Client
	StripNewLines bool
	BatchSize     int
	baseURL       string
	headers       map[string]string
	cookies       map[string]string
	timeout       time.Duration
	poolSize      int
}

var _ embeddings.Embedder = TextEmbeddingsInference{}

func New(opts ...Option) (TextEmbeddingsInference, error) {
	emb, err := applyClientOptions(opts...)
	if err != nil {
		return emb, err
	}
	emb.client = client.NewClient(emb.baseURL, emb.headers, emb.cookies, emb.timeout)

	return emb, nil
}

// EmbedDocuments creates one vector embedding for each of the texts.
func (e TextEmbeddingsInference) EmbedDocuments(_ context.Context, texts []string) ([][]float32, error) {
	batchedTexts := embeddings.BatchTexts(
		embeddings.MaybeRemoveNewLines(texts, e.StripNewLines),
		e.BatchSize,
	)

	emb := make([][]float32, 0, len(texts))

	p := pool.New().WithMaxGoroutines(e.poolSize).WithErrors()

	for _, txt := range batchedTexts {
		p.Go(func() error {
			curTextEmbeddings, err := e.client.Embed(strings.Join(txt, " "), false)
			if err != nil {
				return err
			}

			textLengths := make([]int, 0, len(txt))
			for _, text := range txt {
				textLengths = append(textLengths, len(text))
			}

			combined, err := embeddings.CombineVectors(curTextEmbeddings, textLengths)
			if err != nil {
				return err
			}

			emb = append(emb, combined)

			return nil
		})
	}
	return emb, p.Wait()
}

// EmbedQuery embeds a single text.
func (e TextEmbeddingsInference) EmbedQuery(_ context.Context, text string) ([]float32, error) {
	if e.StripNewLines {
		text = strings.ReplaceAll(text, "\n", " ")
	}

	emb, err := e.client.Embed(text, false)
	if err != nil {
		return nil, err
	}

	return emb[0], nil
}
