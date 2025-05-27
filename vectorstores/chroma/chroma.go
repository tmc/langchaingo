package chroma

import (
    "context"
    "errors"
    "fmt"
    "maps"

    chromago "github.com/amikos-tech/chroma-go"
    chromagov2 "github.com/amikos-tech/chroma-go/pkg/api/v2"
    chromaembedding "github.com/amikos-tech/chroma-go/pkg/embeddings"
    "github.com/amikos-tech/chroma-go/pkg/embeddings/openai"
    chromatypes "github.com/amikos-tech/chroma-go/types"
    "github.com/google/uuid"
    "github.com/tmc/langchaingo/embeddings"
    "github.com/tmc/langchaingo/schema"
    "github.com/tmc/langchaingo/vectorstores"
)

var (
    ErrInvalidScoreThreshold    = errors.New("score threshold must be between 0 and 1")
    ErrUnexpectedResponseLength = errors.New("unexpected length of response")
    ErrNewClient                = errors.New("error creating collection")
    ErrAddDocument              = errors.New("error adding document")
    ErrRemoveCollection         = errors.New("error resetting collection")
    ErrUnsupportedOptions       = errors.New("unsupported options")
)

// Store is a wrapper around the chromaGo API and client.
type Store struct {
    version            string
    client             *chromago.Client
    collection         *chromago.Collection
    distanceFunction   chromatypes.DistanceFunction
    chromaURL          string
    openaiAPIKey       string
    openaiOrganization string

    nameSpace    string
    nameSpaceKey string
    embedder     embeddings.Embedder
    includes     []chromatypes.QueryEnum

    clientV2     chromagov2.Client
    collectionV2 chromagov2.Collection
    includesV2   []chromagov2.Include
}

var _ vectorstores.VectorStore = Store{}

// New creates an active client connection to the (specified, or default) collection in the Chroma server
// and returns the `Store` object needed by the other accessors.
func New(opts ...Option) (Store, error) {
    s, coErr := applyClientOptions(opts...)
    if coErr != nil {
        return s, coErr
    }

    if s.version == ChromaV1 {
        return NewV1(s)
    }

    return NewV2(s)
}

func NewV2(s Store) (Store, error) {
    chromaClient, err := chromagov2.NewHTTPClient(chromagov2.WithBaseURL(s.chromaURL))
    if err != nil {
        return s, err
    }

    if errHb := chromaClient.Heartbeat(context.Background()); errHb != nil {
        return s, errHb
    }
    s.clientV2 = chromaClient

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

    col, errCc := s.clientV2.GetOrCreateCollection(context.Background(), s.nameSpace,
        chromagov2.WithEmbeddingFunctionCreate(embeddingFunction))
    if errCc != nil {
        return s, fmt.Errorf("%w: %w", ErrNewClient, errCc)
    }

    s.collectionV2 = col
    return s, nil
}

func NewV1(s Store) (Store, error) {
    // create the client connection and confirm that we can access the server with it
    chromaClient, err := chromago.NewClient(chromago.WithBasePath(s.chromaURL))
    if err != nil {
        return s, err
    }

    if _, errHb := chromaClient.Heartbeat(context.Background()); errHb != nil {
        return s, errHb
    }
    s.client = chromaClient

    var embeddingFunction chromatypes.EmbeddingFunction
    if s.embedder != nil {
        // inject user's embedding function, if provided
        embeddingFunction = chromaGoEmbedder{Embedder: s.embedder}
    } else {
        // otherwise use standard langchaingo OpenAI embedding function
        var options []openai.Option

        if s.openaiOrganization != "" {
            options = append(options, openai.WithOpenAIOrganizationID(s.openaiOrganization))
        }
        openAiEmbeddingFunc, err := openai.NewOpenAIEmbeddingFunction(s.openaiAPIKey, options...)
        if err != nil {
            return s, err
        }

        embeddingFunction = chromatypes.NewV2EmbeddingFunctionAdapter(openAiEmbeddingFunc)
    }

    col, errCc := s.client.CreateCollection(context.Background(), s.nameSpace, map[string]any{}, true,
        embeddingFunction, s.distanceFunction)
    if errCc != nil {
        return s, fmt.Errorf("%w: %w", ErrNewClient, errCc)
    }

    s.collection = col

    return s, nil
}

func (s Store) AddDocuments(ctx context.Context,
    docs []schema.Document,
    options ...vectorstores.Option,
) ([]string, error) {
    if s.version == ChromaV1 {
        return s.AddDocumentsV1(ctx, docs, options...)
    }

    return s.AddDocumentsV2(ctx, docs, options...)
}

func (s Store) SimilaritySearch(ctx context.Context, query string, numDocuments int,
    options ...vectorstores.Option,
) ([]schema.Document, error) {
    if s.version == ChromaV1 {
        return s.SimilaritySearchV1(ctx, query, numDocuments, options...)
    }

    return s.SimilaritySearchV2(ctx, query, numDocuments, options...)
}

func (s Store) RemoveCollection() error {
    if s.version == ChromaV1 {
        return s.RemoveCollectionV1()
    }

    return s.RemoveCollectionV2()
}

// AddDocumentsV1 adds the text and metadata from the documents to the Chroma collection associated with 'Store'.
// and returns the ids of the added documents.
func (s Store) AddDocumentsV1(ctx context.Context,
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

    ids := make([]string, len(docs))
    texts := make([]string, len(docs))
    metadatas := make([]map[string]any, len(docs))
    for docIdx, doc := range docs {
        ids[docIdx] = uuid.New().String() // TODO (noodnik2): find & use something more meaningful
        texts[docIdx] = doc.PageContent
        mc := make(map[string]any, 0)
        maps.Copy(mc, doc.Metadata)
        metadatas[docIdx] = mc
        if nameSpace != "" {
            metadatas[docIdx][s.nameSpaceKey] = nameSpace
        }
    }

    col := s.collection
    if _, addErr := col.Add(ctx, nil, metadatas, texts, ids); addErr != nil {
        return nil, fmt.Errorf("%w: %w", ErrAddDocument, addErr)
    }
    return ids, nil
}

// SimilaritySearchV1 performs a vector similarity search against the collection using the given query string.
// It returns up to numDocuments that meet the score threshold (if configured), along with their metadata and scores.
func (s Store) SimilaritySearchV1(ctx context.Context, query string, numDocuments int,
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
    qr, queryErr := s.collection.Query(ctx, []string{query}, safeIntToInt32(numDocuments), filter, nil, s.includes)
    if queryErr != nil {
        return nil, queryErr
    }

    if len(qr.Documents) != len(qr.Metadatas) || len(qr.Metadatas) != len(qr.Distances) {
        return nil, fmt.Errorf("%w: qr.Documents[%d], qr.Metadatas[%d], qr.Distances[%d]",
            ErrUnexpectedResponseLength, len(qr.Documents), len(qr.Metadatas), len(qr.Distances))
    }
    var sDocs []schema.Document
    for docsI := range qr.Documents {
        for docI := range qr.Documents[docsI] {
            if score := 1.0 - qr.Distances[docsI][docI]; score >= scoreThreshold {
                sDocs = append(sDocs, schema.Document{
                    Metadata:    qr.Metadatas[docsI][docI],
                    PageContent: qr.Documents[docsI][docI],
                    Score:       score,
                })
            }
        }
    }

    return sDocs, nil
}

// RemoveCollectionV1 deletes the current collection from the Chroma client.
func (s Store) RemoveCollectionV1() error {
    if s.client == nil || s.collection == nil {
        return fmt.Errorf("%w: no collection", ErrRemoveCollection)
    }
    _, errDc := s.client.DeleteCollection(context.Background(), s.collection.Name)
    if errDc != nil {
        return fmt.Errorf("%w(%s): %w", ErrRemoveCollection, s.collection.Name, errDc)
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

func (s Store) getScoreThreshold(opts vectorstores.Options) (float32, error) {
    if opts.ScoreThreshold < 0 || opts.ScoreThreshold > 1 {
        return 0, ErrInvalidScoreThreshold
    }
    return opts.ScoreThreshold, nil
}

func (s Store) getNameSpace(opts vectorstores.Options) string {
    if opts.NameSpace != "" {
        return opts.NameSpace
    }
    return s.nameSpace
}

func (s Store) getNamespacedFilter(opts vectorstores.Options) map[string]any {
    filter, _ := opts.Filters.(map[string]any)

    nameSpace := s.getNameSpace(opts)
    if nameSpace == "" || s.nameSpaceKey == "" {
        return filter
    }

    nameSpaceFilter := map[string]any{s.nameSpaceKey: nameSpace}
    if filter == nil {
        return nameSpaceFilter
    }

    return map[string]any{"$and": []map[string]any{nameSpaceFilter, filter}}
}

func safeIntToInt32(n int) int32 {
    return int32(max(0, n))
}

// AddDocumentsV2 adds the text and metadata from the documents to the Chroma collection associated with 'Store'.
// and returns the ids of the added documents.
func (s Store) AddDocumentsV2(ctx context.Context,
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

    ids := make([]chromagov2.DocumentID, len(docs))
    texts := make([]string, len(docs))
    metadatas := make([]chromagov2.DocumentMetadata, len(docs))
    var err error
    for docIdx, doc := range docs {
        ids[docIdx] = chromagov2.DocumentID(uuid.New().String()) // TODO (noodnik2): find & use something more meaningful
        texts[docIdx] = doc.PageContent
        metadatas[docIdx], err = chromagov2.NewDocumentMetadataFromMap(doc.Metadata)
        if err != nil {
            return nil, fmt.Errorf("get meta data from map fail: %w", err)
        }
        if nameSpace != "" {
            metadatas[docIdx].SetString(s.nameSpaceKey, nameSpace)
        }
    }

    col := s.collectionV2
    if addErr := col.Add(ctx, chromagov2.WithMetadatas(metadatas...), chromagov2.WithTexts(texts...), chromagov2.WithIDs(ids...)); addErr != nil {
        return nil, fmt.Errorf("%w: %w", ErrAddDocument, addErr)
    }

    idsStr := make([]string, len(ids))
    for i, id := range ids {
        idsStr[i] = string(id)
    }

    return idsStr, nil
}

// SimilaritySearchV2 performs a vector similarity search against the collection using the given query string.
// It returns up to numDocuments that meet the score threshold (if configured), along with their metadata and scores.
func (s Store) SimilaritySearchV2(ctx context.Context, query string, numDocuments int,
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

    filter := s.getNamespacedFilterV2(opts)
    qr, queryErr := s.collectionV2.Query(ctx, chromagov2.WithQueryTexts(query), chromagov2.WithNResults(numDocuments),
        chromagov2.WithWhereQuery(filter), chromagov2.WithIncludeQuery(s.includesV2...))
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
                dm, _ := qr.GetMetadatasGroups()[docsI][docI].(*chromagov2.DocumentMetadataImpl)
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

// RemoveCollectionV2 deletes the current collection from the Chroma client.
func (s Store) RemoveCollectionV2() error {
    if s.clientV2 == nil || s.collectionV2 == nil {
        return fmt.Errorf("%w: no collection", ErrRemoveCollection)
    }
    errDc := s.clientV2.DeleteCollection(context.Background(), s.collectionV2.Name())
    if errDc != nil {
        return fmt.Errorf("%w(%s): %w", ErrRemoveCollection, s.collectionV2.Name(), errDc)
    }
    return nil
}

func (s Store) getNamespacedFilterV2(opts vectorstores.Options) chromagov2.WhereFilter {
    filter, _ := opts.Filters.(chromagov2.WhereClause)

    nameSpace := s.getNameSpace(opts)
    if nameSpace == "" || s.nameSpaceKey == "" {
        return filter
    }

    nameSpaceFilter := chromagov2.EqString(s.nameSpaceKey, nameSpace)
    if filter == nil {
        return nameSpaceFilter
    }

    return chromagov2.And([]chromagov2.WhereClause{nameSpaceFilter, filter}...)
}
