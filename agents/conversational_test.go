package agents

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/0xDezzy/langchaingo/chains"
	"github.com/0xDezzy/langchaingo/httputil"
	"github.com/0xDezzy/langchaingo/internal/httprr"
	"github.com/0xDezzy/langchaingo/llms/openai"
	"github.com/0xDezzy/langchaingo/memory"
	"github.com/0xDezzy/langchaingo/tools"
)

func TestConversationalWithMemory(t *testing.T) {
	t.Parallel()
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, httputil.DefaultTransport)
	// Configure OpenAI client with httprr
	opts := []openai.Option{
		openai.WithModel("gpt-4o"),
		openai.WithHTTPClient(rr.Client()),
	}
	if rr.Replaying() {
		opts = append(opts, openai.WithToken("test-api-key"))
	}

	llm, err := openai.New(opts...)
	require.NoError(t, err)

	executor, err := Initialize(
		llm,
		[]tools.Tool{tools.Calculator{}},
		ConversationalReactDescription,
		WithMemory(memory.NewConversationBuffer()),
	)
	require.NoError(t, err)

	ctx := context.Background()
	res, err := chains.Run(ctx, executor, "Hi! my name is Bob and the year I was born is 1987")
	require.NoError(t, err)

	// Verify we got a reasonable response
	require.Contains(t, res, "Bob")
	t.Logf("Agent response: %s", res)
}
