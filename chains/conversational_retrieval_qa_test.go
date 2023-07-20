package chains

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
)

type testConversationalRetriever struct{}

func (t testConversationalRetriever) GetRelevantDocuments(ctx context.Context, query string) ([]schema.Document, error) {
	return []schema.Document{
		{PageContent: "foo is 34"},
		{PageContent: "bar is 1"},
	}, nil
}

var _ schema.Retriever = testConversationalRetriever{}

func TestConversationalRetrievalQA(t *testing.T) {
	t.Parallel()
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	llm, err := openai.New()
	require.NoError(t, err)

	combinedStuffQAChain := LoadStuffQA(llm)
	combinedQuestionGeneratorChain := LoadQuestionGenerator(llm)
	r := testConversationalRetriever{}
	chatHistory := []schema.ChatMessage{
		schema.HumanChatMessage{Text: "Hello, I'm interested to buy the BMW, what do you think about that car?"},
		schema.AIChatMessage{Text: "The BMW is a great car, with great petrol and diesel engines."},
		schema.HumanChatMessage{Text: "What is the difference between petrol and diesel engine?"},
		schema.AIChatMessage{Text: "The diesel engine is lower on fuel consumption."},
	}

	chain := NewConversationalRetrievalQA(combinedStuffQAChain, combinedQuestionGeneratorChain, r, chatHistory)
	result, err := Run(context.Background(), chain, "what is foo? ")
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "34"), "expected 34 in result")
}

func TestConversationalRetrievalQAFromLLM(t *testing.T) {
	t.Parallel()
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	r := testConversationalRetriever{}
	llm, err := openai.New()
	require.NoError(t, err)
	chatHistory := []schema.ChatMessage{
		schema.HumanChatMessage{Text: "Hello, I'm interested to buy the BMW, what do you think about that car?"},
		schema.AIChatMessage{Text: "The BMW is a great car, with great petrol and diesel engines."},
		schema.HumanChatMessage{Text: "What is the difference between petrol and diesel engine?"},
		schema.AIChatMessage{Text: "The diesel engine is lower on fuel consumption."},
	}

	chain := NewConversationalRetrievalQAFromLLM(llm, r, chatHistory)
	result, err := Run(context.Background(), chain, "what is foo? ")
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "34"), "expected 34 in result")
}
