package chroma

import (
	"context"
	"fmt"
	chromago "github.com/amikos-tech/chroma-go/pkg/api/v2"
	chromaembedding "github.com/amikos-tech/chroma-go/pkg/embeddings"
	"github.com/amikos-tech/chroma-go/pkg/embeddings/openai"
	chromatypes "github.com/amikos-tech/chroma-go/types"
	"github.com/google/uuid"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

// StoreV2 is a wrapper around the chromaGo V2 API and client.
type StoreV2 struct {
	client             chromago.Client
	collection         chromago.Collection
	distanceFunction   chromatypes.DistanceFunction
	chromaURL          string
	openaiAPIKey       string
	openaiOrganization string

	nameSpace    string
	nameSpaceKey string
	embedder     embeddings.Embedder
	includes     []chromago.Include
}

var _ vectorstores.VectorStore = StoreV2{}

// NewV2 creates an active client connection to the (specified, or default) collection in the Chroma server
// and returns the `Store` object needed by the other accessors.
func NewV2(opts ...OptionV2) (StoreV2, error) {
	s, coErr := applyV2ClientOptions(opts...)
	if coErr != nil {
		return s, coErr
	}

	chromaClient, err := chromago.NewHTTPClient(chromago.WithBaseURL(s.chromaURL))
	if err != nil {
		return s, err
	}

	if errHb := chromaClient.Heartbeat(context.Background()); errHb != nil {
		return s, errHb
	}
	s.client = chromaClient

	// inject user's embedding function, if provided
	var embeddingFunction chromaembedding.EmbeddingFunction
	if s.embedder != nil {
		// inject user's embedding function, if provided
		embeddingFunction = chromaGoEmbedderV2{Embedder: s.embedder}
	} else {
		// otherwise use standard langchaingo OpenAI embedding function
		var options []openai.Option
		if s.openaiOrganization != "" {
			options = append(options, openai.WithOpenAIOrganizationID(s.openaiOrganization))
		}
		embeddingFunction, err = openai.NewOpenAIEmbeddingFunction(s.openaiAPIKey, options...)
		if err != nil {
			return s, err
		}
	}

	col, errCc := s.client.GetOrCreateCollection(context.Background(), s.nameSpace,
		chromago.WithEmbeddingFunctionCreate(embeddingFunction))
	if errCc != nil {
		return s, fmt.Errorf("%w: %w", ErrNewClient, errCc)
	}

	s.collection = col
	return s, nil
}

// AddDocuments adds the text and metadata from the documents to the Chroma collection associated with 'Store'.
// and returns the ids of the added documents.
func (s StoreV2) AddDocuments(ctx context.Context,
	docs []schema.Document,
	options ...vectorstores.Option,
) ([]string, error) {
	opts := s.getOptions(options...)
	if opts.Embedder != nil || opts.ScoreThreshold != 0 || opts.Filters != nil {
		return nil, ErrUnsupportedOptions
	}

	nameSpace := s.getNameSpace(opts)
	if nameSpace != "" && s.nameSpaceKey == "" {
		return nil, fmt.Errorf("%w: nameSpace without nameSpaceKey", ErrUnsupportedOptions)
	}

	ids := make([]chromago.DocumentID, len(docs))
	texts := make([]string, len(docs))
	metadatas := make([]chromago.DocumentMetadata, len(docs))
	var err error
	for docIdx, doc := range docs {
		ids[docIdx] = chromago.DocumentID(uuid.New().String()) // TODO (noodnik2): find & use something more meaningful
		texts[docIdx] = doc.PageContent
		metadatas[docIdx], err = chromago.NewDocumentMetadataFromMap(doc.Metadata)
		if err != nil {
			return nil, fmt.Errorf("get meta data from map fail: %w", err)
		}
		if nameSpace != "" {
			metadatas[docIdx].SetString(s.nameSpaceKey, nameSpace)
		}
	}

	col := s.collection
	if addErr := col.Add(ctx, chromago.WithMetadatas(metadatas...), chromago.WithTexts(texts...), chromago.WithIDs(ids...)); addErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrAddDocument, addErr)
	}

	idsStr := make([]string, len(ids))
	for i, id := range ids {
		idsStr[i] = string(id)
	}

	return idsStr, nil
}

// SimilaritySearch performs a vector similarity search against the collection using the given query string.
// It returns up to numDocuments that meet the score threshold (if configured), along with their metadata and scores.
func (s StoreV2) SimilaritySearch(ctx context.Context, query string, numDocuments int,
	options ...vectorstores.Option,
) ([]schema.Document, error) {
	opts := s.getOptions(options...)

	if opts.Embedder != nil {
		// embedder is not used by this method, so shouldn't ever be specified
		return nil, fmt.Errorf("%w: Embedder", ErrUnsupportedOptions)
	}

	scoreThreshold, stErr := s.getScoreThreshold(opts)
	if stErr != nil {
		return nil, stErr
	}

	filter := s.getNamespacedFilter(opts)
	qr, queryErr := s.collection.Query(ctx, chromago.WithQueryTexts(query), chromago.WithNResults(numDocuments),
		chromago.WithWhereQuery(filter), chromago.WithIncludeQuery(s.includes...))
	if queryErr != nil {
		return nil, queryErr
	}

	if len(qr.GetDocumentsGroups()) != len(qr.GetMetadatasGroups()) || len(qr.GetMetadatasGroups()) != len(qr.GetDistancesGroups()) {
		return nil, fmt.Errorf("%w: qr.Documents[%d], qr.Metadatas[%d], qr.Distances[%d]",
			ErrUnexpectedResponseLength, len(qr.GetDocumentsGroups()), len(qr.GetMetadatasGroups()), len(qr.GetDistancesGroups()))
	}
	var sDocs []schema.Document
	for docsI := range qr.GetDocumentsGroups() {
		for docI := range qr.GetDocumentsGroups()[docsI] {
			if score := 1.0 - qr.GetDistancesGroups()[docsI][docI]; float32(score) >= scoreThreshold {

				md := make(map[string]any)
				dm, _ := qr.GetMetadatasGroups()[docsI][docI].(*chromago.DocumentMetadataImpl)
				for _, key := range dm.Keys() {
					if raw, ok := dm.GetRaw(key); ok {
						md[key] = raw
					}
				}
				sDocs = append(sDocs, schema.Document{
					Metadata:    md,
					PageContent: qr.GetDocumentsGroups()[docsI][docI].ContentString(),
					Score:       float32(score),
				})
			}
		}
	}

	return sDocs, nil
}

// RemoveCollection deletes the current collection from the Chroma client.
func (s StoreV2) RemoveCollection() error {
	if s.client == nil || s.collection == nil {
		return fmt.Errorf("%w: no collection", ErrRemoveCollection)
	}
	errDc := s.client.DeleteCollection(context.Background(), s.collection.Name())
	if errDc != nil {
		return fmt.Errorf("%w(%s): %w", ErrRemoveCollection, s.collection.Name(), errDc)
	}
	return nil
}

func (s StoreV2) getOptions(options ...vectorstores.Option) vectorstores.Options {
	opts := vectorstores.Options{}
	for _, opt := range options {
		opt(&opts)
	}
	return opts
}

func (s StoreV2) getScoreThreshold(opts vectorstores.Options) (float32, error) {
	if opts.ScoreThreshold < 0 || opts.ScoreThreshold > 1 {
		return 0, ErrInvalidScoreThreshold
	}
	return opts.ScoreThreshold, nil
}

func (s StoreV2) getNameSpace(opts vectorstores.Options) string {
	if opts.NameSpace != "" {
		return opts.NameSpace
	}
	return s.nameSpace
}

func (s StoreV2) getNamespacedFilter(opts vectorstores.Options) chromago.WhereFilter {
	filter, _ := opts.Filters.(chromago.WhereClause)

	nameSpace := s.getNameSpace(opts)
	if nameSpace == "" || s.nameSpaceKey == "" {
		return filter
	}

	nameSpaceFilter := chromago.EqString(s.nameSpaceKey, nameSpace)
	if filter == nil {
		return nameSpaceFilter
	}

	return chromago.And([]chromago.WhereClause{nameSpaceFilter, filter}...)
}
