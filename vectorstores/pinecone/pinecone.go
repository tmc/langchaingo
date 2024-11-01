package pinecone

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/pinecone-io/go-pinecone/pinecone"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	// ErrMissingTextKey is returned in SimilaritySearch if a vector
	// from the query is missing the text key.
	ErrMissingTextKey = errors.New("missing text key in vector metadata")
	// ErrEmbedderWrongNumberVectors is returned when if the embedder returns a number
	// of vectors that is not equal to the number of documents given.
	ErrEmbedderWrongNumberVectors = errors.New(
		"number of vectors from embedder does not match number of documents",
	)
	// ErrEmptyResponse is returned if the API gives an empty response.
	ErrEmptyResponse         = errors.New("empty response")
	ErrInvalidScoreThreshold = errors.New(
		"score threshold must be between 0 and 1")
)

// Store is a wrapper around the pinecone rest API and grpc client.
type Store struct {
	embedder embeddings.Embedder
	client   *pinecone.Client

	host      string
	apiKey    string
	textKey   string
	nameSpace string
}

// New creates a new Store with options. Options for WithAPIKey, WithHost and WithEmbedder must be set.
func New(opts ...Option) (Store, error) {
	s, err := applyClientOptions(opts...)
	if err != nil {
		return Store{}, err
	}

	s.client, err = pinecone.NewClient(pinecone.NewClientParams{ApiKey: s.apiKey})
	if err != nil {
		return Store{}, err
	}

	return s, nil
}

// AddDocuments creates vector embeddings from the documents using the embedder
// and upsert the vectors to the pinecone index and returns the ids of the added documents.
func (s Store) AddDocuments(ctx context.Context,
	docs []schema.Document,
	options ...vectorstores.Option,
) ([]string, error) {
	opts := s.getOptions(options...)

	nameSpace := s.getNameSpace(opts)

	indexConn, err := s.client.IndexWithNamespace(s.host, nameSpace)
	if err != nil {
		return nil, err
	}
	defer indexConn.Close()

	texts := make([]string, 0, len(docs))
	for _, doc := range docs {
		texts = append(texts, doc.PageContent)
	}

	vectors, err := s.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return nil, err
	}

	if len(vectors) != len(docs) {
		return nil, ErrEmbedderWrongNumberVectors
	}

	metadatas := make([]map[string]any, 0, len(docs))
	for i := 0; i < len(docs); i++ {
		metadata := make(map[string]any, len(docs[i].Metadata))
		for key, value := range docs[i].Metadata {
			metadata[key] = value
		}
		metadata[s.textKey] = texts[i]

		metadatas = append(metadatas, metadata)
	}

	pineconeVectors := make([]*pinecone.Vector, 0, len(vectors))

	ids := make([]string, len(vectors))
	for i := 0; i < len(vectors); i++ {
		metadataStruct, err := structpb.NewStruct(metadatas[i])
		if err != nil {
			return nil, err
		}

		id := opts.GenerateDocumentID(ctx, docs[i], ids)
		ids[i] = id
		pineconeVectors = append(
			pineconeVectors,
			&pinecone.Vector{
				Id:       id,
				Values:   vectors[i],
				Metadata: metadataStruct,
			},
		)
	}

	_, err = indexConn.UpsertVectors(&ctx, pineconeVectors)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

// SimilaritySearch creates a vector embedding from the query using the embedder
// and queries to find the most similar documents.
func (s Store) SimilaritySearch(ctx context.Context, query string, numDocuments int, options ...vectorstores.Option) ([]schema.Document, error) { //nolint:lll
	opts := s.getOptions(options...)

	nameSpace := s.getNameSpace(opts)
	indexConn, err := s.client.IndexWithNamespace(s.host, nameSpace)
	if err != nil {
		return nil, err
	}
	defer indexConn.Close()

	var protoFilterStruct *structpb.Struct
	filters := s.getFilters(opts)
	if filters != nil {
		protoFilterStruct, err = s.createProtoStructFilter(filters)
		if err != nil {
			return nil, err
		}
	}

	scoreThreshold, err := s.getScoreThreshold(opts)
	if err != nil {
		return nil, err
	}

	vector, err := s.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	queryResult, err := indexConn.QueryByVectorValues(
		&ctx,
		&pinecone.QueryByVectorValuesRequest{
			Vector:          vector,
			TopK:            uint32(numDocuments),
			Filter:          protoFilterStruct,
			IncludeMetadata: true,
			IncludeValues:   true,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(queryResult.Matches) == 0 {
		return nil, ErrEmptyResponse
	}

	return s.getDocumentsFromMatches(queryResult, scoreThreshold)
}

func (s Store) getDocumentsFromMatches(queryResult *pinecone.QueryVectorsResponse, scoreThreshold float32) ([]schema.Document, error) {
	resultDocuments := make([]schema.Document, 0)
	for _, match := range queryResult.Matches {
		metadata := match.Vector.Metadata.AsMap()
		pageContent, ok := metadata[s.textKey].(string)
		if !ok {
			return nil, ErrMissingTextKey
		}
		delete(metadata, s.textKey)

		doc := schema.Document{
			PageContent: pageContent,
			Metadata:    metadata,
			Score:       match.Score,
		}

		// If scoreThreshold is not 0, we only return matches with a score above the threshold.
		if scoreThreshold != 0 && match.Score >= scoreThreshold {
			resultDocuments = append(resultDocuments, doc)
		} else if scoreThreshold == 0 { // If scoreThreshold is 0, we return all matches.
			resultDocuments = append(resultDocuments, doc)
		}
	}
	return resultDocuments, nil
}

func (s Store) getNameSpace(opts vectorstores.Options) string {
	if opts.NameSpace != "" {
		return opts.NameSpace
	}
	return s.nameSpace
}

func (s Store) getScoreThreshold(opts vectorstores.Options) (float32, error) {
	if opts.ScoreThreshold < 0 || opts.ScoreThreshold > 1 {
		return 0, ErrInvalidScoreThreshold
	}
	return opts.ScoreThreshold, nil
}

func (s Store) getFilters(opts vectorstores.Options) any {
	if opts.Filters != nil {
		return opts.Filters
	}
	return nil
}

func (s Store) getOptions(options ...vectorstores.Option) vectorstores.Options {
	opts := vectorstores.Options{}
	for _, opt := range options {
		opt(&opts)
	}
	return opts
}

func (s Store) createProtoStructFilter(filter any) (*structpb.Struct, error) {
	filterBytes, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}

	var filterStruct structpb.Struct
	err = json.Unmarshal(filterBytes, &filterStruct)
	if err != nil {
		return nil, err
	}

	return &filterStruct, nil
}
