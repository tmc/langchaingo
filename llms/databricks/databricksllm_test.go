package databricks_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/databricks"
	databricksclientsllama31 "github.com/tmc/langchaingo/llms/databricks/clients/llama/v3.1"
	databricksclientsmistralv1 "github.com/tmc/langchaingo/llms/databricks/clients/mistral/v1"
)

func testModelStream(t *testing.T, model databricks.Model, url string) {
	t.Helper()

	const envVarToken = "DATABRICKS_TOKEN"

	if os.Getenv(envVarToken) == "" {
		t.Skipf("%s not set", envVarToken)
	}

	dbllm, err := databricks.New(
		model,
		databricks.WithFullURL(url),
		databricks.WithToken(os.Getenv(envVarToken)),
	)
	require.NoError(t, err, "failed to create databricks LLM")

	ctx := context.Background()
	resp, err := dbllm.GenerateContent(ctx,
		[]llms.MessageContent{
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: "Brazil is a country?"},
				},
			},
		},
		llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			require.NotEmpty(t, chunk, "unexpected empty chunk in streaming response")
			return nil
		}),
	)
	require.NoError(t, err, "failed to generate content")

	assert.NotEmpty(t, resp.Choices, "expected at least one choice from model")
}

func testModel(t *testing.T, model databricks.Model, url string) {
	t.Helper()

	const envVarToken = "DATABRICKS_TOKEN"

	if os.Getenv(envVarToken) == "" {
		t.Skipf("%s not set", envVarToken)
	}

	dbllm, err := databricks.New(
		model,
		databricks.WithFullURL(url),
		databricks.WithToken(os.Getenv(envVarToken)),
	)
	require.NoError(t, err, "failed to create databricks LLM")

	ctx := context.Background()
	resp, err := dbllm.GenerateContent(ctx,
		[]llms.MessageContent{
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: "Brazil is a country?"},
				},
			},
		},
	)
	require.NoError(t, err, "failed to generate content")

	assert.NotEmpty(t, resp.Choices, "expected at least one choice from model")
}

func TestDatabricksLlama31(t *testing.T) {
	t.Parallel()

	const envVar = "DATABRICKS_LLAMA31_URL"
	url := os.Getenv(envVar)
	if url == "" {
		t.Skipf("%s not set", envVar)
	}

	llama31 := databricksclientsllama31.NewLlama31()
	testModelStream(t, llama31, url)
	testModel(t, llama31, url)
}

func TestDatabricksMistral1(t *testing.T) {
	t.Parallel()

	const envVarURL = "DATABRICKS_MISTAL1_URL"
	const envVarModel = "DATABRICKS_MISTAL1_MODEL"

	model := os.Getenv(envVarModel)
	url := os.Getenv(envVarURL)

	if url == "" {
		t.Skipf("%s not set", envVarURL)
	}
	if model == "" {
		t.Skipf("%s not set", envVarModel)
	}

	mistral1 := databricksclientsmistralv1.NewMistral1(model)
	testModelStream(t, mistral1, url)
	testModel(t, mistral1, url)
}
