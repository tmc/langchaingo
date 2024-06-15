package tencentvectordb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/tencent/vectordatabase-sdk-go/tcvectordb"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
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

const (
	fieldID       = "id"
	fieldVector   = "vector"
	fieldText     = "text"
	fieldMetadata = "metadata"
)

const (
	defaultHNSWParamM              = 16
	defaultHNSWParamEfConstruction = 200
	defaultSearchDocParamsEf       = 10
)

// MetaField is a field in the metadata of a document.
type MetaField struct {
	Name        string
	Description string
	// DataType enum: string, uint64, array, vector
	DataType string
	Index    bool
}

func (m MetaField) getFieldType() tcvectordb.FieldType {
	switch m.DataType {
	case "string":
		return tcvectordb.String
	case "uint64":
		return tcvectordb.Uint64
	case "array":
		return tcvectordb.Array
	case "vector":
		return tcvectordb.Vector
	default:
		panic("unsupported data type")
	}
}

// Store is a wrapper around the tencentvectordb rest API and grpc client.
type Store struct {
	embedder   embeddings.Embedder
	database   *tcvectordb.Database
	collection *tcvectordb.Collection
	userOption *tcvectordb.ClientOption
	url        string
	apiKey     string
	userName   string

	databaseName          string
	collectionName        string
	shardNum, replicasNum uint32
	dimension             uint32
	indexType             string
	metricType            string
	collectionDescription string
	embeddingModel        string
	metaFields            []MetaField
}

// New creates a new Store with options. Options for WithAPIKey, WithHost and WithEmbedder must be set.
func New(opts ...Option) (Store, error) {
	s, err := applyClientOptions(opts...)
	if err != nil {
		return Store{}, err
	}
	client, err := tcvectordb.NewClient(s.url, s.userName, s.apiKey, s.userOption)
	if err != nil {
		return Store{}, err
	}
	s.database, err = s.getDatabase(context.Background(), client)
	if err != nil {
		return Store{}, err
	}
	s.collection, err = s.getDBCollection(context.Background(), s.database, s.collectionName)
	if err != nil {
		return Store{}, err
	}
	return s, nil
}

func (s Store) getIndexType() tcvectordb.IndexType {
	switch s.indexType {
	case "FLAT":
		return tcvectordb.FLAT
	case "HNSW":
		return tcvectordb.HNSW
	case "IVF_FLAT":
		return tcvectordb.IVF_FLAT
	case "IVF_PQ":
		return tcvectordb.IVF_PQ
	case "IVF_SQ4":
		return tcvectordb.IVF_SQ4
	case "IVF_SQ8":
		return tcvectordb.IVF_SQ8
	case "IVF_SQ16":
		return tcvectordb.IVF_SQ16
	default:
		panic("unsupported index_type")
	}
}

func (s Store) getMetricType() tcvectordb.MetricType {
	switch s.metricType {
	case "L2":
		return tcvectordb.L2
	case "IP":
		return tcvectordb.IP
	case "COSINE":
		return tcvectordb.COSINE
	default:
		panic("unsupported metric_type")
	}
}

func (s Store) getEmbeddingModel() tcvectordb.EmbeddingModel {
	switch s.embeddingModel {
	case "M3E_BASE":
		return tcvectordb.M3E_BASE
	case "BGE_BASE_ZH":
		return tcvectordb.BGE_BASE_ZH
	case "BGE_LARGE_ZH":
		return tcvectordb.BGE_LARGE_ZH
	case "MULTILINGUAL_E5_BASE":
		return tcvectordb.MULTILINGUAL_E5_BASE
	case "E5_LARGE_V2":
		return tcvectordb.E5_LARGE_V2
	case "TEXT2VEC_LARGE_CHINESE":
		return tcvectordb.TEXT2VEC_LARGE_CHINESE
	default:
		panic("unsupported embedding model")
	}
}

func (s Store) getDatabase(ctx context.Context, client *tcvectordb.Client) (*tcvectordb.Database, error) {
	listDatabaseRsp, err := client.ListDatabase(ctx)
	if err != nil {
		return nil, fmt.Errorf("list database: %w", err)
	}
	for _, db := range listDatabaseRsp.Databases {
		if db.DatabaseName == s.databaseName {
			return &db, nil
		}
	}
	// create database
	createdDB, err := client.CreateDatabase(ctx, s.databaseName)
	if err != nil {
		return nil, fmt.Errorf("create database: %w", err)
	}
	return &createdDB.Database, nil
}

func (s Store) getDBCollection(ctx context.Context, db *tcvectordb.Database, collectionName string) (*tcvectordb.Collection, error) {
	colls, err := db.ListCollection(ctx)
	if err != nil {
		return nil, fmt.Errorf("list collection: %w", err)
	}
	for _, coll := range colls.Collections {
		if coll.CollectionName == collectionName {
			return coll, nil
		}
	}
	// create collection
	return s.createCollection(ctx, db, collectionName)
}

// DropCollection drops the collection from the database.
func (s Store) DropCollection(ctx context.Context, collectionName string) error {
	_, err := s.database.DropCollection(ctx, collectionName)
	return err
}

func (s Store) createCollection(ctx context.Context, db *tcvectordb.Database, collectionName string) (*tcvectordb.Collection, error) {
	// define index
	dim := s.dimension
	if dim == 0 && s.embedder != nil {
		// get dimension from embedder
		v, err := s.embedder.EmbedQuery(ctx, "test")
		if err != nil {
			return nil, err
		}
		dim = uint32(len(v))
	}
	// create vector index and filter index
	index := tcvectordb.Indexes{
		VectorIndex: []tcvectordb.VectorIndex{
			{
				FilterIndex: tcvectordb.FilterIndex{
					FieldName: fieldVector,
					FieldType: tcvectordb.Vector,
					IndexType: s.getIndexType(),
				},
				Dimension:  dim,
				MetricType: s.getMetricType(),
				Params: &tcvectordb.HNSWParam{
					M:              defaultHNSWParamM,
					EfConstruction: defaultHNSWParamEfConstruction,
				},
			},
		},
		// add text field as filter index
		FilterIndex: []tcvectordb.FilterIndex{
			{
				FieldName: fieldID,
				FieldType: tcvectordb.String,
				IndexType: tcvectordb.PRIMARY,
			},
			{
				FieldName: fieldText,
				FieldType: tcvectordb.String,
				IndexType: tcvectordb.FILTER,
			},
		},
	}
	// Add metadata indexes
	if len(s.metaFields) > 0 {
		for _, metaField := range s.metaFields {
			if metaField.Index {
				index.FilterIndex = append(index.FilterIndex, tcvectordb.FilterIndex{
					FieldName: metaField.Name,
					FieldType: metaField.getFieldType(),
					IndexType: tcvectordb.FILTER,
				})
			}
		}
	} else {
		index.FilterIndex = append(index.FilterIndex, tcvectordb.FilterIndex{
			FieldName: fieldMetadata,
			FieldType: tcvectordb.String,
			IndexType: tcvectordb.FILTER,
		})
	}
	if s.embedder != nil {
		return db.CreateCollection(ctx, collectionName, s.shardNum, s.replicasNum, s.collectionDescription, index)
	}
	// use vectordb Embedding
	ebd := &tcvectordb.Embedding{VectorField: fieldVector, Field: fieldText, Model: s.getEmbeddingModel()}
	return db.CreateCollection(ctx, collectionName, s.shardNum, s.replicasNum, s.collectionDescription, index,
		&tcvectordb.CreateCollectionParams{
			Embedding: ebd,
		})
}

func (s Store) getMeta(result map[string]tcvectordb.Field) map[string]any {
	if len(s.metaFields) > 0 {
		meta := make(map[string]any)
		for _, field := range s.metaFields {
			meta[field.Name] = result[field.Name]
		}
		return meta
	}
	rawMeta, ok := result[fieldMetadata]
	if ok {
		if rawMetaStr, ok := rawMeta.Val.(string); ok {
			meta := make(map[string]any)
			_ = json.Unmarshal([]byte(rawMetaStr), &meta)
			return meta
		}
	}
	return map[string]any{}
}

// AddDocuments creates vector embeddings from the documents using the embedder
// and upsert the vectors to the tencentvectordb index and returns the ids of the added documents.
func (s Store) AddDocuments(ctx context.Context,
	docs []schema.Document,
	options ...vectorstores.Option,
) ([]string, error) {
	opts := s.getOptions(options...)
	if len(docs) == 0 {
		return nil, nil
	}

	texts := make([]string, 0, len(docs))
	for _, doc := range docs {
		texts = append(texts, doc.PageContent)
	}
	// embed documents
	embedder := s.embedder
	if opts.Embedder != nil {
		embedder = opts.Embedder
	}
	var err error
	vectors := make([][]float32, 0, len(docs))
	if embedder != nil {
		vectors, err = embedder.EmbedDocuments(ctx, texts)
		if err != nil {
			return nil, err
		}

		if len(vectors) != len(docs) {
			return nil, ErrEmbedderWrongNumberVectors
		}
	}
	// convert documents to tcvectordb.Document
	tcvDocs := make([]tcvectordb.Document, 0, len(docs))
	ids := make([]string, len(docs))
	for i, doc := range docs {
		fields := s.convertDocument2Field(doc)
		id := uuid.New().String()
		ids[i] = id
		tcvDoc := tcvectordb.Document{
			Id:     id,
			Fields: fields,
		}
		if embedder != nil {
			tcvDoc.Vector = vectors[i]
		}
		tcvDocs = append(tcvDocs, tcvDoc)
	}
	// upsert documents
	_, err = s.collection.Upsert(ctx, tcvDocs)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (s Store) convertDocument2Field(doc schema.Document) map[string]tcvectordb.Field {
	metadata := make(map[string]tcvectordb.Field, len(doc.Metadata))
	for key, value := range doc.Metadata {
		metadata[key] = tcvectordb.Field{Val: value}
	}
	metadata[fieldText] = tcvectordb.Field{Val: doc.PageContent}
	return metadata
}

func (s Store) convertField2Document(fields map[string]tcvectordb.Field) schema.Document {
	doc := schema.Document{PageContent: fields[fieldText].Val.(string)} //nolint
	metaData := make(map[string]any)
	for key, value := range fields {
		if key == fieldText {
			continue
		}
		if key == fieldMetadata {
			metaData = s.getMeta(fields)
			continue
		}
		metaData[key] = value.Val
	}
	doc.Metadata = metaData
	return doc
}

// SimilaritySearch creates a vector embedding from the query using the embedder
// and queries to find the most similar documents.
func (s Store) SimilaritySearch(ctx context.Context, query string, numDocuments int, options ...vectorstores.Option) ([]schema.Document, error) {
	// get options
	opts := s.getOptions(options...)
	scoreThreshold, err := s.getScoreThreshold(opts)
	if err != nil {
		return nil, err
	}
	// filter
	filter := s.getFilters(opts)

	searchParams := &tcvectordb.SearchDocumentParams{
		Params:         &tcvectordb.SearchDocParams{Ef: defaultSearchDocParamsEf},
		RetrieveVector: false,
		Limit:          int64(numDocuments),
		Filter:         filter,
	}
	var result *tcvectordb.SearchDocumentResult
	embedder := s.embedder
	if opts.Embedder != nil {
		embedder = opts.Embedder
	}
	// search by text
	if embedder == nil {
		result, err = s.collection.SearchByText(ctx, map[string][]string{fieldText: {query}}, searchParams)
	} else {
		// search by vector
		vector, err1 := embedder.EmbedQuery(ctx, query)
		if err1 != nil {
			return nil, err1
		}
		result, err = s.collection.Search(ctx, [][]float32{vector}, searchParams)
	}
	if err != nil {
		return nil, err
	}

	if len(result.Documents) == 0 {
		return nil, ErrEmptyResponse
	}

	return s.getDocumentsFromMatches(result, scoreThreshold)
}

// getDocumentsFromMatches filter by scoreThreshold and convert tcvectordb.Document to schema.Document.
func (s Store) getDocumentsFromMatches(queryResult *tcvectordb.SearchDocumentResult, scoreThreshold float32) ([]schema.Document, error) {
	resultDocuments := make([]schema.Document, 0)
	for _, item := range queryResult.Documents {
		for _, d := range item {
			if scoreThreshold == 0 || d.Score >= scoreThreshold {
				doc := s.convertField2Document(d.Fields)
				doc.Score = d.Score
				resultDocuments = append(resultDocuments, doc)
			}
		}
	}
	return resultDocuments, nil
}

func (s Store) getScoreThreshold(opts vectorstores.Options) (float32, error) {
	if opts.ScoreThreshold < 0 || opts.ScoreThreshold > 1 {
		return 0, ErrInvalidScoreThreshold
	}
	return opts.ScoreThreshold, nil
}

func (s Store) getFilters(opts vectorstores.Options) *tcvectordb.Filter {
	if opts.Filters != nil {
		return tcvectordb.NewFilter((opts.Filters).(string)) //nolint
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
