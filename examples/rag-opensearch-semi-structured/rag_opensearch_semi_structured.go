package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	_ "embed"

	"github.com/aws/aws-sdk-go-v2/config"
	opensearchgo "github.com/opensearch-project/opensearch-go/v2"
	requestsigner "github.com/opensearch-project/opensearch-go/v2/signer/awsv2"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/exp/detectschema"
	detectschemaopensearch "github.com/tmc/langchaingo/exp/detectschema/translator/opensearch"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/vectorstores/opensearch"
)

//go:embed movies.csv
var _moviesCSV string //nolint:gochecknoglobals

func main() {
	// set OPENAI_API_KEY, OPENAI_BASE_URL, OPENSEARCH_ENDPOINT, AWS_PROFILE
	ctx := context.TODO()
	llm := setLLM()
	opensearchVectorstore := getOpensearchVectorStore((os.Getenv("OPENSEARCH_ENDPOINT")), os.Getenv("AWS_PROFILE"), llm)
	defer opensearchVectorstore.DeleteIndex(ctx, "movies", nil)
	schemadetector := detectschema.New(llm)

	// we need to detect the schema of the csv file, and then create the index with the schema.
	attributes, err := schemadetector.GetAttributeInfo(ctx, "movies.csv", "csv", _moviesCSV)
	if err != nil {
		panic(err)
	}

	// let's translate the attribute info to opensearch mapping.
	schemadetectoropensearch := detectschemaopensearch.New(opensearchVectorstore)
	opensearchAttributes, err := schemadetectoropensearch.TranslateAttributeInfo(attributes)
	if err != nil {
		panic(err)
	}

	// let's create the index with the schema.
	if err := opensearchVectorstore.CreateIndex(ctx, "movies", opensearch.WithMetadata(opensearchAttributes)); err != nil {
		panic(err)
	}

	// convert CSV to Documents.
	splitter := textsplitter.RecursiveCharacter{
		Separators:   []string{"\n\n", "\n", " ", ""},
		ChunkSize:    800,
		ChunkOverlap: 200,
	}
	loader := documentloaders.NewCSV(strings.NewReader(_moviesCSV))
	documents, err := loader.LoadAndSplit(ctx, splitter)
	if err != nil {
		panic(err)
	}
	fmt.Printf("documents: %+v\n", documents)

}

func setLLM() *openai.LLM {

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
		panic(err)
	}

	return llm
}

func getOpensearchVectorStore(endpoint, profile string, embedderClient embeddings.EmbedderClient) opensearch.Store {

	ctx := context.Background()

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		panic(err)
	}

	// Create an AWS request Signer and load AWS configuration using default config folder or env vars.
	signer, err := requestsigner.NewSignerWithService(awsCfg, "es")
	if err != nil {
		panic(err)
	}

	// Create an opensearch client and use the request-signer
	client, err := opensearchgo.NewClient(opensearchgo.Config{
		Addresses: []string{endpoint},
		Signer:    signer,
	})
	if err != nil {
		panic(err)
	}

	e, err := embeddings.NewEmbedder(embedderClient)
	if err != nil {
		panic(err)
	}

	vectorstore, err := opensearch.New(
		client,
		opensearch.WithEmbedder(e),
	)
	if err != nil {
		panic(err)
	}
	return vectorstore
}
