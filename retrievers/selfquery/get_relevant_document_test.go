package selfquery_test

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	opensearchgo "github.com/opensearch-project/opensearch-go/v2"
	requestsigner "github.com/opensearch-project/opensearch-go/v2/signer/awsv2"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/retrievers/selfquery"
	selfqueryopensearch "github.com/tmc/langchaingo/retrievers/selfquery/opensearch"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/opensearch"
)

func getEnvVariables(t *testing.T) (string, string) {
	t.Helper()

	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		t.Skipf("Must set %s to run test", "OPENAI_API_KEY")
	}

	opensearchEndpoint := os.Getenv("OPENSEARCH_ENDPOINT")
	if opensearchEndpoint == "" {
		t.Skipf("Must set %s to run test", "OPENSEARCH_ENDPOINT")
	}

	awsProfile := os.Getenv("AWS_PROFILE")
	if awsProfile == "" {
		t.Skipf("Must set %s to run test", "AWS_PROFILE")
	}

	return opensearchEndpoint, awsProfile
}

func setLLM(t *testing.T) *openai.LLM {
	t.Helper()
	openaiOpts := []openai.Option{}

	if openAIBaseURL := os.Getenv("OPENAI_BASE_URL"); openAIBaseURL != "" {
		openaiOpts = append(openaiOpts,
			openai.WithBaseURL(openAIBaseURL),
			openai.WithAPIType(openai.APITypeAzure),
			openai.WithEmbeddingModel("text-embedding-ada-002"),
			openai.WithModel("gpt-4"),
		)
	}

	llm, err := openai.New(openaiOpts...)
	if err != nil {
		t.Fatalf("error setting openAI embedded: %v\n", err)
	}

	return llm
}

func setAndFillIndex(
	t *testing.T,
	client *opensearch.Store,
	indexName string,
) {
	t.Helper()
	_, err := client.CreateIndex(context.TODO(), indexName)
	require.NoError(t, err)

	_, err = client.AddDocuments(context.TODO(), []schema.Document{
		{
			PageContent: "A bunch of scientists bring back dinosaurs and mayhem breaks loose",
			Metadata: map[string]any{
				"year": 1993, "rating": 7.7, "genre": "science fiction",
			},
		},
		{
			PageContent: "Leo DiCaprio gets lost in a dream within a dream within a dream within a ...",
			Metadata: map[string]any{
				"year": 2010, "director": "Christopher Nolan", "rating": 8.2,
			},
		},
		{
			PageContent: "A psychologist / detective gets lost in a series of dreams within dreams within dreams and Inception reused the idea",
			Metadata: map[string]any{
				"year": 2006, "director": "Satoshi Kon", "rating": 8.6,
			},
		},
		{
			PageContent: "A bunch of normal-sized women are supremely wholesome and some men pine after them",
			Metadata: map[string]any{
				"year": 2019, "director": "Greta Gerwig", "rating": 8.3,
			},
		},
		{
			PageContent: "Toys come alive and have a blast doing so",
			Metadata: map[string]any{
				"year": 1995, "genre": "animated",
			},
		},
		{
			PageContent: "Three men walk into the Zone, three men walk out of the Zone",
			Metadata: map[string]any{
				"year":     1979,
				"rating":   9.9,
				"director": "Andrei Tarkovsky",
				"genre":    "science fiction",
			},
		},
	}, vectorstores.WithNameSpace(indexName))
	require.NoError(t, err)
}

func getOpensearchVectorStore(t *testing.T, endpoint, profile string, embedderClient embeddings.EmbedderClient) opensearch.Store {
	t.Helper()
	ctx := context.Background()

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		t.Fail()
	}

	// Create an AWS request Signer and load AWS configuration using default config folder or env vars.
	signer, err := requestsigner.NewSignerWithService(awsCfg, "es")
	if err != nil {
		t.Fail()
	}

	// Create an opensearch client and use the request-signer
	client, err := opensearchgo.NewClient(opensearchgo.Config{
		Addresses: []string{endpoint},
		Signer:    signer,
	})
	if err != nil {
		t.Fail()
	}

	e, err := embeddings.NewEmbedder(embedderClient)
	require.NoError(t, err)

	vectorstore, err := opensearch.New(
		client,
		opensearch.WithEmbedder(e),
	)
	if err != nil {
		t.Fail()
	}
	return vectorstore
}

func TestParser(t *testing.T) {
	t.Parallel()
	indexName := "selfquery_test"
	opensearchEndpoint, awsProfile := getEnvVariables(t)
	llm := setLLM(t)

	vectorstore := getOpensearchVectorStore(t, opensearchEndpoint, awsProfile, llm)
	// fmt.Printf("vectorstore: %v\n", vectorstore)
	defer func() {
		_, err := vectorstore.DeleteIndex(context.TODO(), indexName)
		require.NoError(t, err)
	}()

	setAndFillIndex(t, &vectorstore, indexName)

	enableLimit := true
	store := selfqueryopensearch.New(vectorstore, indexName)
	retriever := selfquery.FromLLM(selfquery.FromLLMArgs{
		LLM:              llm,
		Store:            store,
		DocumentContents: "Brief summary of a movie",
		MetadataFieldInfo: []schema.AttributeInfo{
			{
				Name:        "genre",
				Description: "The genre of the movie",
				Type:        "string or list[string]",
			},
			{
				Name:        "year",
				Description: "The year the movie was release",
				Type:        "integer",
			},
			{
				Name:        "director",
				Description: "The name of the movie director",
				Type:        "string",
			},
			{
				Name:        "rating",
				Description: "A 1-10 rating for the movie",
				Type:        "float",
			},
		},
		EnableLimit: &enableLimit,
	})

	documents, err := retriever.GetRelevantDocuments(context.TODO(), "I want to watch a movie rated higher than 8.5")
	require.NoError(t, err)
	if len(documents) != 2 {
		t.Fail()
	}

	documents, err = retriever.GetRelevantDocuments(context.TODO(), "Has Greta Gerwig directed any movies about women")
	require.NoError(t, err)
	if len(documents) != 1 {
		t.Fail()
	}

	documents, err = retriever.GetRelevantDocuments(context.TODO(), "What's a highly rated (above 8.5) science fiction film?")
	require.NoError(t, err)
	if len(documents) != 2 {
		t.Fail()
	}
}
