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
	_metadataSuffix = ".metadata.json"
	// _documentsBatchSize is the number of documents per IngestKnowledgeBaseDocuments and
	// GetKnowledgeBaseDocuments request; GetKnowledgeBaseDocuments accepts at most 10.
	_documentsBatchSize = 10
)

// Variables rather than constants, to let the tests shorten the waits.
var (
	_initialPollInterval = time.Second
	_maxPollInterval     = 10 * time.Second
)

// ErrIngestionJobInProgress is returned by AddDocuments and AddNamedDocuments when an ingestion job
// is running on the data source.
var ErrIngestionJobInProgress = errors.New("an ingestion job is in progress on the data source")

var _runningIngestionJobStatuses = []types.IngestionJobStatus{
	types.IngestionJobStatusStarting,
	types.IngestionJobStatusInProgress,
	types.IngestionJobStatusStopping,
}

// dataSource represents a data source for a knowledge base.
type dataSource struct {
	ID        string
	BucketARN string
}

func (kb *KnowledgeBase) hash(docs []NamedDocument) string {
	var hashInput bytes.Buffer
	for _, doc := range docs {
		hashInput.WriteString(doc.Document.PageContent)
	}

	hasher := blake3.New()
	_, _ = hasher.Write(hashInput.Bytes()) // hash.Hash.Write never returns an error
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
					ID:        aws.ToString(res.DataSource.DataSourceId),
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

	for i := 0; i < len(docs); i += _documentsBatchSize {
		end := i + _documentsBatchSize
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

// StartSync triggers a Bedrock ingestion job for the given data source and waits for it to complete.
// Call this once after all documents have been uploaded across all ingestion tasks.
// The wait lasts until ctx is done; the returned error then wraps the error of ctx.
func (kb *KnowledgeBase) StartSync(ctx context.Context, datasourceID string) error {
	ingestionJobID, err := kb.startIngestionJob(ctx, datasourceID)
	if err != nil {
		return fmt.Errorf(
			"failed to start ingestion job, please retry or manually sync from the AWS console: %w",
			err,
		)
	}
	if err := kb.checkIngestionJobStatus(ctx, datasourceID, ingestionJobID); err != nil {
		return fmt.Errorf(
			"failed to check ingestion status, please retry or manually sync from the AWS console: %w",
			err,
		)
	}
	return nil
}

func (kb *KnowledgeBase) checkIngestionJobStatus(ctx context.Context, datasourceID, ingestionJobID string) error {
	return poll(ctx, func(ctx context.Context) (bool, error) {
		result, err := kb.bedrockAgent.GetIngestionJob(ctx, &bedrockagent.GetIngestionJobInput{
			KnowledgeBaseId: aws.String(kb.knowledgeBaseID),
			DataSourceId:    aws.String(datasourceID),
			IngestionJobId:  aws.String(ingestionJobID),
		})
		if err != nil {
			return false, fmt.Errorf("failed to get ingestion job: %w", err)
		}
		switch result.IngestionJob.Status {
		case types.IngestionJobStatusFailed:
			return false, fmt.Errorf("ingestion job failed: %v", result.IngestionJob.FailureReasons)
		case types.IngestionJobStatusComplete:
			return true, nil
		case types.IngestionJobStatusStopped:
			return false, errors.New("ingestion job stopped")
		}
		return false, nil
	})
}

// poll calls check until it reports done or fails, or until ctx is done. The first call happens after
// _initialPollInterval, and the interval doubles after each call up to _maxPollInterval.
func poll(ctx context.Context, check func(ctx context.Context) (done bool, err error)) error {
	interval := _initialPollInterval
	timer := time.NewTimer(interval)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			done, err := check(ctx)
			if err != nil || done {
				return err
			}
			interval = min(2*interval, _maxPollInterval)
			timer.Reset(interval)
		}
	}
}

func (kb *KnowledgeBase) hasRunningIngestionJob(ctx context.Context, datasourceID string) (bool, error) {
	for _, status := range _runningIngestionJobStatuses {
		found, err := kb.hasIngestionJob(ctx, datasourceID, status)
		if err != nil || found {
			return found, err
		}
	}
	return false, nil
}

func (kb *KnowledgeBase) hasCompletedIngestionJob(ctx context.Context, datasourceID string) (bool, error) {
	return kb.hasIngestionJob(ctx, datasourceID, types.IngestionJobStatusComplete)
}

// hasIngestionJob reports whether the data source has at least one ingestion job in status.
// ListIngestionJobs matches a status filter against its first value only, so each status needs its own call.
func (kb *KnowledgeBase) hasIngestionJob(ctx context.Context, datasourceID string, status types.IngestionJobStatus) (bool, error) {
	result, err := kb.bedrockAgent.ListIngestionJobs(ctx, &bedrockagent.ListIngestionJobsInput{
		KnowledgeBaseId: aws.String(kb.knowledgeBaseID),
		DataSourceId:    aws.String(datasourceID),
		Filters: []types.IngestionJobFilter{{
			Attribute: types.IngestionJobFilterAttributeStatus,
			Operator:  types.IngestionJobFilterOperatorEq,
			Values:    []string{string(status)},
		}},
		MaxResults: aws.Int32(1),
	})
	if err != nil {
		return false, fmt.Errorf("failed to list ingestion jobs: %w", err)
	}
	return len(result.IngestionJobSummaries) > 0, nil
}

// waitForIndexedDocuments waits until every document of docs has a final ingestion status. It returns an
// error that lists the documents whose final status is not INDEXED, with the reason Bedrock gives.
func (kb *KnowledgeBase) waitForIndexedDocuments(ctx context.Context, ds dataSource, docs []NamedDocument) error {
	bucketName := kb.getBucketName(ds.BucketARN)
	pending := make([]string, len(docs))
	for i, doc := range docs {
		pending[i] = "s3://" + bucketName + "/" + doc.Name
	}
	var notIndexed []string

	return poll(ctx, func(ctx context.Context) (bool, error) {
		var stillPending []string
		for i := 0; i < len(pending); i += _documentsBatchSize {
			batch := pending[i:min(i+_documentsBatchSize, len(pending))]
			identifiers := make([]types.DocumentIdentifier, len(batch))
			for j, uri := range batch {
				identifiers[j] = types.DocumentIdentifier{
					DataSourceType: types.ContentDataSourceTypeS3,
					S3:             &types.S3Location{Uri: aws.String(uri)},
				}
			}
			result, err := kb.bedrockAgent.GetKnowledgeBaseDocuments(ctx, &bedrockagent.GetKnowledgeBaseDocumentsInput{
				KnowledgeBaseId:     aws.String(kb.knowledgeBaseID),
				DataSourceId:        aws.String(ds.ID),
				DocumentIdentifiers: identifiers,
			})
			if err != nil {
				return false, fmt.Errorf("failed to get documents: %w", err)
			}

			reported := make(map[string]bool, len(result.DocumentDetails))
			for _, detail := range result.DocumentDetails {
				uri := aws.ToString(detail.Identifier.S3.Uri)
				reported[uri] = true
				switch detail.Status {
				case types.DocumentStatusStarting, types.DocumentStatusPending, types.DocumentStatusInProgress:
					stillPending = append(stillPending, uri)
				case types.DocumentStatusIndexed:
				default:
					notIndexed = append(notIndexed, fmt.Sprintf("%s (%s %s)", uri, detail.Status, aws.ToString(detail.StatusReason)))
				}
			}
			for _, uri := range batch {
				if !reported[uri] {
					stillPending = append(stillPending, uri)
				}
			}
		}

		pending = stillPending
		if len(pending) > 0 {
			return false, nil
		}
		if len(notIndexed) > 0 {
			return true, fmt.Errorf("documents not indexed: %s", strings.Join(notIndexed, ", "))
		}
		return true, nil
	})
}
