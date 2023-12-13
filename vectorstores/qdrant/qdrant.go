package qdrant

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	qdrant "github.com/qdrant/go-client/qdrant"
)

var (
	ErrEmbedderWrongNumberVectors = errors.New("number of vectors from embedder does not match number of documents")
	ErrInvalidScoreThreshold      = errors.New("score threshold must be between 0 and 1")
	ErrInvalidFilters             = errors.New("invalid filters")

	ErrUnsupportedOptions = errors.New("unsupported options")
)

// Store is a wrapper around the pgvector client.
type Store struct {
	embedder embeddings.Embedder

	collectionName string

	grpcAddr    string
	grpcOptions []grpc.DialOption
	grpcConn    *grpc.ClientConn

	qdrantClient      qdrant.QdrantClient
	collectionsClient qdrant.CollectionsClient
	pointsClient      qdrant.PointsClient
}

var _ vectorstores.VectorStore = Store{}

// New creates a new Store with options.
func New(ctx context.Context, opts ...Option) (Store, error) {
	s := Store{}
	for _, opt := range opts {
		opt(&s)
	}

	if s.grpcConn == nil {
		if len(s.grpcOptions) == 0 {
			s.grpcOptions = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
		}
		conn, err := grpc.DialContext(ctx, s.grpcAddr, s.grpcOptions...)
		if err != nil {
			return Store{}, fmt.Errorf("failed to connect to grpc: %w", err)
		}
		s.grpcConn = conn
	}
	if s.embedder == nil {
		return Store{}, errors.New("embedder must be set")
	}
	s.qdrantClient = qdrant.NewQdrantClient(s.grpcConn)
	s.collectionsClient = qdrant.NewCollectionsClient(s.grpcConn)
	s.pointsClient = qdrant.NewPointsClient(s.grpcConn)
	return s, nil
}

func (s Store) AddDocuments(ctx context.Context, docs []schema.Document, options ...vectorstores.Option) error {
	opts := s.getOptions(options...)
	if opts.ScoreThreshold != 0 || opts.Filters != nil || opts.NameSpace != "" {
		return ErrUnsupportedOptions
	}

	texts := make([]string, 0, len(docs))
	for _, doc := range docs {
		texts = append(texts, doc.PageContent)
	}

	embedder := s.embedder
	if opts.Embedder != nil {
		embedder = opts.Embedder
	}
	vectors, err := embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return err
	}

	if len(vectors) != len(docs) {
		return ErrEmbedderWrongNumberVectors
	}

	points := []*qdrant.PointStruct{}
	for docIdx := range docs {
		points = append(points, &qdrant.PointStruct{
			// Id
			// Payload
			Vectors: &qdrant.Vectors{
				VectorsOptions: &qdrant.Vectors_Vector{
					Vector: &qdrant.Vector{
						Data: vectors[docIdx],
						// Indices
					},
				},
			},
		})
	}

	resp, err := s.pointsClient.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: s.collectionName,
		Points:         points,
		Wait:           ref(true),
		// Ordering
		// ShardKeySelector
	})
	_ = resp
	return err
}

//nolint:cyclop
func (s Store) SimilaritySearch(
	ctx context.Context,
	query string,
	numDocuments int,
	options ...vectorstores.Option,
) ([]schema.Document, error) {
	opts := s.getOptions(options...)
	// collectionName := s.getNameSpace(opts)
	scoreThreshold, err := s.getScoreThreshold(opts)
	if err != nil {
		return nil, err
	}
	// filter, err := s.getFilters(opts)
	// if err != nil {
	// 	return nil, err
	// }
	embedder := s.embedder
	if opts.Embedder != nil {
		embedder = opts.Embedder
	}
	embedderData, err := embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}
	_ = embedderData
	whereQuerys := make([]string, 0)
	if scoreThreshold != 0 {
		whereQuerys = append(whereQuerys, fmt.Sprintf("data.distance < %f", 1-scoreThreshold))
	}
	// for k, v := range filter {
	// 	whereQuerys = append(whereQuerys, fmt.Sprintf("(data.cmetadata ->> '%s') = '%s'", k, v))
	// }
	whereQuery := strings.Join(whereQuerys, " AND ")
	if len(whereQuery) == 0 {
		whereQuery = "TRUE"
	}
	//rows, err := s.grpcConn.Query(ctx, sql, pgvector.NewVector(embedderData), numDocuments)
	if err != nil {
		return nil, err
	}
	docs := make([]schema.Document, 0)
	// for rows.Next() {
	// 	doc := schema.Document{}
	// 	if err := rows.Scan(&doc.PageContent, &doc.Metadata, &doc.Score); err != nil {
	// 		return nil, err
	// 	}
	// 	docs = append(docs, doc)
	// }
	return docs, nil
}

// Close closes the connection.
func (s Store) Close(ctx context.Context) error {
	return s.grpcConn.Close()
}

func (s Store) RemoveCollection(ctx context.Context) error {
	// _, err := s.grpcConn.Exec(ctx, fmt.Sprintf(`DELETE FROM %s WHERE name = $1`, s.collectionTableName), s.collectionName)
	_, err := s.collectionsClient.Delete(ctx, &qdrant.DeleteCollection{
		CollectionName: s.collectionName,
	})
	return err
}

// getOptions applies given options to default Options and returns it
// This uses options pattern so clients can easily pass options without changing function signature.
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

// ref is a generic that returns a pointer to a value:
func ref[T any](v T) *T {
	return &v
}
