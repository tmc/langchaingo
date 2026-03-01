package duckdb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	duckdb "github.com/duckdb/duckdb-go/v2"
	"github.com/google/uuid"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

var (
	ErrEmbedderWrongNumberVectors = errors.New("number of vectors from embedder does not match number of documents")
	ErrInvalidScoreThreshold      = errors.New("score threshold must be between 0 and 1")
	ErrInvalidFilters             = errors.New("invalid filters")
	ErrInvalidFilterKey           = errors.New("invalid filter key: must contain only letters, digits, underscores, and dots")
	ErrUnsupportedOptions         = errors.New("unsupported options")
)

// validFilterKeyRe matches safe JSON path keys (letters, digits, underscores, dots).
var validFilterKeyRe = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_.]*$`)

// Store is a wrapper around the DuckDB client.
type Store struct {
	embedder            embeddings.Embedder
	connURL             string
	connURLSet          bool
	db                  *sql.DB
	embeddingTableName  string
	collectionTableName string
	collectionName      string
	collectionUUID      string
	collectionMetadata  map[string]any
	preDeleteCollection bool
	vectorDimensions    int
	distanceMetric      string
}

var _ vectorstores.VectorStore = Store{}

// New creates a new Store with options.
func New(ctx context.Context, opts ...Option) (Store, error) {
	store, err := applyClientOptions(opts...)
	if err != nil {
		return Store{}, err
	}

	if store.db == nil {
		connector, cErr := duckdb.NewConnector(store.connURL, func(execer driver.ExecerContext) error {
			queries := []string{
				"INSTALL vss",
				"LOAD vss",
				"SET hnsw_enable_experimental_persistence = true",
			}
			for _, q := range queries {
				if _, qErr := execer.ExecContext(ctx, q, nil); qErr != nil {
					return qErr
				}
			}
			return nil
		})
		if cErr != nil {
			return Store{}, cErr
		}
		store.db = sql.OpenDB(connector)
	}

	if err = store.db.PingContext(ctx); err != nil {
		return Store{}, err
	}
	if err = (&store).init(ctx); err != nil {
		return Store{}, err
	}
	return store, nil
}

// Close closes the database connection.
func (s Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *Store) init(ctx context.Context) error {
	if err := s.createCollectionTableIfNotExists(ctx); err != nil {
		return err
	}
	if err := s.createEmbeddingTableIfNotExists(ctx); err != nil {
		return err
	}
	if s.preDeleteCollection {
		if err := s.RemoveCollection(ctx); err != nil {
			return err
		}
	}
	return s.createOrGetCollection(ctx)
}

func (s Store) createCollectionTableIfNotExists(ctx context.Context) error {
	//nolint:gosec
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
	uuid VARCHAR PRIMARY KEY,
	name VARCHAR UNIQUE,
	cmetadata JSON)`, s.collectionTableName)
	_, err := s.db.ExecContext(ctx, query)
	return err
}

func (s Store) createEmbeddingTableIfNotExists(ctx context.Context) error {
	//nolint:gosec
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
	uuid VARCHAR PRIMARY KEY,
	collection_id VARCHAR,
	document VARCHAR,
	cmetadata JSON,
	embedding FLOAT[%d])`, s.embeddingTableName, s.vectorDimensions)
	if _, err := s.db.ExecContext(ctx, query); err != nil {
		return err
	}

	// DuckDB does not support CREATE INDEX IF NOT EXISTS, so check the catalog.
	idxName := fmt.Sprintf("%s_embedding_hnsw", s.embeddingTableName)
	//nolint:gosec
	checkQuery := fmt.Sprintf(
		`SELECT COUNT(*) FROM duckdb_indexes() WHERE index_name = '%s'`, idxName)
	var count int
	if err := s.db.QueryRowContext(ctx, checkQuery).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		//nolint:gosec
		idxQuery := fmt.Sprintf(
			`CREATE INDEX %s ON %s USING HNSW (embedding) WITH (metric = '%s')`,
			idxName, s.embeddingTableName, s.distanceMetric)
		if _, err := s.db.ExecContext(ctx, idxQuery); err != nil {
			return err
		}
	}
	return nil
}

// RemoveCollection removes the collection from the collection table.
func (s Store) RemoveCollection(ctx context.Context) error {
	//nolint:gosec
	_, err := s.db.ExecContext(ctx,
		fmt.Sprintf(`DELETE FROM %s WHERE name = ?`, s.collectionTableName),
		s.collectionName)
	return err
}

func (s *Store) createOrGetCollection(ctx context.Context) error {
	jsonMetadata, err := json.Marshal(s.collectionMetadata)
	if err != nil {
		return err
	}

	// Try to get existing collection first.
	//nolint:gosec
	selectQuery := fmt.Sprintf(
		`SELECT uuid FROM %s WHERE name = ? LIMIT 1`, s.collectionTableName)
	err = s.db.QueryRowContext(ctx, selectQuery, s.collectionName).Scan(&s.collectionUUID)
	if err == nil {
		// Collection exists, update metadata.
		//nolint:gosec
		updateQuery := fmt.Sprintf(
			`UPDATE %s SET cmetadata = ? WHERE uuid = ?`, s.collectionTableName)
		_, err = s.db.ExecContext(ctx, updateQuery, string(jsonMetadata), s.collectionUUID)
		return err
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	// Collection doesn't exist, create it.
	s.collectionUUID = uuid.New().String()
	//nolint:gosec
	insertQuery := fmt.Sprintf(
		`INSERT INTO %s (uuid, name, cmetadata) VALUES (?, ?, ?)`, s.collectionTableName)
	_, err = s.db.ExecContext(ctx, insertQuery,
		s.collectionUUID, s.collectionName, string(jsonMetadata))
	return err
}

// AddDocuments adds documents to the DuckDB collection associated with 'Store'
// and returns the ids of the added documents.
func (s Store) AddDocuments(
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

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	//nolint:gosec
	query := fmt.Sprintf(`INSERT INTO %s
		(uuid, collection_id, document, cmetadata, embedding)
		VALUES (?, ?, ?, ?, ?)`, s.embeddingTableName)

	ids := make([]string, len(docs))
	for docIdx, doc := range docs {
		id := uuid.New().String()
		ids[docIdx] = id

		jsonMetadata, mErr := json.Marshal(doc.Metadata)
		if mErr != nil {
			return nil, mErr
		}

		if _, err = tx.ExecContext(ctx, query,
			id, s.collectionUUID, doc.PageContent,
			string(jsonMetadata), vectors[docIdx],
		); err != nil {
			return nil, err
		}
	}
	return ids, tx.Commit()
}

//nolint:cyclop
func (s Store) SimilaritySearch(
	ctx context.Context,
	query string,
	numDocuments int,
	options ...vectorstores.Option,
) ([]schema.Document, error) {
	opts := s.getOptions(options...)
	collectionName := s.getCollectionName(opts)
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
	queryVector, err := embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	whereClause, filterArgs, err := buildFilterWhereClause("data", scoreThreshold, filter)
	if err != nil {
		return nil, err
	}

	//nolint:gosec
	sqlQuery := fmt.Sprintf(`SELECT
	data.document,
	data.cmetadata,
	(1.0 - data.distance) AS score
FROM (
	SELECT
		e.document,
		e.cmetadata::VARCHAR AS cmetadata,
		%s(e.embedding, ?::FLOAT[%d]) AS distance
	FROM %s AS e
	JOIN %s AS c ON e.collection_id = c.uuid
	WHERE c.name = ?
) AS data
WHERE %s
ORDER BY data.distance
LIMIT ?`,
		s.distanceFunction(),
		s.vectorDimensions,
		s.embeddingTableName,
		s.collectionTableName,
		whereClause,
	)

	// Build args: queryVector, collectionName, ...filterValues, numDocuments.
	args := make([]any, 0, 3+len(filterArgs))
	args = append(args, queryVector, collectionName)
	args = append(args, filterArgs...)
	args = append(args, numDocuments)

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDocuments(rows, true)
}

func (s Store) Search(
	ctx context.Context,
	numDocuments int,
	options ...vectorstores.Option,
) ([]schema.Document, error) {
	opts := s.getOptions(options...)
	collectionName := s.getCollectionName(opts)
	filter, err := s.getFilters(opts)
	if err != nil {
		return nil, err
	}

	whereClause, filterArgs, err := buildFilterWhereClause("e", 0, filter)
	if err != nil {
		return nil, err
	}

	//nolint:gosec
	sqlQuery := fmt.Sprintf(`SELECT
	e.document,
	e.cmetadata::VARCHAR
FROM %s AS e
JOIN %s AS c ON e.collection_id = c.uuid
WHERE c.name = ? AND %s
LIMIT ?`,
		s.embeddingTableName,
		s.collectionTableName,
		whereClause,
	)

	// Build args: collectionName, ...filterValues, numDocuments.
	args := make([]any, 0, 2+len(filterArgs))
	args = append(args, collectionName)
	args = append(args, filterArgs...)
	args = append(args, numDocuments)

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDocuments(rows, false)
}

// DropTables drops the embedding and collection tables.
func (s Store) DropTables(ctx context.Context) error {
	//nolint:gosec
	if _, err := s.db.ExecContext(ctx,
		fmt.Sprintf(`DROP TABLE IF EXISTS %s`, s.embeddingTableName)); err != nil {
		return err
	}
	//nolint:gosec
	_, err := s.db.ExecContext(ctx,
		fmt.Sprintf(`DROP TABLE IF EXISTS %s`, s.collectionTableName))
	return err
}

// buildFilterWhereClause builds a parameterized WHERE clause from metadata filters.
// alias is the table/subquery prefix for cmetadata (e.g. "data" or "e").
func buildFilterWhereClause(
	alias string,
	scoreThreshold float32,
	filter map[string]any,
) (string, []any, error) {
	whereConditions := make([]string, 0)
	args := make([]any, 0)
	if scoreThreshold != 0 {
		whereConditions = append(whereConditions,
			fmt.Sprintf("%s.distance < %f", alias, 1-scoreThreshold))
	}
	for k, v := range filter {
		if !validFilterKeyRe.MatchString(k) {
			return "", nil, fmt.Errorf("%w: %q", ErrInvalidFilterKey, k)
		}
		whereConditions = append(whereConditions,
			fmt.Sprintf("json_extract_string(%s.cmetadata, '$.%s') = ?", alias, k))
		args = append(args, fmt.Sprint(v))
	}
	whereClause := "TRUE"
	if len(whereConditions) > 0 {
		whereClause = strings.Join(whereConditions, " AND ")
	}
	return whereClause, args, nil
}

// scanDocuments scans rows into a slice of schema.Document.
// If withScore is true, it expects a third column with the similarity score.
func scanDocuments(rows *sql.Rows, withScore bool) ([]schema.Document, error) {
	docs := make([]schema.Document, 0)
	for rows.Next() {
		var content string
		var metadataStr string
		var score float64

		var err error
		if withScore {
			err = rows.Scan(&content, &metadataStr, &score)
		} else {
			err = rows.Scan(&content, &metadataStr)
		}
		if err != nil {
			return nil, err
		}

		var metadataMap map[string]any
		if metadataStr != "" {
			if err := json.Unmarshal([]byte(metadataStr), &metadataMap); err != nil {
				return nil, err
			}
		}

		docs = append(docs, schema.Document{
			PageContent: content,
			Metadata:    metadataMap,
			Score:       float32(score),
		})
	}
	return docs, rows.Err()
}

func (s Store) distanceFunction() string {
	switch s.distanceMetric {
	case "l2sq":
		return "array_distance"
	case "ip":
		return "array_negative_inner_product"
	default:
		return "array_cosine_distance"
	}
}

func (s Store) getOptions(options ...vectorstores.Option) vectorstores.Options {
	opts := vectorstores.Options{}
	for _, opt := range options {
		opt(&opts)
	}
	return opts
}

func (s Store) getCollectionName(opts vectorstores.Options) string {
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

// getFilters return metadata filters, now only support map[key]value pattern.
func (s Store) getFilters(opts vectorstores.Options) (map[string]any, error) {
	if opts.Filters != nil {
		if filters, ok := opts.Filters.(map[string]any); ok {
			return filters, nil
		}
		return nil, ErrInvalidFilters
	}
	return map[string]any{}, nil
}

func (s Store) deduplicate(
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
