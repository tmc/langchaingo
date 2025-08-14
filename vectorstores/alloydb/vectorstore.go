package alloydb

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/0xDezzy/langchaingo/embeddings"
	"github.com/0xDezzy/langchaingo/schema"
	"github.com/0xDezzy/langchaingo/util/alloydbutil"
	"github.com/0xDezzy/langchaingo/vectorstores"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pgvector/pgvector-go"
)

const (
	defaultIndexNameSuffix = "langchainvectorindex"
)

type VectorStore struct {
	engine             alloydbutil.PostgresEngine
	embedder           embeddings.Embedder
	tableName          string
	schemaName         string
	idColumn           string
	metadataJSONColumn string
	contentColumn      string
	embeddingColumn    string
	metadataColumns    []string
	k                  int
	distanceStrategy   distanceStrategy
}

type BaseIndex struct {
	name             string
	indexType        string
	options          Index
	distanceStrategy distanceStrategy
	partialIndexes   []string
}

type SearchDocument struct {
	Content           string
	LangchainMetadata string
	Distance          float32
}

var _ vectorstores.VectorStore = &VectorStore{}

// NewVectorStore creates a new VectorStore with options.
func NewVectorStore(engine alloydbutil.PostgresEngine,
	embedder embeddings.Embedder,
	tableName string,
	opts ...VectorStoreOption,
) (VectorStore, error) {
	vs, err := applyAlloyDBVectorStoreOptions(engine, embedder, tableName, opts...)
	if err != nil {
		return VectorStore{}, err
	}
	return vs, nil
}

// AddDocuments adds documents to the Postgres collection, and returns the ids
// of the added documents.
func (vs *VectorStore) AddDocuments(ctx context.Context, docs []schema.Document, _ ...vectorstores.Option) ([]string, error) {
	texts := make([]string, 0, len(docs))
	for _, doc := range docs {
		texts = append(texts, doc.PageContent)
	}
	embeddings, err := vs.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("failed embed documents: %w", err)
	}
	// If no ids provided, generate them.
	ids := make([]string, len(texts))
	for i, doc := range docs {
		if val, ok := doc.Metadata["id"].(string); ok {
			ids[i] = val
		} else {
			ids[i] = uuid.New().String()
		}
	}
	// If no metadata provided, initialize with empty maps
	metadatas := make([]map[string]any, len(docs))
	for i := range docs {
		if docs[i].Metadata == nil {
			metadatas[i] = make(map[string]any)
		} else {
			metadatas[i] = docs[i].Metadata
		}
	}
	b := &pgx.Batch{}

	for i := range texts {
		id := ids[i]
		content := texts[i]
		embedding := pgvector.NewVector(embeddings[i]).String()
		metadata := metadatas[i]
		query, values, err := vs.generateAddDocumentsQuery(id, content, embedding, metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to generate query: %w", err)
		}
		b.Queue(query, values...)
	}

	batchResults := vs.engine.Pool.SendBatch(ctx, b)
	if err := batchResults.Close(); err != nil {
		return nil, fmt.Errorf("failed to execute batch: %w", err)
	}

	return ids, nil
}

func (vs *VectorStore) generateAddDocumentsQuery(id, content, embedding string, metadata map[string]any) (string, []any, error) {
	// Construct metadata column names if present
	metadataColNames := ""
	if len(vs.metadataColumns) > 0 {
		metadataColNames = ", " + strings.Join(vs.metadataColumns, ", ")
	}

	if vs.metadataJSONColumn != "" {
		metadataColNames += ", " + vs.metadataJSONColumn
	}

	insertStmt := fmt.Sprintf(`INSERT INTO %q.%q (%s, %s, %s%s)`,
		vs.schemaName, vs.tableName, vs.idColumn, vs.contentColumn, vs.embeddingColumn, metadataColNames)
	valuesStmt := "VALUES ($1, $2, $3"
	values := []any{id, content, embedding}

	// Add metadata
	for _, metadataColumn := range vs.metadataColumns {
		if val, ok := metadata[metadataColumn]; ok {
			valuesStmt += fmt.Sprintf(", $%d", len(values)+1)
			values = append(values, val)
		} else {
			valuesStmt += ", NULL"
		}
	}
	// Add JSON column and/or close statement
	if vs.metadataJSONColumn != "" {
		valuesStmt += fmt.Sprintf(", $%d", len(values)+1)
		metadataJSON, err := json.Marshal(metadata)
		if err != nil {
			return "", nil, fmt.Errorf("failed to transform metadata to json: %w", err)
		}
		values = append(values, metadataJSON)
	}
	valuesStmt += ")"
	query := insertStmt + valuesStmt
	return query, values, nil
}

// SimilaritySearch performs a similarity search on the database using the
// query vector.
func (vs *VectorStore) SimilaritySearch(ctx context.Context, query string, _ int, options ...vectorstores.Option) ([]schema.Document, error) {
	opts := applyOpts(options...)
	var documents []schema.Document
	embedding, err := vs.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed embed query: %w", err)
	}
	operator := vs.distanceStrategy.operator()
	searchFunction := vs.distanceStrategy.similaritySearchFunction()

	columns := []string{}
	columns = append(columns, vs.contentColumn)
	if vs.metadataJSONColumn != "" {
		columns = append(columns, vs.metadataJSONColumn)
	}
	columnNames := strings.Join(columns, `, `)
	whereClause := ""
	if opts.Filters != nil {
		whereClause = fmt.Sprintf("WHERE %s", opts.Filters)
	}
	vector := pgvector.NewVector(embedding)
	stmt := fmt.Sprintf(`
        SELECT %s, %s(%s, '%s') AS distance FROM "%s"."%s" %s ORDER BY %s %s '%s' LIMIT $1::int;`,
		columnNames, searchFunction, vs.embeddingColumn, vector.String(), vs.schemaName, vs.tableName, whereClause, vs.embeddingColumn, operator, vector.String())

	results, err := vs.executeSQLQuery(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("failed to execute sql query: %w", err)
	}
	documents, err = vs.processResultsToDocuments(results)
	if err != nil {
		return nil, fmt.Errorf("failed to process Results to Documents with Scores: %w", err)
	}
	return documents, nil
}

func (vs *VectorStore) executeSQLQuery(ctx context.Context, stmt string) ([]SearchDocument, error) {
	rows, err := vs.engine.Pool.Query(ctx, stmt, vs.k)
	if err != nil {
		return nil, fmt.Errorf("failed to execute similar search query: %w", err)
	}
	defer rows.Close()

	var results []SearchDocument
	for rows.Next() {
		doc := SearchDocument{}

		err = rows.Scan(&doc.Content, &doc.LangchainMetadata, &doc.Distance)
		if err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}
		results = append(results, doc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return results, nil
}

func (*VectorStore) processResultsToDocuments(results []SearchDocument) ([]schema.Document, error) {
	documents := make([]schema.Document, 0, len(results))
	for _, result := range results {
		mapMetadata := map[string]any{}
		err := json.Unmarshal([]byte(result.LangchainMetadata), &mapMetadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal langchain metadata: %w", err)
		}
		doc := schema.Document{
			PageContent: result.Content,
			Metadata:    mapMetadata,
			Score:       result.Distance,
		}
		documents = append(documents, doc)
	}
	return documents, nil
}

// ApplyVectorIndex creates an index in the table of the embeddings.
func (vs *VectorStore) ApplyVectorIndex(ctx context.Context, index BaseIndex, name string, concurrently bool) error {
	if index.indexType == "exactnearestneighbor" {
		return vs.DropVectorIndex(ctx, name)
	}
	function := index.distanceStrategy.searchFunction()
	if index.indexType == "ScaNN" {
		_, err := vs.engine.Pool.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS alloydb_scann")
		if err != nil {
			return fmt.Errorf("failed to create alloydb scann extension: %w", err)
		}
	}
	filter := ""
	if len(index.partialIndexes) > 0 {
		filter = fmt.Sprintf("WHERE %s", index.partialIndexes)
	}
	optsString := index.indexOptions()
	params := fmt.Sprintf("WITH %s", optsString)

	if name == "" {
		if index.name == "" {
			index.name = vs.tableName + defaultIndexNameSuffix
		}
		name = index.name
	}

	concurrentlyStr := ""
	if concurrently {
		concurrentlyStr = "CONCURRENTLY"
	}

	stmt := fmt.Sprintf(`CREATE INDEX %s %s ON "%s"."%s" USING %s (%s %s) %s %s`,
		concurrentlyStr, name, vs.schemaName, vs.tableName, index.indexType, vs.embeddingColumn, function, params, filter)

	_, err := vs.engine.Pool.Exec(ctx, stmt)
	if err != nil {
		return fmt.Errorf("failed to execute creation of index: %w", err)
	}

	return nil
}

// ReIndex recreates the index on the VectorStore.
func (vs *VectorStore) ReIndex(ctx context.Context) error {
	indexName := vs.tableName + defaultIndexNameSuffix
	return vs.ReIndexWithName(ctx, indexName)
}

// ReIndex recreates the index on the VectorStore by name.
func (vs *VectorStore) ReIndexWithName(ctx context.Context, indexName string) error {
	query := fmt.Sprintf("REINDEX INDEX %s;", indexName)
	_, err := vs.engine.Pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to reindex: %w", err)
	}

	return nil
}

// DropVectorIndex drops the vector index from the VectorStore.
func (vs *VectorStore) DropVectorIndex(ctx context.Context, indexName string) error {
	if indexName == "" {
		indexName = vs.tableName + defaultIndexNameSuffix
	}
	query := fmt.Sprintf("DROP INDEX IF EXISTS %s;", indexName)
	_, err := vs.engine.Pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to drop vector index: %w", err)
	}

	return nil
}

// IsValidIndex checks if index exists in the VectorStore.
func (vs *VectorStore) IsValidIndex(ctx context.Context, indexName string) (bool, error) {
	if indexName == "" {
		indexName = vs.tableName + defaultIndexNameSuffix
	}
	query := fmt.Sprintf("SELECT tablename, indexname  FROM pg_indexes WHERE tablename = '%s' AND schemaname = '%s' AND indexname = '%s';",
		vs.tableName, vs.schemaName, indexName)
	var tablename, indexnameFromDB string
	err := vs.engine.Pool.QueryRow(ctx, query).Scan(&tablename, &indexnameFromDB)
	if err != nil {
		return false, fmt.Errorf("failed to check if index exists: %w", err)
	}

	return indexnameFromDB == indexName, nil
}

func (*VectorStore) NewBaseIndex(indexName, indexType string, strategy distanceStrategy, partialIndexes []string, opts Index) BaseIndex {
	return BaseIndex{
		name:             indexName,
		indexType:        indexType,
		distanceStrategy: strategy,
		partialIndexes:   partialIndexes,
		options:          opts,
	}
}
