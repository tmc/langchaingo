package embeddings

import (
	"math"
	"strings"

	"github.com/tmc/langchaingo/llms/openai"
)

type OpenAI struct {
	client        *openai.LLM
	StripNewLines bool
	BatchSize     int
}

func NewOpenAI() (OpenAI, error) {
	client, err := openai.New()
	if err != nil {
		return OpenAI{}, err
	}

	return OpenAI{
		client:        client,
		StripNewLines: true,
		BatchSize:     512,
	}, nil
}

func (e OpenAI) EmbedDocuments(texts []string) ([][]float64, error) {
	removedNewLines := make([]string, 0)

	for i := 0; i < len(texts); i++ {
		curText := texts[i]
		if e.StripNewLines {
			curText = strings.ReplaceAll(curText, "\n", " ")
		}

		removedNewLines = append(removedNewLines, curText)
	}

	subPrompts := chunkArray(removedNewLines, e.BatchSize)

	embeddings := make([][]float64, 0)
	for i := 0; i < len(subPrompts); i++ {
		curEmbeddings, err := e.client.CreateEmbedding(subPrompts[i])
		if err != nil {
			return [][]float64{}, err
		}

		embeddings = append(embeddings, curEmbeddings...)
	}

	return embeddings, nil
}

func (e OpenAI) EmbedQuery(text string) ([]float64, error) {
	if e.StripNewLines {
		text = strings.ReplaceAll(text, "\n", " ")
	}

	embeddings, err := e.client.CreateEmbedding([]string{text})
	if err != nil {
		return []float64{}, err
	}

	return embeddings[0], nil
}

func chunkArray[T any](arr []T, chunkSize int) [][]T {
	var chunks [][]T
	var chunkIndex int
	for i, elem := range arr {
		chunkIndex = int(math.Floor(float64(i) / float64(chunkSize)))
		if chunkIndex >= len(chunks) {
			chunks = append(chunks, []T{})
		}
		chunks[chunkIndex] = append(chunks[chunkIndex], elem)
	}
	return chunks
}
