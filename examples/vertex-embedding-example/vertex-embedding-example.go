// Set the VERTEX_PROJECT to your GCP project with Vertex AI APIs enabled.
// Set VERTEX_LOCATION to a GCP location (region); if you're not sure about
// the location, set us-central1
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/googleai/vertex"
)

func main() {
	ctx := context.Background()
	project := os.Getenv("VERTEX_PROJECT")
	location := os.Getenv("VERTEX_LOCATION")
	llm, err := vertex.New(ctx, googleai.WithCloudProject(project), googleai.WithCloudLocation(location))
	if err != nil {
		log.Fatal(err)
	}

	embeddings, err := llm.CreateEmbedding(ctx, []string{"I am a human"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(embeddings)
}
