package ollama

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama/internal/ollamaclient"
)

func newTestClient(t *testing.T, opts ...Option) *LLM {
	t.Helper()
	var ollamaModel string
	if ollamaModel = os.Getenv("OLLAMA_TEST_MODEL"); ollamaModel == "" {
		t.Skip("OLLAMA_TEST_MODEL not set")
		return nil
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

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "feet", strings.ToLower(c1.Content))
}

func TestWithFormat(t *testing.T) {
	t.Parallel()
	llm := newTestClient(t, WithFormat("json"))

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

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "feet", strings.ToLower(c1.Content))

	// check whether we got *any* kind of JSON object.
	var result map[string]any
	err = json.Unmarshal([]byte(c1.Content), &result)
	require.NoError(t, err)
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

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "feet", strings.ToLower(c1.Content))
	assert.Regexp(t, "feet", strings.ToLower(sb.String()))
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

	assert.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]
	assert.Regexp(t, "feet", strings.ToLower(c1.Content))

	vector, err := llm.CreateEmbedding(context.Background(), []string{"test embedding with keep_alive"})
	require.NoError(t, err)
	assert.NotEmpty(t, vector)
}

func TestWithTools(t *testing.T) {
	t.Parallel()
	llm := newTestClient(t, WithTools([]ollamaclient.Tool{
		{
			Type: "function",
			Function: ollamaclient.ToolFunction{
				Name:        "getCurrentWeather",
				Description: "Get the current weather in a given location",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "The city, e.g. San Francisco",
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}))

	parts := []llms.ContentPart{
		llms.TextContent{Text: "What is the weather like in London ?"},
	}
	msgs := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	resp, err := llm.GenerateContent(context.Background(), msgs)
	require.NoError(t, err)

	assert.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]

	assert.NotEmpty(t, c1.ToolCalls)

	tool := c1.ToolCalls[0]

	assert.Equal(t, "getCurrentWeather", tool.FunctionCall.Name)
	assert.Equal(t, "{\"location\":\"london\"}", strings.ToLower(tool.FunctionCall.Arguments))

	toolCallingResponse := llms.TextParts(llms.ChatMessageTypeAI, c1.Content)
	for _, tc := range c1.ToolCalls {
		toolCallingResponse.Parts = append(toolCallingResponse.Parts, tc)
	}
	msgs = append(msgs, toolCallingResponse)

	toolExecResult := "{\"location\":\"london\", \"forecast\":\"63°F and cloudy\"}"

	weatherCallResponse := llms.MessageContent{
		Role: llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{
			llms.ToolCallResponse{
				Name:    tool.FunctionCall.Name,
				Content: toolExecResult,
			},
		},
	}
	msgs = append(msgs, weatherCallResponse)

	resp, err = llm.GenerateContent(context.Background(), msgs)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Choices)
	c1Lower := strings.ToLower(resp.Choices[0].Content)

	assert.Equal(t, "63", c1Lower)
	assert.Equal(t, "cloudy", c1Lower)
	assert.Equal(t, "london", c1Lower)
}
