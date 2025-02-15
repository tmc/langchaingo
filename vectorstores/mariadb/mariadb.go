package mariadb

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	// Importing the MySQL driver for its side-effects (registering mysql driver).
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

var (
	ErrEmbedderWrongNumberVectors = errors.New("number of vectors from embedder does not match number of documents")
	ErrUnsupportedOptions         = errors.New("unsupported options")
	ErrInvalidScoreThreshold      = errors.New("score threshold must be between 0 and 1")
	ErrInvalidDocumentsNumber     = errors.New("documents number must be > 0")
	ErrInvalidFilters             = errors.New("invalid filters")
)

type Store struct {
	embedder            embeddings.Embedder
	dsn                 string
	db                  *sql.DB
	embeddingTableName  string
	collectionTableName string
	vectorDimensions    int
	collectionName      string
	collectionUUID      string
	collectionMetadata  map[string]any
	preDeleteCollection bool
	hnswIndex           *HNSWIndex

	preparedUpdateStatement *sql.Stmt
}

var _ vectorstores.VectorStore = &Store{}

func New(ctx context.Context, opts ...Option) (*Store, error) {
	store, err := applyClientOptions(opts...)
	if err != nil {
		return nil, err
	}

	store.db, err = sql.Open("mysql", store.dsn)
	if err != nil {
		return nil, err
	}

	if err := store.db.PingContext(ctx); err != nil {
		return nil, err
	}

	if err := store.init(ctx); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *Store) SimilaritySearch(
	ctx context.Context,
	query string,
	numDocuments int,
	options ...vectorstores.Option,
) ([]schema.Document, error) {
	if numDocuments <= 0 {
		return nil, ErrInvalidDocumentsNumber
	}

	opts := s.getOptions(options...)
	scoreThreshold, err := s.getScoreThreshold(opts)
	if err != nil {
		return nil, err
	}

	filters, err := s.getFilters(opts)
	if err != nil {
		return nil, err
	}

	embedder := s.embedder
	if opts.Embedder != nil {
		embedder = opts.Embedder
	}

	embeddedVector, err := embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	havingStatement := ""
	if scoreThreshold != 0 {
		havingStatement = fmt.Sprintf("HAVING score >= %f", scoreThreshold)
	}

	filterStatement, filterValues := prepareFilterConditions(filters)

	// nolint:gosec
	// Table name is controlled by configuration in code, so it's safe.
	// Distance function can be only euclidean or cosine, so it's also safe.
	// SQL parameters cannot be used for table identifiers.
	sql := fmt.Sprintf(`
	SELECT document, metadata, (1 - vec_distance_%s(embedding, vec_fromtext(?))) as score
	FROM %s
	WHERE collection_id = ? %s
	%s
	ORDER BY score DESC
	LIMIT %d;
	`, s.hnswIndex.DistanceFunc, s.embeddingTableName, filterStatement, havingStatement, numDocuments)

	params := []any{vectorToString(embeddedVector), s.collectionUUID}
	params = append(params, filterValues...)

	rows, err := s.db.QueryContext(ctx, sql, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	docs := make([]schema.Document, 0)

	for rows.Next() {
		doc := schema.Document{}
		var metadata []byte
		if err := rows.Scan(&doc.PageContent, &metadata, &doc.Score); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(metadata, &doc.Metadata); err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}

	return docs, rows.Err()
}

// Close closes the DB connection.
func (s *Store) Close() error {
	if err := s.preparedUpdateStatement.Close(); err != nil {
		return err
	}
	return s.db.Close()
}

// RemoveCollection removes the existing collection.
func (s *Store) RemoveCollection(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, fmt.Sprintf(`DELETE FROM %s WHERE name = ?`, s.collectionTableName), s.collectionName)
	return err
}

// AddDocuments adds documents to the MariaDB collection associated with 'Store'.
// and returns the ids of the added documents.
func (s *Store) AddDocuments(
	ctx context.Context,
	docs []schema.Document,
	options ...vectorstores.Option,
) ([]string, error) {
	opts := s.getOptions(options...)
	if opts.ScoreThreshold != 0 || opts.Filters != nil || opts.NameSpace != "" {
		return nil, ErrUnsupportedOptions
	}
	docs = s.deduplicate(ctx, opts, docs)

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
		return nil, err
	}

	if len(vectors) != len(docs) {
		return nil, ErrEmbedderWrongNumberVectors
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("INSERT INTO %s (uuid, collection_id, metadata, document, embedding) VALUES ", s.embeddingTableName))
	values := make([]interface{}, 0, len(docs)*5)

	ids := make([]string, len(docs))
	for docIdx, doc := range docs {
		if docIdx > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("(?, ?, ?, ?, vec_fromtext(?))")

		id := uuid.New().String()
		ids[docIdx] = id
		metadataJSON, err := json.Marshal(doc.Metadata)
		if err != nil {
			return []string{}, err
		}
		values = append(values, id, s.collectionUUID, metadataJSON, doc.PageContent, vectorToString(vectors[docIdx]))
	}

	sql := sb.String()

	if _, err := s.db.ExecContext(ctx, sql, values...); err != nil {
		return []string{}, err
	}

	return ids, nil
}

// UpdateDocument updates a document by its id.
func (s *Store) UpdateDocument(
	ctx context.Context,
	id string,
	doc schema.Document,
	options ...vectorstores.Option,
) error {
	opts := s.getOptions(options...)
	embedder := s.embedder
	if opts.Embedder != nil {
		embedder = opts.Embedder
	}

	vectors, err := embedder.EmbedDocuments(ctx, []string{doc.PageContent})
	if err != nil {
		return err
	}

	metadataJSON, err := json.Marshal(doc.Metadata)
	if err != nil {
		return err
	}

	_, err = s.preparedUpdateStatement.ExecContext(
		ctx,
		doc.PageContent,
		metadataJSON,
		vectorToString(vectors[0]),
		id,
		s.collectionUUID,
	)

	return err
}

// Search searches for documents that match the given metadata filters.
func (s *Store) Search(
	ctx context.Context,
	filters map[string]any,
	numDocuments int,
) ([]schema.Document, error) {
	if numDocuments <= 0 {
		return nil, ErrInvalidDocumentsNumber
	}

	filterStatement, filterValues := prepareFilterConditions(filters)

	// nolint:gosec
	sql := fmt.Sprintf(`
    SELECT document, metadata
    FROM %s
    WHERE collection_id = ? %s
    LIMIT %d;
    `, s.embeddingTableName, filterStatement, numDocuments)

	params := append([]any{s.collectionUUID}, filterValues...)

	rows, err := s.db.QueryContext(ctx, sql, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	docs := make([]schema.Document, 0)

	for rows.Next() {
		doc := schema.Document{}
		var metadata []byte
		if err := rows.Scan(&doc.PageContent, &metadata); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(metadata, &doc.Metadata); err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}

	return docs, rows.Err()
}

// DeleteDocumentsByFilter deletes all documents that match the given metadata filters.
// The method returns the number of deleted records.
func (s *Store) DeleteDocumentsByFilter(ctx context.Context, filters map[string]any) (int64, error) {
	filterStatement, filterValues := prepareFilterConditions(filters)

	//nolint:gosec
	sqlStmt := fmt.Sprintf(`
		DELETE FROM %s 
		WHERE collection_id = ? %s;
	`, s.embeddingTableName, filterStatement)

	params := append([]any{s.collectionUUID}, filterValues...)

	res, err := s.db.ExecContext(ctx, sqlStmt, params...)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

func (s *Store) init(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := s.createCollectionTableIfNotExists(ctx, tx); err != nil {
		return err
	}

	if err := s.createEmbeddingTableIfNotExists(ctx, tx); err != nil {
		return err
	}

	if s.preDeleteCollection {
		if err := s.RemoveCollection(ctx, tx); err != nil {
			return err
		}
	}

	if err := s.createOrGetCollection(ctx, tx); err != nil {
		return err
	}
	if err := s.prepareUpdateStatement(ctx); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) createCollectionTableIfNotExists(ctx context.Context, tx *sql.Tx) error {
	sql := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		uuid UUID NOT NULL DEFAULT UUID_v7() PRIMARY KEY,
		name VARCHAR(256) UNIQUE,
		metadata JSON
	);
	`, s.collectionTableName)

	if _, err := tx.ExecContext(ctx, sql); err != nil {
		return err
	}

	return nil
}

func (s *Store) prepareUpdateStatement(ctx context.Context) error {
	var err error
	//nolint:gosec
	sqlStmt := fmt.Sprintf(`
        UPDATE %s 
        SET document = ?, metadata = ?, embedding = vec_fromtext(?)
        WHERE uuid = ? AND collection_id = ?;
    `, s.embeddingTableName)

	s.preparedUpdateStatement, err = s.db.PrepareContext(ctx, sqlStmt)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) createEmbeddingTableIfNotExists(ctx context.Context, tx *sql.Tx) error {
	// nolint:gosec
	// Table name is controlled by configuration in code, so it's safe.
	// SQL parameters (?) cannot be used for table identifiers.
	sql := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
	uuid UUID NOT NULL DEFAULT UUID_v7() PRIMARY KEY,
	collection_id UUID REFERENCES %s(uuid) ON DELETE CASCADE,
	metadata JSON,
	document TEXT,
	embedding VECTOR(%d) NOT NULL,
	VECTOR INDEX(embedding) M = %d DISTANCE = %s
	);
	`, s.embeddingTableName, s.collectionTableName, s.vectorDimensions, s.hnswIndex.M, s.hnswIndex.DistanceFunc)

	if _, err := tx.ExecContext(ctx, sql); err != nil {
		return err
	}

	return nil
}

func (s *Store) createOrGetCollection(ctx context.Context, tx *sql.Tx) error {
	// nolint:gosec
	// Table name is controlled by configuration in code, so it's safe.
	// SQL parameters cannot be used for table identifiers.
	sql := fmt.Sprintf(`INSERT INTO %s (uuid, name, metadata)
		VALUES(?, ?, ?) on DUPLICATE KEY UPDATE metadata = VALUES(metadata)`, s.collectionTableName)

	metadataJSON, err := json.Marshal(s.collectionMetadata)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, sql, uuid.New().String(), s.collectionName, metadataJSON); err != nil {
		return err
	}

	sql = fmt.Sprintf(`SELECT uuid FROM %s WHERE name = ? ORDER BY name LIMIT 1`, s.collectionTableName)
	return tx.QueryRowContext(ctx, sql, s.collectionName).Scan(&s.collectionUUID)
}

// vectorToString converts a slice of float32 to a string.
func vectorToString(vector []float32) string {
	strs := make([]string, len(vector))
	for i, v := range vector {
		strs[i] = fmt.Sprintf("%f", v)
	}
	return fmt.Sprintf("[%s]", strings.Join(strs, ","))
}

func prepareFilterConditions(filters map[string]any) (string, []any) {
	filterConditions := ""
	filterValues := make([]any, 0)
	if len(filters) > 0 {
		conditions := make([]string, 0, len(filters))

		for k, v := range filters {
			operator := "="
			if strings.Contains(k, " ") {
				parts := strings.SplitN(k, " ", 2)
				k = parts[0]
				operator = parts[1]
			}
			conditions = append(conditions, fmt.Sprintf("JSON_EXTRACT(metadata, '$.%s') %s ?", k, operator))
			filterValues = append(filterValues, fmt.Sprintf("%v", v))
		}
		filterConditions = " AND " + strings.Join(conditions, " AND ")
	}

	return filterConditions, filterValues
}

// getOptions applies given options to default Options and returns it
// This uses options pattern so clients can easily pass options without changing function signature.
func (s *Store) getOptions(options ...vectorstores.Option) vectorstores.Options {
	opts := vectorstores.Options{}
	for _, opt := range options {
		opt(&opts)
	}
	return opts
}

func (s *Store) getScoreThreshold(opts vectorstores.Options) (float32, error) {
	if opts.ScoreThreshold < 0 || opts.ScoreThreshold > 1 {
		return 0, ErrInvalidScoreThreshold
	}
	return opts.ScoreThreshold, nil
}

// getFilters return metadata filters.
func (s *Store) getFilters(opts vectorstores.Options) (map[string]any, error) {
	if opts.Filters != nil {
		if filters, ok := opts.Filters.(map[string]any); ok {
			return filters, nil
		}
		return nil, ErrInvalidFilters
	}
	return map[string]any{}, nil
}

func (s *Store) deduplicate(
	ctx context.Context,
	opts vectorstores.Options,
	docs []schema.Document,
) []schema.Document {
	if opts.Deduplicater == nil {
		return docs
	}

	filtered := make([]schema.Document, 0, len(docs))
	for _, doc := range docs {
		if !opts.Deduplicater(ctx, doc) {
			filtered = append(filtered, doc)
		}
	}

	return filtered
}
