package weaviate

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/auth"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/filters"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
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
	ErrInvalidResponse       = errors.New("invalid response")
	ErrInvalidScoreThreshold = errors.New(
		"score threshold must be between 0 and 1")
	ErrInvalidFilter = errors.New("invalid filter")
)

// Store is a wrapper around the weaviate client.
type Store struct {
	embedder embeddings.Embedder
	client   *weaviate.Client

	textKey      string
	nameSpaceKey string

	indexName string
	nameSpace string
	host      string
	scheme    string

	// optional
	apiKey *string
	// optional
	authConfig auth.Config
	// optional
	connectionClient *http.Client

	// optional
	queryAttrs []string
}

var _ vectorstores.VectorStore = Store{}

// New creates a new Store with options.
// When using weaviate,
// the properties in the Class of weaviate must have properties with the values set by textKey and nameSpaceKey.
func New(opts ...Option) (Store, error) {
	s, err := applyClientOptions(opts...)
	if err != nil {
		return Store{}, err
	}
	headers := make(map[string]string)
	if s.apiKey != nil {
		headers["Authorization"] = fmt.Sprintf("Bearer %s", *s.apiKey)
	}
	s.client = weaviate.New(weaviate.Config{
		Scheme:           s.scheme,
		Host:             s.host,
		Headers:          headers,
		AuthConfig:       s.authConfig,
		ConnectionClient: s.connectionClient,
	})

	return s, nil
}

func (s Store) AddDocuments(ctx context.Context, docs []schema.Document, options ...vectorstores.Option) error {
	opts := s.getOptions(options...)
	nameSpace := s.getNameSpace(opts)

	texts := make([]string, 0, len(docs))
	for _, doc := range docs {
		texts = append(texts, doc.PageContent)
	}

	vectors, err := s.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return err
	}

	if len(vectors) != len(docs) {
		return ErrEmbedderWrongNumberVectors
	}

	metadatas := make([]map[string]any, 0, len(docs))
	for i := 0; i < len(docs); i++ {
		metadata := make(map[string]any, len(docs[i].Metadata))
		for key, value := range docs[i].Metadata {
			metadata[key] = value
		}
		metadata[s.textKey] = texts[i]
		metadata[s.nameSpaceKey] = nameSpace

		metadatas = append(metadatas, metadata)
	}

	objects := make([]*models.Object, 0, len(docs))
	for i := range docs {
		objects = append(objects, &models.Object{
			Class:      s.indexName,
			ID:         strfmt.UUID(uuid.New().String()),
			Vector:     convertVector(vectors[i]),
			Properties: metadatas[i],
		})
	}
	if _, err := s.client.Batch().ObjectsBatcher().WithObjects(objects...).Do(ctx); err != nil {
		return err
	}
	return nil
}

func (s Store) SimilaritySearch(
	ctx context.Context,
	query string,
	numDocuments int,
	options ...vectorstores.Option,
) ([]schema.Document, error) {
	opts := s.getOptions(options...)
	nameSpace := s.getNameSpace(opts)
	scoreThreshold, err := s.getScoreThreshold(opts)
	if err != nil {
		return nil, err
	}
	filter := s.getFilters(opts)
	whereBuilder, err := s.createWhereBuilder(nameSpace, filter)
	if err != nil {
		return nil, err
	}

	vector, err := s.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	res, err := s.client.GraphQL().
		Get().
		WithNearVector(s.client.GraphQL().
			NearVectorArgBuilder().
			WithVector(convertVector(vector)).
			WithCertainty(scoreThreshold),
		).
		WithWhere(whereBuilder).
		WithClassName(s.indexName).
		WithLimit(numDocuments).
		WithFields(s.createFields()...).Do(ctx)
	if err != nil {
		return nil, err
	}
	return s.parseDocumentsByGraphQLResponse(res)
}

func (s Store) parseDocumentsByGraphQLResponse(res *models.GraphQLResponse) ([]schema.Document, error) {
	if len(res.Errors) > 0 {
		messages := make([]string, 0, len(res.Errors))
		for _, e := range res.Errors {
			messages = append(messages, e.Message)
		}
		return nil, fmt.Errorf("%w: %s", ErrInvalidResponse, strings.Join(messages, ", "))
	}

	data, ok := res.Data["Get"].(map[string]any)[s.indexName]
	if !ok || data == nil {
		return nil, ErrEmptyResponse
	}
	items, ok := data.([]any)
	if !ok || len(items) == 0 {
		return nil, ErrEmptyResponse
	}
	docs := make([]schema.Document, 0, len(items))
	for _, item := range items {
		itemMap, ok := item.(map[string]any)
		if !ok {
			return nil, ErrInvalidResponse
		}
		pageContent, ok := itemMap[s.textKey].(string)
		if !ok {
			return nil, ErrMissingTextKey
		}
		delete(itemMap, s.textKey)
		doc := schema.Document{
			PageContent: pageContent,
			Metadata:    itemMap,
		}
		docs = append(docs, doc)
	}
	return docs, nil
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
	f32 := float32(opts.ScoreThreshold)
	return f32, nil
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

func (s Store) createWhereBuilder(namespace string, filter any) (*filters.WhereBuilder, error) {
	if filter == nil {
		return filters.Where().WithPath([]string{s.nameSpaceKey}).WithOperator(filters.Equal).WithValueString(namespace), nil
	}

	whereFilter, ok := filter.(*filters.WhereBuilder)
	if !ok {
		return nil, ErrInvalidFilter
	}
	return filters.Where().WithOperator(filters.And).WithOperands([]*filters.WhereBuilder{
		filters.Where().WithPath([]string{s.nameSpaceKey}).WithOperator(filters.Equal).WithValueString(namespace),
		whereFilter,
	}), nil
}

func (s Store) createFields() []graphql.Field {
	fields := make([]graphql.Field, 0, len(s.queryAttrs))
	for _, attr := range s.queryAttrs {
		fields = append(fields, graphql.Field{
			Name: attr,
		})
	}
	fields = append(fields, graphql.Field{
		Name: "_additional",
		Fields: []graphql.Field{
			{Name: "certainty"},
		},
	})
	return fields
}

func convertVector(v []float64) []float32 {
	v32 := make([]float32, len(v))
	for i, f := range v {
		v32[i] = float32(f)
	}
	return v32
}
