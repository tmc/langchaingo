// Set the VERTEX_PROJECT env var to your GCP project with Vertex AI APIs
// enabled. Set VERTEX_LOCATION to a GCP location (region); if you're not sure
// about the location, set us-central1
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai/vertex"
)

func main() {
	ctx := context.Background()
	project := os.Getenv("VERTEX_PROJECT")
	location := os.Getenv("VERTEX_LOCATION")
	llm, err := vertex.New(ctx, vertex.WithCloudProject(project), vertex.WithCloudLocation(location))
	if err != nil {
		log.Fatal(err)
	}

	prompt := "Who was the second person to walk on the moon?"
	answer, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(answer)
}
