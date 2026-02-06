package chroma

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	chromaembed "github.com/amikos-tech/chroma-go/pkg/embeddings"
	"github.com/google/uuid"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

var (
	ErrInvalidScoreThreshold    = errors.New("score threshold must be between 0 and 1")
	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
	ErrNewClient                = errors.New("error creating collection")
	ErrAddDocument              = errors.New("error adding document")
	ErrRemoveCollection         = errors.New("error resetting collection")
	ErrUnsupportedOptions       = errors.New("unsupported options")
)

// Store is a wrapper around the chromaGo API and client.
type Store struct {
	client           chroma.Client
	collection       chroma.Collection
	distanceFunction chromaembed.DistanceMetric
	chromaURL        string
	cloudAPIKey      string
	tenant           string
	database         string
	nameSpace        string
	nameSpaceKey     string
	embedder         embeddings.Embedder
	includes         []chroma.Include
}

var _ vectorstores.VectorStore = Store{}

// New creates an active client connection to the (specified, or default) collection in the Chroma server
// and returns the `Store` object needed by the other accessors.
func New(opts ...Option) (Store, error) {
	s, coErr := applyClientOptions(opts...)
	if coErr != nil {
		return s, coErr
	}

	if s.client == nil {
		var err error
		if s.cloudAPIKey != "" || os.Getenv(ChromaAPIKeyEnvVarName) != "" {
			s.client, err = createCloudClient(&s)
		} else {
			s.client, err = createHTTPClient(&s)
		}
		if err != nil {
			return s, err
		}
	}

	if errHb := s.client.Heartbeat(context.Background()); errHb != nil {
		return s, errHb
	}

	ef := chromaGoEmbedder{Embedder: s.embedder}

	col, errCc := s.client.GetOrCreateCollection(context.Background(), s.nameSpace,
		chroma.WithEmbeddingFunctionCreate(ef),
		chroma.WithHNSWSpaceCreate(s.distanceFunction),
	)
	if errCc != nil {
		return s, fmt.Errorf("%w: %w", ErrNewClient, errCc)
	}

	s.collection = col
	return s, nil
}

func createHTTPClient(s *Store) (chroma.Client, error) {
	return chroma.NewHTTPClient(chroma.WithBaseURL(s.chromaURL))
}

func createCloudClient(s *Store) (chroma.Client, error) {
	var opts []chroma.ClientOption
	apiKey := s.cloudAPIKey
	if apiKey == "" {
		apiKey = os.Getenv(ChromaAPIKeyEnvVarName)
	}
	if apiKey != "" {
		opts = append(opts, chroma.WithCloudAPIKey(apiKey))
	}
	if s.tenant != "" && s.database != "" {
		opts = append(opts, chroma.WithDatabaseAndTenant(s.database, s.tenant))
	}
	return chroma.NewCloudClient(opts...)
}

// Close releases client resources.
func (s Store) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

// Collection returns the underlying Chroma collection for advanced operations.
func (s Store) Collection() chroma.Collection {
	return s.collection
}

// AddDocuments adds the text and metadata from the documents to the Chroma collection associated with 'Store'.
// and returns the ids of the added documents.
func (s Store) AddDocuments(ctx context.Context,
	docs []schema.Document,
	options ...vectorstores.Option,
) ([]string, error) {
	opts := s.getOptions(options...)
	if opts.Embedder != nil || opts.ScoreThreshold != 0 || opts.Filters != nil {
		return nil, ErrUnsupportedOptions
	}

	nameSpace := s.getNameSpace(opts)
	if nameSpace != "" && s.nameSpaceKey == "" {
		return nil, fmt.Errorf("%w: nameSpace without nameSpaceKey", ErrUnsupportedOptions)
	}

	ids := make([]chroma.DocumentID, len(docs))
	strIDs := make([]string, len(docs))
	texts := make([]string, len(docs))
	metadatas := make([]chroma.DocumentMetadata, len(docs))

	for i, doc := range docs {
		id := uuid.New().String()
		ids[i] = chroma.DocumentID(id)
		strIDs[i] = id
		texts[i] = doc.PageContent
		mc := make(map[string]any)
		maps.Copy(mc, doc.Metadata)
		if nameSpace != "" {
			mc[s.nameSpaceKey] = nameSpace
		}
		dm, err := chroma.NewDocumentMetadataFromMap(mc)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrAddDocument, err)
		}
		metadatas[i] = dm
	}

	col := s.collection
	if addErr := col.Add(ctx,
		chroma.WithIDs(ids...),
		chroma.WithTexts(texts...),
		chroma.WithMetadatas(metadatas...),
	); addErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrAddDocument, addErr)
	}
	return strIDs, nil
}

func (s Store) SimilaritySearch(ctx context.Context, query string, numDocuments int,
	options ...vectorstores.Option,
) ([]schema.Document, error) {
	opts := s.getOptions(options...)

	if opts.Embedder != nil {
		return nil, fmt.Errorf("%w: Embedder", ErrUnsupportedOptions)
	}

	scoreThreshold, stErr := s.getScoreThreshold(opts)
	if stErr != nil {
		return nil, stErr
	}

	filter := s.getNamespacedFilter(opts)

	queryOpts := []chroma.CollectionQueryOption{
		chroma.WithQueryTexts(query),
		chroma.WithNResults(numDocuments),
	}

	if filter != nil {
		queryOpts = append(queryOpts, chroma.WithWhere(&rawWhereFilter{data: filter}))
	}

	if len(s.includes) > 0 {
		queryOpts = append(queryOpts, chroma.WithInclude(s.includes...))
	}

	qr, queryErr := s.collection.Query(ctx, queryOpts...)
	if queryErr != nil {
		return nil, queryErr
	}

	docGroups := qr.GetDocumentsGroups()
	metaGroups := qr.GetMetadatasGroups()
	distGroups := qr.GetDistancesGroups()

	if len(docGroups) != len(metaGroups) || len(metaGroups) != len(distGroups) {
		return nil, fmt.Errorf("%w: documents[%d], metadatas[%d], distances[%d]",
			ErrUnexpectedResponseLength, len(docGroups), len(metaGroups), len(distGroups))
	}

	var sDocs []schema.Document
	for g := range docGroups {
		for i := range docGroups[g] {
			dist := float64(distGroups[g][i])
			if score := 1.0 - dist; score >= float64(scoreThreshold) {
				sDocs = append(sDocs, schema.Document{
					Metadata:    documentMetadataToMap(metaGroups[g][i]),
					PageContent: docGroups[g][i].ContentString(),
					Score:       float32(score),
				})
			}
		}
	}

	return sDocs, nil
}

func (s Store) RemoveCollection() error {
	if s.client == nil || s.collection == nil {
		return fmt.Errorf("%w: no collection", ErrRemoveCollection)
	}
	name := s.collection.Name()
	if errDc := s.client.DeleteCollection(context.Background(), name); errDc != nil {
		return fmt.Errorf("%w(%s): %w", ErrRemoveCollection, name, errDc)
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

func (s Store) getScoreThreshold(opts vectorstores.Options) (float32, error) {
	if opts.ScoreThreshold < 0 || opts.ScoreThreshold > 1 {
		return 0, ErrInvalidScoreThreshold
	}
	return opts.ScoreThreshold, nil
}

func (s Store) getNameSpace(opts vectorstores.Options) string {
	if opts.NameSpace != "" {
		return opts.NameSpace
	}
	return s.nameSpace
}

func (s Store) getNamespacedFilter(opts vectorstores.Options) map[string]any {
	filter, _ := opts.Filters.(map[string]any)

	nameSpace := s.getNameSpace(opts)
	if nameSpace == "" || s.nameSpaceKey == "" {
		return filter
	}

	nameSpaceFilter := map[string]any{s.nameSpaceKey: nameSpace}
	if filter == nil {
		return nameSpaceFilter
	}

	return map[string]any{"$and": []map[string]any{nameSpaceFilter, filter}}
}

func (s Store) getNamespacedSearchFilter(filter map[string]any) map[string]any {
	if s.nameSpace == "" || s.nameSpaceKey == "" {
		return filter
	}

	nameSpaceFilter := map[string]any{s.nameSpaceKey: s.nameSpace}
	if filter == nil {
		return nameSpaceFilter
	}

	return map[string]any{"$and": []map[string]any{nameSpaceFilter, filter}}
}

// rawWhereFilter wraps a map[string]any to implement chroma.WhereFilter,
// allowing users to pass arbitrary filter maps through the vectorstores API.
type rawWhereFilter struct {
	data map[string]any
}

func (r *rawWhereFilter) String() string {
	b, _ := json.Marshal(r.data)
	return string(b)
}

func (r *rawWhereFilter) Validate() error { return nil }

func (r *rawWhereFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.data)
}

func (r *rawWhereFilter) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &r.data)
}

// rawWhereClause wraps a map[string]any to implement chroma.WhereClause
// for use with the Search API filtering.
type rawWhereClause struct {
	data map[string]any
}

func (r *rawWhereClause) Operator() chroma.WhereFilterOperator { return "" }
func (r *rawWhereClause) Key() string                          { return "" }
func (r *rawWhereClause) Operand() any                         { return r.data }

func (r *rawWhereClause) String() string {
	b, _ := json.Marshal(r.data)
	return string(b)
}

func (r *rawWhereClause) Validate() error              { return nil }
func (r *rawWhereClause) MarshalJSON() ([]byte, error) { return json.Marshal(r.data) }
func (r *rawWhereClause) UnmarshalJSON(b []byte) error { return json.Unmarshal(b, &r.data) }

func documentMetadataToMap(dm chroma.DocumentMetadata) map[string]any {
	if dm == nil {
		return nil
	}
	b, err := json.Marshal(dm)
	if err != nil {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}
	return m
}
