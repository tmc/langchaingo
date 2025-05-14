package bedrockknowledgebases

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime/document"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithyDocument "github.com/aws/smithy-go/document"
	"github.com/google/uuid"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

type bedrockAgentAPI interface {
	GetKnowledgeBase(ctx context.Context, params *bedrockagent.GetKnowledgeBaseInput, optFns ...func(*bedrockagent.Options)) (*bedrockagent.GetKnowledgeBaseOutput, error)
	ListDataSources(ctx context.Context, params *bedrockagent.ListDataSourcesInput, optFns ...func(*bedrockagent.Options)) (*bedrockagent.ListDataSourcesOutput, error)
	GetDataSource(ctx context.Context, params *bedrockagent.GetDataSourceInput, optFns ...func(*bedrockagent.Options)) (*bedrockagent.GetDataSourceOutput, error)
	IngestKnowledgeBaseDocuments(ctx context.Context, params *bedrockagent.IngestKnowledgeBaseDocumentsInput, optFns ...func(*bedrockagent.Options)) (*bedrockagent.IngestKnowledgeBaseDocumentsOutput, error)
	StartIngestionJob(ctx context.Context, params *bedrockagent.StartIngestionJobInput, optFns ...func(*bedrockagent.Options)) (*bedrockagent.StartIngestionJobOutput, error)
	GetIngestionJob(ctx context.Context, params *bedrockagent.GetIngestionJobInput, optFns ...func(*bedrockagent.Options)) (*bedrockagent.GetIngestionJobOutput, error)
}

type bedrockAgentRuntimeAPI interface {
	Retrieve(ctx context.Context, params *bedrockagentruntime.RetrieveInput, optFns ...func(*bedrockagentruntime.Options)) (*bedrockagentruntime.RetrieveOutput, error)
}

type s3Config struct {
	client         s3API
	maxConcurrency int64
}

type KnowledgeBase struct {
	knowledgeBaseID     string
	bedrockAgent        bedrockAgentAPI
	bedrockAgentRuntime bedrockAgentRuntimeAPI
	s3Config            s3Config
}

var _ vectorstores.VectorStore = &KnowledgeBase{}

func New(ctx context.Context, knowledgeBaseID string) (*KnowledgeBase, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	bedrockAgent := bedrockagent.NewFromConfig(cfg)
	bedrockAgentRuntime := bedrockagentruntime.NewFromConfig(cfg)
	s3Client := s3.NewFromConfig(cfg)

	return newFromClients(knowledgeBaseID, bedrockAgent, bedrockAgentRuntime, s3Client), nil
}

func newFromClients(
	knowledgeBaseID string,
	bedrockAgent bedrockAgentAPI,
	bedrockAgentRuntime bedrockAgentRuntimeAPI,
	s3Client s3API,
) *KnowledgeBase {
	return &KnowledgeBase{
		knowledgeBaseID:     knowledgeBaseID,
		bedrockAgent:        bedrockAgent,
		bedrockAgentRuntime: bedrockAgentRuntime,
		s3Config: s3Config{
			client:         s3Client,
			maxConcurrency: 20,
		},
	}
}

type NamedDocument struct {
	Name     string
	Document schema.Document
}

func (kb *KnowledgeBase) AddDocuments(ctx context.Context, docs []schema.Document, options ...vectorstores.Option) ([]string, error) {
	namedDocs := make([]NamedDocument, len(docs))
	for i, doc := range docs {
		namedDocs[i] = NamedDocument{
			Name:     "kb_doc_" + uuid.NewString(),
			Document: doc,
		}
	}
	return kb.addDocuments(ctx, namedDocs, options...)
}

func (kb *KnowledgeBase) AddNamedDocuments(ctx context.Context, docs []NamedDocument, options ...vectorstores.Option) ([]string, error) {
	return kb.addDocuments(ctx, docs, options...)
}

// shouldRemoveMetadata returns true if the value should be removed from metadata.
func shouldRemoveMetadata(v any) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	// Note: we're explicitly handling the cases we care about and have a default for all others
	// nolint:exhaustive
	switch rv.Kind() {
	case reflect.Map, reflect.Slice:
		return rv.Len() == 0
	case reflect.String:
		return v == ""
	default:
		// For all other kinds, keep the value
		return false
	}
}

func (kb *KnowledgeBase) filterMetadata(docs []NamedDocument) {
	for i, doc := range docs {
		if doc.Document.Metadata == nil {
			continue
		}

		// Create a list of keys to remove
		keysToRemove := []string{}
		for k, v := range doc.Document.Metadata {
			if shouldRemoveMetadata(v) {
				keysToRemove = append(keysToRemove, k)
			}
		}

		// Remove keys
		for _, k := range keysToRemove {
			delete(doc.Document.Metadata, k)
		}

		// If no metadata left, set to nil
		if len(doc.Document.Metadata) == 0 {
			doc.Document.Metadata = nil
		}

		docs[i] = doc
	}
}

// validateDataSources checks if there are compatible data sources and returns an error if not.
func (kb *KnowledgeBase) validateDataSources(compatibleDs, incompatibleDs []dataSource) error {
	if len(compatibleDs) > 0 {
		return nil
	}

	switch {
	case len(incompatibleDs) > 0:
		return fmt.Errorf(
			"found data sources but none with S3 type, please create a data source with S3 type for the knowledge base with id: %s in the AWS console",
			kb.knowledgeBaseID,
		)
	default:
		return fmt.Errorf(
			"no data sources with S3 type found, please create a data source with S3 type for the knowledge base with id: %s in the AWS console",
			kb.knowledgeBaseID,
		)
	}
}

// findDataSource finds the appropriate data source based on options.
func (kb *KnowledgeBase) findDataSource(compatibleDs []dataSource, nameSpace string) (string, string, error) {
	if nameSpace != "" {
		for _, ds := range compatibleDs {
			if ds.ID == nameSpace {
				return ds.ID, ds.BucketARN, nil
			}
		}
		return "", "", fmt.Errorf("data source with S3 type with id %s not found", nameSpace)
	}

	if len(compatibleDs) == 1 {
		return compatibleDs[0].ID, compatibleDs[0].BucketARN, nil
	}

	return "", "", fmt.Errorf("multiple data sources with S3 type found, please specify which one you want to use by passing its id with the `vectorstores.WithNameSpace` option")
}

// collectDocumentNames extracts document names into a slice.
func (kb *KnowledgeBase) collectDocumentNames(docs []NamedDocument) []string {
	names := make([]string, len(docs))
	for i, doc := range docs {
		names[i] = doc.Name
	}
	return names
}

func (kb *KnowledgeBase) addDocuments(ctx context.Context, docs []NamedDocument, options ...vectorstores.Option) ([]string, error) {
	opts := kb.getOptions(options...)

	kb.filterMetadata(docs)
	if err := kb.checkKnowledgeBase(ctx); err != nil {
		return nil, fmt.Errorf("failed to validate knowledge base: %w", err)
	}

	compatibleDs, incompatibleDs, err := kb.listDataSources(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list data sources: %w", err)
	}

	if err := kb.validateDataSources(compatibleDs, incompatibleDs); err != nil {
		return nil, err
	}

	datasourceID, bucketARN, err := kb.findDataSource(compatibleDs, opts.NameSpace)
	if err != nil {
		return nil, err
	}

	if err := kb.addToS3(ctx, bucketARN, docs); err != nil {
		return nil, fmt.Errorf("failed to upload documents to S3: %w", err)
	}

	if err := kb.ingestDocuments(ctx, datasourceID, bucketARN, docs); err != nil {
		kb.removeFromS3(ctx, bucketARN, docs)
		return nil, fmt.Errorf("failed to ingest documents: %w", err)
	}

	ingestionJobID, err := kb.startIngestionJob(ctx, datasourceID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to start ingestion job. Documents are correctly loaded, please retry or manually sync the knowledgebase from the AWS console. error: %w",
			err,
		)
	}

	if err := kb.checkIngestionJobStatus(ctx, datasourceID, ingestionJobID); err != nil {
		return nil, fmt.Errorf(
			"failed to check ingestion status. Documents are correctly loaded, please retry or manually sync the knowledgebase from the AWS console. error: %w",
			err,
		)
	}

	return kb.collectDocumentNames(docs), nil
}

func (kb *KnowledgeBase) SimilaritySearch(ctx context.Context, query string, numDocuments int, options ...vectorstores.Option) (
	[]schema.Document, error,
) {
	opts := kb.getOptions(options...)

	query = strings.TrimSpace(query)
	docs := []schema.Document{}

	retrieveInput := bedrockagentruntime.RetrieveInput{
		KnowledgeBaseId: aws.String(kb.knowledgeBaseID),
		RetrievalQuery: &types.KnowledgeBaseQuery{
			Text: aws.String(query),
		},
		RetrievalConfiguration: &types.KnowledgeBaseRetrievalConfiguration{
			VectorSearchConfiguration: &types.KnowledgeBaseVectorSearchConfiguration{
				NumberOfResults: aws.Int32(int32(numDocuments)),
			},
		},
	}

	if opts.Filters != nil {
		filters, err := kb.getFilters(opts.Filters)
		if err != nil && !errors.Is(err, ErrNoFilters) {
			return nil, err
		}
		if filters != nil {
			retrieveInput.RetrievalConfiguration.VectorSearchConfiguration.Filter = filters
		}
	}

	p := bedrockagentruntime.NewRetrievePaginator(kb.bedrockAgentRuntime, &retrieveInput)

	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, result := range page.RetrievalResults {
			metadata, err := kb.parseMetadata(result)
			if err != nil {
				return nil, fmt.Errorf("failed to parse metadata: %w", err)
			}

			score := float32(*result.Score)

			if opts.ScoreThreshold > 0 && score < opts.ScoreThreshold {
				continue
			}
			docs = append(docs, schema.Document{
				PageContent: aws.ToString(result.Content.Text),
				Metadata:    metadata,
				Score:       score,
			})
		}
	}

	return docs, nil
}

// buildFilterAttribute creates a FilterAttribute from key and value.
func buildFilterAttribute(key string, value any) types.FilterAttribute {
	return types.FilterAttribute{
		Key:   aws.String(key),
		Value: document.NewLazyDocument(value),
	}
}

// processFilters processes a list of filters and returns a filtered list.
func (kb *KnowledgeBase) processFilters(filters []Filter) ([]types.RetrievalFilter, error) {
	var filtersList []types.RetrievalFilter
	for _, f := range filters {
		filter, err := kb.getFilters(f)
		if err != nil {
			return nil, fmt.Errorf("failed to get filter: %w", err)
		}
		if filter != nil {
			filtersList = append(filtersList, filter)
		}
	}
	return filtersList, nil
}

// ErrEmptyFilterList is returned when the filter list is empty.
var ErrEmptyFilterList = fmt.Errorf("empty filter list")

// combineFilters combines a list of filters based on the number of filters.
func combineFilters(filtersList []types.RetrievalFilter, combiner func([]types.RetrievalFilter) types.RetrievalFilter) (types.RetrievalFilter, error) {
	switch len(filtersList) {
	case 0:
		return nil, ErrEmptyFilterList
	case 1:
		return filtersList[0], nil
	default:
		return combiner(filtersList), nil
	}
}

// ErrNoFilters is returned when no filters are provided.
var ErrNoFilters = fmt.Errorf("no filters provided")

func (kb *KnowledgeBase) getFilters(filters any) (types.RetrievalFilter, error) {
	if filters == nil {
		return nil, ErrNoFilters
	}

	switch f := filters.(type) {
	case EqualsFilter:
		return &types.RetrievalFilterMemberEquals{
			Value: buildFilterAttribute(f.Key, f.Value),
		}, nil
	case NotEqualsFilter:
		return &types.RetrievalFilterMemberNotEquals{
			Value: buildFilterAttribute(f.Key, f.Value),
		}, nil
	case ContainsFilter:
		return &types.RetrievalFilterMemberListContains{
			Value: buildFilterAttribute(f.Key, f.Value),
		}, nil
	case AllFilter:
		filtersList, err := kb.processFilters(f.Filters)
		if err != nil {
			return nil, err
		}
		return combineFilters(filtersList, func(filters []types.RetrievalFilter) types.RetrievalFilter {
			return &types.RetrievalFilterMemberAndAll{Value: filters}
		})
	case AnyFilter:
		filtersList, err := kb.processFilters(f.Filters)
		if err != nil {
			return nil, err
		}
		return combineFilters(filtersList, func(filters []types.RetrievalFilter) types.RetrievalFilter {
			return &types.RetrievalFilterMemberOrAll{Value: filters}
		})
	default:
		return nil, fmt.Errorf("unsupported filter type: %T", filters)
	}
}

func (kb *KnowledgeBase) parseMetadata(retrievalResult types.KnowledgeBaseRetrievalResult) (map[string]any, error) {
	var res map[string]any

	if retrievalResult.Metadata != nil {
		res = make(map[string]any)
		keyCount := 0
		for k, v := range retrievalResult.Metadata {
			value, err := kb.unmarshalMetadataValue(v)
			if err != nil {
				return nil, err
			}
			res[k] = value
			if !strings.HasPrefix(k, "x-amz-bedrock-kb") {
				keyCount++
			}
		}
		if keyCount > 0 {
			res["metadata-source-uri"] = aws.ToString(retrievalResult.Location.S3Location.Uri) + _metadataSuffix
		}
	}

	return res, nil
}

func (kb *KnowledgeBase) unmarshalMetadataValue(value document.Interface) (any, error) {
	var v any
	if err := value.UnmarshalSmithyDocument(&v); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata value: %w", err)
	}

	// Handle specific types that need conversion
	if number, ok := v.(smithyDocument.Number); ok {
		floatValue, err := number.Float32()
		if err != nil {
			return nil, err
		}
		return float32(floatValue), nil
	}

	return v, nil
}
