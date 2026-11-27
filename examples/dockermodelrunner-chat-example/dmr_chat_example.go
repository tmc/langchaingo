package main

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go/modules/dockermodelrunner"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	const (
		modelNamespace = "ai"
		modelName      = "smollm2"
		modelTag       = "360M-Q4_K_M"
		fqName         = modelNamespace + "/" + modelName + ":" + modelTag
	)

	dmrCtr, err := dockermodelrunner.Run(context.Background(), dockermodelrunner.WithModel(fqName))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := dmrCtr.Terminate(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()

	opts := []openai.Option{
		openai.WithBaseURL(dmrCtr.OpenAIEndpoint()),
		openai.WithModel(fqName),
		openai.WithToken("foo"), // No API key needed for Model Runner
	}

	llm, err := openai.New(opts...)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a useful AI agent expert with TV series."),
		llms.TextParts(llms.ChatMessageTypeHuman, "Tell me about the Anime series called Attack on Titan"),
	}

	if _, err := llm.GenerateContent(ctx, content,
		llms.WithMaxTokens(1024),
		llms.WithTemperature(0.8),
		llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			fmt.Print(string(chunk))
			return nil
		})); err != nil {
		log.Fatal(err)
	}
}
