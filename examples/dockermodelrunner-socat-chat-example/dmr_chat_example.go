package main

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/socat"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	const modelRunnerPort = 80

	socatCtr, err := socat.Run(
		context.Background(), "alpine/socat:1.8.0.1",
		testcontainers.WithWaitStrategy(wait.ForListeningPort("80/tcp")),
		socat.WithTarget(socat.NewTarget(modelRunnerPort, "model-runner.docker.internal")),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := socatCtr.Terminate(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()

	dmrURL := socatCtr.TargetURL(modelRunnerPort)

	opts := []openai.Option{
		openai.WithBaseURL(dmrURL.String() + "/engines/v1"),
		openai.WithModel("ai/llama3.2:latest"),
	}

	llm, err := openai.New(opts...)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a company branding design wizard."),
		llms.TextParts(llms.ChatMessageTypeHuman, "What would be a good company name a company that makes colorful clothes for whales?"),
	}

	if _, err := llm.GenerateContent(ctx, content,
		llms.WithMaxTokens(1024),
		llms.WithTemperature(0.8),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			fmt.Print(string(chunk))
			return nil
		})); err != nil {
		log.Fatal(err)
	}
}
