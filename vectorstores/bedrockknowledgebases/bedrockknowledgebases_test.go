package bedrockknowledgebases

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
	runtimetypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

type testModel struct{}

var _ llms.Model = testModel{}

func (testModel) GenerateContent(_ context.Context, _ []llms.MessageContent, _ ...llms.CallOption) (*llms.ContentResponse, error) {
	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{Content: "orange"},
		},
	}, nil
}

func (testModel) Call(_ context.Context, _ string, _ ...llms.CallOption) (string, error) {
	return "orange", nil
}

func TestMain(m *testing.M) {
	_initialPollInterval = time.Millisecond
	_maxPollInterval = time.Millisecond
	os.Exit(m.Run())
}

type testBedrockAgent struct {
	calls int
	mu    sync.Mutex

	ingestionJobRunning    bool
	neverSynced            bool
	ingestionJobStatus     types.IngestionJobStatus
	documentStatus         types.DocumentStatus
	ingestErr              error
	onIngest               func()
	ingestDocumentsCalls   int
	startIngestionJobCalls int
}

var _ bedrockAgentAPI = &testBedrockAgent{}

func (t *testBedrockAgent) GetKnowledgeBase(_ context.Context, _ *bedrockagent.GetKnowledgeBaseInput, _ ...func(*bedrockagent.Options)) (*bedrockagent.GetKnowledgeBaseOutput, error) {
	t.calls++
	return &bedrockagent.GetKnowledgeBaseOutput{
		KnowledgeBase: &types.KnowledgeBase{
			KnowledgeBaseId: aws.String("testKbId"),
		},
	}, nil
}

func (t *testBedrockAgent) ListDataSources(_ context.Context, params *bedrockagent.ListDataSourcesInput, _ ...func(*bedrockagent.Options)) (*bedrockagent.ListDataSourcesOutput, error) {
	t.calls++
	switch aws.ToString(params.KnowledgeBaseId) {
	case "testKbWithoutDatasources":
		return &bedrockagent.ListDataSourcesOutput{}, nil
	case "testKbWithOneS3Datasource":
		return &bedrockagent.ListDataSourcesOutput{
			DataSourceSummaries: []types.DataSourceSummary{
				{
					DataSourceId: aws.String("testS3DatasourceID"),
				},
			},
		}, nil
	case "testKbWithTwoS3Datasource":
		return &bedrockagent.ListDataSourcesOutput{
			DataSourceSummaries: []types.DataSourceSummary{
				{
					DataSourceId: aws.String("testS3DatasourceID"),
				},
				{
					DataSourceId: aws.String("testS3DatasourceID"),
				},
			},
		}, nil
	case "testKbWithOneNotS3Datasource":
		return &bedrockagent.ListDataSourcesOutput{
			DataSourceSummaries: []types.DataSourceSummary{
				{
					DataSourceId: aws.String("testNotS3DatasourceID"),
				},
			},
		}, nil
	case "testKbWithTwoMixedDatasources":
		return &bedrockagent.ListDataSourcesOutput{
			DataSourceSummaries: []types.DataSourceSummary{
				{
					DataSourceId: aws.String("testS3DatasourceID"),
				},
				{
					DataSourceId: aws.String("testNotS3DatasourceID"),
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown knowledge base id: %s", aws.ToString(params.KnowledgeBaseId))
	}
}

func (t *testBedrockAgent) GetDataSource(_ context.Context, params *bedrockagent.GetDataSourceInput, _ ...func(*bedrockagent.Options)) (*bedrockagent.GetDataSourceOutput, error) {
	t.mu.Lock()
	t.calls++
	t.mu.Unlock()
	if aws.ToString(params.DataSourceId) == "testS3DatasourceID" {
		return &bedrockagent.GetDataSourceOutput{
			DataSource: &types.DataSource{
				DataSourceId: aws.String("testS3DatasourceID"),
				DataSourceConfiguration: &types.DataSourceConfiguration{
					Type: types.DataSourceTypeS3,
					S3Configuration: &types.S3DataSourceConfiguration{
						BucketArn: aws.String("arn:aws:s3:::testBucket"),
					},
				},
			},
		}, nil
	}
	return &bedrockagent.GetDataSourceOutput{
		DataSource: &types.DataSource{
			DataSourceConfiguration: &types.DataSourceConfiguration{},
		},
	}, nil
}

func (t *testBedrockAgent) IngestKnowledgeBaseDocuments(_ context.Context, _ *bedrockagent.IngestKnowledgeBaseDocumentsInput, _ ...func(*bedrockagent.Options)) (*bedrockagent.IngestKnowledgeBaseDocumentsOutput, error) {
	t.calls++
	t.ingestDocumentsCalls++
	if t.onIngest != nil {
		t.onIngest()
	}
	if t.ingestErr != nil {
		return nil, t.ingestErr
	}
	return &bedrockagent.IngestKnowledgeBaseDocumentsOutput{}, nil
}

func (t *testBedrockAgent) StartIngestionJob(_ context.Context, _ *bedrockagent.StartIngestionJobInput, _ ...func(*bedrockagent.Options)) (*bedrockagent.StartIngestionJobOutput, error) {
	t.calls++
	t.startIngestionJobCalls++
	return &bedrockagent.StartIngestionJobOutput{
		IngestionJob: &types.IngestionJob{
			IngestionJobId: aws.String("testIngestionJobId"),
		},
	}, nil
}

func (t *testBedrockAgent) GetIngestionJob(_ context.Context, _ *bedrockagent.GetIngestionJobInput, _ ...func(*bedrockagent.Options)) (*bedrockagent.GetIngestionJobOutput, error) {
	t.calls++
	status := t.ingestionJobStatus
	if status == "" {
		status = types.IngestionJobStatusComplete
	}
	return &bedrockagent.GetIngestionJobOutput{
		IngestionJob: &types.IngestionJob{
			Status: status,
		},
	}, nil
}

func (t *testBedrockAgent) ListIngestionJobs(_ context.Context, params *bedrockagent.ListIngestionJobsInput, _ ...func(*bedrockagent.Options)) (*bedrockagent.ListIngestionJobsOutput, error) {
	t.calls++
	values := params.Filters[0].Values
	if len(values) != 1 {
		return nil, fmt.Errorf("ListIngestionJobs matches a status filter against its first value only, got %v", values)
	}
	status := types.IngestionJobStatus(values[0])
	if (status == types.IngestionJobStatusInProgress && t.ingestionJobRunning) ||
		(status == types.IngestionJobStatusComplete && !t.neverSynced) {
		return &bedrockagent.ListIngestionJobsOutput{
			IngestionJobSummaries: []types.IngestionJobSummary{{IngestionJobId: aws.String("testIngestionJobId")}},
		}, nil
	}
	return &bedrockagent.ListIngestionJobsOutput{}, nil
}

func (t *testBedrockAgent) GetKnowledgeBaseDocuments(_ context.Context, params *bedrockagent.GetKnowledgeBaseDocumentsInput, _ ...func(*bedrockagent.Options)) (*bedrockagent.GetKnowledgeBaseDocumentsOutput, error) {
	t.calls++
	status := t.documentStatus
	if status == "" {
		status = types.DocumentStatusIndexed
	}
	details := make([]types.KnowledgeBaseDocumentDetail, len(params.DocumentIdentifiers))
	for i, identifier := range params.DocumentIdentifiers {
		details[i] = types.KnowledgeBaseDocumentDetail{
			Identifier: &identifier,
			Status:     status,
		}
	}
	return &bedrockagent.GetKnowledgeBaseDocumentsOutput{DocumentDetails: details}, nil
}

type testBedrockAgentRuntime struct{ calls int }

var _ bedrockAgentRuntimeAPI = &testBedrockAgentRuntime{}

func (t *testBedrockAgentRuntime) Retrieve(_ context.Context, _ *bedrockagentruntime.RetrieveInput, _ ...func(*bedrockagentruntime.Options)) (*bedrockagentruntime.RetrieveOutput, error) {
	t.calls++
	return &bedrockagentruntime.RetrieveOutput{
		RetrievalResults: []runtimetypes.KnowledgeBaseRetrievalResult{
			{
				Content: &runtimetypes.RetrievalResultContent{
					Text: aws.String("The color of the house is blue."),
				},
				Score: aws.Float64(0.9),
			},
			{
				Content: &runtimetypes.RetrievalResultContent{
					Text: aws.String("The color of the car is red."),
				},
				Score: aws.Float64(0.8),
			},
			{
				Content: &runtimetypes.RetrievalResultContent{
					Text: aws.String("The color of the desk is orange."),
				},
				Score: aws.Float64(0.7),
			},
		},
	}, nil
}

type testSearchableRuntime struct {
	calls       int
	hiddenCalls int
	requested   [][]string
}

var _ bedrockAgentRuntimeAPI = &testSearchableRuntime{}

func (t *testSearchableRuntime) Retrieve(_ context.Context, params *bedrockagentruntime.RetrieveInput, _ ...func(*bedrockagentruntime.Options)) (*bedrockagentruntime.RetrieveOutput, error) {
	t.calls++
	filter := params.RetrievalConfiguration.VectorSearchConfiguration.Filter
	in, ok := filter.(*runtimetypes.RetrievalFilterMemberIn)
	if !ok || aws.ToString(in.Value.Key) != "x-amz-bedrock-kb-source-uri" {
		return nil, fmt.Errorf("expected an in filter on the source URI, got %#v", filter)
	}
	value, err := in.Value.Value.MarshalSmithyDocument()
	if err != nil {
		return nil, err
	}
	var uris []string
	if err := json.Unmarshal(value, &uris); err != nil {
		return nil, err
	}
	t.requested = append(t.requested, uris)
	if t.calls <= t.hiddenCalls {
		return &bedrockagentruntime.RetrieveOutput{}, nil
	}
	return &bedrockagentruntime.RetrieveOutput{
		RetrievalResults: []runtimetypes.KnowledgeBaseRetrievalResult{{
			Location: &runtimetypes.RetrievalResultLocation{
				S3Location: &runtimetypes.RetrievalResultS3Location{Uri: aws.String(uris[0])},
			},
		}},
	}, nil
}

type testS3Client struct {
	calls int
	mu    sync.Mutex

	failingPutKey string
	deleteErr     error
	deleted       []string
}

var _ s3API = &testS3Client{}

func (t *testS3Client) PutObject(_ context.Context, params *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.calls++
	if aws.ToString(params.Key) == t.failingPutKey {
		return nil, errors.New("put failed")
	}
	return &s3.PutObjectOutput{}, nil
}

func (t *testS3Client) DeleteObject(_ context.Context, params *s3.DeleteObjectInput, _ ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.calls++
	if t.deleteErr != nil {
		return nil, t.deleteErr
	}
	t.deleted = append(t.deleted, aws.ToString(params.Key))
	return &s3.DeleteObjectOutput{}, nil
}

func TestKnowledgeBaseAddDocuments(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	testBedrockAgent := &testBedrockAgent{}
	s3Client := &testS3Client{}
	kb := newFromClients("testKbWithOneS3Datasource", testBedrockAgent, &testBedrockAgentRuntime{}, s3Client)

	ids, err := kb.AddDocuments(ctx, []schema.Document{{PageContent: "Mock"}})
	require.NoError(t, err)
	require.Equal(t, 9, testBedrockAgent.calls, "expected 9 calls to testBedrockAgent")
	require.Equal(t, 1, s3Client.calls, "expected 1 call to s3Client")
	require.Len(t, ids, 1, "expected 1 id")
}

func TestKnowledgeBaseAddDocumentsWithoutDs(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	testBedrockAgent := &testBedrockAgent{}
	s3Client := &testS3Client{}
	kb := newFromClients("testKbWithoutDatasources", testBedrockAgent, &testBedrockAgentRuntime{}, s3Client)

	_, err := kb.AddDocuments(ctx, []schema.Document{{PageContent: "Mock"}})
	require.Error(t, err, "expected error because knowledge base has no datasources")
}

func TestKnowledgeBaseAddDocumentsWithWrongDsId(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	testBedrockAgent := &testBedrockAgent{}
	s3Client := &testS3Client{}
	kb := newFromClients("testKbWithOneS3Datasource", testBedrockAgent, &testBedrockAgentRuntime{}, s3Client)

	_, err := kb.AddDocuments(ctx, []schema.Document{{PageContent: "Mock"}}, vectorstores.WithNameSpace("wrongDatasourceID"))
	require.Error(t, err, "expected error because wrongDatasourceID is not valid")
}

func TestKnowledgeBaseAddDocumentsWithMultipleDs(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	testBedrockAgent := &testBedrockAgent{}
	s3Client := &testS3Client{}
	kb := newFromClients("testKbWithTwoS3Datasource", testBedrockAgent, &testBedrockAgentRuntime{}, s3Client)

	ids, err := kb.AddDocuments(ctx, []schema.Document{{PageContent: "Mock"}}, vectorstores.WithNameSpace("testS3DatasourceID"))
	require.NoError(t, err)
	require.Equal(t, 10, testBedrockAgent.calls, "expected 10 calls to testBedrockAgent")
	require.Equal(t, 1, s3Client.calls, "expected 1 call to s3Client")
	require.Len(t, ids, 1, "expected 1 id")
}

func TestKnowledgeBaseAddDocumentsWithMultipleMixedDs(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	testBedrockAgent := &testBedrockAgent{}
	s3Client := &testS3Client{}
	kb := newFromClients("testKbWithTwoMixedDatasources", testBedrockAgent, &testBedrockAgentRuntime{}, s3Client)

	ids, err := kb.AddDocuments(ctx, []schema.Document{{PageContent: "Mock"}})
	require.NoError(t, err)
	require.Equal(t, 10, testBedrockAgent.calls, "expected 10 calls to testBedrockAgent")
	require.Equal(t, 1, s3Client.calls, "expected 1 call to s3Client")
	require.Len(t, ids, 1, "expected 1 id")
}

func TestKnowledgeBaseAddDocumentsWithMultipleS3DsAndWithoutDsID(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	kb := newFromClients("testKbWithTwoS3Datasource", &testBedrockAgent{}, &testBedrockAgentRuntime{}, &testS3Client{})

	_, err := kb.AddDocuments(ctx, []schema.Document{{PageContent: "Mock"}})
	require.Error(t, err, "expected error because vectorstores.WithNameSpace is required")
}

func TestKnowledgeBaseAddNamedDocuments(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	testBedrockAgent := &testBedrockAgent{}
	s3Client := &testS3Client{}
	kb := newFromClients("testKbWithOneS3Datasource", testBedrockAgent, &testBedrockAgentRuntime{}, s3Client)

	ids, err := kb.AddNamedDocuments(
		ctx,
		[]NamedDocument{
			{
				Document: schema.Document{PageContent: "Mock"},
				Name:     "Mock",
			},
		},
	)
	require.NoError(t, err)
	require.Equal(t, 9, testBedrockAgent.calls, "expected 9 calls to testBedrockAgent")
	require.Equal(t, 1, s3Client.calls, "expected 1 call to s3Client")
	require.Len(t, ids, 1, "expected 1 id")
	require.Equal(t, "Mock", ids[0], "expected id to be 'Mock'")
}

func TestStartSync(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	agent := &testBedrockAgent{}
	kb := newFromClients("testKbId", agent, &testBedrockAgentRuntime{}, &testS3Client{})

	err := kb.StartSync(ctx, "testS3DatasourceID")
	require.NoError(t, err)
	require.Equal(t, 2, agent.calls, "expected 2 calls: StartIngestionJob + GetIngestionJob")
}

func TestStartSyncWithContextDone(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	t.Parallel()

	agent := &testBedrockAgent{ingestionJobStatus: types.IngestionJobStatusInProgress}
	kb := newFromClients("testKbId", agent, &testBedrockAgentRuntime{}, &testS3Client{})

	err := kb.StartSync(ctx, "testS3DatasourceID")
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestKnowledgeBaseAddDocumentsWithIngestionJobInProgress(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	testBedrockAgent := &testBedrockAgent{ingestionJobRunning: true}
	s3Client := &testS3Client{}
	kb := newFromClients("testKbWithOneS3Datasource", testBedrockAgent, &testBedrockAgentRuntime{}, s3Client)

	_, err := kb.AddDocuments(ctx, []schema.Document{{PageContent: "Mock"}})
	require.ErrorIs(t, err, ErrIngestionJobInProgress)
	require.Equal(t, 0, s3Client.calls, "expected no upload to S3")
}

func TestKnowledgeBaseAddDocumentsWithDataSourceNeverSynced(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	testBedrockAgent := &testBedrockAgent{neverSynced: true}
	kb := newFromClients("testKbWithOneS3Datasource", testBedrockAgent, &testBedrockAgentRuntime{}, &testS3Client{})

	ids, err := kb.AddDocuments(ctx, []schema.Document{{PageContent: "Mock"}})
	require.NoError(t, err)
	require.Len(t, ids, 1, "expected 1 id")
	require.Equal(t, 1, testBedrockAgent.startIngestionJobCalls, "expected the first sync of the data source")
	require.Equal(t, 0, testBedrockAgent.ingestDocumentsCalls, "expected no direct ingestion")
}

func TestKnowledgeBaseAddDocumentsWithDocumentNotIndexed(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	testBedrockAgent := &testBedrockAgent{documentStatus: types.DocumentStatusIgnored}
	s3Client := &testS3Client{}
	kb := newFromClients("testKbWithOneS3Datasource", testBedrockAgent, &testBedrockAgentRuntime{}, s3Client)

	_, err := kb.AddNamedDocuments(ctx, []NamedDocument{{Name: "Mock", Document: schema.Document{PageContent: "Mock"}}})
	require.ErrorContains(t, err, "s3://testBucket/Mock (IGNORED")
	require.Empty(t, s3Client.deleted, "expected no removal from S3")
}

func TestKnowledgeBaseAddDocumentsWithIngestionFailure(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	testBedrockAgent := &testBedrockAgent{ingestErr: errors.New("ingestion failed")}
	s3Client := &testS3Client{}
	kb := newFromClients("testKbWithOneS3Datasource", testBedrockAgent, &testBedrockAgentRuntime{}, s3Client)

	_, err := kb.AddNamedDocuments(ctx, []NamedDocument{{Name: "Mock", Document: schema.Document{PageContent: "Mock"}}})
	require.ErrorContains(t, err, "ingestion failed")
	require.Equal(t, []string{"Mock"}, s3Client.deleted, "expected the uploaded document to be removed from S3")
}

func TestKnowledgeBaseAddDocumentsWithIngestionFailureAfterContextDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	t.Parallel()

	testBedrockAgent := &testBedrockAgent{ingestErr: context.Canceled, onIngest: cancel}
	s3Client := &testS3Client{}
	kb := newFromClients("testKbWithOneS3Datasource", testBedrockAgent, &testBedrockAgentRuntime{}, s3Client)

	_, err := kb.AddNamedDocuments(ctx, []NamedDocument{{Name: "Mock", Document: schema.Document{PageContent: "Mock"}}})
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, []string{"Mock"}, s3Client.deleted, "expected the uploaded document to be removed from S3")
}

func TestKnowledgeBaseAddDocumentsWithFailedRemoval(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	testBedrockAgent := &testBedrockAgent{ingestErr: errors.New("ingestion failed")}
	s3Client := &testS3Client{deleteErr: errors.New("delete failed")}
	kb := newFromClients("testKbWithOneS3Datasource", testBedrockAgent, &testBedrockAgentRuntime{}, s3Client)

	_, err := kb.AddNamedDocuments(ctx, []NamedDocument{{Name: "Mock", Document: schema.Document{PageContent: "Mock"}}})
	require.ErrorContains(t, err, "ingestion failed")
	require.ErrorContains(t, err, "Mock, left in bucket testBucket")
}

func TestKnowledgeBaseAddDocumentsWithoutWaitUntilSearchable(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	runtime := &testSearchableRuntime{}
	kb := newFromClients("testKbWithOneS3Datasource", &testBedrockAgent{}, runtime, &testS3Client{})

	_, err := kb.AddDocuments(ctx, []schema.Document{{PageContent: "Mock"}})
	require.NoError(t, err)
	require.Equal(t, 0, runtime.calls, "expected no Retrieve request")
}

func TestKnowledgeBaseAddDocumentsWithWaitUntilSearchable(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	runtime := &testSearchableRuntime{hiddenCalls: 1}
	kb := newFromClients("testKbWithOneS3Datasource", &testBedrockAgent{}, runtime, &testS3Client{})
	WithWaitUntilSearchable()(kb)

	_, err := kb.AddNamedDocuments(ctx, []NamedDocument{
		{Name: "Mock1", Document: schema.Document{PageContent: "Mock"}},
		{Name: "Mock2", Document: schema.Document{PageContent: "Mock"}},
	})
	require.NoError(t, err)
	require.Equal(t, [][]string{
		{"s3://testBucket/Mock1", "s3://testBucket/Mock2"},
		{"s3://testBucket/Mock1", "s3://testBucket/Mock2"},
		{"s3://testBucket/Mock2"},
	}, runtime.requested, "expected each request to ask only for the documents not searchable yet")
}

func TestKnowledgeBaseAddDocumentsWithDataSourceNeverSyncedAndWaitUntilSearchable(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	runtime := &testSearchableRuntime{}
	kb := newFromClients("testKbWithOneS3Datasource", &testBedrockAgent{neverSynced: true}, runtime, &testS3Client{})
	WithWaitUntilSearchable()(kb)

	_, err := kb.AddDocuments(ctx, []schema.Document{{PageContent: "Mock"}})
	require.NoError(t, err)
	require.Equal(t, 1, runtime.calls, "expected the first sync to wait until the document is searchable")
}

func TestKnowledgeBaseAddDocumentsWithWaitUntilSearchableAfterContextDone(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	t.Parallel()

	kb := newFromClients("testKbWithOneS3Datasource", &testBedrockAgent{}, &testSearchableRuntime{hiddenCalls: math.MaxInt}, &testS3Client{})
	WithWaitUntilSearchable()(kb)

	_, err := kb.AddNamedDocuments(ctx, []NamedDocument{{Name: "Mock", Document: schema.Document{PageContent: "Mock"}}})
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.ErrorContains(t, err, "s3://testBucket/Mock")
}

func TestKnowledgeBaseUploadNamedDocumentsWithFailedUpload(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	s3Client := &testS3Client{failingPutKey: "Mock" + _metadataSuffix}
	kb := newFromClients("testKbWithOneS3Datasource", &testBedrockAgent{}, &testBedrockAgentRuntime{}, s3Client)

	_, err := kb.UploadNamedDocuments(ctx, []NamedDocument{{
		Name:     "Mock",
		Document: schema.Document{PageContent: "Mock", Metadata: map[string]any{"type": "tutorial"}},
	}})
	require.ErrorContains(t, err, "put failed")
	require.Equal(t, []string{"Mock"}, s3Client.deleted, "expected only the uploaded object to be removed from S3")
}

func TestKnowledgeBaseUploadDocuments(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	testBedrockAgent := &testBedrockAgent{}
	s3Client := &testS3Client{}
	kb := newFromClients("testKbWithOneS3Datasource", testBedrockAgent, &testBedrockAgentRuntime{}, s3Client)

	ids, err := kb.UploadDocuments(ctx, []schema.Document{{PageContent: "Mock"}})
	require.NoError(t, err)
	require.Len(t, ids, 1, "expected 1 id")
	require.True(t, strings.HasPrefix(ids[0], "kb_doc_"), "expected a generated name")
	require.Equal(t, 1, s3Client.calls, "expected 1 call to s3Client")
	require.Equal(t, 0, testBedrockAgent.ingestDocumentsCalls, "expected no direct ingestion")
}

func TestKnowledgeBaseSimilaritySearch(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	kb := newFromClients("testKbId", &testBedrockAgent{}, &testBedrockAgentRuntime{}, &testS3Client{})

	docs, err := kb.SimilaritySearch(ctx, "What color is the desk?", 5)
	require.NoError(t, err)
	require.Len(t, docs, 3, "expected 3 documents")
	require.Equal(t, "The color of the house is blue.", docs[0].PageContent, "expected content to be 'The color of the house is blue.'")
	require.Equal(t, "The color of the car is red.", docs[1].PageContent, "expected content to be 'The color of the car is red.'")
	require.Equal(t, "The color of the desk is orange.", docs[2].PageContent, "expected content to be 'The color of the desk is orange.'")
}

func TestKnowledgeBaseSimilaritySearchWithScoreThreshold(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	kb := newFromClients("testKbId", &testBedrockAgent{}, &testBedrockAgentRuntime{}, &testS3Client{})

	docs, err := kb.SimilaritySearch(ctx, "What color is the desk?", 5, vectorstores.WithScoreThreshold(0.8))
	require.NoError(t, err)
	require.Len(t, docs, 2, "expected 2 documents")
	require.Equal(t, "The color of the house is blue.", docs[0].PageContent, "expected content to be 'The color of the house is blue.'")
	require.Equal(t, "The color of the car is red.", docs[1].PageContent, "expected content to be 'The color of the car is red.'")
}

type retrieveResponseClient struct{ body string }

func (c retrieveResponseClient) Do(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(c.body)),
	}, nil
}

func TestKnowledgeBaseSimilaritySearchWithNumericMetadata(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	runtimeClient := bedrockagentruntime.New(bedrockagentruntime.Options{
		Region:      "us-east-1",
		Credentials: aws.AnonymousCredentials{},
		HTTPClient: retrieveResponseClient{body: `{"retrievalResults": [{
			"content": {"text": "The color of the house is blue."},
			"location": {"type": "S3", "s3Location": {"uri": "s3://testBucket/house.txt"}},
			"metadata": {"document-chunk-index": 2},
			"score": 0.9
		}]}`},
	})
	kb := newFromClients("testKbId", &testBedrockAgent{}, runtimeClient, &testS3Client{})

	docs, err := kb.SimilaritySearch(ctx, "What color is the house?", 5)
	require.NoError(t, err)
	require.Len(t, docs, 1, "expected 1 document")
	require.Equal(t, float32(2), docs[0].Metadata["document-chunk-index"], "expected numeric metadata as float32")
}

func TestKnowledgeBaseSimilaritySearchWithFilter(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	kb := newFromClients("testKbId", &testBedrockAgent{}, &testBedrockAgentRuntime{}, &testS3Client{})

	_, err := kb.SimilaritySearch(ctx, "What color is the desk?", 5, vectorstores.WithFilters(EqualsFilter{Key: "color", Value: "orange"}))
	require.NoError(t, err)
}

func TestKnowledgeBaseSimilaritySearchWrongWithFilter(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	kb := newFromClients("testKbId", &testBedrockAgent{}, &testBedrockAgentRuntime{}, &testS3Client{})

	_, err := kb.SimilaritySearch(ctx, "What color is the desk?", 5, vectorstores.WithFilters("wrongFilter"))
	require.Error(t, err, "expected error because of wrong filter format")
}

type trackSearchResults struct {
	*KnowledgeBase
	track []schema.Document
}

var _ vectorstores.VectorStore = &trackSearchResults{}

func newTrackSearchResults(kb *KnowledgeBase) *trackSearchResults {
	return &trackSearchResults{kb, nil}
}

func (t *trackSearchResults) SimilaritySearch(ctx context.Context, query string, k int, options ...vectorstores.Option) ([]schema.Document, error) {
	res, err := t.KnowledgeBase.SimilaritySearch(ctx, query, k, options...)
	if err != nil {
		return nil, err
	}
	t.track = res
	return res, nil
}

func TestKnowledgeBaseAsRetriever(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	testBedrockAgentRuntime := &testBedrockAgentRuntime{}
	kb := newTrackSearchResults(newFromClients("testKbId", &testBedrockAgent{}, testBedrockAgentRuntime, &testS3Client{}))

	result, err := chains.Run(
		ctx,
		chains.NewRetrievalQAFromLLM(
			testModel{},
			vectorstores.ToRetriever(kb, 1),
		),
		"What color is the desk?",
	)
	require.NoError(t, err)
	require.Equal(t, "orange", result, "expected result to be orange")
	require.Equal(t, 1, testBedrockAgentRuntime.calls, "expected testBedrockAgentRuntime to be called once")
	require.Len(t, kb.track, 3, "expected 3 documents")
	require.Equal(t, "The color of the house is blue.", kb.track[0].PageContent, "expected content to be 'The color of the house is blue.'")
	require.Equal(t, "The color of the car is red.", kb.track[1].PageContent, "expected content to be 'The color of the car is red.'")
	require.Equal(t, "The color of the desk is orange.", kb.track[2].PageContent, "expected content to be 'The color of the desk is orange.'")
}

func TestKnowledgeBaseAsRetrieverWithScoreThreshold(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	testBedrockAgentRuntime := &testBedrockAgentRuntime{}
	kb := newTrackSearchResults(newFromClients("testKbId", &testBedrockAgent{}, testBedrockAgentRuntime, &testS3Client{}))

	result, err := chains.Run(
		ctx,
		chains.NewRetrievalQAFromLLM(
			testModel{},
			vectorstores.ToRetriever(kb, 1, vectorstores.WithScoreThreshold(0.8)),
		),
		"What color is the desk?",
	)
	require.NoError(t, err)
	require.Equal(t, "orange", result, "expected result to be orange")
	require.Equal(t, 1, testBedrockAgentRuntime.calls, "expected testBedrockAgentRuntime to be called once")
	require.Len(t, kb.track, 2, "expected 2 documents")
	require.Equal(t, "The color of the house is blue.", kb.track[0].PageContent, "expected content to be 'The color of the house is blue.'")
	require.Equal(t, "The color of the car is red.", kb.track[1].PageContent, "expected content to be 'The color of the car is red.'")
}
