package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	opensearchgo "github.com/opensearch-project/opensearch-go/v2"
	requestsigner "github.com/opensearch-project/opensearch-go/v2/signer/awsv2"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/exp/detectschema"
	detectschemaopensearch "github.com/tmc/langchaingo/exp/detectschema/translator/opensearch"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/retrievers/selfquery"
	selfqueryopensearch "github.com/tmc/langchaingo/retrievers/selfquery/opensearch"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/opensearch"
)

//go:embed movies.csv
var _moviesCSV string //nolint:gochecknoglobals

func main() {
	// set OPENAI_API_KEY, OPENAI_BASE_URL, OPENSEARCH_ENDPOINT, AWS_PROFILE.
	ctx := context.TODO()
	llm := setLLM()
	indexName := "movies"
	opensearchVectorstore := getOpensearchVectorStore((os.Getenv("OPENSEARCH_ENDPOINT")), os.Getenv("AWS_PROFILE"), llm)
	defer func() {
		err := opensearchVectorstore.DeleteIndex(ctx, indexName, nil)
		if err != nil {
			panic(err)
		}
	}()

	// first we need to create the rag.
	attributeInfos := ragCreation(ctx, llm, indexName, opensearchVectorstore)

	// now we can search the rag.
	ragSearch(ctx, llm, indexName, opensearchVectorstore, attributeInfos)
}

func ragCreation(ctx context.Context, llm *openai.LLM, indexName string, opensearchVectorstore opensearch.Store) []schema.AttributeInfo {
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

	// // let's create the index with the schema.
	if err := opensearchVectorstore.CreateIndex(
		ctx,
		indexName,
		nil,
		opensearch.WithMetadata(opensearchAttributes),
	); err != nil {
		panic(err)
	}

	// convert CSV to Documents.
	splitter := textsplitter.RecursiveCharacter{
		Separators:   []string{"\n\n", "\n", " ", ""},
		ChunkSize:    800,
		ChunkOverlap: 200,
	}

	loader := documentloaders.NewCSV(strings.NewReader(_moviesCSV), documentloaders.WithRowPropertiesInMetadata())
	documents, err := loader.LoadAndSplit(ctx, splitter)
	if err != nil {
		panic(err)
	}

	// let's index the documents.
	if _, err := opensearchVectorstore.AddDocuments(ctx, documents, vectorstores.WithNameSpace(indexName)); err != nil {
		panic(err)
	}

	return attributes
}

func ragSearch(ctx context.Context, llm *openai.LLM, indexName string, opensearchVectorstore opensearch.Store, attributeInfos []schema.AttributeInfo) {
	// let's search for the documents.
	retriever := selfquery.FromLLM(selfquery.FromLLMArgs{
		LLM:               llm,
		Store:             selfqueryopensearch.New(opensearchVectorstore, indexName),
		MetadataFieldInfo: attributeInfos,
		DocumentContents:  "all the properties of a movie",
		EnableLimit:       true,
		Debug:             true,
	})

	log.Printf("%s\n", "movies released after 2007 with a audience rating higher than 85")
	documents, err := retriever.GetRelevantDocuments(ctx, "movies released after 2007 with a audience rating higher than 85")
	if err != nil {
		panic(err)
	}
	fmt.Printf("documents: %v\n", len(documents))

	for _, d := range documents {
		log.Printf("%s | %s | %s \n", d.Metadata["Film"], d.Metadata["Year"], d.Metadata["Audience score %"])
	}

	// filter: and(gt("Year", 2007), gt("Audience score %", 85))
	// result:
	// A Dangerous Method | 2011 | 89
	// WALL-E | 2008 | 89
	// Tangled | 2010 | 88

	log.Printf("%s\n", "10 movies with rotten tomatoes rating higher than 60")

	documents, err = retriever.GetRelevantDocuments(ctx, "10 movies with rotten tomatoes rating higher than 60")
	if err != nil {
		panic(err)
	}
	fmt.Printf("documents: %v\n", len(documents))

	for _, d := range documents {
		log.Printf("%s | %s | %s \n", d.Metadata["Film"], d.Metadata["Year"], d.Metadata["Rotten Tomatoes %"])
	}
	// filter: gt("Rotten Tomatoes %", 60)
	// result:
	// WALL-E | 2008 | 96
	// Jane Eyre | 2011 | 85
	// A Dangerous Method | 2011 | 79
	// The Curious Case of Benjamin Button | 2008 | 73
	// Nick and Norah's Infinite Playlist | 2008 | 73
	// High School Musical 3: Senior Year | 2008 | 65
	// Zack and Miri Make a Porno | 2008 | 64
	// Miss Pettigrew Lives for a Day | 2008 | 78
	// Midnight in Paris | 2011 | 93
	// Knocked Up | 2007 | 91
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
