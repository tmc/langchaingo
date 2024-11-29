package ollama

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcollama "github.com/testcontainers/testcontainers-go/modules/ollama"
	"github.com/tmc/langchaingo/llms"
)

const (
	ollamaVersion  string = "0.3.13"
	llamaModel     string = "llama3.2"
	llamaTag       string = "1b" // the 1b model is the smallest model, that fits in CPUs instead of GPUs.
	llamaModelName string = llamaModel + ":" + llamaTag

	// ollamaImage is the Docker image to use for the test container.
	// See https://hub.docker.com/r/mdelapenya/llama3.2/tags
	ollamaImage string = "mdelapenya/" + llamaModel + ":" + ollamaVersion + "-" + llamaTag
)

func runOllama(t *testing.T) (string, error) {
	t.Helper()

	ctx := context.Background()

	// the Ollama container is reused across tests, that's why it defines a fixed container name and reuses it.
	ollamaContainer, err := tcollama.RunContainer(
		ctx,
		testcontainers.WithImage(ollamaImage),
		testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Name: "ollama-model",
			},
			Reuse: true,
		},
		))
	if err != nil {
		return "", err
	}

	url, err := ollamaContainer.ConnectionString(ctx)
	if err != nil {
		return "", err
	}
	return url, nil
}

func newTestClient(t *testing.T, opts ...Option) *LLM {
	t.Helper()
	var ollamaModel string
	if ollamaModel = os.Getenv("OLLAMA_TEST_MODEL"); ollamaModel == "" {
		address, err := runOllama(t)
		if err != nil {
			t.Skip("OLLAMA_TEST_MODEL not set")
			return nil
		}
		ollamaModel = llamaModelName
		opts = append(opts, WithServerURL(address))
	}

	opts = append([]Option{WithModel(ollamaModel)}, opts...)

	c, err := New(opts...)
	require.NoError(t, err)
	return c
}

func TestGenerateContent(t *testing.T) {
	t.Parallel()
	llm := newTestClient(t)

	parts := []llms.ContentPart{
		llms.TextContent{Text: "How many feet are in a nautical mile?"},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	rsp, err := llm.GenerateContent(context.Background(), content)
	require.NoError(t, err)

	require.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	require.Regexp(t, "feet", strings.ToLower(c1.Content))
}

func TestWithFormat(t *testing.T) {
	t.Parallel()
	llm := newTestClient(t, WithFormat("json"))

	parts := []llms.ContentPart{
		llms.TextContent{Text: `Please respond with a JSON object in this format:
{
    "feet": <number>,
    "explanation": "<string explaining the conversion>"
}

How many feet are in a nautical mile?`},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	rsp, err := llm.GenerateContent(context.Background(), content)
	require.NoError(t, err)

	require.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	require.Regexp(t, "feet", strings.ToLower(c1.Content))

	// check whether we got *any* kind of JSON object.
	var result map[string]any
	err = json.Unmarshal([]byte(c1.Content), &result)
	require.NoError(t, err)

	// Verify the response contains the expected fields
	require.Contains(t, result, "feet", "Response should contain 'feet' field")
	require.Contains(t, result, "explanation", "Response should contain 'explanation' field")
}

func TestWithStreaming(t *testing.T) {
	t.Parallel()
	llm := newTestClient(t)

	parts := []llms.ContentPart{
		llms.TextContent{Text: "How many feet are in a nautical mile?"},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	var sb strings.Builder
	rsp, err := llm.GenerateContent(context.Background(), content,
		llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			sb.Write(chunk)
			return nil
		}))
	require.NoError(t, err)

	require.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	require.Regexp(t, "feet", strings.ToLower(c1.Content))
	require.Regexp(t, "feet", strings.ToLower(sb.String()))
}

func TestWithKeepAlive(t *testing.T) {
	t.Parallel()
	llm := newTestClient(t, WithKeepAlive("1m"))

	parts := []llms.ContentPart{
		llms.TextContent{Text: "How many feet are in a nautical mile?"},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	resp, err := llm.GenerateContent(context.Background(), content)
	require.NoError(t, err)

	require.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]
	require.Regexp(t, "feet", strings.ToLower(c1.Content))

	vector, err := llm.CreateEmbedding(context.Background(), []string{"test embedding with keep_alive"})
	require.NoError(t, err)
	require.NotEmpty(t, vector)
}
