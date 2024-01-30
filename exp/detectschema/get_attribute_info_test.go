package detectschema_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/exp/detectschema"
	"github.com/tmc/langchaingo/llms/openai"
)

func getEnvVariables(t *testing.T) {
	t.Helper()

	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		t.Skipf("Must set %s to run test", "OPENAI_API_KEY")
	}
}

func setLLM(t *testing.T) *openai.LLM {
	t.Helper()
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
		t.Fatalf("error setting openAI embedded: %v\n", err)
	}

	return llm
}

func TestGetAttributeInfo(t *testing.T) {
	t.Parallel()
	getEnvVariables(t)
	llm := setLLM(t)

	detector := detectschema.New(llm)

	result, err := detector.GetAttributeInfo(context.TODO(), "clients.csv", "csv", `id,first_name,last_name,email,gender,ip_address,newsletter
	1,Vonnie,Lidden,vlidden0@yellowbook.com,Female,159.74.52.219,false
	2,Kerr,Fallens,kfallens1@booking.com,Male,30.98.27.209,true`)
	require.NoError(t, err)

	if len(result) != 7 {
		t.Fail()
	}

	if result[0].Type != detectschema.AllowedTypeInt ||
		result[1].Type != detectschema.AllowedTypeString ||
		result[2].Type != detectschema.AllowedTypeString ||
		result[3].Type != detectschema.AllowedTypeString ||
		result[4].Type != detectschema.AllowedTypeString ||
		result[5].Type != detectschema.AllowedTypeString ||
		result[6].Type != detectschema.AllowedTypeBool {
		t.Fail()
	}
}
