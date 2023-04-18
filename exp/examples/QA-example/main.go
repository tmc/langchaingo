package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/exp/chains"
	"github.com/tmc/langchaingo/exp/documentLoaders"
	"github.com/tmc/langchaingo/exp/embeddings"
	"github.com/tmc/langchaingo/exp/textSplitters"
	"github.com/tmc/langchaingo/exp/vectorStores/pinecone"
	"github.com/tmc/langchaingo/llms/openai"
)

var pineconeEnv = "us-central1-gcp"
var textFile = "./The_Art_Of_War.txt"
var indexName = "database"
var dimensions = 1536
var numDocsInReq = 5

func main() {
	// We start with splitting the input text file into smaller documents
	splitter := textSplitters.NewRecursiveCharactersSplitter()
	docs, err := documentLoaders.NewTextLoaderFromFile(textFile).LoadAndSplit(splitter)
	if err != nil {
		log.Fatalf("Error loading and splitting document: %s", err.Error())
	}

	// Next we need an embeddings model to get the embeddings of all of the documents
	embedding, err := embeddings.NewOpenAI()
	if err != nil {
		log.Fatal(err.Error())
	}

	// We also need to create a vector database to store these embeddings for queries. Here is how it's done using pinecone
	// Because pinecone takes time to initialize indexes, this should be an index that already exists
	p, err := pinecone.NewPinecone(embedding, pineconeEnv, indexName, dimensions)
	if err != nil {
		log.Fatal(err.Error())
	}

	p.AddDocuments(docs, []string{}, textFile)

	llm, err := openai.New()
	if err != nil {
		log.Fatal(err.Error())
	}

	// Now we can create a RetrievalQAChain using the pinecone index and a llm
	chain := chains.NewRetrievalQAChainFromLLM(llm, p.ToRetriever(numDocsInReq, textFile))

	for {
		fmt.Print("Enter query: ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("An error occurred while reading input. Please try again", err)
			continue
		}

		result, err := chains.Call(chain, map[string]any{
			"query": input,
		})
		if err != nil {
			log.Fatal(err.Error())
		}

		fmt.Println(result["text"])
	}
}
