package tencentvectordb

import (
	"context"
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
	fieldId       = "id"
	fieldVector   = "vector"
	fieldText     = "text"
	fieldMetadata = "metadata"
)

// MetaField is a field in the metadata of a document.
type MetaField struct {
	Name        string
	Description string
	DataType    string
	Index       bool
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
	embedder    embeddings.Embedder
	database    *tcvectordb.Database
	collections map[string]*tcvectordb.Collection
	userOption  *tcvectordb.ClientOption
	url         string
	apiKey      string
	userName    string

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
	s.collections = make(map[string]*tcvectordb.Collection)
	s.collections[s.collectionName], err = s.getDbCollection(context.Background(), s.database, s.collectionName)
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
	//create database
	createdDb, err := client.CreateDatabase(ctx, s.databaseName)
	if err != nil {
		return nil, fmt.Errorf("create database: %w", err)
	}
	return &createdDb.Database, nil
}
func (s Store) getCollection(ctx context.Context, collectionName string) (*tcvectordb.Collection, error) {
	if len(collectionName) == 0 {
		collectionName = s.collectionName
	}
	if coll, ok := s.collections[collectionName]; ok {
		return coll, nil
	}
	coll, err := s.getDbCollection(ctx, s.database, collectionName)
	if err != nil {
		return nil, err
	}
	s.collections[collectionName] = coll
	return coll, nil
}

func (s Store) getDbCollection(ctx context.Context, db *tcvectordb.Database, collectionName string) (*tcvectordb.Collection, error) {
	colls, err := db.ListCollection(ctx)
	if err != nil {
		return nil, fmt.Errorf("list collection: %w", err)

	}
	for _, coll := range colls.Collections {
		if coll.CollectionName == collectionName {
			return coll, nil
		}
	}
	//create collection
	return s.createCollection(ctx, db, collectionName)
}
func (s Store) DropCollection(ctx context.Context, collectionName string) error {
	_, err := s.database.DropCollection(ctx, collectionName)
	delete(s.collections, collectionName)
	return err
}

func (s Store) createCollection(ctx context.Context, db *tcvectordb.Database, collectionName string) (*tcvectordb.Collection, error) {
	//define index
	dim := uint32(0)
	if s.embedder != nil {
		dim = s.dimension
	}
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
					M:              16,
					EfConstruction: 200,
				},
			},
		},

		FilterIndex: []tcvectordb.FilterIndex{
			{
				FieldName: fieldId,
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
	for _, metaField := range s.metaFields {
		if metaField.Index {
			index.FilterIndex = append(index.FilterIndex, tcvectordb.FilterIndex{
				FieldName: metaField.Name,
				FieldType: metaField.getFieldType(),
				IndexType: tcvectordb.FILTER,
			})
		}
	}
	if s.embedder != nil {
		return db.CreateCollection(ctx, collectionName, s.shardNum, s.replicasNum, s.collectionDescription, index)
	}
	//use vectordb Embedding
	ebd := &tcvectordb.Embedding{VectorField: fieldVector, Field: fieldText, Model: s.getEmbeddingModel()}
	return db.CreateCollection(ctx, collectionName, s.shardNum, s.replicasNum, s.collectionDescription, index,
		&tcvectordb.CreateCollectionParams{
			Embedding: ebd,
		})
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
	collection, err := s.getCollection(ctx, opts.NameSpace)
	if err != nil {
		return nil, err
	}
	texts := make([]string, 0, len(docs))
	for _, doc := range docs {
		texts = append(texts, doc.PageContent)
	}
	//embed documents
	embedder := s.embedder
	if opts.Embedder != nil {
		embedder = opts.Embedder
	}
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
	//convert documents to tcvectordb.Document
	tcvDocs := make([]tcvectordb.Document, 0, len(docs))
	ids := make([]string, len(docs))
	for i, doc := range docs {
		fields := convert(doc)
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
	//upsert documents
	_, err = collection.Upsert(ctx, tcvDocs)
	if err != nil {
		return nil, err
	}

	return ids, nil
}
func convert(doc schema.Document) map[string]tcvectordb.Field {
	metadata := make(map[string]tcvectordb.Field, len(doc.Metadata))
	for key, value := range doc.Metadata {
		metadata[key] = tcvectordb.Field{Val: value}
	}
	metadata[fieldText] = tcvectordb.Field{Val: doc.PageContent}
	return metadata
}
func convertField(fields map[string]tcvectordb.Field) schema.Document {
	doc := schema.Document{
		PageContent: fields[fieldText].Val.(string),
	}
	metaData := make(map[string]any)
	for key, value := range fields {
		if key == fieldText {
			continue
		}
		metaData[key] = value.Val
	}
	doc.Metadata = metaData
	return doc
}

// SimilaritySearch creates a vector embedding from the query using the embedder
// and queries to find the most similar documents.
func (s Store) SimilaritySearch(ctx context.Context, query string, numDocuments int, options ...vectorstores.Option) (
	[]schema.Document, error) {
	opts := s.getOptions(options...)
	scoreThreshold, err := s.getScoreThreshold(opts)
	if err != nil {
		return nil, err
	}
	collection, err := s.getCollection(ctx, opts.NameSpace)
	if err != nil {
		return nil, err
	}
	searchParams := &tcvectordb.SearchDocumentParams{
		Params:         &tcvectordb.SearchDocParams{Ef: 100},
		RetrieveVector: false,
		Limit:          int64(numDocuments),
	}
	var result *tcvectordb.SearchDocumentResult
	embedder := s.embedder
	if opts.Embedder != nil {
		embedder = opts.Embedder
	}
	//search by text
	if embedder == nil {
		result, err = collection.SearchByText(ctx, map[string][]string{fieldText: {query}}, searchParams)

	} else {
		//search by vector
		vector, err := embedder.EmbedQuery(ctx, query)
		if err != nil {
			return nil, err
		}
		result, err = collection.Search(ctx, [][]float32{vector}, searchParams)
	}
	if err != nil {
		return nil, err
	}

	if len(result.Documents) == 0 {
		return nil, ErrEmptyResponse
	}

	return s.getDocumentsFromMatches(result, scoreThreshold)
}

func (s Store) getDocumentsFromMatches(queryResult *tcvectordb.SearchDocumentResult, scoreThreshold float32) (
	[]schema.Document, error) {
	resultDocuments := make([]schema.Document, 0)
	for _, item := range queryResult.Documents {
		for _, d := range item {
			if scoreThreshold == 0 || d.Score >= scoreThreshold {
				doc := convertField(d.Fields)
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
