package agents

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/vxcontrol/langchaingo/chains"
	"github.com/vxcontrol/langchaingo/httputil"
	"github.com/vxcontrol/langchaingo/internal/httprr"
	"github.com/vxcontrol/langchaingo/llms/openai"
	"github.com/vxcontrol/langchaingo/memory"
	"github.com/vxcontrol/langchaingo/tools"

	"github.com/stretchr/testify/require"
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
	_, err = chains.Run(ctx, executor, "Hi! my name is Bob and the year I was born is 1987")
	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}

	res, err := chains.Run(t.Context(), executor, "What is the year I was born times 34")
	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}
	expectedRe := "67,?558"
	if !regexp.MustCompile(expectedRe).MatchString(res) {
		t.Errorf("result does not contain the crrect answer '67558', got: %s", res)
	}
}
