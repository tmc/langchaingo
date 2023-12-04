package tei

import (
	"context"
	"time"

	client "github.com/gage-technologies/tei-go"
	"github.com/sourcegraph/conc/pool"
)

type TextEmbeddingsInference struct {
	client   *client.Client
	truncate bool
	baseURL  string
	headers  map[string]string
	cookies  map[string]string
	timeout  time.Duration
	poolSize int
}

func New(opts ...Option) (TextEmbeddingsInference, error) {
	emb, err := applyClientOptions(opts...)
	if err != nil {
		return emb, err
	}
	emb.client = client.NewClient(emb.baseURL, emb.headers, emb.cookies, emb.timeout)

	return emb, nil
}

// CreateEmbedding creates one vector embedding for each of the texts.
func (e TextEmbeddingsInference) CreateEmbedding(_ context.Context, inputTexts []string) ([][]float32, error) {
	p := pool.NewWithResults[[]float32]().
		WithMaxGoroutines(e.poolSize).
		WithErrors()
	for _, txt := range inputTexts {
		p.Go(func() ([]float32, error) {
			res, err := e.client.Embed(txt, e.truncate)
			return res[0], err
		})
	}
	return p.Wait()
}
