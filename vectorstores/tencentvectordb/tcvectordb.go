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
	"google.golang.org/protobuf/types/known/structpb"
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

// Store is a wrapper around the tencentvectordb rest API and grpc client.
type Store struct {
	embedder   embeddings.Embedder
	client     *tcvectordb.Client
	collection *tcvectordb.Collection
	userOption *tcvectordb.ClientOption
	url        string
	apiKey     string
	userName   string

	database              string
	collectionName        string
	shardNum, replicasNum uint32
	dimension             int
	indexType             string
	metricType            string
	collectionDescription string
	vdbEmbedding          string
}

// New creates a new Store with options. Options for WithAPIKey, WithHost and WithEmbedder must be set.
func New(opts ...Option) (Store, error) {
	s, err := applyClientOptions(opts...)
	if err != nil {
		return Store{}, err
	}

	s.client, err = tcvectordb.NewClient(s.url, s.userName, s.apiKey, s.userOption)
	if err != nil {
		return Store{}, err
	}
	db, err := s.getDatabase(context.Background())
	if err != nil {
		return Store{}, err
	}
	s.collection, err = s.getCollection(context.Background(), db)
	if err != nil {
		return Store{}, err
	}
	return s, nil
}

func (s Store) getDatabase(ctx context.Context) (*tcvectordb.Database, error) {
	listDatabaseRsp, err := s.client.ListDatabase(ctx)
	if err != nil {
		return nil, fmt.Errorf("list database: %w", err)
	}
	for _, db := range listDatabaseRsp.Databases {
		if db.DatabaseName == s.database {
			return &db, nil
		}
	}
	//create database
	createdDb, err := s.client.CreateDatabase(ctx, s.database)
	if err != nil {
		return nil, fmt.Errorf("create database: %w", err)
	}
	return &createdDb.Database, nil
}
func (s Store) getCollection(ctx context.Context, db *tcvectordb.Database) (*tcvectordb.Collection, error) {
	colls, err := db.ListCollection(ctx)
	if err != nil {
		return nil, fmt.Errorf("list collection: %w", err)

	}
	for _, coll := range colls.Collections {
		if coll.CollectionName == s.collectionName {
			return coll, nil
		}
	}
	//create collection
	return s.createCollection(ctx, db)
}
func (s Store) createCollection(ctx context.Context, db *tcvectordb.Database) (*tcvectordb.Collection, error) {
	//TODO : index
	index := tcvectordb.Indexes{}
	ebd := &tcvectordb.Embedding{VectorField: "vector", Field: "text", Model: tcvectordb.BGE_BASE_ZH}
	//TODO: Embedding
	return db.CreateCollection(ctx, s.collectionName, s.shardNum, s.replicasNum, s.collectionDescription, index,
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

	nameSpace := s.getNameSpace(opts)

	indexConn, err := s.client.IndexWithNamespace(s.url, nameSpace)
	if err != nil {
		return nil, err
	}
	defer indexConn.Close()

	texts := make([]string, 0, len(docs))
	for _, doc := range docs {
		texts = append(texts, doc.PageContent)
	}

	vectors, err := s.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return nil, err
	}

	if len(vectors) != len(docs) {
		return nil, ErrEmbedderWrongNumberVectors
	}

	metadatas := make([]map[string]any, 0, len(docs))
	for i := 0; i < len(docs); i++ {
		metadata := make(map[string]any, len(docs[i].Metadata))
		for key, value := range docs[i].Metadata {
			metadata[key] = value
		}
		metadata[s.textKey] = texts[i]

		metadatas = append(metadatas, metadata)
	}

	tcvDocs := make([]tcvectordb.Document, 0, len(vectors))

	ids := make([]string, len(vectors))
	for i := 0; i < len(vectors); i++ {
		metadataStruct, err := structpb.NewStruct(metadatas[i])
		if err != nil {
			return nil, err
		}

		id := uuid.New().String()
		ids[i] = id
		tcvDocs = append(
			tcvDocs,
			tcvectordb.Document{
				Id:     id,
				Vector: vectors[i],
				Fields: metadataStruct,
			},
		)
	}

	_, err = s.collection.Upsert(ctx, tcvDocs)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

// SimilaritySearch creates a vector embedding from the query using the embedder
// and queries to find the most similar documents.
func (s Store) SimilaritySearch(ctx context.Context, query string, numDocuments int, options ...vectorstores.Option) ([]schema.Document, error) { //nolint:lll
	opts := s.getOptions(options...)

	nameSpace := s.getNameSpace(opts)
	indexConn, err := s.client.IndexWithNamespace(s.url, nameSpace)
	if err != nil {
		return nil, err
	}
	defer indexConn.Close()

	var protoFilterStruct *structpb.Struct
	filters := s.getFilters(opts)
	if filters != nil {
		protoFilterStruct, err = s.createProtoStructFilter(filters)
		if err != nil {
			return nil, err
		}
	}

	scoreThreshold, err := s.getScoreThreshold(opts)
	if err != nil {
		return nil, err
	}
	var result *tcvectordb.SearchDocumentResult
	//search by text
	if s.embedder == nil {
		result, err = s.collection.SearchByText(ctx, query, uint32(numDocuments), protoFilterStruct, true, true)
	} else {
		//search by vector
		vector, err := s.embedder.EmbedQuery(ctx, query)
		if err != nil {
			return nil, err
		}
		result, err = s.collection.Search(ctx, [][]float32{vector}, &tcvectordb.SearchDocumentParams{
			Params:         &tcvectordb.SearchDocParams{Ef: 100}, // 若使用HNSW索引，则需要指定参数ef，ef越大，召回率越高，但也会影响检索速度
			RetrieveVector: false,                                // 是否需要返回向量字段，False：不返回，True：返回
			Limit:          10,                                   // 指定 Top K 的 K 值
		})
	}
	if err != nil {
		return nil, err
	}

	if len(result.Documents) == 0 {
		return nil, ErrEmptyResponse
	}

	return s.getDocumentsFromMatches(result, scoreThreshold)
}

func (s Store) getDocumentsFromMatches(queryResult *tcvectordb.SearchDocumentResult, scoreThreshold float32) ([]schema.Document, error) {
	resultDocuments := make([]schema.Document, 0)
	for _, item := range queryResult.Documents {
		for _, d := range item {

			doc := schema.Document{
				PageContent: d.Fields[s.textKey].Val.(string),
				Metadata:    d.Fields[s.metadataKey].Val.(map[string]any),
				Score:       d.Score,
			}
			resultDocuments = append(resultDocuments, doc)
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

func (s Store) createProtoStructFilter(filter any) (*structpb.Struct, error) {
	filterBytes, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}

	var filterStruct structpb.Struct
	err = json.Unmarshal(filterBytes, &filterStruct)
	if err != nil {
		return nil, err
	}

	return &filterStruct, nil
}
