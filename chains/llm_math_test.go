package chains

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/0xDezzy/langchaingo/internal/httprr"
	"github.com/0xDezzy/langchaingo/llms/openai"
	"github.com/stretchr/testify/require"
)

func TestLLMMath(t *testing.T) {
	ctx := context.Background()
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	// Only run tests in parallel when not recording
	if rr.Replaying() {
		t.Parallel()
	}

	opts := []openai.Option{
		openai.WithHTTPClient(rr.Client()),
	}

	// Only add fake token when NOT recording (i.e., during replay)
	if rr.Replaying() {
		opts = append(opts, openai.WithToken("test-api-key"))
	}
	// When recording, openai.New() will read OPENAI_API_KEY from environment

	llm, err := openai.New(opts...)
	require.NoError(t, err)

	chain := NewLLMMathChain(llm)
	q := "what is forty plus three? take that then multiply it by ten thousand divided by 7324.3"
	result, err := Run(ctx, chain, q)
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "58.708"), "expected 58.708 in result")
}
