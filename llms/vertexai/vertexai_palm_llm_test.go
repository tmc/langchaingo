package vertexai

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/vertexai/internal/aiplatformclient"
	"github.com/tmc/langchaingo/llms/vertexai/internal/genaiclient"
	"github.com/tmc/langchaingo/schema"
	"os"
	"testing"
)

type MockHandler struct {
}

func (m *MockHandler) HandleLLMStart(ctx context.Context, prompts []string) {
	// Mock implementation
	return
}

func (m *MockHandler) HandleLLMEnd(ctx context.Context, p llms.LLMResult) {
	// Mock implementation
	return
}

func TestMain(m *testing.M) {
	// Load .env to get required keys
	_ = godotenv.Load()
	os.Exit(m.Run())
}

type testCase struct {
	model string
}

func TestLLM_Call(t *testing.T) {
	if len(os.Getenv(projectIDEnvVarName)) == 0 {
		t.Skipf("Missing env var: %s, skipping test", projectIDEnvVarName)
	}

	testCases := []testCase{
		{model: aiplatformclient.TextModelName},
		{model: genaiclient.TextModelName},
	}

	llm, err := New()
	assert.NoError(t, err)

	for _, tt := range testCases {
		t.Run(tt.model, func(t *testing.T) {
			output, err := llm.Call(context.Background(), "Write the word Panama and nothing else.", llms.WithModel(tt.model))
			assert.NoError(t, err)
			assert.Contains(t, output, "Panama")
		})
	}

}

// New function to test chat functionality
func TestLLM_Chat(t *testing.T) {
	if len(os.Getenv(projectIDEnvVarName)) == 0 {
		t.Skipf("Missing env var: %s, skipping test", projectIDEnvVarName)
	}

	testCases := []testCase{
		{model: aiplatformclient.ChatModelName},
		{model: genaiclient.ChatModelName},
	}

	llm, err := NewChat()
	assert.NoError(t, err)

	messages := []schema.ChatMessage{
		schema.HumanChatMessage{
			Content: "Please say the word 'hello' in all lower case",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.model, func(t *testing.T) {
			output, err := llm.Call(context.Background(), messages, llms.WithModel(tt.model))
			assert.NoError(t, err)
			assert.Contains(t, output.Content, "hello")
		})
	}

}

func TestBaseLLM_CreateEmbedding(t *testing.T) {
	if len(os.Getenv(projectIDEnvVarName)) == 0 {
		t.Skipf("Missing env var: %s, skipping test", projectIDEnvVarName)
	}

	llm, err := New()
	assert.NoError(t, err)

	text := []string{"some text"}

	vector, err := llm.CreateEmbedding(context.Background(), text)
	assert.NoError(t, err)
	assert.NotEmpty(t, vector)
}
