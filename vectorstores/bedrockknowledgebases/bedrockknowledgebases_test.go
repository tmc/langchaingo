package bedrockknowledgebases

import (
	"context"
	"fmt"
	"sync"
	"testing"
	
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
	runtimetypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
	"github.com/yincongcyincong/langchaingo/chains"
	"github.com/yincongcyincong/langchaingo/llms"
	"github.com/yincongcyincong/langchaingo/schema"
	"github.com/yincongcyincong/langchaingo/vectorstores"
)

type testModel struct{}

var _ llms.Model = testModel{}

func (testModel) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{Content: "orange"},
		},
	}, nil
}

func (testModel) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return "orange", nil
}

type testBedrockAgent struct {
	calls int
	mu    sync.Mutex
}

var _ bedrockAgentAPI = &testBedrockAgent{}

func (t *testBedrockAgent) GetKnowledgeBase(ctx context.Context, params *bedrockagent.GetKnowledgeBaseInput, optFns ...func(*bedrockagent.Options)) (*bedrockagent.GetKnowledgeBaseOutput, error) {
	t.calls++
	return &bedrockagent.GetKnowledgeBaseOutput{
		KnowledgeBase: &types.KnowledgeBase{
			KnowledgeBaseId: aws.String("testKbId"),
		},
	}, nil
}

func (t *testBedrockAgent) ListDataSources(ctx context.Context, params *bedrockagent.ListDataSourcesInput, optFns ...func(*bedrockagent.Options)) (*bedrockagent.ListDataSourcesOutput, error) {
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

func (t *testBedrockAgent) GetDataSource(ctx context.Context, params *bedrockagent.GetDataSourceInput, optFns ...func(*bedrockagent.Options)) (*bedrockagent.GetDataSourceOutput, error) {
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

func (t *testBedrockAgent) IngestKnowledgeBaseDocuments(ctx context.Context, params *bedrockagent.IngestKnowledgeBaseDocumentsInput, optFns ...func(*bedrockagent.Options)) (*bedrockagent.IngestKnowledgeBaseDocumentsOutput, error) {
	t.calls++
	return &bedrockagent.IngestKnowledgeBaseDocumentsOutput{}, nil
}

func (t *testBedrockAgent) StartIngestionJob(ctx context.Context, params *bedrockagent.StartIngestionJobInput, optFns ...func(*bedrockagent.Options)) (*bedrockagent.StartIngestionJobOutput, error) {
	t.calls++
	return &bedrockagent.StartIngestionJobOutput{
		IngestionJob: &types.IngestionJob{
			IngestionJobId: aws.String("testIngestionJobId"),
		},
	}, nil
}

func (t *testBedrockAgent) GetIngestionJob(ctx context.Context, params *bedrockagent.GetIngestionJobInput, optFns ...func(*bedrockagent.Options)) (*bedrockagent.GetIngestionJobOutput, error) {
	t.calls++
	return &bedrockagent.GetIngestionJobOutput{
		IngestionJob: &types.IngestionJob{
			Status: types.IngestionJobStatusComplete,
		},
	}, nil
}

type testBedrockAgentRuntime struct{ calls int }

var _ bedrockAgentRuntimeAPI = &testBedrockAgentRuntime{}

func (t *testBedrockAgentRuntime) Retrieve(ctx context.Context, params *bedrockagentruntime.RetrieveInput, optFns ...func(*bedrockagentruntime.Options)) (*bedrockagentruntime.RetrieveOutput, error) {
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

type testS3Client struct{ calls int }

var _ s3API = &testS3Client{}

func (t *testS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	t.calls++
	return &s3.PutObjectOutput{}, nil
}
func (t *testS3Client) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	t.calls++
	return &s3.DeleteObjectOutput{}, nil
}

func TestKnowledgeBaseAddDocuments(t *testing.T) {
	t.Parallel()
	
	testBedrockAgent := &testBedrockAgent{}
	s3Client := &testS3Client{}
	ctx := context.TODO()
	kb := newFromClients("testKbWithOneS3Datasource", testBedrockAgent, &testBedrockAgentRuntime{}, s3Client)
	
	ids, err := kb.AddDocuments(ctx, []schema.Document{{PageContent: "Mock"}})
	require.NoError(t, err)
	require.Equal(t, 6, testBedrockAgent.calls, "expected 6 calls to testBedrockAgent")
	require.Equal(t, 1, s3Client.calls, "expected 1 call to s3Client")
	require.Len(t, ids, 1, "expected 1 id")
}

func TestKnowledgeBaseAddDocumentsWithoutDs(t *testing.T) {
	t.Parallel()
	
	testBedrockAgent := &testBedrockAgent{}
	s3Client := &testS3Client{}
	ctx := context.TODO()
	kb := newFromClients("testKbWithoutDatasources", testBedrockAgent, &testBedrockAgentRuntime{}, s3Client)
	
	_, err := kb.AddDocuments(ctx, []schema.Document{{PageContent: "Mock"}})
	require.Error(t, err, "expected error because knowledge base has no datasources")
}

func TestKnowledgeBaseAddDocumentsWithWrongDsId(t *testing.T) {
	t.Parallel()
	
	testBedrockAgent := &testBedrockAgent{}
	s3Client := &testS3Client{}
	ctx := context.TODO()
	kb := newFromClients("testKbWithOneS3Datasource", testBedrockAgent, &testBedrockAgentRuntime{}, s3Client)
	
	_, err := kb.AddDocuments(ctx, []schema.Document{{PageContent: "Mock"}}, vectorstores.WithNameSpace("wrongDatasourceID"))
	require.Error(t, err, "expected error because wrongDatasourceID is not valid")
}

func TestKnowledgeBaseAddDocumentsWithMultipleDs(t *testing.T) {
	t.Parallel()
	
	testBedrockAgent := &testBedrockAgent{}
	s3Client := &testS3Client{}
	ctx := context.TODO()
	kb := newFromClients("testKbWithTwoS3Datasource", testBedrockAgent, &testBedrockAgentRuntime{}, s3Client)
	
	ids, err := kb.AddDocuments(ctx, []schema.Document{{PageContent: "Mock"}}, vectorstores.WithNameSpace("testS3DatasourceID"))
	require.NoError(t, err)
	require.Equal(t, 7, testBedrockAgent.calls, "expected 7 calls to testBedrockAgent")
	require.Equal(t, 1, s3Client.calls, "expected 1 call to s3Client")
	require.Len(t, ids, 1, "expected 1 id")
}

func TestKnowledgeBaseAddDocumentsWithMultipleMixedDs(t *testing.T) {
	t.Parallel()
	
	testBedrockAgent := &testBedrockAgent{}
	s3Client := &testS3Client{}
	ctx := context.TODO()
	kb := newFromClients("testKbWithTwoMixedDatasources", testBedrockAgent, &testBedrockAgentRuntime{}, s3Client)
	
	ids, err := kb.AddDocuments(ctx, []schema.Document{{PageContent: "Mock"}})
	require.NoError(t, err)
	require.Equal(t, 7, testBedrockAgent.calls, "expected 7 calls to testBedrockAgent")
	require.Equal(t, 1, s3Client.calls, "expected 1 call to s3Client")
	require.Len(t, ids, 1, "expected 1 id")
}

func TestKnowledgeBaseAddDocumentsWithMultipleS3DsAndWithoutDsID(t *testing.T) {
	t.Parallel()
	
	ctx := context.TODO()
	kb := newFromClients("testKbWithTwoS3Datasource", &testBedrockAgent{}, &testBedrockAgentRuntime{}, &testS3Client{})
	
	_, err := kb.AddDocuments(ctx, []schema.Document{{PageContent: "Mock"}})
	require.Error(t, err, "expected error because vectorstores.WithNameSpace is required")
}

func TestKnowledgeBaseAddNamedDocuments(t *testing.T) {
	t.Parallel()
	
	testBedrockAgent := &testBedrockAgent{}
	s3Client := &testS3Client{}
	ctx := context.TODO()
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
	require.Equal(t, 6, testBedrockAgent.calls, "expected 6 calls to testBedrockAgent")
	require.Equal(t, 1, s3Client.calls, "expected 1 call to s3Client")
	require.Len(t, ids, 1, "expected 1 id")
	require.Equal(t, "Mock", ids[0], "expected id to be 'Mock'")
}

func TestKnowledgeBaseSimilaritySearch(t *testing.T) {
	t.Parallel()
	
	ctx := context.TODO()
	kb := newFromClients("testKbId", &testBedrockAgent{}, &testBedrockAgentRuntime{}, &testS3Client{})
	
	docs, err := kb.SimilaritySearch(ctx, "What color is the desk?", 5)
	require.NoError(t, err)
	require.Len(t, docs, 3, "expected 3 documents")
	require.Equal(t, "The color of the house is blue.", docs[0].PageContent, "expected content to be 'The color of the house is blue.'")
	require.Equal(t, "The color of the car is red.", docs[1].PageContent, "expected content to be 'The color of the car is red.'")
	require.Equal(t, "The color of the desk is orange.", docs[2].PageContent, "expected content to be 'The color of the desk is orange.'")
}

func TestKnowledgeBaseSimilaritySearchWithScoreThreshold(t *testing.T) {
	t.Parallel()
	
	ctx := context.TODO()
	kb := newFromClients("testKbId", &testBedrockAgent{}, &testBedrockAgentRuntime{}, &testS3Client{})
	
	docs, err := kb.SimilaritySearch(ctx, "What color is the desk?", 5, vectorstores.WithScoreThreshold(0.8))
	require.NoError(t, err)
	require.Len(t, docs, 2, "expected 2 documents")
	require.Equal(t, "The color of the house is blue.", docs[0].PageContent, "expected content to be 'The color of the house is blue.'")
	require.Equal(t, "The color of the car is red.", docs[1].PageContent, "expected content to be 'The color of the car is red.'")
}

func TestKnowledgeBaseSimilaritySearchWithFilter(t *testing.T) {
	t.Parallel()
	
	ctx := context.TODO()
	kb := newFromClients("testKbId", &testBedrockAgent{}, &testBedrockAgentRuntime{}, &testS3Client{})
	
	_, err := kb.SimilaritySearch(ctx, "What color is the desk?", 5, vectorstores.WithFilters(EqualsFilter{Key: "color", Value: "orange"}))
	require.NoError(t, err)
}

func TestKnowledgeBaseSimilaritySearchWrongWithFilter(t *testing.T) {
	t.Parallel()
	
	ctx := context.TODO()
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
	t.Parallel()
	
	ctx := context.TODO()
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
	t.Parallel()
	
	ctx := context.TODO()
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
