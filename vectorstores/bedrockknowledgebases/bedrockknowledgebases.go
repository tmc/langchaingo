package bedrockknowledgebases

import (
	"context"
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
	"github.com/0xDezzy/langchaingo/schema"
	"github.com/0xDezzy/langchaingo/vectorstores"
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

func (kb *KnowledgeBase) filterMetadata(docs []NamedDocument) {
	for i, doc := range docs {
		if doc.Document.Metadata != nil {
			for k, v := range doc.Document.Metadata {
				if v == nil {
					delete(doc.Document.Metadata, k)
					continue
				}

				rv := reflect.ValueOf(v)
				switch rv.Kind() {
				case reflect.Map:
					if rv.Len() == 0 {
						delete(doc.Document.Metadata, k)
					}
				case reflect.Slice:
					if rv.Len() == 0 {
						delete(doc.Document.Metadata, k)
					}
				case reflect.String:
					if v == "" {
						delete(doc.Document.Metadata, k)
					}
				}
			}

			if len(doc.Document.Metadata) == 0 {
				doc.Document.Metadata = nil
			}

			docs[i] = doc
		}
	}
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
	if len(compatibleDs) == 0 {
		if len(incompatibleDs) > 0 {
			return nil, fmt.Errorf(
				"found data sources but none with S3 type, please create a data source with S3 type for the knowledge base with id: %s in the AWS console",
				kb.knowledgeBaseID,
			)
		}
		return nil, fmt.Errorf(
			"no data sources with S3 type found, please create a data source with S3 type for the knowledge base with id: %s in the AWS console",
			kb.knowledgeBaseID,
		)
	}

	var datasourceID string
	var bucketARN string
	if opts.NameSpace != "" {
		for _, ds := range compatibleDs {
			if ds.ID == opts.NameSpace {
				datasourceID = ds.ID
				bucketARN = ds.BucketARN
				break
			}
		}
		if datasourceID == "" {
			return nil, fmt.Errorf("data source with S3 type with id %s not found", opts.NameSpace)
		}
	} else if len(compatibleDs) == 1 {
		datasourceID = compatibleDs[0].ID
		bucketARN = compatibleDs[0].BucketARN
	} else {
		return nil, fmt.Errorf("multiple data sources with S3 type found, please specify which one you want to use by passing its id with the `vectorstores.WithNameSpace` option")
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

	names := make([]string, len(docs))
	for i, doc := range docs {
		names[i] = doc.Name
	}
	return names, nil
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

	if filters, err := kb.getFilters(opts.Filters); err != nil {
		return nil, err
	} else if filters != nil {
		retrieveInput.RetrievalConfiguration.VectorSearchConfiguration.Filter = filters
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

func (kb *KnowledgeBase) getFilters(filters any) (types.RetrievalFilter, error) {
	if filters == nil {
		return nil, nil
	}

	switch filters := filters.(type) {
	case EqualsFilter:
		return &types.RetrievalFilterMemberEquals{
			Value: types.FilterAttribute{
				Key:   aws.String(filters.Key),
				Value: document.NewLazyDocument(filters.Value),
			},
		}, nil
	case NotEqualsFilter:
		return &types.RetrievalFilterMemberNotEquals{
			Value: types.FilterAttribute{
				Key:   aws.String(filters.Key),
				Value: document.NewLazyDocument(filters.Value),
			},
		}, nil
	case ContainsFilter:
		return &types.RetrievalFilterMemberListContains{
			Value: types.FilterAttribute{
				Key:   aws.String(filters.Key),
				Value: document.NewLazyDocument(filters.Value),
			},
		}, nil
	case AllFilter:
		return kb.buildCompositeFilter(filters.Filters, true)
	case AnyFilter:
		return kb.buildCompositeFilter(filters.Filters, false)
	default:
		return nil, fmt.Errorf("unsupported filter type: %T", filters)
	}
}

func (kb *KnowledgeBase) buildCompositeFilter(filters []Filter, isAll bool) (types.RetrievalFilter, error) {
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

	switch len(filtersList) {
	case 0:
		return nil, nil
	case 1:
		return filtersList[0], nil
	default:
		if isAll {
			return &types.RetrievalFilterMemberAndAll{Value: filtersList}, nil
		}
		return &types.RetrievalFilterMemberOrAll{Value: filtersList}, nil
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
	switch value := v.(type) {
	// convert to float 32 for easier handling by the user
	case smithyDocument.Number:
		floatValue, err := value.Float32()
		if err != nil {
			return nil, err
		}
		return float32(floatValue), nil
	}
	return v, nil
}
