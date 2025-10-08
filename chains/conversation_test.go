package chains

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"

	z "github.com/getzep/zep-go"
	zClient "github.com/getzep/zep-go/client"
	zOption "github.com/getzep/zep-go/option"
	"github.com/stretchr/testify/require"
	"github.com/vendasta/langchaingo/llms/openai"
	"github.com/vendasta/langchaingo/memory"
	"github.com/vendasta/langchaingo/memory/zep"
)

func TestConversation(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	// Only run tests in parallel when not recording (to avoid rate limits)
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

	c := NewConversation(llm, memory.NewConversationBuffer())
	_, err = Run(ctx, c, "Hi! I'm Jim")
	require.NoError(t, err)

	res, err := Run(ctx, c, "What is my name?")
	require.NoError(t, err)
	require.True(t, strings.Contains(res, "Jim"), `result does not contain the keyword 'Jim'`)
}

func TestConversationWithZepMemory(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	// Only run tests in parallel when not recording (to avoid rate limits)
	if rr.Replaying() {
		t.Parallel()
	}
	zepAPIKey := os.Getenv("ZEP_API_KEY")
	sessionID := os.Getenv("ZEP_SESSION_ID")
	if zepAPIKey == "" || sessionID == "" {
		t.Skip("ZEP_API_KEY or ZEP_SESSION_ID not set")
	}
	if zepKey := os.Getenv("ZEP_API_KEY"); zepKey == "" {
		t.Skip("ZEP_API_KEY not set")
	}
	if sessionID := os.Getenv("ZEP_SESSION_ID"); sessionID == "" {
		t.Skip("ZEP_SESSION_ID not set")
	}

	llm, err := openai.New()
	require.NoError(t, err)

	zc := zClient.NewClient(
		zOption.WithAPIKey(zepAPIKey),
	)

	c := NewConversation(
		llm,
		zep.NewMemory(
			zc,
			sessionID,
			zep.WithMemoryType(z.MemoryGetRequestMemoryTypePerpetual),
			zep.WithHumanPrefix("Joe"),
			zep.WithAIPrefix("Robot"),
		),
	)
	_, err = Run(ctx, c, "Hi! I'm Jim")
	require.NoError(t, err)

	res, err := Run(ctx, c, "What is my name?")
	require.NoError(t, err)
	require.True(t, strings.Contains(res, "Jim"), `result does not contain the keyword 'Jim'`)
}

func TestConversationWithChatLLM(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	// Only run tests in parallel when not recording (to avoid rate limits)
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

	c := NewConversation(llm, memory.NewConversationTokenBuffer(llm, 2000))
	_, err = Run(ctx, c, "Hi! I'm Jim")
	require.NoError(t, err)

	res, err := Run(ctx, c, "What is my name?")
	require.NoError(t, err)
	require.True(t, strings.Contains(res, "Jim"), `result does contain the keyword 'Jim'`)

	// this message will hit the maxTokenLimit and will initiate the prune of the messages to fit the context
	res, err = Run(ctx, c, "Are you sure that my name is Jim?")
	require.NoError(t, err)
	require.True(t, strings.Contains(res, "Jim"), `result does contain the keyword 'Jim'`)
}
