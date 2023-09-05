package chroma

import (
	"context"
	"errors"
	"fmt"

	chromago "github.com/amikos-tech/chroma-go"
	"github.com/amikos-tech/chroma-go/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

var (
	// TODO (noodnik2): confirm additional errors used by other adapters are irrelevant here (for consistency)
	ErrInvalidScoreThreshold = errors.New("score threshold must be between 0 and 1")
)

// Store is a wrapper around the chromaGo API and client
type Store struct {
	collection       *chromago.Collection
	distanceFunction chromago.DistanceFunction
	resetChroma      bool
	chromaUrl        string
	openaiApiKey     string
	collectionName   string // noodnik2: preferred this to "nameSpace" since it's germaine to Chroma
	// TODO (noodnik2): clarify need for / support of the following fields
	//embedder    embeddings.Embedder
	//grpcConn    *grpc.ClientConn
	//client      pinecone_grpc.VectorServiceClient
	//indexName   string
	//projectName string
	//environment string
	//apiKey      string
	//textKey     string
	//nameSpace   string
	//useGRPC     bool
}

// TODO (noodnik2): (why) is this needed?
//var _ vectorstores.VectorStore = Store{}

func New(_ context.Context, opts ...Option) (Store, error) {
	store, coErr := applyClientOptions(opts...)
	if coErr != nil {
		return Store{}, coErr
	}
	col, cgErr := newChromaGo(store)
	if cgErr != nil {
		return Store{}, cgErr
	}
	store.collection = col
	return store, nil
}

func (s Store) AddDocuments(_ context.Context, docs []schema.Document, _ ...vectorstores.Option) error {
	ids := make([]string, len(docs))
	texts := make([]string, len(docs))
	metadatas := make([]map[string]any, len(docs))
	for i, doc := range docs {
		texts[i] = doc.PageContent
		metadatas[i] = doc.Metadata
		ids[i] = fmt.Sprintf("id%d", i+1) // TODO (noodnik2): clarify meaning / use of "ids"
	}

	col := s.collection
	if _, addErr := col.Add(nil, metadatas, texts, ids); addErr != nil {
		return fmt.Errorf("adding documents: %w", addErr)
	}
	//countDocs, countErr := col.Count()
	//if countErr != nil {
	//	return fmt.Errorf("counting documents: %s", countErr)
	//}
	//fmt.Printf("document count: %v\n", countDocs)
	return nil
}

func (s Store) SimilaritySearch(_ context.Context, query string, numDocuments int, options ...vectorstores.Option) ([]schema.Document, error) {
	//fmt.Printf("querying query(%v)\n", query)
	opts := s.getOptions(options...)
	scoreThreshold, stErr := s.getScoreThreshold(opts)
	if stErr != nil {
		return nil, stErr
	}
	filters, _ := s.getFilters(opts).(map[string]any)
	qr, queryErr := s.collection.Query([]string{query}, int32(numDocuments), filters, nil, nil)
	if queryErr != nil {
		return nil, queryErr
	}
	//fmt.Printf("qr: %v\n", qr.Documents[0][0])

	if len(qr.Documents) != len(qr.Metadatas) || len(qr.Metadatas) != len(qr.Distances) {
		return nil, fmt.Errorf("unexpected lengths: qr.Documents[%d], qr.Metadatas[%d], qr.Distances[%d]\n",
			len(qr.Documents), len(qr.Metadatas), len(qr.Distances))
	}
	var sDocs []schema.Document
	for docsI := range qr.Documents {
		for docI := range qr.Documents[docsI] {
			distanceFound := float64(qr.Distances[docsI][docI])
			//fmt.Printf("df(%f), st(%f)\n", distanceFound, scoreThreshold)
			if (1.0 - distanceFound) >= scoreThreshold {
				sDocs = append(sDocs, schema.Document{
					Metadata:    qr.Metadatas[docsI][docI],
					PageContent: qr.Documents[docsI][docI],
				})
			}
		}
	}

	return sDocs, nil
}

func newChromaGo(s Store) (*chromago.Collection, error) {
	client := chromago.NewClient(s.chromaUrl)
	if s.resetChroma {
		// TODO (noodnik2): is this really needed?
		if _, errRest := client.Reset(); errRest != nil {
			return nil, fmt.Errorf("resetting database: %w", errRest)
		}
	}
	embeddingFunction := openai.NewOpenAIEmbeddingFunction(s.openaiApiKey)
	// TODO (noodnik2): integrate "embedding function" similar to the other vectorstore adapters
	col, errCreate := client.CreateCollection(s.collectionName, map[string]any{}, true, embeddingFunction, s.distanceFunction)
	if errCreate != nil {
		return nil, fmt.Errorf("creating collection: %w", errCreate)
	}
	return col, nil
}

// TODO (noodnik2): does this map to chroma.Store.collectionName?  E.g., for consistency with
//  the existing model, or to leave it in the local "chroma.Option" where it is now?
//func (s Store) getNameSpace(opts vectorstores.Options) string {
//	if opts.NameSpace != "" {
//		return opts.NameSpace
//	}
//	return s.nameSpace
//}

func (s Store) getScoreThreshold(opts vectorstores.Options) (float64, error) {
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
