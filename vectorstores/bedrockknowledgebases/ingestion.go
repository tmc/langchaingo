package bedrockknowledgebases

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/zeebo/blake3"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	eg "golang.org/x/sync/errgroup"
)

const (
	_metadataSuffix                      = ".metadata.json"
	_initialIngestionJobStatusRetryDelay = 1*time.Second + 250*time.Millisecond
)

// dataSource represents a data source for a knowledge base.
type dataSource struct {
	Id        string
	BucketARN string
}

func (kb *KnowledgeBase) hash(docs []NamedDocument) string {
	var hashInput bytes.Buffer
	for _, doc := range docs {
		hashInput.WriteString(doc.Document.PageContent)
	}

	hasher := blake3.New()
	hasher.Write(hashInput.Bytes())
	return hex.EncodeToString(hasher.Sum(nil))
}

func (kb *KnowledgeBase) checkKnowledgeBase(ctx context.Context) error {
	_, err := kb.bedrockAgent.GetKnowledgeBase(ctx, &bedrockagent.GetKnowledgeBaseInput{
		KnowledgeBaseId: aws.String(kb.knowledgeBaseID),
	})
	if err != nil {
		return fmt.Errorf("failed to get knowledge base: %w", err)
	}
	return nil
}

// listDataSources retrieves the list of data sources from Bedrock and returns the compatible and incompatible ones.
func (kb *KnowledgeBase) listDataSources(ctx context.Context) (compatible, incompatible []dataSource, err error) {
	result, err := kb.bedrockAgent.ListDataSources(ctx, &bedrockagent.ListDataSourcesInput{
		KnowledgeBaseId: aws.String(kb.knowledgeBaseID),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list data sources: %w", err)
	}

	getDataSources, ctx := eg.WithContext(ctx)
	var mu sync.Mutex
	for _, ds := range result.DataSourceSummaries {
		getDataSources.Go(func() error {
			res, err := kb.bedrockAgent.GetDataSource(ctx, &bedrockagent.GetDataSourceInput{
				KnowledgeBaseId: aws.String(kb.knowledgeBaseID),
				DataSourceId:    ds.DataSourceId,
			})
			if err != nil {
				return fmt.Errorf("failed to get datasource: %w", err)
			}

			mu.Lock()
			defer mu.Unlock()
			if res.DataSource.DataSourceConfiguration.Type == types.DataSourceTypeS3 {
				compatible = append(compatible, dataSource{
					Id:        aws.ToString(res.DataSource.DataSourceId),
					BucketARN: aws.ToString(res.DataSource.DataSourceConfiguration.S3Configuration.BucketArn),
				})
			} else {
				incompatible = append(incompatible, dataSource{})
			}
			return nil
		})
	}

	if err := getDataSources.Wait(); err != nil {
		return nil, nil, fmt.Errorf("failed to get data sources: %w", err)
	}
	return compatible, incompatible, nil
}

type ingestDocumentsRetryer struct {
	*retry.Standard
}

// hasReachedMaxNumberRequests checks if the error message indicates that the max number of requests has been reached.
func (r ingestDocumentsRetryer) hasReachedMaxNumberRequests(msg string) bool {
	return strings.Contains(msg, "reached max number")
}

func (r ingestDocumentsRetryer) hasReachedMaxConcurrency(msg string) bool {
	return strings.Contains(msg, "sum of concurrent")
}

func (r ingestDocumentsRetryer) IsErrorRetryable(err error) bool {
	var responseError *smithyhttp.ResponseError
	if ok := errors.As(err, &responseError); ok {
		if responseError.HTTPStatusCode() == http.StatusBadRequest &&
			(r.hasReachedMaxNumberRequests(responseError.Error()) || r.hasReachedMaxConcurrency(responseError.Error())) {
			return true
		}
	}
	return false
}

// ingestDocuments sends a request to Bedrock to ingest the provided documents in batches of 10.
// Returns an error if any request fails.
func (kb *KnowledgeBase) ingestDocuments(ctx context.Context, datasourceID, bucketArn string, docs []NamedDocument) error {
	bucketName := kb.getBucketName(bucketArn)
	const batchSize = 10

	for i := 0; i < len(docs); i += batchSize {
		end := i + batchSize
		if end > len(docs) {
			end = len(docs)
		}
		batchDocs := docs[i:end]

		docsToIngest := make([]types.KnowledgeBaseDocument, len(batchDocs))
		for j, doc := range batchDocs {
			docToIngest := types.KnowledgeBaseDocument{
				Content: &types.DocumentContent{
					DataSourceType: "S3",
					S3: &types.S3Content{
						S3Location: &types.S3Location{
							Uri: aws.String("s3://" + bucketName + "/" + doc.Name),
						},
					},
				},
			}
			if doc.Document.Metadata != nil {
				docToIngest.Metadata = &types.DocumentMetadata{
					Type: "S3_LOCATION",
					S3Location: &types.CustomS3Location{
						Uri: aws.String("s3://" + bucketName + "/" + doc.Name + _metadataSuffix),
					},
				}
			}
			docsToIngest[j] = docToIngest
		}

		// Create a hash for the current batch to guarantee idempotency.
		hash := kb.hash(batchDocs)

		_, err := kb.bedrockAgent.IngestKnowledgeBaseDocuments(
			ctx,
			&bedrockagent.IngestKnowledgeBaseDocumentsInput{
				KnowledgeBaseId: aws.String(kb.knowledgeBaseID),
				DataSourceId:    aws.String(datasourceID),
				ClientToken:     aws.String(hash),
				Documents:       docsToIngest,
			},
			func(o *bedrockagent.Options) {
				o.Retryer = ingestDocumentsRetryer{
					Standard: retry.NewStandard(func(o *retry.StandardOptions) {
						o.MaxAttempts = 10
						o.MaxBackoff = 30 * time.Second
					}),
				}
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

type startIngestionJobRetryer struct {
	*retry.Standard
}

// hasOngoingIngestDocsRequest checks if the error message indicates an ongoing KnowledgeBaseDocuments request.
func (r startIngestionJobRetryer) hasOngoingIngestDocsRequest(msg string) bool {
	return strings.Contains(msg, "ongoing KnowledgeBaseDocuments")
}

func (r startIngestionJobRetryer) IsErrorRetryable(err error) bool {
	var responseError *smithyhttp.ResponseError
	if ok := errors.As(err, &responseError); ok {
		if responseError.HTTPStatusCode() == http.StatusBadRequest &&
			r.hasOngoingIngestDocsRequest(responseError.Error()) {
			return true
		}
	}
	return false
}

// startIngestionJob sends an ingestion job request to Bedrock and retries with exponential backoff if
// the request fails because the KnowledgeBaseDocuments request is still ongoing.
func (kb *KnowledgeBase) startIngestionJob(ctx context.Context, datasourceID string) (string, error) {
	result, err := kb.bedrockAgent.StartIngestionJob(
		ctx,
		&bedrockagent.StartIngestionJobInput{
			KnowledgeBaseId: aws.String(kb.knowledgeBaseID),
			DataSourceId:    aws.String(datasourceID),
		},
		func(o *bedrockagent.Options) {
			o.Retryer = startIngestionJobRetryer{
				Standard: retry.NewStandard(func(o *retry.StandardOptions) {
					o.MaxAttempts = 6
					o.MaxBackoff = 7 * time.Second
				}),
			}
		})
	if err != nil {
		return "", fmt.Errorf("failed to start ingestion job: %w", err)
	}
	return aws.ToString(result.IngestionJob.IngestionJobId), nil
}

func (kb *KnowledgeBase) checkIngestionJobStatus(ctx context.Context, datasourceID, ingestionJobID string) error {
	maxRetries := 8
	delay := _initialIngestionJobStatusRetryDelay

	// If the ingestion job is still running, retry with exponential backoff
	for attempt := 1; attempt <= maxRetries; attempt++ {
		time.Sleep(delay)

		result, err := kb.bedrockAgent.GetIngestionJob(ctx, &bedrockagent.GetIngestionJobInput{
			KnowledgeBaseId: aws.String(kb.knowledgeBaseID),
			DataSourceId:    aws.String(datasourceID),
			IngestionJobId:  aws.String(ingestionJobID),
		})
		if err != nil {
			return fmt.Errorf("failed to get ingestion job: %w", err)
		}
		if result.IngestionJob.Status == types.IngestionJobStatusComplete {
			return nil
		}
		if result.IngestionJob.Status == "COMPLETE" {
			return nil
		}
		delay *= 2
	}

	return fmt.Errorf("exceeded maximum number of retries (%d)", maxRetries)
}
