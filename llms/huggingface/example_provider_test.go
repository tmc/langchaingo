package huggingface_test

import (
	"context"
	"fmt"
	"log"

	"github.com/vendasta/langchaingo/llms"
	"github.com/vendasta/langchaingo/llms/huggingface"
)

func ExampleNew_withInferenceProvider() {
	// Create a new HuggingFace LLM with inference provider
	llm, err := huggingface.New(
		huggingface.WithModel("deepseek-ai/DeepSeek-R1-0528"),
		huggingface.WithInferenceProvider("hyperbolic"),
		// Token will be read from HF_TOKEN or HUGGINGFACEHUB_API_TOKEN environment variable
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Use the LLM
	result, err := llm.Call(ctx, "What is the capital of France?",
		llms.WithTemperature(0.5),
		llms.WithMaxLength(50),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)
}

func ExampleNew_standardInference() {
	// Create a new HuggingFace LLM with standard inference API
	llm, err := huggingface.New(
		huggingface.WithModel("HuggingFaceH4/zephyr-7b-beta"),
		// Token will be read from HF_TOKEN or HUGGINGFACEHUB_API_TOKEN environment variable
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Use the LLM
	result, err := llm.Call(ctx, "Hello, how are you?",
		llms.WithTemperature(0.5),
		llms.WithMaxLength(50),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)
}
