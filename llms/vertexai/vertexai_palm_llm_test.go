package vertexai

import (
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/vertexai/internal/aiplatformclient"
	"github.com/tmc/langchaingo/llms/vertexai/internal/genaiclient"
	"github.com/tmc/langchaingo/schema"
)

func TestMain(m *testing.M) {
	// Load .env to get required keys
	_ = godotenv.Load()
	os.Exit(m.Run())
}

type testCase struct {
	model string
}

func TestLLM_Call(t *testing.T) {
	t.Parallel()

	if len(os.Getenv(projectIDEnvVarName)) == 0 {
		t.Skipf("Missing env var: %s, skipping test", projectIDEnvVarName)
	}

	testCases := []testCase{
		{model: aiplatformclient.TextModelName},
		{model: genaiclient.TextModelName},
	}

	llm, err := New()
	require.NoError(t, err)

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.model, func(t *testing.T) {
			t.Parallel()

			output, err := llm.Call(context.Background(), "Write the word Panama and nothing else.", llms.WithModel(tt.model))
			require.NoError(t, err)
			assert.Contains(t, output, "Panama")
		})
	}
}

func TestLLM_Chat(t *testing.T) {
	t.Parallel()

	if len(os.Getenv(projectIDEnvVarName)) == 0 {
		t.Skipf("Missing env var: %s, skipping test", projectIDEnvVarName)
	}

	testCases := []testCase{
		{model: aiplatformclient.ChatModelName},
		{model: genaiclient.ChatModelName},
	}

	llm, err := NewChat()
	require.NoError(t, err)

	messages := []schema.ChatMessage{
		schema.HumanChatMessage{
			Content: "Please say the word 'hello' in all lower case",
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.model, func(t *testing.T) {
			t.Parallel()

			output, err := llm.Call(context.Background(), messages, llms.WithModel(tt.model))
			require.NoError(t, err)
			assert.Contains(t, output.Content, "hello")
		})
	}
}

func TestBaseLLM_CreateEmbedding(t *testing.T) {
	t.Parallel()

	if len(os.Getenv(projectIDEnvVarName)) == 0 {
		t.Skipf("Missing env var: %s, skipping test", projectIDEnvVarName)
	}

	testCases := []testCase{
		{model: aiplatformclient.ChatModelName},
		{model: genaiclient.ChatModelName},
	}

	text := []string{"some text"}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.model, func(t *testing.T) {
			t.Parallel()

			llm, err := New(WithModel(tt.model))
			require.NoError(t, err)

			vector, err := llm.CreateEmbedding(context.Background(), text)
			require.NoError(t, err)
			assert.NotEmpty(t, vector)
		})
	}
}
