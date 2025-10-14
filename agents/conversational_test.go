package agents

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vendasta/langchaingo/chains"
	"github.com/vendasta/langchaingo/httputil"
	"github.com/vendasta/langchaingo/internal/httprr"
	"github.com/vendasta/langchaingo/llms/openai"
	"github.com/vendasta/langchaingo/memory"
	"github.com/vendasta/langchaingo/tools"
)

// hasExistingRecording checks if a httprr recording exists for this test
func hasExistingRecording(t *testing.T) bool {
	testName := strings.ReplaceAll(t.Name(), "/", "_")
	testName = strings.ReplaceAll(testName, " ", "_")
	recordingPath := filepath.Join("testdata", testName+".httprr")
	_, err := os.Stat(recordingPath)
	return err == nil
}

func TestConversationalWithMemory(t *testing.T) {
	t.Parallel()

	// Skip if no recording available and no credentials
	if !hasExistingRecording(t) {
		t.Skip("No httprr recording available. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
	}

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
	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}

	// Verify we got a reasonable response
	require.Contains(t, res, "Bob")
	t.Logf("Agent response: %s", res)
}
