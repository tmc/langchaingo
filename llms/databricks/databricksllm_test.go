package databricks_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/databricks"
	databricksclientsllama31 "github.com/tmc/langchaingo/llms/databricks/clients/llama/v3.1"
	databricksclientsmistralv1 "github.com/tmc/langchaingo/llms/databricks/clients/mistral/v1"
)

func testModel(t *testing.T, model databricks.Model, url string) {
	t.Helper()

	const envVarToken = "DATABRICKS_TOKEN"

	if os.Getenv(envVarToken) == "" {
		t.Skipf("%s not set", envVarToken)
	}

	dbllm, err := databricks.New(model, databricks.WithFullURL(url), databricks.WithToken(os.Getenv(envVarToken)))
	if err != nil {
		t.Fatalf("failed to create databricks LLM: %v", err)
	}

	ctx := context.Background()
	resp, err := dbllm.GenerateContent(ctx, []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Brazil is a country? the answer should just be yes or no"},
			},
		},
	}, llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
		fmt.Printf("string(chunk): %v\n", string(chunk))
		return nil
	}))
	if err != nil {
		t.Fatalf("failed to generate content: %v", err)
	}

	if len(resp.Choices) < 1 {
		t.Fatalf("empty response from model")
	}
}

func TestDatabricksLlama31(t *testing.T) {
	t.Parallel()

	const envVar = "DATABRICKS_LLAMA31_URL"

	if os.Getenv(envVar) == "" {
		t.Skipf("%s not set", envVar)
	}

	testModel(t, databricksclientsllama31.NewLlama31(), os.Getenv(envVar))

	t.Error()
}

func TestDatabricksMistal1(t *testing.T) {
	t.Parallel()

	const envVarURL = "DATABRICKS_MISTAL1_URL"
	const envVarModel = "DATABRICKS_MISTAL1_MODEL"

	if os.Getenv(envVarURL) == "" {
		t.Skipf("%s not set", envVarURL)
	}

	if os.Getenv(envVarModel) == "" {
		t.Skipf("%s not set", envVarModel)
	}

	testModel(t, databricksclientsmistralv1.NewMistral1(os.Getenv(envVarModel)), os.Getenv(envVarURL))

	t.Error()
}
