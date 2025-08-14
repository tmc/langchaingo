package llamafile

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/0xDezzy/langchaingo/llms"
	"github.com/0xDezzy/langchaingo/llms/llamafile/internal/llamafileclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// isLlamafileAvailable checks if the llamafile server is available
func isLlamafileAvailable() bool {
	// Check if CI environment variable is set - skip if in CI
	if os.Getenv("CI") != "" {
		return false
	}

	// Check if LLAMAFILE_HOST is set
	host := os.Getenv("LLAMAFILE_HOST")
	if host == "" {
		host = "http://127.0.0.1:8080"
	}

	// Try to connect to llamafile server
	client := &http.Client{
		Timeout: 2 * time.Second,
		// Don't use system proxy for local llamafile check
		Transport: &http.Transport{
			Proxy: nil,
		},
	}

	// Try the /v1/models endpoint first (standard OpenAI API endpoint)
	resp, err := client.Get(host + "/v1/models")
	if err == nil {
		defer resp.Body.Close()
		return resp.StatusCode == 200
	}

	// Try /health endpoint
	resp, err = client.Get(host + "/health")
	if err == nil {
		defer resp.Body.Close()
		return resp.StatusCode < 500
	}

	// Try root endpoint as last resort
	resp, err = client.Get(host)
	if err == nil {
		defer resp.Body.Close()
		return resp.StatusCode < 500
	}

	// Server is not available
	return false
}

func newTestClient(t *testing.T) *LLM {
	t.Helper()
	options := []Option{
		WithEmbeddingSize(2048),
		WithTemperature(0.8),
	}
	c, err := New(options...)
	require.NoError(t, err)
	return c
}

func TestGenerateContent(t *testing.T) {
	if !isLlamafileAvailable() {
		t.Skip("llamafile is not available")
	}
	t.Parallel()
	ctx := context.Background()
	llm := newTestClient(t)

	parts := []llms.ContentPart{
		llms.TextContent{Text: "Brazil is a country? the answer should just be yes or no"},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	rsp, err := llm.GenerateContent(ctx, content)
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "yes", strings.ToLower(c1.Content))
}

func TestWithStreaming(t *testing.T) {
	if !isLlamafileAvailable() {
		t.Skip("llamafile is not available")
	}
	t.Parallel()
	ctx := context.Background()
	llm := newTestClient(t)

	parts := []llms.ContentPart{
		llms.TextContent{Text: "Brazil is a country? answer yes or no"},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	var sb strings.Builder
	rsp, err := llm.GenerateContent(ctx, content,
		llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			sb.Write(chunk)
			return nil
		}))
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "yes", strings.ToLower(c1.Content))
	assert.Regexp(t, "yes", strings.ToLower(sb.String()))
}

func TestCreateEmbedding(t *testing.T) {
	t.Parallel()
	if !isLlamafileAvailable() {
		t.Skip("llamafile is not available")
	}
	ctx := context.Background()
	llm := newTestClient(t)

	embeddings, err := llm.CreateEmbedding(ctx, []string{"hello", "world"})
	require.NoError(t, err)
	assert.Len(t, embeddings, 2)
}

// Unit tests that don't require external dependencies

func TestNew_UnitTests(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr bool
	}{
		{
			name: "with default options",
			opts: []Option{},
		},
		{
			name: "with multiple options",
			opts: []Option{
				WithModel("llama2"),
				WithTemperature(0.7),
				WithTopP(0.9),
				WithRepeatPenalty(1.1),
			},
		},
		{
			name: "with all options",
			opts: []Option{
				WithFrequencyPenalty(0.5),
				WithGrammar("grammar"),
				WithIgnoreEOS(true),
				WithMinP(0.05),
				WithMirostat(2),
				WithMirostatEta(0.1),
				WithMirostatTau(5.0),
				WithModel("llama2"),
				WithLogitBias([]interface{}{1, 2}),
				WithPenaltyPromptTokens([]interface{}{3, 4}),
				WithPresencePenalty(0.6),
				WithRepeatLastN(64),
				WithRepeatPenalty(1.1),
				WithSeed(42),
				WithStop([]string{"</s>"}),
				WithStream(true),
				WithTemperature(0.8),
				WithTfsZ(1.0),
				WithTopK(40),
				WithTopP(0.95),
				WithTypicalP(1.0),
				WithUsePenaltyPromptTokens(true),
				WithNPredict(128),
				WithNProbs(0),
				WithPenalizeNL(true),
				WithNKeep(0),
				WithNCtx(2048),
				WithEmbeddingSize(384),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm, err := New(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && llm == nil {
				t.Error("New() returned nil LLM without error")
			}
		})
	}
}

func TestTypeToRole(t *testing.T) {
	tests := []struct {
		name     string
		typ      llms.ChatMessageType
		expected string
	}{
		{
			name:     "system message",
			typ:      llms.ChatMessageTypeSystem,
			expected: "system",
		},
		{
			name:     "AI message",
			typ:      llms.ChatMessageTypeAI,
			expected: "assistant",
		},
		{
			name:     "human message",
			typ:      llms.ChatMessageTypeHuman,
			expected: "user",
		},
		{
			name:     "generic message",
			typ:      llms.ChatMessageTypeGeneric,
			expected: "user",
		},
		{
			name:     "function message",
			typ:      llms.ChatMessageTypeFunction,
			expected: "function",
		},
		{
			name:     "tool message",
			typ:      llms.ChatMessageTypeTool,
			expected: "tool",
		},
		{
			name:     "unknown message type",
			typ:      llms.ChatMessageType("unknown"),
			expected: "user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := typeToRole(tt.typ)
			if result != tt.expected {
				t.Errorf("typeToRole() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMakeLlamaOptionsFromOptions(t *testing.T) { //nolint:funlen // comprehensive test
	tests := []struct {
		name     string
		input    *llamafileclient.ChatRequest
		opts     llms.CallOptions
		validate func(t *testing.T, result *llamafileclient.ChatRequest)
	}{
		{
			name:  "basic options",
			input: &llamafileclient.ChatRequest{},
			opts: llms.CallOptions{
				Model:             "llama2",
				Temperature:       0.8,
				MaxTokens:         100,
				TopK:              40,
				TopP:              0.9,
				FrequencyPenalty:  0.5,
				PresencePenalty:   0.6,
				RepetitionPenalty: 1.1,
				Seed:              42,
				StopWords:         []string{"</s>", "<|im_end|>"},
				MinLength:         10,
				N:                 2048,
			},
			validate: func(t *testing.T, result *llamafileclient.ChatRequest) {
				if result.Model != "llama2" {
					t.Errorf("Model = %v, want %v", result.Model, "llama2")
				}
				if result.Temperature != 0.8 {
					t.Errorf("Temperature = %v, want %v", result.Temperature, 0.8)
				}
				if result.NPredict != 100 {
					t.Errorf("NPredict = %v, want %v", result.NPredict, 100)
				}
				if result.TopK != 40 {
					t.Errorf("TopK = %v, want %v", result.TopK, 40)
				}
				if result.TopP != 0.9 {
					t.Errorf("TopP = %v, want %v", result.TopP, 0.9)
				}
				if result.FrequencyPenalty != 0.5 {
					t.Errorf("FrequencyPenalty = %v, want %v", result.FrequencyPenalty, 0.5)
				}
				if result.PresencePenalty != 0.6 {
					t.Errorf("PresencePenalty = %v, want %v", result.PresencePenalty, 0.6)
				}
				if result.RepeatPenalty != 1.1 {
					t.Errorf("RepeatPenalty = %v, want %v", result.RepeatPenalty, 1.1)
				}
				if result.Seed != 42 {
					t.Errorf("Seed = %v, want %v", result.Seed, 42)
				}
				if len(result.Stop) != 2 {
					t.Errorf("Stop length = %v, want %v", len(result.Stop), 2)
				}
				if result.MinP != 10.0 {
					t.Errorf("MinP = %v, want %v", result.MinP, 10.0)
				}
				if result.NCtx != 2048 {
					t.Errorf("NCtx = %v, want %v", result.NCtx, 2048)
				}
				if result.Stream == nil || *result.Stream != false {
					t.Error("Stream should be false when StreamingFunc is nil")
				}
			},
		},
		{
			name:  "with streaming",
			input: &llamafileclient.ChatRequest{},
			opts: llms.CallOptions{
				StreamingFunc: func(ctx context.Context, chunk []byte) error {
					return nil
				},
			},
			validate: func(t *testing.T, result *llamafileclient.ChatRequest) {
				if result.Stream == nil || *result.Stream != true {
					t.Error("Stream should be true when StreamingFunc is provided")
				}
			},
		},
		{
			name:  "empty options",
			input: &llamafileclient.ChatRequest{},
			opts:  llms.CallOptions{},
			validate: func(t *testing.T, result *llamafileclient.ChatRequest) {
				if result.Stream == nil || *result.Stream != false {
					t.Error("Stream should be false by default")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeLlamaOptionsFromOptions(tt.input, tt.opts)
			tt.validate(t, result)
		})
	}
}

func TestGenerateContent_UnitValidation(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		messages      []llms.MessageContent
		wantErr       bool
		expectedError string
	}{
		{
			name: "multiple text parts error",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "Hello"},
						llms.TextContent{Text: "World"},
					},
				},
			},
			wantErr:       true,
			expectedError: "expecting a single Text content",
		},
		{
			name: "unsupported content type",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.ImageURLContent{URL: "http://example.com/image.jpg"},
					},
				},
			},
			wantErr:       true,
			expectedError: "only support Text and BinaryContent parts right now",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := &LLM{}

			_, err := llm.GenerateContent(ctx, tt.messages)

			if tt.wantErr && tt.expectedError != "" {
				if err == nil {
					t.Errorf("GenerateContent() expected error containing %q, got nil", tt.expectedError)
				} else if !containsString(err.Error(), tt.expectedError) {
					t.Errorf("GenerateContent() error = %v, want error containing %q", err, tt.expectedError)
				}
			}
		})
	}
}

func TestErrorConstants(t *testing.T) {
	if ErrEmptyResponse == nil {
		t.Error("ErrEmptyResponse should not be nil")
	}
	if ErrEmptyResponse.Error() != "no response" {
		t.Errorf("ErrEmptyResponse.Error() = %v, want %v", ErrEmptyResponse.Error(), "no response")
	}

	if ErrIncompleteEmbedding == nil {
		t.Error("ErrIncompleteEmbedding should not be nil")
	}
	if ErrIncompleteEmbedding.Error() != "not all input got embedded" {
		t.Errorf("ErrIncompleteEmbedding.Error() = %v, want %v", ErrIncompleteEmbedding.Error(), "not all input got embedded")
	}
}

func TestLLMImplementsModel(t *testing.T) {
	// This test verifies that LLM implements the llms.Model interface
	var _ llms.Model = (*LLM)(nil)
}

// Test all option functions
func TestOptions(t *testing.T) { //nolint:funlen // comprehensive test
	tests := []struct {
		name     string
		option   Option
		validate func(t *testing.T, g *llamafileclient.GenerationSettings)
	}{
		{
			name:   "WithFrequencyPenalty",
			option: WithFrequencyPenalty(0.7),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if g.FrequencyPenalty != 0.7 {
					t.Errorf("FrequencyPenalty = %v, want %v", g.FrequencyPenalty, 0.7)
				}
			},
		},
		{
			name:   "WithGrammar",
			option: WithGrammar("test-grammar"),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if g.Grammar != "test-grammar" {
					t.Errorf("Grammar = %v, want %v", g.Grammar, "test-grammar")
				}
			},
		},
		{
			name:   "WithIgnoreEOS",
			option: WithIgnoreEOS(true),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if g.IgnoreEOS != true {
					t.Errorf("IgnoreEOS = %v, want %v", g.IgnoreEOS, true)
				}
			},
		},
		{
			name:   "WithMinP",
			option: WithMinP(0.05),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if g.MinP != 0.05 {
					t.Errorf("MinP = %v, want %v", g.MinP, 0.05)
				}
			},
		},
		{
			name:   "WithMirostat",
			option: WithMirostat(2),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if g.Mirostat != 2 {
					t.Errorf("Mirostat = %v, want %v", g.Mirostat, 2)
				}
			},
		},
		{
			name:   "WithMirostatEta",
			option: WithMirostatEta(0.1),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if g.MirostatEta != 0.1 {
					t.Errorf("MirostatEta = %v, want %v", g.MirostatEta, 0.1)
				}
			},
		},
		{
			name:   "WithMirostatTau",
			option: WithMirostatTau(5.0),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if g.MirostatTau != 5.0 {
					t.Errorf("MirostatTau = %v, want %v", g.MirostatTau, 5.0)
				}
			},
		},
		{
			name:   "WithModel",
			option: WithModel("llama2-7b"),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if g.Model != "llama2-7b" {
					t.Errorf("Model = %v, want %v", g.Model, "llama2-7b")
				}
			},
		},
		{
			name:   "WithLogitBias",
			option: WithLogitBias([]interface{}{1, 2, 3}),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if len(g.LogitBias) != 3 {
					t.Errorf("LogitBias length = %v, want %v", len(g.LogitBias), 3)
				}
			},
		},
		{
			name:   "WithPenaltyPromptTokens",
			option: WithPenaltyPromptTokens([]interface{}{10, 20}),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if len(g.PenaltyPromptTokens) != 2 {
					t.Errorf("PenaltyPromptTokens length = %v, want %v", len(g.PenaltyPromptTokens), 2)
				}
			},
		},
		{
			name:   "WithPresencePenalty",
			option: WithPresencePenalty(0.6),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if g.PresencePenalty != 0.6 {
					t.Errorf("PresencePenalty = %v, want %v", g.PresencePenalty, 0.6)
				}
			},
		},
		{
			name:   "WithRepeatLastN",
			option: WithRepeatLastN(64),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if g.RepeatLastN != 64 {
					t.Errorf("RepeatLastN = %v, want %v", g.RepeatLastN, 64)
				}
			},
		},
		{
			name:   "WithRepeatPenalty",
			option: WithRepeatPenalty(1.1),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if g.RepeatPenalty != 1.1 {
					t.Errorf("RepeatPenalty = %v, want %v", g.RepeatPenalty, 1.1)
				}
			},
		},
		{
			name:   "WithSeed",
			option: WithSeed(42),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if g.Seed != 42 {
					t.Errorf("Seed = %v, want %v", g.Seed, 42)
				}
			},
		},
		{
			name:   "WithStop",
			option: WithStop([]string{"</s>", "<|im_end|>"}),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if len(g.Stop) != 2 {
					t.Errorf("Stop length = %v, want %v", len(g.Stop), 2)
				}
			},
		},
		{
			name:   "WithStream",
			option: WithStream(true),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if g.Stream != true {
					t.Errorf("Stream = %v, want %v", g.Stream, true)
				}
			},
		},
		{
			name:   "WithTemperature",
			option: WithTemperature(0.8),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if g.Temperature != 0.8 {
					t.Errorf("Temperature = %v, want %v", g.Temperature, 0.8)
				}
			},
		},
		{
			name:   "WithTfsZ",
			option: WithTfsZ(1.0),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if g.TfsZ != 1.0 {
					t.Errorf("TfsZ = %v, want %v", g.TfsZ, 1.0)
				}
			},
		},
		{
			name:   "WithTopK",
			option: WithTopK(40),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if g.TopK != 40 {
					t.Errorf("TopK = %v, want %v", g.TopK, 40)
				}
			},
		},
		{
			name:   "WithTopP",
			option: WithTopP(0.95),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if g.TopP != 0.95 {
					t.Errorf("TopP = %v, want %v", g.TopP, 0.95)
				}
			},
		},
		{
			name:   "WithTypicalP",
			option: WithTypicalP(1.0),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if g.TypicalP != 1.0 {
					t.Errorf("TypicalP = %v, want %v", g.TypicalP, 1.0)
				}
			},
		},
		{
			name:   "WithUsePenaltyPromptTokens",
			option: WithUsePenaltyPromptTokens(true),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if g.UsePenaltyPromptTokens != true {
					t.Errorf("UsePenaltyPromptTokens = %v, want %v", g.UsePenaltyPromptTokens, true)
				}
			},
		},
		{
			name:   "WithNPredict",
			option: WithNPredict(128),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if g.NPredict != 128 {
					t.Errorf("NPredict = %v, want %v", g.NPredict, 128)
				}
			},
		},
		{
			name:   "WithNProbs",
			option: WithNProbs(10),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if g.NProbs != 10 {
					t.Errorf("NProbs = %v, want %v", g.NProbs, 10)
				}
			},
		},
		{
			name:   "WithPenalizeNL",
			option: WithPenalizeNL(true),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if g.PenalizeNL != true {
					t.Errorf("PenalizeNL = %v, want %v", g.PenalizeNL, true)
				}
			},
		},
		{
			name:   "WithNKeep",
			option: WithNKeep(5),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if g.NKeep != 5 {
					t.Errorf("NKeep = %v, want %v", g.NKeep, 5)
				}
			},
		},
		{
			name:   "WithNCtx",
			option: WithNCtx(2048),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if g.NCtx != 2048 {
					t.Errorf("NCtx = %v, want %v", g.NCtx, 2048)
				}
			},
		},
		{
			name:   "WithEmbeddingSize",
			option: WithEmbeddingSize(384),
			validate: func(t *testing.T, g *llamafileclient.GenerationSettings) {
				if g.EmbeddingSize != 384 {
					t.Errorf("EmbeddingSize = %v, want %v", g.EmbeddingSize, 384)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &llamafileclient.GenerationSettings{}
			tt.option(g)
			tt.validate(t, g)
		})
	}
}

// Helper function
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}
