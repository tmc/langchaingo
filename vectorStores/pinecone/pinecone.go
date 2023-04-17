package pinecone

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorStores/pinecone/internal/pineconeClient"
)

const pineconeEnvVrName = "PINECONE_API_KEY"

var ErrMissingToken = errors.New("missing the Pinecone API key, set it in the PINECONE_API_KEY environment variable")

type Pinecone struct {
	client     pineconeClient.Client
	embeddings embeddings.Embeddings
	nameSpace  string
	textKey    string
}

// Environment for project is found in the pinecone console. Index name must not be larger then 45 characters.
func NewPinecone(embeddings embeddings.Embeddings, environment, indexName, nameSpace string, dimensions int) (Pinecone, error) {
	token := os.Getenv(pineconeEnvVrName)
	if token == "" {
		return Pinecone{}, ErrMissingToken
	}

	p, err := pineconeClient.New(
		pineconeClient.WithApiKey(token),
		pineconeClient.WithEnvironment(environment),
		pineconeClient.WithIndexName(indexName),
		pineconeClient.WithDimensions(dimensions),
	)

	return Pinecone{
		client:     p,
		embeddings: embeddings,
		nameSpace:  nameSpace,
		textKey:    "text",
	}, err
}

// If the length of the documentIds slice is 0 uuids will be used as ids.
func (p Pinecone) AddDocuments(documents []schema.Document, documentIds []string) error {
	if len(documentIds) == 0 {
		for i := 0; i < len(documents); i++ {
			documentIds = append(documentIds, uuid.New().String())
		}
	}

	if len(documentIds) != len(documents) {
		return fmt.Errorf("Number of documents and number of document ids must match")
	}

	texts := make([]string, 0)
	for i := 0; i < len(documents); i++ {
		texts = append(texts, documents[i].PageContent)
	}

	vectorData, err := p.embeddings.EmbedDocuments(texts)
	if err != nil {
		return err
	}

	vectors := make([]pineconeClient.Vector, 0)
	for i := 0; i < len(vectorData); i++ {
		curMetadata := make(map[string]string, 0)
		for key, value := range documents[i].Metadata {
			curMetadata[key] = fmt.Sprintf("%s", value)
		}

		curMetadata[p.textKey] = documents[i].PageContent

		vectors = append(vectors, pineconeClient.Vector{
			Values:   vectorData[i],
			Metadata: curMetadata,
			ID:       documentIds[i],
		})
	}

	return p.client.Upsert(context.Background(), vectors, p.nameSpace)
}

func (p Pinecone) SimilaritySearch(query string, numDocuments int) ([]schema.Document, error) {
	vector, err := p.embeddings.EmbedQuery(query)
	if err != nil {
		return []schema.Document{}, err
	}

	queryResponse, err := p.client.Query(context.Background(), vector, numDocuments, p.nameSpace)
	if err != nil {
		return []schema.Document{}, err
	}

	resultDocuments := make([]schema.Document, 0)
	for _, match := range queryResponse.Matches {
		pageContent, ok := match.Metadata[p.textKey]
		if !ok {
			return []schema.Document{}, fmt.Errorf("Missing textKey %s in query response match", p.textKey)
		}

		metadata := make(map[string]any)
		for key, value := range match.Metadata {
			metadata[key] = value
		}

		resultDocuments = append(resultDocuments, schema.Document{
			PageContent: pageContent,
			Metadata:    metadata,
		})
	}

	return resultDocuments, nil
}

func (p Pinecone) ToRetriever(numDocs int) PineconeRetriever {
	return PineconeRetriever{
		p:       p,
		numDocs: numDocs,
	}
}

type PineconeRetriever struct {
	p       Pinecone
	numDocs int
}

func (r PineconeRetriever) GetRelevantDocuments(query string) ([]schema.Document, error) {
	return r.p.SimilaritySearch(query, r.numDocs)
}
