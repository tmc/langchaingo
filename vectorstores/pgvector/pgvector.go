package pgvector

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pgvector/pgvector-go"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

const (
	// pgLockIDEmbeddingTable is used for advisor lock to fix issue arising from concurrent
	// creation of the embedding table.The same value represents the same lock.
	pgLockIDEmbeddingTable = 1573678846307946494
	// pgLockIDCollectionTable is used for advisor lock to fix issue arising from concurrent
	// creation of the collection table.The same value represents the same lock.
	pgLockIDCollectionTable = 1573678846307946495
	// pgLockIDExtension is used for advisor lock to fix issue arising from concurrent creation
	// of the vector extension. The value is deliberately set to the same as python langchain
	// https://github.com/langchain-ai/langchain/blob/v0.0.340/libs/langchain/langchain/vectorstores/pgvector.py#L167
	pgLockIDExtension = 1573678846307946496
)

var (
	ErrEmbedderWrongNumberVectors = errors.New("number of vectors from embedder does not match number of documents")
	ErrInvalidScoreThreshold      = errors.New("score threshold must be between 0 and 1")
	ErrInvalidFilters             = errors.New("invalid filters")
	ErrUnsupportedOptions         = errors.New("unsupported options")
)

// Store is a wrapper around the pgvector client.
type Store struct {
	embedder              embeddings.Embedder
	conn                  *pgx.Conn
	postgresConnectionURL string
	embeddingTableName    string
	collectionTableName   string
	collectionName        string
	collectionUUID        string
	collectionMetadata    map[string]any
	preDeleteCollection   bool
}

var _ vectorstores.VectorStore = Store{}

// New creates a new Store with options.
func New(ctx context.Context, opts ...Option) (Store, error) {
	store, err := applyClientOptions(opts...)
	if err != nil {
		return Store{}, err
	}
	store.conn, err = pgx.Connect(ctx, store.postgresConnectionURL)
	if err != nil {
		return Store{}, err
	}

	if err = store.conn.Ping(ctx); err != nil {
		return Store{}, err
	}

	if err = store.createVectorExtensionIfNotExists(ctx); err != nil {
		return Store{}, err
	}
	if err = store.createCollectionTableIfNotExists(ctx); err != nil {
		return Store{}, err
	}
	if err = store.createEmbeddingTableIfNotExists(ctx); err != nil {
		return Store{}, err
	}
	if store.preDeleteCollection {
		if err = store.RemoveCollection(ctx); err != nil {
			return Store{}, err
		}
	}
	if err = store.createOrGetCollection(ctx); err != nil {
		return Store{}, err
	}
	return store, nil
}

func (s Store) createVectorExtensionIfNotExists(ctx context.Context) error {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	// inspired by
	// https://github.com/langchain-ai/langchain/blob/v0.0.340/libs/langchain/langchain/vectorstores/pgvector.py#L167
	// The advisor lock fixes issue arising from concurrent
	// creation of the vector extension.
	// https://github.com/langchain-ai/langchain/issues/12933
	// For more information see:
	// https://www.postgresql.org/docs/16/explicit-locking.html#ADVISORY-LOCKS
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", pgLockIDExtension); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS vector"); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s Store) createCollectionTableIfNotExists(ctx context.Context) error {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	// inspired by
	// https://github.com/langchain-ai/langchain/blob/v0.0.340/libs/langchain/langchain/vectorstores/pgvector.py#L167
	// The advisor lock fixes issue arising from concurrent
	// creation of the vector extension.
	// https://github.com/langchain-ai/langchain/issues/12933
	// For more information see:
	// https://www.postgresql.org/docs/16/explicit-locking.html#ADVISORY-LOCKS
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", pgLockIDCollectionTable); err != nil {
		return err
	}
	sql := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
	name varchar,
	cmetadata json,
	"uuid" uuid NOT NULL,
	PRIMARY KEY (uuid))`, s.collectionTableName)
	if _, err = tx.Exec(ctx, sql); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s Store) createEmbeddingTableIfNotExists(ctx context.Context) error {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	// inspired by
	// https://github.com/langchain-ai/langchain/blob/v0.0.340/libs/langchain/langchain/vectorstores/pgvector.py#L167
	// The advisor lock fixes issue arising from concurrent
	// creation of the vector extension.
	// https://github.com/langchain-ai/langchain/issues/12933
	// For more information see:
	// https://www.postgresql.org/docs/16/explicit-locking.html#ADVISORY-LOCKS
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", pgLockIDEmbeddingTable); err != nil {
		return err
	}
	sql := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
	collection_id uuid,
	embedding vector,
	document varchar,
	cmetadata json,
	custom_id varchar,
	"uuid" uuid NOT NULL,
	CONSTRAINT langchain_pg_embedding_collection_id_fkey 
	FOREIGN KEY (collection_id) REFERENCES %s (uuid) ON DELETE CASCADE,
PRIMARY KEY (uuid))`, s.embeddingTableName, s.collectionTableName)
	if _, err = tx.Exec(ctx, sql); err != nil {
		return err
	}
	return tx.Commit(ctx)
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
	customID := uuid.New().String()
	b := &pgx.Batch{}
	sql := fmt.Sprintf(`INSERT INTO %s (uuid, document, embedding, cmetadata, custom_id, collection_id)
		VALUES($1, $2, $3, $4, $5, $6)`, s.embeddingTableName)
	for docIdx, doc := range docs {
		id := uuid.New().String()
		b.Queue(sql, id, doc.PageContent, pgvector.NewVector(vectors[docIdx]), doc.Metadata, customID, s.collectionUUID)
	}
	return s.conn.SendBatch(ctx, b).Close()
}

//nolint:cyclop
func (s Store) SimilaritySearch(
	ctx context.Context,
	query string,
	numDocuments int,
	options ...vectorstores.Option,
) ([]schema.Document, error) {
	opts := s.getOptions(options...)
	collectionName := s.getNameSpace(opts)
	scoreThreshold, err := s.getScoreThreshold(opts)
	if err != nil {
		return nil, err
	}
	filter, err := s.getFilters(opts)
	if err != nil {
		return nil, err
	}
	embedder := s.embedder
	if opts.Embedder != nil {
		embedder = opts.Embedder
	}
	embedderData, err := embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}
	whereQuerys := make([]string, 0)
	if scoreThreshold != 0 {
		whereQuerys = append(whereQuerys, fmt.Sprintf("data.distance < %f", 1-scoreThreshold))
	}
	for k, v := range filter {
		whereQuerys = append(whereQuerys, fmt.Sprintf("(data.cmetadata ->> '%s') = '%s'", k, v))
	}
	whereQuery := strings.Join(whereQuerys, " AND ")
	if len(whereQuery) == 0 {
		whereQuery = "TRUE"
	}
	sql := fmt.Sprintf(`SELECT
	data.document,
	data.cmetadata,
	data.distance
FROM (
	SELECT
		%s.*,
		embedding <=> $1 AS distance
	FROM
		%s
		JOIN %s ON %s.collection_id=%s.uuid WHERE %s.name='%s') AS data
WHERE %s 
ORDER BY
	data.distance
LIMIT $2`, s.embeddingTableName,
		s.embeddingTableName,
		s.collectionTableName, s.embeddingTableName, s.collectionTableName, s.collectionTableName, collectionName,
		whereQuery)
	rows, err := s.conn.Query(ctx, sql, pgvector.NewVector(embedderData), numDocuments)
	if err != nil {
		return nil, err
	}
	docs := make([]schema.Document, 0)
	for rows.Next() {
		doc := schema.Document{}
		if err := rows.Scan(&doc.PageContent, &doc.Metadata, &doc.Score); err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

// Close closes the connection.
func (s Store) Close(ctx context.Context) error {
	return s.conn.Close(ctx)
}

func (s Store) DropTables(ctx context.Context) error {
	if _, err := s.conn.Exec(ctx, fmt.Sprintf(`DROP TABLE IF EXISTS %s`, s.collectionTableName)); err != nil {
		return err
	}
	if _, err := s.conn.Exec(ctx, fmt.Sprintf(`DROP TABLE IF EXISTS %s`, s.embeddingTableName)); err != nil {
		return err
	}
	return nil
}

func (s Store) RemoveCollection(ctx context.Context) error {
	_, err := s.conn.Exec(ctx, fmt.Sprintf(`DELETE FROM %s WHERE name = $1`, s.collectionTableName), s.collectionName)
	return err
}

func (s *Store) createOrGetCollection(ctx context.Context) error {
	sql := fmt.Sprintf(`INSERT INTO %s (uuid, name, cmetadata)
		VALUES($1, $2, $3) ON CONFLICT DO NOTHING`, s.collectionTableName)
	if _, err := s.conn.Exec(ctx, sql, uuid.New().String(), s.collectionName, s.collectionMetadata); err != nil {
		return err
	}
	sql = fmt.Sprintf(`SELECT uuid FROM %s WHERE name = $1 ORDER BY name limit 1`, s.collectionTableName)
	if err := s.conn.QueryRow(ctx, sql, s.collectionName).Scan(&s.collectionUUID); err != nil {
		return err
	}
	return nil
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

func (s Store) getNameSpace(opts vectorstores.Options) string {
	if opts.NameSpace != "" {
		return opts.NameSpace
	}
	return s.collectionName
}

func (s Store) getScoreThreshold(opts vectorstores.Options) (float32, error) {
	if opts.ScoreThreshold < 0 || opts.ScoreThreshold > 1 {
		return 0, ErrInvalidScoreThreshold
	}
	return opts.ScoreThreshold, nil
}

// getFilters return metadata filters, now only support map[key]value pattern
// TODO: should support more types like {"key1": {"key2":"values2"}} or {"key": ["value1", "values2"]}.
func (s Store) getFilters(opts vectorstores.Options) (map[string]any, error) {
	if opts.Filters != nil {
		if filters, ok := opts.Filters.(map[string]any); ok {
			return filters, nil
		}
		return nil, ErrInvalidFilters
	}
	return map[string]any{}, nil
}
