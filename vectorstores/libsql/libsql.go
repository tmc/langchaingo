package libsql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	_ "github.com/tursodatabase/go-libsql"
)

type Store struct {
	db        *sql.DB
	embedder  embeddings.Embedder
	table     string
	column    string
	vectorDim int
}

var _ vectorstores.VectorStore = (*Store)(nil)

func New(dsn string, embedder embeddings.Embedder, table, column string, vectorDim int) (*Store, error) {
	if table == "" {
		//default will be a langchain
		table = "langchain"
	}

	if column == "" {
		column = "EMBEDDING_COLUMN"
	}

	db, err := sql.Open("libsql", dsn)

	if err != nil {
		return nil, err
	}

	store := &Store{db: db, embedder: embedder, table: table, column: column, vectorDim: vectorDim}

	ctx := context.Background()
	err = store.init(ctx)

	if err != nil {
		return nil, err
	}

	return store, nil
}

func (s *Store) GetDB() *sql.DB {
	return s.db
}
func (s *Store) ColumnName() string {
	return s.column
}

func (s *Store) TableName() string {
	return s.table
}

func (s *Store) VectorDim() int {
	return s.vectorDim
}

func (s *Store) init(ctx context.Context) error {

	// #nosec G201 -- table name and index name are from internal config, not user input
	create := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id TEXT PRIMARY KEY,
		content TEXT,
		metadata TEXT,
		%s F32_BLOB (%s)
	);`, s.table, s.column, strconv.Itoa(s.vectorDim))
	if _, err := s.db.ExecContext(ctx, create); err != nil {
		return err
	}

	// Create vector index
	// #nosec G201 -- table name and index name are from internal config, not user input
	index := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_%s ON %s(libsql_vector_idx(%s));`, s.table, s.column, s.table, s.column)
	_, err := s.db.ExecContext(ctx, index)
	return err
}

func (s *Store) AddDocuments(ctx context.Context, docs []schema.Document, options ...vectorstores.Option) ([]string, error) {
	texts := make([]string, len(docs))
	for i, d := range docs {
		texts[i] = d.PageContent
	}
	vecs, err := s.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return nil, err
	}
	return s.AddVectors(ctx, vecs, docs)
}

// AddVectors inserts embeddings + docs
func (s *Store) AddVectors(ctx context.Context, vectors [][]float32, docs []schema.Document) ([]string, error) {
	if len(vectors) != len(docs) {
		return nil, errors.New("vectors length mismatch")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	// #nosec G201 -- table name and index name are from internal config, not user input
	stmt, err := tx.PrepareContext(ctx, fmt.Sprintf(`INSERT INTO %s (id, content, metadata, %s) VALUES (?, ?, ?, vector(?))`, s.table, s.column))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	ids := make([]string, len(docs))
	for i, doc := range docs {
		id := uuid.New().String()
		ids[i] = id
		vecStr := make([]string, len(vectors[i]))
		for j, v := range vectors[i] {
			vecStr[j] = fmt.Sprintf("%f", v)
		}
		embedding := "[" + strings.Join(vecStr, ",") + "]"
		meta, _ := json.Marshal(doc.Metadata)
		if _, err := stmt.ExecContext(ctx, id, doc.PageContent, string(meta), embedding); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return ids, nil
}

// SimilaritySearch finds k nearest neighbors
func (s *Store) SimilaritySearch(ctx context.Context, query string, k int, options ...vectorstores.Option) ([]schema.Document, error) {
	vec, err := s.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}
	vecStr := make([]string, len(vec))
	for i, v := range vec {
		vecStr[i] = fmt.Sprintf("%f", v)
	}
	queryVector := "[" + strings.Join(vecStr, ",") + "]"

	// #nosec G201 -- table name and index name are from internal config, not user input
	sqlStr := fmt.Sprintf(`
        SELECT t.id, t.content, t.metadata,
               vector_distance_cos(t.%s, vector(?)) AS distance
        FROM vector_top_k('idx_%s_%s', vector(?), CAST(? AS INTEGER)) AS top_k
        JOIN %s t ON top_k.rowid = t.rowid
        ORDER BY distance ASC;`,
		s.column, s.table, s.column, s.table,
	)
	rows, err := s.db.QueryContext(ctx, sqlStr, queryVector, queryVector, k)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []schema.Document
	for rows.Next() {
		var id, content, metaStr string
		var score float64
		if err := rows.Scan(&id, &content, &metaStr, &score); err != nil {
			return nil, err
		}
		var meta map[string]interface{}
		_ = json.Unmarshal([]byte(metaStr), &meta)
		docs = append(docs, schema.Document{
			PageContent: content,
			Metadata:    meta,
			Score:       float32(score),
		})
	}
	return docs, rows.Err()
}

// SimilaritySearchWithScore finds k nearest neighbors and returns documents with their similarity scores
func (s *Store) SimilaritySearchWithScore(ctx context.Context, query string, k int, options ...vectorstores.Option) ([]struct {
	Doc   schema.Document
	Score float32
}, error) {
	vec, err := s.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	vecStr := make([]string, len(vec))
	for i, v := range vec {
		vecStr[i] = fmt.Sprintf("%f", v)
	}
	queryVector := "[" + strings.Join(vecStr, ",") + "]"

	// #nosec G201 -- table name and index name are from internal config, not user input
	sqlStr := fmt.Sprintf(`
    SELECT t.id, t.content, t.metadata,
           1 - vector_distance_cos(t.%s, vector(?)) AS score
    FROM vector_top_k('idx_%s_%s', vector(?), CAST(? AS INTEGER)) AS top_k
    JOIN %s t ON top_k.rowid = t.rowid
    ORDER BY score DESC;`,
		s.column, s.table, s.column, s.table,
	)
	rows, err := s.db.QueryContext(ctx, sqlStr, queryVector, queryVector, k)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []struct {
		Doc   schema.Document
		Score float32
	}

	for rows.Next() {
		var id, content, metaStr string
		var score float64
		if err := rows.Scan(&id, &content, &metaStr, &score); err != nil {
			return nil, err
		}

		var meta map[string]interface{}
		_ = json.Unmarshal([]byte(metaStr), &meta)

		results = append(results, struct {
			Doc   schema.Document
			Score float32
		}{
			Doc: schema.Document{
				PageContent: content,
				Metadata:    meta,
			},
			Score: float32(score),
		})
	}

	return results, rows.Err()
}

// Delete removes by ids or all
func (s *Store) Delete(ctx context.Context, ids []string, deleteAll bool) error {
	if deleteAll {
		_, err := s.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", s.table))
		return err
	}
	if len(ids) == 0 {
		return errors.New("no ids provided")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE id = ?", s.table))
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, id := range ids {
		if _, err := stmt.ExecContext(ctx, id); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
