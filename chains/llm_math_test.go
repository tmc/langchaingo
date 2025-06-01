package chains

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/llms/openai"
)

func TestLLMMath(t *testing.T) {
	t.Parallel()
	httprr.SkipIfNoCredentialsOrRecording(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	t.Cleanup(func() { rr.Close() })
	llm, err := openai.New(openai.WithHTTPClient(rr.Client()))
	require.NoError(t, err)

	chain := NewLLMMathChain(llm)
	q := "what is forty plus three? take that then multiply it by ten thousand divided by 7324.3"
	result, err := Run(t.Context(), chain, q)
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "58.708"), "expected 58.708 in result")
}
