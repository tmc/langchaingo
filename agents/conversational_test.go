package agents

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/httputil"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/tools"
)

func TestConversationalWithMemory(t *testing.T) {
	t.Parallel()
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, httputil.DefaultTransport)
	t.Cleanup(func() { rr.Close() })
	// Configure OpenAI client based on recording vs replay mode
	opts := []openai.Option{
		openai.WithModel("gpt-4o"),
		openai.WithHTTPClient(rr.Client()),
	}

	// Only add fake token when NOT recording (i.e., during replay)
	if !rr.Recording() {
		opts = append(opts, openai.WithToken("fake-api-key-for-testing"))
	}
	// When recording, openai.New() will read OPENAI_API_KEY from environment

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
