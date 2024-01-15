package selfquery_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	opensearchgo "github.com/opensearch-project/opensearch-go/v2"
	requestsigner "github.com/opensearch-project/opensearch-go/v2/signer/awsv2"
	"github.com/tmc/langchaingo/exp/retrievers/selfquery"
	selfquery_opensearch "github.com/tmc/langchaingo/exp/retrievers/selfquery/opensearch"
	"github.com/tmc/langchaingo/exp/tools/queryconstructor"
	"github.com/tmc/langchaingo/llms/openai"
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

func getOpensearchVectorStore(t *testing.T, endpoint, profile string) vectorstores.VectorStore {
	t.Helper()
	ctx := context.Background()

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		fmt.Printf("err loading config: %v\n", err)
		t.Fail()
	}

	// Create an AWS request Signer and load AWS configuration using default config folder or env vars.
	signer, err := requestsigner.NewSignerWithService(awsCfg, "es")
	if err != nil {
		fmt.Printf("err setting signer: %v\n", err)
		t.Fail()
	}

	// Create an opensearch client and use the request-signer
	client, err := opensearchgo.NewClient(opensearchgo.Config{
		Addresses: []string{endpoint},
		Signer:    signer,
	})
	if err != nil {
		fmt.Printf("client creation err: %v\n", err)
		t.Fail()
	}

	vectorstore, err := opensearch.New(client)
	if err != nil {
		fmt.Printf("vectorstore creation err: %v\n", err)
		t.Fail()
	}
	return vectorstore
}

func TestParser(t *testing.T) {
	opensearchEndpoint, awsProfile := getEnvVariables(t)
	llm := setLLM(t)

	vectorstore := getOpensearchVectorStore(t, opensearchEndpoint, awsProfile)
	// fmt.Printf("vectorstore: %v\n", vectorstore)
	enableLimit := true
	store := selfquery_opensearch.New(vectorstore)
	retriever := selfquery.FromLLM(selfquery.FromLLMArgs{
		LLM:   llm,
		Store: store,
		// VectorStore:      vectorstore,
		DocumentContents: "Brief summary of a movie",
		MetadataFieldInfo: []queryconstructor.AttributeInfo{
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
	// fmt.Printf("retriever: %v\n", retriever)
	documents, err := retriever.GetRelevantDocuments(context.TODO(), "Give me all new good movies")
	if err != nil {
		panic(err)
	}

	fmt.Printf("documents: %v\n", documents)
	t.Fail()
}
