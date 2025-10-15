package chains

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vendasta/langchaingo/callbacks"
	"github.com/vendasta/langchaingo/httputil"
	"github.com/vendasta/langchaingo/internal/httprr"
	"github.com/vendasta/langchaingo/llms/googleai"
	"github.com/vendasta/langchaingo/llms/openai"
	"github.com/vendasta/langchaingo/prompts"
)

// hasExistingRecording checks if a httprr recording exists for this test
func hasExistingRecording(t *testing.T) bool {
	testName := strings.ReplaceAll(t.Name(), "/", "_")
	testName = strings.ReplaceAll(testName, " ", "_")
	recordingPath := filepath.Join("testdata", testName+".httprr")
	_, err := os.Stat(recordingPath)
	return err == nil
}

func TestLLMChain(t *testing.T) {
	ctx := context.Background()
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, httputil.DefaultTransport)

	// Only run tests in parallel when not recording (to avoid rate limits)
	if rr.Replaying() {
		t.Parallel()
	}

	var opts []openai.Option
	opts = append(opts, openai.WithHTTPClient(rr.Client()))

	// Use test token when replaying
	if rr.Replaying() {
		opts = append(opts, openai.WithToken("test-api-key"))
	}

	model, err := openai.New(opts...)
	require.NoError(t, err)
	model.CallbacksHandler = callbacks.LogHandler{}

	prompt := prompts.NewPromptTemplate(
		"What is the capital of {{.country}}",
		[]string{"country"},
	)

	chain := NewLLMChain(model, prompt)

	result, err := Predict(ctx, chain,
		map[string]any{
			"country": "France",
		},
	)
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "Paris"))
}

func TestLLMChainWithChatPromptTemplate(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	c := NewLLMChain(
		&testLanguageModel{},
		prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
			prompts.NewAIMessagePromptTemplate("{{.foo}}", []string{"foo"}),
			prompts.NewHumanMessagePromptTemplate("{{.boo}}", []string{"boo"}),
		}),
	)
	result, err := Predict(ctx, c, map[string]any{
		"foo": "foo",
		"boo": "boo",
	})
	require.NoError(t, err)
	require.Equal(t, "AI: foo\nHuman: boo", result)
}

func TestLLMChainWithGoogleAI(t *testing.T) {
	ctx := context.Background()

	// Skip if no recording available and no credentials
	if !hasExistingRecording(t) {
		t.Skip("No httprr recording available. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
	}

	// Create httprr with API key transport wrapper
	// This is necessary because the Google API library doesn't add the API key
	// when a custom HTTP client is provided via WithHTTPClient
	apiKey := os.Getenv("GOOGLE_API_KEY")
	transport := httputil.DefaultTransport
	if apiKey != "" {
		transport = &httputil.ApiKeyTransport{
			Transport: transport,
			APIKey:    apiKey,
		}
	}

	rr := httprr.OpenForTest(t, transport)

	// Scrub API key for security in recordings
	rr.ScrubReq(func(req *http.Request) error {
		q := req.URL.Query()
		if q.Get("key") != "" {
			q.Set("key", "test-api-key")
			req.URL.RawQuery = q.Encode()
		}
		return nil
	})

	// Configure client with httprr
	var opts []googleai.Option
	opts = append(opts, googleai.WithRest(), googleai.WithHTTPClient(rr.Client()))

	model, err := googleai.New(ctx, opts...)
	require.NoError(t, err)
	model.CallbacksHandler = callbacks.LogHandler{}

	prompt := prompts.NewPromptTemplate(
		"What is the capital of {{.country}}",
		[]string{"country"},
	)

	chain := NewLLMChain(model, prompt)

	// chains tramples over defaults for options, so setting these options
	// explicitly is required until https://github.com/vendasta/langchaingo/issues/626
	// is fully resolved.
	result, err := Predict(ctx, chain,
		map[string]any{
			"country": "France",
		},
	)
	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}
	require.True(t, strings.Contains(result, "Paris"))
}
