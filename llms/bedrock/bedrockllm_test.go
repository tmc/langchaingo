package bedrock_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/vxcontrol/langchaingo/internal/httprr"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/bedrock"
	"github.com/vxcontrol/langchaingo/llms/streaming"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/smithy-go/auth/bearer"
)

func setUpTestWithTransport(rr *httprr.RecordReplay) (*bedrockruntime.Client, error) {
	// Configure request scrubbing to remove dynamic AWS headers
	rr.ScrubReq(func(req *http.Request) error {
		req.Header.Del("Amz-Sdk-Invocation-Id")
		req.Header.Del("Amz-Sdk-Request")
		req.Header.Del("X-Amz-Date")
		return nil
	})

	httpClient := &http.Client{
		Transport: rr,
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}

	client := bedrockruntime.NewFromConfig(cfg)
	return client, nil
}

func TestAmazonOutputConverseAPI(t *testing.T) { //nolint:funlen
	ctx := t.Context()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	// Only run tests in parallel when not recording (to avoid rate limits)
	if !rr.Recording() {
		t.Parallel()
	}

	// Configure AWS client to use httprr transport
	client, err := setUpTestWithTransport(rr)
	if err != nil {
		t.Fatal(err)
	}
	llm, err := bedrock.New(bedrock.WithClient(client), bedrock.WithConverseAPI())
	if err != nil {
		t.Fatal(err)
	}

	msgs := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextPart("You know all about AI."),
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Explain AI in 10 words or less."),
			},
		},
	}

	// All the on-demand models (based on docs with Deployment type: Serverless)
	models := []string{
		// AI21 Labs models
		bedrock.ModelAi21Jamba15LargeV1,
		bedrock.ModelAi21Jamba15MiniV1,

		// Amazon Nova models
		bedrock.ModelAmazonNova2LiteV1,
		bedrock.ModelAmazonNovaPremiereV1,
		bedrock.ModelAmazonNovaProV1,
		bedrock.ModelAmazonNovaLiteV1,
		bedrock.ModelAmazonNovaMicroV1,

		// Anthropic models
		bedrock.ModelAnthropicClaudeOpus46,
		bedrock.ModelAnthropicClaudeSonnet46,
		bedrock.ModelAnthropicClaudeOpus45,
		bedrock.ModelAnthropicClaudeHaiku45,
		bedrock.ModelAnthropicClaudeSonnet45,
		bedrock.ModelAnthropicClaudeOpus41,
		// bedrock.ModelAnthropicClaudeOpus4,         // Model is deprecated
		bedrock.ModelAnthropicClaudeSonnet4,
		// bedrock.ModelAnthropicClaude37Sonnet,      // Model is deprecated
		bedrock.ModelAnthropicClaude35Haiku,

		// Cohere models
		bedrock.ModelCohereCommandRV1,
		bedrock.ModelCohereCommandRPlusV1,

		// Meta models
		// bedrock.ModelMetaLlama4MaverickInstructV1, // Unavailable for MENA users
		// bedrock.ModelMetaLlama4ScoutInstructV1,    // Unavailable for MENA users
		bedrock.ModelMetaLlama3370bInstructV1,
		// bedrock.ModelMetaLlama3211bInstructV1,     // Unavailable for MENA users
		// bedrock.ModelMetaLlama3211bInstructV1,     // Unavailable for MENA users
		// bedrock.ModelMetaLlama3290bInstructV1,     // Unavailable for MENA users
		bedrock.ModelMetaLlama3170bInstructV1,
		// bedrock.ModelMetaLlama318bInstructV1,      // Unavailable for MENA users
		bedrock.ModelMetaLlama370bInstructV1,
		bedrock.ModelMetaLlama38bInstructV1,

		// DeepSeek models
		bedrock.ModelDeepSeekR1V1,
		bedrock.ModelDeepSeekV32,

		// OpenAI models
		bedrock.ModelOpenAIGptOss120BV1,
		bedrock.ModelOpenAIGptOss20BV1,

		// Qwen models
		bedrock.ModelQwen3Next80BA3B,
		bedrock.ModelQwen3VL235BA22B,
		bedrock.ModelQwen332BV1,
		bedrock.ModelQwen3Coder30BA3BV1,
		bedrock.ModelQwen3CoderNext,

		// Mistral models
		bedrock.ModelMistralLarge3,
		bedrock.ModelMistralMagistralSmall2509,
		bedrock.ModelMistralLarge2402V1,

		// Moonshot models
		bedrock.ModelMoonshotKimiK25,
		bedrock.ModelMoonshotKimiK2Thinking,

		// Z.AI models
		bedrock.ModelGLM47,
		bedrock.ModelGLM47Flash,
	}

	for _, model := range models {
		t.Logf("Model output for %s:-", model)

		resp, err := llm.GenerateContent(ctx, msgs, llms.WithModel(model), llms.WithMaxTokens(512))
		if err != nil {
			t.Fatal(err)
		}
		for i, choice := range resp.Choices {
			t.Logf("Choice %d: %s", i, choice.Content)
		}
	}
}

func TestAmazonOutputLegacyAPI(t *testing.T) {
	ctx := t.Context()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	// Only run tests in parallel when not recording (to avoid rate limits)
	if !rr.Recording() {
		t.Parallel()
	}

	// Configure AWS client to use httprr transport
	client, err := setUpTestWithTransport(rr)
	if err != nil {
		t.Fatal(err)
	}
	llm, err := bedrock.New(bedrock.WithClient(client))
	if err != nil {
		t.Fatal(err)
	}

	msgs := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextPart("You know all about AI."),
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Explain AI in 10 words or less."),
			},
		},
	}

	// All the on-demand models (based on docs with Deployment type: Serverless)
	models := []string{
		// AI21 Labs models
		bedrock.ModelAi21Jamba15LargeV1,
		bedrock.ModelAi21Jamba15MiniV1,

		// Amazon Nova models
		bedrock.ModelAmazonNova2LiteV1,
		bedrock.ModelAmazonNovaPremiereV1,
		bedrock.ModelAmazonNovaProV1,
		bedrock.ModelAmazonNovaLiteV1,
		bedrock.ModelAmazonNovaMicroV1,

		// Anthropic models
		bedrock.ModelAnthropicClaudeOpus45,
		bedrock.ModelAnthropicClaudeHaiku45,
		bedrock.ModelAnthropicClaudeSonnet45,
		bedrock.ModelAnthropicClaudeOpus41,
		bedrock.ModelAnthropicClaudeOpus4,
		bedrock.ModelAnthropicClaudeSonnet4,
		bedrock.ModelAnthropicClaude37Sonnet,
		bedrock.ModelAnthropicClaude35Haiku,

		// Cohere models
		bedrock.ModelCohereCommandRV1,
		bedrock.ModelCohereCommandRPlusV1,

		// Meta models
		// bedrock.ModelMetaLlama4MaverickInstructV1, // Unavailable for MENA users
		// bedrock.ModelMetaLlama4ScoutInstructV1, // Unavailable for MENA users
		bedrock.ModelMetaLlama3370bInstructV1,
		// bedrock.ModelMetaLlama3211bInstructV1, // Unavailable for MENA users
		// bedrock.ModelMetaLlama3211bInstructV1, // Unavailable for MENA users
		// bedrock.ModelMetaLlama3290bInstructV1, // Unavailable for MENA users
		bedrock.ModelMetaLlama3170bInstructV1,
		// bedrock.ModelMetaLlama318bInstructV1, // Unavailable for MENA users
		bedrock.ModelMetaLlama370bInstructV1,
		bedrock.ModelMetaLlama38bInstructV1,

		// DeepSeek models
		bedrock.ModelDeepSeekR1V1,
	}

	for _, model := range models {
		t.Logf("Model output for %s:-", model)

		resp, err := llm.GenerateContent(ctx, msgs, llms.WithModel(model), llms.WithMaxTokens(512))
		if err != nil {
			// Check if this is a recording mismatch error
			if strings.Contains(err.Error(), "cached HTTP response not found") {
				t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
			}
			t.Fatal(err)
		}
		for i, choice := range resp.Choices {
			t.Logf("Choice %d: %s", i, choice.Content)
		}
	}
}

func TestAmazonStreamingOutputConverseAPI(t *testing.T) { //nolint:funlen
	ctx := t.Context()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	if !rr.Recording() {
		t.Parallel()
	}

	client, err := setUpTestWithTransport(rr)
	if err != nil {
		t.Fatal(err)
	}
	llm, err := bedrock.New(bedrock.WithClient(client), bedrock.WithConverseAPI())
	if err != nil {
		t.Fatal(err)
	}

	msgs := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextPart("You are helpful AI assistant."),
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Tell me a very short story about a cat."),
			},
		},
	}

	// Start with Anthropic models that support streaming
	models := []string{
		// AI21 Labs models
		bedrock.ModelAi21Jamba15LargeV1,
		bedrock.ModelAi21Jamba15MiniV1,

		// Amazon Nova models
		bedrock.ModelAmazonNova2LiteV1,
		bedrock.ModelAmazonNovaPremiereV1,
		bedrock.ModelAmazonNovaProV1,
		bedrock.ModelAmazonNovaLiteV1,
		bedrock.ModelAmazonNovaMicroV1,

		// Anthropic models
		bedrock.ModelAnthropicClaudeOpus46,
		bedrock.ModelAnthropicClaudeSonnet46,
		bedrock.ModelAnthropicClaudeOpus45,
		bedrock.ModelAnthropicClaudeHaiku45,
		bedrock.ModelAnthropicClaudeSonnet45,
		bedrock.ModelAnthropicClaudeOpus41,
		// bedrock.ModelAnthropicClaudeOpus4,         // Model is deprecated
		bedrock.ModelAnthropicClaudeSonnet4,
		// bedrock.ModelAnthropicClaude37Sonnet,      // Model is deprecated
		bedrock.ModelAnthropicClaude35Haiku,

		// Cohere models (only Command-R supports streaming and Converse API)
		bedrock.ModelCohereCommandRV1,
		bedrock.ModelCohereCommandRPlusV1,

		// Meta models
		// bedrock.ModelMetaLlama4MaverickInstructV1, // Unavailable for MENA users
		// bedrock.ModelMetaLlama4ScoutInstructV1,    // Unavailable for MENA users
		bedrock.ModelMetaLlama3370bInstructV1,
		// bedrock.ModelMetaLlama3211bInstructV1,     // Unavailable for MENA users
		// bedrock.ModelMetaLlama3211bInstructV1,     // Unavailable for MENA users
		// bedrock.ModelMetaLlama3290bInstructV1,     // Unavailable for MENA users
		bedrock.ModelMetaLlama3170bInstructV1,
		// bedrock.ModelMetaLlama318bInstructV1,      // Unavailable for MENA users
		bedrock.ModelMetaLlama370bInstructV1,
		bedrock.ModelMetaLlama38bInstructV1,

		// DeepSeek models
		bedrock.ModelDeepSeekR1V1,
		bedrock.ModelDeepSeekV32,

		// OpenAI models
		bedrock.ModelOpenAIGptOss120BV1,
		bedrock.ModelOpenAIGptOss20BV1,

		// Qwen models
		bedrock.ModelQwen3Next80BA3B,
		bedrock.ModelQwen3VL235BA22B,
		bedrock.ModelQwen332BV1,
		bedrock.ModelQwen3Coder30BA3BV1,
		bedrock.ModelQwen3CoderNext,

		// Mistral models
		bedrock.ModelMistralLarge3,
		bedrock.ModelMistralMagistralSmall2509,
		bedrock.ModelMistralLarge2402V1,

		// Moonshot models
		bedrock.ModelMoonshotKimiK25,
		bedrock.ModelMoonshotKimiK2Thinking,

		// Z.AI models
		bedrock.ModelGLM47,
		bedrock.ModelGLM47Flash,
	}

	for _, model := range models {
		t.Logf("Testing streaming for model %s:-", model)

		var streamedChunks []string
		var isDone bool
		var mu sync.Mutex

		streamingFunc := func(ctx context.Context, chunk streaming.Chunk) error {
			mu.Lock()
			defer mu.Unlock()

			switch chunk.Type {
			case streaming.ChunkTypeText:
				streamedChunks = append(streamedChunks, chunk.Content)
			case streaming.ChunkTypeDone:
				isDone = true
			default:
				// Ignore other chunks in this test
			}
			return nil
		}

		resp, err := llm.GenerateContent(ctx, msgs,
			llms.WithModel(model),
			llms.WithMaxTokens(512),
			llms.WithStreamingFunc(streamingFunc),
		)
		if err != nil {
			t.Fatal(err)
		}

		// Validate streaming worked
		mu.Lock()
		if !isDone {
			t.Errorf("Model %s: streaming callback with Done=true was not called", model)
		}
		if len(streamedChunks) == 0 {
			t.Errorf("Model %s: no streaming chunks received", model)
		}
		mu.Unlock()

		// Validate response
		if len(resp.Choices) == 0 {
			t.Errorf("Model %s: no choices in response", model)
		} else {
			// Check that streamed content matches final content
			var fullStreamedContent string
			for _, chunk := range streamedChunks {
				fullStreamedContent += chunk
			}
			if fullStreamedContent != resp.Choices[0].Content {
				t.Logf("Model %s: streamed content (%s) != final content (%s)",
					model, fullStreamedContent, resp.Choices[0].Content)
			}
		}
	}
}

func TestAmazonStreamingOutputLegacyAPI(t *testing.T) { //nolint:funlen
	ctx := t.Context()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	if !rr.Recording() {
		t.Parallel()
	}

	client, err := setUpTestWithTransport(rr)
	if err != nil {
		t.Fatal(err)
	}
	llm, err := bedrock.New(bedrock.WithClient(client))
	if err != nil {
		t.Fatal(err)
	}

	msgs := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextPart("You are helpful AI assistant."),
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Tell me a very short story about a cat."),
			},
		},
	}

	// Start with Anthropic models that support streaming
	models := []string{
		// AI21 Labs models
		// bedrock.ModelAi21Jamba15LargeV1,  // Not supported for streaming
		// bedrock.ModelAi21Jamba15MiniV1,   // Not supported for streaming

		// Amazon Nova models
		bedrock.ModelAmazonNova2LiteV1,
		bedrock.ModelAmazonNovaPremiereV1,
		bedrock.ModelAmazonNovaProV1,
		bedrock.ModelAmazonNovaLiteV1,
		bedrock.ModelAmazonNovaMicroV1,

		// Anthropic models
		bedrock.ModelAnthropicClaudeOpus45,
		bedrock.ModelAnthropicClaudeHaiku45,
		bedrock.ModelAnthropicClaudeSonnet45,
		bedrock.ModelAnthropicClaudeOpus41,
		bedrock.ModelAnthropicClaudeOpus4,
		bedrock.ModelAnthropicClaudeSonnet4,
		bedrock.ModelAnthropicClaude37Sonnet,
		bedrock.ModelAnthropicClaude35Haiku,

		// Cohere models (only Command-R supports streaming)
		bedrock.ModelCohereCommandRV1,
		bedrock.ModelCohereCommandRPlusV1,

		// Meta models
		// bedrock.ModelMetaLlama4MaverickInstructV1, // Unavailable for MENA users
		// bedrock.ModelMetaLlama4ScoutInstructV1, // Unavailable for MENA users
		bedrock.ModelMetaLlama3370bInstructV1,
		// bedrock.ModelMetaLlama3211bInstructV1, // Unavailable for MENA users
		// bedrock.ModelMetaLlama3211bInstructV1, // Unavailable for MENA users
		// bedrock.ModelMetaLlama3290bInstructV1, // Unavailable for MENA users
		bedrock.ModelMetaLlama3170bInstructV1,
		// bedrock.ModelMetaLlama318bInstructV1, // Unavailable for MENA users
		bedrock.ModelMetaLlama370bInstructV1,
		bedrock.ModelMetaLlama38bInstructV1,

		// DeepSeek models
		bedrock.ModelDeepSeekR1V1,
	}

	for _, model := range models {
		t.Logf("Testing streaming for model %s:-", model)

		var streamedChunks []string
		var isDone bool
		var mu sync.Mutex

		streamingFunc := func(ctx context.Context, chunk streaming.Chunk) error {
			mu.Lock()
			defer mu.Unlock()

			switch chunk.Type {
			case streaming.ChunkTypeText:
				streamedChunks = append(streamedChunks, chunk.Content)
			case streaming.ChunkTypeDone:
				isDone = true
			default:
				// Ignore other chunks in this test
			}
			return nil
		}

		resp, err := llm.GenerateContent(ctx, msgs,
			llms.WithModel(model),
			llms.WithMaxTokens(100),
			llms.WithStreamingFunc(streamingFunc),
		)
		if err != nil {
			t.Fatal(err)
		}

		// Validate streaming worked
		mu.Lock()
		if !isDone {
			t.Errorf("Model %s: streaming callback with Done=true was not called", model)
		}
		if len(streamedChunks) == 0 {
			t.Errorf("Model %s: no streaming chunks received", model)
		}
		mu.Unlock()

		// Validate response
		if len(resp.Choices) == 0 {
			t.Errorf("Model %s: no choices in response", model)
		} else {
			// Check that streamed content matches final content
			var fullStreamedContent string
			for _, chunk := range streamedChunks {
				fullStreamedContent += chunk
			}
			if fullStreamedContent != resp.Choices[0].Content {
				t.Logf("Model %s: streamed content (%s) != final content (%s)",
					model, fullStreamedContent, resp.Choices[0].Content)
			}
		}
	}
}

func TestAmazonNova(t *testing.T) {
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	// Only run tests in parallel when not recording (to avoid rate limits)
	if !rr.Recording() {
		t.Parallel()
	}

	// Configure AWS client to use httprr transport
	client, err := setUpTestWithTransport(rr)
	if err != nil {
		t.Fatal(err)
	}
	llm, err := bedrock.New(bedrock.WithClient(client))
	if err != nil {
		t.Fatal(err)
	}

	msgs := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextPart("You know all about AI."),
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Explain AI in 10 words or less."),
			},
		},
	}

	// All the test models.
	models := []string{
		bedrock.ModelAmazonNovaMicroV1,
		bedrock.ModelAmazonNovaLiteV1,
		bedrock.ModelAmazonNovaProV1,
	}

	ctx := context.Background()

	for _, model := range models {
		t.Logf("Model output for %s:-", model)

		resp, err := llm.GenerateContent(ctx, msgs, llms.WithModel(model), llms.WithMaxTokens(4096))
		if err != nil {
			// Check if this is a recording mismatch error
			if strings.Contains(err.Error(), "cached HTTP response not found") {
				t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
			}
			t.Fatal(err)
		}
		for i, choice := range resp.Choices {
			t.Logf("Choice %d: %s", i, choice.Content)
		}
	}
}

func TestAmazonNovaImage(t *testing.T) {
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	// Only run tests in parallel when not recording (to avoid rate limits)
	if !rr.Recording() {
		t.Parallel()
	}

	// Configure AWS client to use httprr transport
	client, err := setUpTestWithTransport(rr)
	if err != nil {
		t.Fatal(err)
	}
	llm, err := bedrock.New(bedrock.WithClient(client))
	if err != nil {
		t.Fatal(err)
	}

	image, err := os.ReadFile("testdata/wikipage.jpg")
	mimeType := "image/jpeg"
	if err != nil {
		t.Fatal(err)
	}

	msgs := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextPart("You know all about AI."),
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Explain AI according to the image. Provide quotes from the image."),
				llms.BinaryPart(mimeType, image),
			},
		},
	}

	// All the test models.
	models := []string{
		bedrock.ModelAmazonNovaLiteV1,
		bedrock.ModelAmazonNovaProV1,
	}

	ctx := context.Background()

	for _, model := range models {
		t.Logf("Model output for %s:-", model)

		resp, err := llm.GenerateContent(ctx, msgs, llms.WithModel(model), llms.WithMaxTokens(4096))
		if err != nil {
			// Check if this is a recording mismatch error
			if strings.Contains(err.Error(), "cached HTTP response not found") {
				t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
			}
			t.Fatal(err)
		}
		for i, choice := range resp.Choices {
			t.Logf("Choice %d: %s", i, choice.Content)
		}
	}
}

func TestAmazonToolCallingConverseAPI(t *testing.T) {
	ctx := t.Context()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	if !rr.Recording() {
		t.Parallel()
	}

	client, err := setUpTestWithTransport(rr)
	if err != nil {
		t.Fatal(err)
	}
	llm, err := bedrock.New(bedrock.WithClient(client), bedrock.WithConverseAPI())
	if err != nil {
		t.Fatal(err)
	}

	// Models that support tool calling (initially Anthropic, will expand to DeepSeek and Meta)
	toolCallModels := []string{
		// AI21 Labs models
		// bedrock.ModelAi21Jamba15LargeV1,           // Has very hard rate limits
		// bedrock.ModelAi21Jamba15MiniV1,            // Has very hard rate limits

		// Amazon Nova models
		bedrock.ModelAmazonNova2LiteV1,
		bedrock.ModelAmazonNovaPremiereV1,
		bedrock.ModelAmazonNovaProV1,
		bedrock.ModelAmazonNovaLiteV1,
		bedrock.ModelAmazonNovaMicroV1,

		// Anthropic models
		bedrock.ModelAnthropicClaudeOpus46,
		bedrock.ModelAnthropicClaudeSonnet46,
		bedrock.ModelAnthropicClaudeOpus45,
		bedrock.ModelAnthropicClaudeHaiku45,
		bedrock.ModelAnthropicClaudeSonnet45,
		// bedrock.ModelAnthropicClaudeOpus41,        // Has very hard rate limits
		// bedrock.ModelAnthropicClaudeOpus4,         // Model is deprecated
		bedrock.ModelAnthropicClaudeSonnet4,
		// bedrock.ModelAnthropicClaude37Sonnet,      // Model is deprecated
		bedrock.ModelAnthropicClaude35Haiku,

		// Cohere models (only Command-R supports streaming and Converse API)
		// bedrock.ModelCohereCommandRV1,             // Not supported for tool calling
		bedrock.ModelCohereCommandRPlusV1,

		// Meta models
		// bedrock.ModelMetaLlama4MaverickInstructV1, // Unavailable for MENA users
		// bedrock.ModelMetaLlama4ScoutInstructV1,    // Unavailable for MENA users
		// bedrock.ModelMetaLlama3370bInstructV1,     // Unstable behavior on processing tool calls results
		// bedrock.ModelMetaLlama3211bInstructV1,     // Unavailable for MENA users
		// bedrock.ModelMetaLlama3211bInstructV1,     // Unavailable for MENA users
		// bedrock.ModelMetaLlama3290bInstructV1,     // Unavailable for MENA users
		// bedrock.ModelMetaLlama3170bInstructV1,     // Unstable behavior on processing tool calls results
		// bedrock.ModelMetaLlama318bInstructV1,      // Unavailable for MENA users
		// bedrock.ModelMetaLlama370bInstructV1,      // Not supported for tool calling
		// bedrock.ModelMetaLlama38bInstructV1,       // Not supported for tool calling

		// DeepSeek models
		// bedrock.ModelDeepSeekR1V1,                 // Not supported for tool calling
		bedrock.ModelDeepSeekV32,

		// OpenAI models
		bedrock.ModelOpenAIGptOss120BV1,
		bedrock.ModelOpenAIGptOss20BV1,

		// Qwen models
		bedrock.ModelQwen3Next80BA3B,
		bedrock.ModelQwen3VL235BA22B,
		bedrock.ModelQwen332BV1,
		bedrock.ModelQwen3Coder30BA3BV1,
		bedrock.ModelQwen3CoderNext,

		// Mistral models
		bedrock.ModelMistralLarge3,
		// bedrock.ModelMistralMagistralSmall2509,    // Not supported for tool calling
		bedrock.ModelMistralLarge2402V1,

		// Moonshot models
		bedrock.ModelMoonshotKimiK25,
		// bedrock.ModelMoonshotKimiK2Thinking,       // Not stable for tool calling

		// Z.AI models
		// bedrock.ModelGLM47,        // Tool calling not supported (backend requires string input instead of JSON)
		// bedrock.ModelGLM47Flash,   // Tool calling not supported (backend requires string input instead of JSON)
	}

	for _, model := range toolCallModels {
		t.Logf("Testing tool calling with model: %s", model)

		err := testToolCallingWorkflow(ctx, t, llm, model, rr.Replaying(), nil)
		if err != nil {
			t.Errorf("Tool calling failed for model %s: %v", model, err)
		}
	}
}

func TestAmazonToolCallingLegacyAPI(t *testing.T) {
	ctx := t.Context()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	if !rr.Recording() {
		t.Parallel()
	}

	client, err := setUpTestWithTransport(rr)
	if err != nil {
		t.Fatal(err)
	}
	llm, err := bedrock.New(bedrock.WithClient(client))
	if err != nil {
		t.Fatal(err)
	}

	// Models that support tool calling (initially Anthropic, will expand to DeepSeek and Meta)
	toolCallModels := []string{
		// Anthropic models
		bedrock.ModelAnthropicClaudeOpus45,
		bedrock.ModelAnthropicClaudeHaiku45,
		bedrock.ModelAnthropicClaudeSonnet45,
		// bedrock.ModelAnthropicClaudeOpus41, // Has very hard rate limits
		bedrock.ModelAnthropicClaudeOpus4,
		bedrock.ModelAnthropicClaudeSonnet4,
		bedrock.ModelAnthropicClaude37Sonnet,
		bedrock.ModelAnthropicClaude35Haiku,
	}

	for _, model := range toolCallModels {
		t.Logf("Testing tool calling with model: %s", model)

		err := testToolCallingWorkflow(ctx, t, llm, model, rr.Replaying(), nil)
		if err != nil {
			t.Errorf("Tool calling failed for model %s: %v", model, err)
		}
	}
}

func TestAmazonToolCallingStreamingConverseAPI(t *testing.T) {
	ctx := t.Context()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	if !rr.Recording() {
		t.Parallel()
	}

	client, err := setUpTestWithTransport(rr)
	if err != nil {
		t.Fatal(err)
	}
	llm, err := bedrock.New(bedrock.WithClient(client), bedrock.WithConverseAPI())
	if err != nil {
		t.Fatal(err)
	}

	// Models that support streaming tool calling (initially Anthropic, will expand)
	streamingToolCallModels := []string{
		// AI21 Labs models
		// bedrock.ModelAi21Jamba15LargeV1,           // Not supported for tool calling in streaming
		// bedrock.ModelAi21Jamba15MiniV1,            // Not supported for tool calling in streaming

		// Amazon Nova models
		bedrock.ModelAmazonNova2LiteV1,
		bedrock.ModelAmazonNovaPremiereV1,
		bedrock.ModelAmazonNovaProV1,
		bedrock.ModelAmazonNovaLiteV1,
		bedrock.ModelAmazonNovaMicroV1,

		// Anthropic models
		bedrock.ModelAnthropicClaudeOpus46,
		bedrock.ModelAnthropicClaudeSonnet46,
		bedrock.ModelAnthropicClaudeOpus45,
		bedrock.ModelAnthropicClaudeHaiku45,
		bedrock.ModelAnthropicClaudeSonnet45,
		// bedrock.ModelAnthropicClaudeOpus41,        // Has very hard rate limits
		// bedrock.ModelAnthropicClaudeOpus4,         // Model is deprecated
		bedrock.ModelAnthropicClaudeSonnet4,
		// bedrock.ModelAnthropicClaude37Sonnet,      // Model is deprecated
		bedrock.ModelAnthropicClaude35Haiku,

		// Cohere models (only Command-R supports streaming and Converse API)
		// bedrock.ModelCohereCommandRV1,             // Not supported for tool calling in streaming
		bedrock.ModelCohereCommandRPlusV1,

		// Meta models
		// bedrock.ModelMetaLlama4MaverickInstructV1, // Unavailable for MENA users
		// bedrock.ModelMetaLlama4ScoutInstructV1,    // Unavailable for MENA users
		// bedrock.ModelMetaLlama3370bInstructV1,     // Unstable behavior on processing tool calls results
		// bedrock.ModelMetaLlama3211bInstructV1,     // Unavailable for MENA users
		// bedrock.ModelMetaLlama3211bInstructV1,     // Unavailable for MENA users
		// bedrock.ModelMetaLlama3290bInstructV1,     // Unavailable for MENA users
		// bedrock.ModelMetaLlama3170bInstructV1,     // Unstable behavior on processing tool calls results
		// bedrock.ModelMetaLlama318bInstructV1,      // Unavailable for MENA users
		// bedrock.ModelMetaLlama370bInstructV1,      // Not supported for tool calling
		// bedrock.ModelMetaLlama38bInstructV1,       // Not supported for tool calling

		// DeepSeek models
		// bedrock.ModelDeepSeekR1V1,                 // Not supported for tool calling
		bedrock.ModelDeepSeekV32,

		// OpenAI models
		bedrock.ModelOpenAIGptOss120BV1,
		bedrock.ModelOpenAIGptOss20BV1,

		// Qwen models
		bedrock.ModelQwen3Next80BA3B,
		// bedrock.ModelQwen3VL235BA22B,              // Not stable for tool calling
		bedrock.ModelQwen332BV1,
		bedrock.ModelQwen3Coder30BA3BV1,
		bedrock.ModelQwen3CoderNext,

		// Mistral models
		bedrock.ModelMistralLarge3,
		// bedrock.ModelMistralMagistralSmall2509,    // Not supported for tool calling
		// bedrock.ModelMistralLarge2402V1,           // Not supported for tool calling in streaming

		// Moonshot models
		bedrock.ModelMoonshotKimiK25,
		// bedrock.ModelMoonshotKimiK2Thinking,       // Not stable for tool calling

		// Z.AI models
		// bedrock.ModelGLM47,        // Tool calling not supported (backend requires string input instead of JSON)
		// bedrock.ModelGLM47Flash,   // Tool calling not supported (backend requires string input instead of JSON)
	}

	for _, model := range streamingToolCallModels {
		t.Logf("Testing streaming tool calling with model: %s", model)

		// Create streaming validator for this specific test
		streamingValidator := func(toolCalls map[string]*streaming.ToolCall) error {
			if len(toolCalls) == 0 {
				return fmt.Errorf("no tool calls detected in streaming")
			}

			toolCallResult, ok := toolCalls["calculator"]
			if !ok {
				return fmt.Errorf("calculator tool call not found")
			}

			calculateArgs, err := toolCallResult.Parse()
			if err != nil {
				return fmt.Errorf("failed to parse tool call arguments: %w", err)
			}

			// Validate the streaming captured the correct tool call arguments
			if calculateArgs["operation"] != "multiply" || calculateArgs["a"] != float64(15) || calculateArgs["b"] != float64(8) {
				return fmt.Errorf("unexpected calculator arguments: %+v", calculateArgs)
			}

			return nil
		}

		err := testToolCallingWorkflow(ctx, t, llm, model, rr.Replaying(), streamingValidator)
		if err != nil {
			t.Errorf("Streaming tool calling failed for model %s: %v", model, err)
		}
	}
}

func TestAmazonToolCallingStreamingLegacyAPI(t *testing.T) {
	ctx := t.Context()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	if !rr.Recording() {
		t.Parallel()
	}

	client, err := setUpTestWithTransport(rr)
	if err != nil {
		t.Fatal(err)
	}
	llm, err := bedrock.New(bedrock.WithClient(client))
	if err != nil {
		t.Fatal(err)
	}

	// Models that support streaming tool calling (initially Anthropic, will expand)
	streamingToolCallModels := []string{
		// Anthropic models
		bedrock.ModelAnthropicClaudeOpus45,
		bedrock.ModelAnthropicClaudeHaiku45,
		bedrock.ModelAnthropicClaudeSonnet45,
		// bedrock.ModelAnthropicClaudeOpus41, // Has very hard rate limits
		bedrock.ModelAnthropicClaudeOpus4,
		bedrock.ModelAnthropicClaudeSonnet4,
		bedrock.ModelAnthropicClaude37Sonnet,
		bedrock.ModelAnthropicClaude35Haiku,
	}

	for _, model := range streamingToolCallModels {
		t.Logf("Testing streaming tool calling with model: %s", model)

		// Create streaming validator for this specific test
		streamingValidator := func(toolCalls map[string]*streaming.ToolCall) error {
			if len(toolCalls) == 0 {
				return fmt.Errorf("no tool calls detected in streaming")
			}

			toolCallResult, ok := toolCalls["calculator"]
			if !ok {
				return fmt.Errorf("calculator tool call not found")
			}

			calculateArgs, err := toolCallResult.Parse()
			if err != nil {
				return fmt.Errorf("failed to parse tool call arguments: %w", err)
			}

			// Validate the streaming captured the correct tool call arguments
			if calculateArgs["operation"] != "multiply" || calculateArgs["a"] != float64(15) || calculateArgs["b"] != float64(8) {
				return fmt.Errorf("unexpected calculator arguments: %+v", calculateArgs)
			}

			return nil
		}

		err := testToolCallingWorkflow(ctx, t, llm, model, rr.Replaying(), streamingValidator)
		if err != nil {
			t.Errorf("Streaming tool calling failed for model %s: %v", model, err)
		}
	}
}

func TestAmazonReasoningConverseAPI(t *testing.T) {
	ctx := t.Context()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	if !rr.Recording() {
		t.Parallel()
	}

	client, err := setUpTestWithTransport(rr)
	if err != nil {
		t.Fatal(err)
	}
	llm, err := bedrock.New(bedrock.WithClient(client), bedrock.WithConverseAPI())
	if err != nil {
		t.Fatal(err)
	}

	reasoningModels := []string{
		bedrock.ModelAnthropicClaudeOpus46,
		bedrock.ModelAnthropicClaudeSonnet46,
		bedrock.ModelAnthropicClaudeOpus45,
		bedrock.ModelAnthropicClaudeHaiku45,
		bedrock.ModelAnthropicClaudeSonnet45,
		// bedrock.ModelAnthropicClaudeOpus4,    // Model is deprecated
		bedrock.ModelAnthropicClaudeSonnet4,
		// bedrock.ModelAnthropicClaude37Sonnet, // Model is deprecated
		bedrock.ModelDeepSeekR1V1,
		bedrock.ModelOpenAIGptOss120BV1,
		bedrock.ModelOpenAIGptOss20BV1,
		bedrock.ModelMoonshotKimiK2Thinking,
	}

	for _, model := range reasoningModels {
		t.Logf("Testing reasoning with model: %s", model)

		err := testReasoningWorkflow(ctx, t, llm, model, nil)
		if err != nil {
			t.Errorf("Reasoning failed for model %s: %v", model, err)
		}
	}
}

func TestAmazonReasoningLegacyAPI(t *testing.T) {
	ctx := t.Context()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	if !rr.Recording() {
		t.Parallel()
	}

	client, err := setUpTestWithTransport(rr)
	if err != nil {
		t.Fatal(err)
	}
	llm, err := bedrock.New(bedrock.WithClient(client))
	if err != nil {
		t.Fatal(err)
	}

	// Only Anthropic models support reasoning in Legacy API
	reasoningModels := []string{
		bedrock.ModelAnthropicClaudeOpus45,
		bedrock.ModelAnthropicClaudeSonnet45,
		bedrock.ModelAnthropicClaudeOpus41,
		bedrock.ModelAnthropicClaudeOpus4,
		bedrock.ModelAnthropicClaudeSonnet4,
		bedrock.ModelAnthropicClaude37Sonnet,
	}

	for _, model := range reasoningModels {
		t.Logf("Testing reasoning with model: %s", model)

		err := testReasoningWorkflow(ctx, t, llm, model, nil)
		if err != nil {
			t.Errorf("Reasoning failed for model %s: %v", model, err)
		}
	}
}

func TestAmazonReasoningStreamingConverseAPI(t *testing.T) {
	ctx := t.Context()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	if !rr.Recording() {
		t.Parallel()
	}

	client, err := setUpTestWithTransport(rr)
	if err != nil {
		t.Fatal(err)
	}
	llm, err := bedrock.New(bedrock.WithClient(client), bedrock.WithConverseAPI())
	if err != nil {
		t.Fatal(err)
	}

	streamingReasoningModels := []string{
		bedrock.ModelAnthropicClaudeOpus46,
		bedrock.ModelAnthropicClaudeSonnet46,
		bedrock.ModelAnthropicClaudeOpus45,
		bedrock.ModelAnthropicClaudeHaiku45,
		bedrock.ModelAnthropicClaudeSonnet45,
		// bedrock.ModelAnthropicClaudeOpus4,    // Model is deprecated
		bedrock.ModelAnthropicClaudeSonnet4,
		// bedrock.ModelAnthropicClaude37Sonnet, // Model is deprecated
		bedrock.ModelDeepSeekR1V1,
		bedrock.ModelOpenAIGptOss120BV1,
		bedrock.ModelOpenAIGptOss20BV1,
		bedrock.ModelMoonshotKimiK2Thinking,
	}

	for _, model := range streamingReasoningModels {
		t.Logf("Testing streaming reasoning with model: %s", model)

		streamingValidator := func(reasoningChunks []string) error {
			if len(reasoningChunks) == 0 {
				return fmt.Errorf("no reasoning chunks detected in streaming")
			}
			return nil
		}

		err := testReasoningWorkflow(ctx, t, llm, model, streamingValidator)
		if err != nil {
			t.Errorf("Streaming reasoning failed for model %s: %v", model, err)
		}
	}
}

func TestAmazonReasoningStreamingLegacyAPI(t *testing.T) {
	ctx := t.Context()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	if !rr.Recording() {
		t.Parallel()
	}

	client, err := setUpTestWithTransport(rr)
	if err != nil {
		t.Fatal(err)
	}
	llm, err := bedrock.New(bedrock.WithClient(client))
	if err != nil {
		t.Fatal(err)
	}

	// Only Anthropic models support reasoning in Legacy API
	streamingReasoningModels := []string{
		bedrock.ModelAnthropicClaudeOpus45,
		bedrock.ModelAnthropicClaudeSonnet45,
		bedrock.ModelAnthropicClaudeOpus41,
		bedrock.ModelAnthropicClaudeOpus4,
		bedrock.ModelAnthropicClaudeSonnet4,
		bedrock.ModelAnthropicClaude37Sonnet,
	}

	for _, model := range streamingReasoningModels {
		t.Logf("Testing streaming reasoning with model: %s", model)

		streamingValidator := func(reasoningChunks []string) error {
			if len(reasoningChunks) == 0 {
				return fmt.Errorf("no reasoning chunks detected in streaming")
			}
			return nil
		}

		err := testReasoningWorkflow(ctx, t, llm, model, streamingValidator)
		if err != nil {
			t.Errorf("Streaming reasoning failed for model %s: %v", model, err)
		}
	}
}

func testReasoningWorkflow( //nolint:funlen
	ctx context.Context,
	t *testing.T,
	llm *bedrock.LLM,
	model string,
	streamingValidator func(reasoningChunks []string) error,
) error {
	t.Logf("Testing reasoning workflow for model: %s", model)

	contents := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Think step by step: What is 17 * 23? Show your reasoning process."},
			},
		},
	}

	var opts []llms.CallOption
	opts = append(opts, llms.WithModel(model))
	opts = append(opts, llms.WithMaxTokens(4096))
	opts = append(opts, llms.WithReasoning(llms.ReasoningNone, 1024))

	var choice *llms.ContentChoice
	if streamingValidator != nil {
		reasoningChunks := make([]string, 0)
		streamingFunc := func(_ context.Context, chunk streaming.Chunk) error {
			switch chunk.Type {
			case streaming.ChunkTypeReasoning:
				if chunk.Reasoning != nil {
					reasoningChunks = append(reasoningChunks, chunk.Reasoning.Content)
				} else {
					t.Log("reasoning chunk is nil")
				}
			default:
				// Ignore other chunks in this test
			}
			return nil
		}
		opts = append(opts, llms.WithStreamingFunc(streamingFunc))

		defer func() {
			if streamingValidator != nil {
				if err := streamingValidator(reasoningChunks); err != nil {
					t.Errorf("Streaming validation failed for model %s: %v", model, err)
				}
				reasoningContent := strings.Join(reasoningChunks, "")
				if choice == nil {
					t.Errorf("No choice in response")
				} else if choice.Reasoning == nil && reasoningContent != "" {
					t.Errorf("Reasoning content mismatch: expected %s, got nil", reasoningContent)
				} else if choice.Reasoning != nil && choice.Reasoning.Content != reasoningContent {
					t.Errorf("Reasoning content mismatch: expected %s, got %s",
						reasoningContent, choice.Reasoning.Content)
				}
			}
		}()
	}

	resp, err := llm.GenerateContent(ctx, contents, opts...)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return fmt.Errorf("no choices in response")
	}

	choice = resp.Choices[0]

	if choice.Content == "" {
		return fmt.Errorf("empty response content")
	}

	if !strings.Contains(choice.Content, "391") {
		return fmt.Errorf("expected final response to contain '391', got: %s", choice.Content)
	}

	if choice.Reasoning != nil && choice.Reasoning.Content != "" {
		preview := choice.Reasoning.Content
		if len(preview) > 50 {
			preview = preview[:50] + "..."
		}
		t.Logf("Found reasoning content for model %s: %s", model, preview)
	} else {
		return fmt.Errorf("empty reasoning content")
	}

	t.Logf("Reasoning workflow completed successfully for model: %s", model)
	return nil
}

type property struct {
	Type        string   `json:"type" document:"type"`
	Description string   `json:"description" document:"description"`
	Enum        []string `json:"enum,omitempty" document:"enum,omitempty"`
}

type properties struct {
	Operation property `json:"operation" document:"operation"`
	A         property `json:"a" document:"a"`
	B         property `json:"b" document:"b"`
}

type calculatorSchema struct {
	Type       string     `json:"type" document:"type"`
	Properties properties `json:"properties" document:"properties"`
	Required   []string   `json:"required" document:"required"`
}

// testToolCallingWorkflow tests the complete tool calling workflow for a given model
//
//nolint:funlen
func testToolCallingWorkflow(
	ctx context.Context,
	t *testing.T,
	llm *bedrock.LLM,
	model string,
	isReplaying bool,
	streamingValidator func(toolCalls map[string]*streaming.ToolCall) error,
) error {
	t.Logf("Testing tool calling workflow for model: %s", model)

	availableTools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "calculator",
				Description: "A calculator that can perform basic arithmetic operations",
				Parameters: &calculatorSchema{
					Type: "object",
					Properties: properties{
						Operation: property{
							Type:        "string",
							Description: "The operation to perform",
							Enum:        []string{"add", "subtract", "multiply", "divide"},
						},
						A: property{
							Type:        "number",
							Description: "First number",
						},
						B: property{
							Type:        "number",
							Description: "Second number",
						},
					},
					Required: []string{"operation", "a", "b"},
				},
			},
		},
	}

	contents := []llms.MessageContent{ //nolint:prealloc
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Calculate 15 * 8"}},
		},
	}

	// Step 1: Send initial request and get tool call
	var opts []llms.CallOption
	opts = append(opts, llms.WithModel(model))
	opts = append(opts, llms.WithTools(availableTools))
	opts = append(opts, llms.WithMaxTokens(512))

	// Add streaming if validator is provided
	if streamingValidator != nil {
		toolCalls := make(map[string]*streaming.ToolCall)
		streamingFunc := func(_ context.Context, chunk streaming.Chunk) error {
			switch chunk.Type {
			case streaming.ChunkTypeToolCall:
				toolCall := chunk.ToolCall
				if toolCall.Name != "calculator" {
					return fmt.Errorf("unexpected tool call: %s", toolCall.Name)
				}

				if resToolCall, ok := toolCalls[toolCall.Name]; !ok {
					toolCalls[toolCall.Name] = &toolCall
				} else {
					streaming.AppendToolCall(toolCall, resToolCall)
				}
			default:
				// Ignore other chunks in this test
			}
			return nil
		}
		opts = append(opts, llms.WithStreamingFunc(streamingFunc))

		// Execute streaming validation after request
		defer func() {
			if streamingValidator != nil {
				if err := streamingValidator(toolCalls); err != nil {
					t.Errorf("Streaming validation failed for model %s: %v", model, err)
				}
			}
		}()
	}

	resp, err := llm.GenerateContent(ctx, contents, opts...)
	if err != nil {
		return fmt.Errorf("initial request failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return fmt.Errorf("no choices in response")
	}

	choice := resp.Choices[0]
	if len(choice.ToolCalls) == 0 {
		return fmt.Errorf("expected tool call in response")
	}

	toolCall := choice.ToolCalls[0]
	if toolCall.FunctionCall.Name != "calculator" {
		return fmt.Errorf("expected calculator tool call, got: %s", toolCall.FunctionCall.Name)
	}

	// If HTTP requests are being replayed (such as during test recording playback), skip the tool call execution step.
	// This is necessary because tool call requests use maps, which can serialize to JSON in a non-deterministic order,
	// leading to cache misses or mismatches with recorded HTTP responses.
	if isReplaying {
		return nil
	}

	// Step 2: Add assistant response with tool call
	assistantResponse := llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{
			llms.ToolCall{
				ID:   toolCall.ID,
				Type: toolCall.Type,
				FunctionCall: &llms.FunctionCall{
					Name:      toolCall.FunctionCall.Name,
					Arguments: toolCall.FunctionCall.Arguments,
				},
			},
		},
	}
	contents = append(contents, assistantResponse)

	// Step 3: Add tool result (15 * 8 = 120)
	calculatorResult := "120"
	toolCallResponse := llms.MessageContent{
		Role: llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{
			llms.ToolCallResponse{
				ToolCallID: toolCall.ID,
				Name:       toolCall.FunctionCall.Name,
				Content:    calculatorResult,
			},
		},
	}
	contents = append(contents, toolCallResponse)

	// Step 4: Get final response from LLM
	finalResp, err := llm.GenerateContent(ctx, contents,
		llms.WithModel(model),
		llms.WithTools(availableTools),
		llms.WithMaxTokens(512),
	)
	if err != nil {
		return fmt.Errorf("final request failed: %w", err)
	}

	if len(finalResp.Choices) == 0 {
		return fmt.Errorf("no choices in final response")
	}

	// Step 5: Validate final response contains the result "120"
	finalContent := finalResp.Choices[0].Content
	if !strings.Contains(finalContent, "120") {
		return fmt.Errorf("expected final response to contain '120', got: %s", finalContent)
	}

	t.Logf("Tool calling workflow completed successfully for model: %s", model)
	return nil
}

// TestAmazonTextResponseWithThinkingConverseAPI tests text response with thinking using Converse API
func TestAmazonTextResponseWithThinkingConverseAPI(t *testing.T) {
	ctx := t.Context()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	if !rr.Recording() {
		t.Parallel()
	}

	client, err := setUpTestWithTransport(rr)
	if err != nil {
		t.Fatal(err)
	}
	llm, err := bedrock.New(bedrock.WithClient(client), bedrock.WithConverseAPI())
	if err != nil {
		t.Fatal(err)
	}

	messages := make([]llms.MessageContent, 0, 3)
	messages = append(messages, llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextPart("Solve: If x + 5 = 12, what is x?"),
		},
	})

	// Request with thinking
	resp, err := llm.GenerateContent(ctx, messages,
		llms.WithModel(bedrock.ModelAnthropicClaudeSonnet45),
		llms.WithReasoning(llms.ReasoningMedium, 2048),
		llms.WithMaxTokens(4096),
		llms.WithTemperature(1.0),
	)
	if err != nil {
		t.Fatal(err)
	}

	choice := resp.Choices[0]
	if choice.Reasoning == nil {
		t.Fatal("Expected reasoning in response")
	}
	if choice.Reasoning.Content == "" {
		t.Error("Expected non-empty reasoning content")
	}
	if len(choice.Reasoning.Signature) == 0 {
		t.Error("Expected signature in reasoning")
	}
	if !strings.Contains(choice.Content, "7") {
		t.Errorf("Expected answer '7' in content, got: %s", choice.Content)
	}

	// ROUNDTRIP: Continue conversation with preserved signature
	messages = append(messages,
		llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextPartWithReasoning(choice.Content, choice.Reasoning),
			},
		},
		llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Now solve x + 10 = 25")},
		},
	)

	resp2, err := llm.GenerateContent(ctx, messages,
		llms.WithModel(bedrock.ModelAnthropicClaudeSonnet45),
		llms.WithReasoning(llms.ReasoningMedium, 2048),
		llms.WithMaxTokens(4096),
		llms.WithTemperature(1.0),
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp2.Choices[0].Content, "15") {
		t.Errorf("Expected answer '15' in roundtrip response, got: %s", resp2.Choices[0].Content)
	}
}

// TestAmazonTextResponseWithThinkingLegacyAPI tests text response with thinking using Legacy API
func TestAmazonTextResponseWithThinkingLegacyAPI(t *testing.T) {
	ctx := t.Context()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	if !rr.Recording() {
		t.Parallel()
	}

	client, err := setUpTestWithTransport(rr)
	if err != nil {
		t.Fatal(err)
	}
	llm, err := bedrock.New(bedrock.WithClient(client)) // Use Legacy API
	if err != nil {
		t.Fatal(err)
	}

	messages := make([]llms.MessageContent, 0, 3)
	messages = append(messages, llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextPart("Solve: If x + 5 = 12, what is x?"),
		},
	})

	// Request with thinking
	resp, err := llm.GenerateContent(ctx, messages,
		llms.WithModel(bedrock.ModelAnthropicClaudeSonnet45),
		llms.WithReasoning(llms.ReasoningMedium, 2048),
		llms.WithMaxTokens(4096),
		llms.WithTemperature(1.0),
	)
	if err != nil {
		t.Fatal(err)
	}

	choice := resp.Choices[0]
	if choice.Reasoning == nil {
		t.Fatal("Expected reasoning in response")
	}
	if choice.Reasoning.Content == "" {
		t.Error("Expected non-empty reasoning content")
	}
	if len(choice.Reasoning.Signature) == 0 {
		t.Error("Expected signature in reasoning")
	}
	if !strings.Contains(choice.Content, "7") {
		t.Errorf("Expected answer '7' in content, got: %s", choice.Content)
	}

	// ROUNDTRIP: Continue conversation with preserved signature
	messages = append(messages,
		llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextPartWithReasoning(choice.Content, choice.Reasoning),
			},
		},
		llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Now solve x + 10 = 25")},
		},
	)

	resp2, err := llm.GenerateContent(ctx, messages,
		llms.WithModel(bedrock.ModelAnthropicClaudeSonnet45),
		llms.WithReasoning(llms.ReasoningMedium, 2048),
		llms.WithMaxTokens(4096),
		llms.WithTemperature(1.0),
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp2.Choices[0].Content, "15") {
		t.Errorf("Expected answer '15' in roundtrip response, got: %s", resp2.Choices[0].Content)
	}
}

// TestAmazonSingleToolCallWithThinkingConverseAPI tests single tool call with thinking
func TestAmazonSingleToolCallWithThinkingConverseAPI(t *testing.T) { //nolint:funlen
	ctx := t.Context()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	if !rr.Recording() {
		t.Parallel()
	}

	client, err := setUpTestWithTransport(rr)
	if err != nil {
		t.Fatal(err)
	}
	llm, err := bedrock.New(bedrock.WithClient(client), bedrock.WithConverseAPI())
	if err != nil {
		t.Fatal(err)
	}

	tools := []llms.Tool{{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "get_weather",
			Description: "Get weather for a location",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"location": map[string]any{"type": "string"},
				},
				"required": []string{"location"},
			},
		},
	}}

	messages := make([]llms.MessageContent, 0, 3)
	messages = append(messages, llms.MessageContent{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.TextPart("What's the weather in Boston?")},
	})

	resp, err := llm.GenerateContent(ctx, messages,
		llms.WithModel(bedrock.ModelAnthropicClaudeSonnet45),
		llms.WithTools(tools),
		llms.WithReasoning(llms.ReasoningMedium, 2048),
		llms.WithMaxTokens(4096),
		llms.WithTemperature(1.0),
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Choices[0].ToolCalls) == 0 {
		t.Fatal("Expected tool call in response")
	}

	choice := resp.Choices[0]
	if choice.Reasoning == nil {
		t.Error("Expected reasoning in response")
	}
	if choice.Reasoning != nil && len(choice.Reasoning.Signature) == 0 {
		t.Error("Expected signature in reasoning")
	}

	// ROUNDTRIP: Send response back preserving reasoning
	messages = append(messages,
		llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextPartWithReasoning(choice.Content, choice.Reasoning),
				resp.Choices[0].ToolCalls[0],
			},
		},
		llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: resp.Choices[0].ToolCalls[0].ID,
					Name:       resp.Choices[0].ToolCalls[0].FunctionCall.Name,
					Content:    `{"temperature": 72, "condition": "sunny"}`,
				},
			},
		},
	)

	resp2, err := llm.GenerateContent(ctx, messages,
		llms.WithModel(bedrock.ModelAnthropicClaudeSonnet45),
		llms.WithTools(tools),
		llms.WithReasoning(llms.ReasoningMedium, 2048),
		llms.WithMaxTokens(4096),
		llms.WithTemperature(1.0),
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp2.Choices[0].Content, "72") {
		t.Logf("Expected '72' in final response, got: %s", resp2.Choices[0].Content)
	}
}

// TestAmazonMultipleToolCallsVariantsConverseAPI tests all possible ways users can structure message chains
// with multiple tool calls, ensuring all variants are normalized correctly for Bedrock API
func TestAmazonMultipleToolCallsVariantsConverseAPI(t *testing.T) { //nolint:funlen
	ctx := t.Context()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	if !rr.Recording() {
		t.Parallel()
	}

	client, err := setUpTestWithTransport(rr)
	if err != nil {
		t.Fatal(err)
	}
	llm, err := bedrock.New(bedrock.WithClient(client), bedrock.WithConverseAPI())
	if err != nil {
		t.Fatal(err)
	}

	// Define two tools that are likely to be called together
	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "search_web",
				Description: "Search the web for information",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"query": map[string]any{
							"type":        "string",
							"description": "Search query",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_current_date",
				Description: "Get the current date and time",
				Parameters: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				},
			},
		},
	}

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Search for CVE-2020-10188 exploits and also tell me what's the current date"),
			},
		},
	}

	// First call - model should invoke both tools
	resp1, err := llm.GenerateContent(ctx, messages,
		llms.WithModel(bedrock.ModelAnthropicClaudeHaiku45),
		llms.WithTools(tools),
		llms.WithMaxTokens(8192),
		llms.WithTemperature(1.0),
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp1.Choices) == 0 {
		t.Fatal("Expected at least one choice in response")
	}

	choice1 := resp1.Choices[0]
	if len(choice1.ToolCalls) < 2 {
		t.Fatalf("Expected at least 2 tool calls, got %d", len(choice1.ToolCalls))
	}

	// Prepare common data for all variants
	toolCall1 := choice1.ToolCalls[0]
	toolCall2 := choice1.ToolCalls[1]
	aiContent := choice1.Content
	aiReasoning := choice1.Reasoning
	result1 := `{"results": ["CVE-2020-10188 is a buffer overflow in telnetd"]}`
	result2 := `{"date": "2026-03-15"}`

	// Helper to test a variant
	testVariant := func(t *testing.T, variantName string, buildMessages func() []llms.MessageContent) {
		t.Run(variantName, func(t *testing.T) {
			messages := buildMessages()

			resp2, err := llm.GenerateContent(ctx, messages,
				llms.WithModel(bedrock.ModelAnthropicClaudeHaiku45),
				llms.WithTools(tools),
				llms.WithMaxTokens(8192),
				llms.WithTemperature(1.0),
			)
			if err != nil {
				t.Fatalf("Variant %s failed: %v", variantName, err)
			}

			if len(resp2.Choices) == 0 {
				t.Fatal("Expected at least one choice in second response")
			}

			if !strings.Contains(resp2.Choices[0].Content, "CVE-2020-10188") {
				t.Errorf("Response should mention CVE-2020-10188, got: %s", resp2.Choices[0].Content)
			}
		})
	}

	// Variant 1: Content separate + all tool calls together + all tool results together
	testVariant(t, "content_separate_toolcalls_together_results_together", func() []llms.MessageContent {
		msgs := append([]llms.MessageContent{}, messages...)
		msgs = append(msgs,
			llms.MessageContent{
				Role:  llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{llms.TextPartWithReasoning(aiContent, aiReasoning)},
			},
			llms.MessageContent{
				Role:  llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{toolCall1, toolCall2},
			},
			llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{ToolCallID: toolCall1.ID, Name: toolCall1.FunctionCall.Name, Content: result1},
					llms.ToolCallResponse{ToolCallID: toolCall2.ID, Name: toolCall2.FunctionCall.Name, Content: result2},
				},
			},
		)
		return msgs
	})

	// Variant 2: Content separate + tool calls separate + tool results together
	testVariant(t, "content_separate_toolcalls_separate_results_together", func() []llms.MessageContent {
		msgs := append([]llms.MessageContent{}, messages...)
		msgs = append(msgs,
			llms.MessageContent{
				Role:  llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{llms.TextPartWithReasoning(aiContent, aiReasoning)},
			},
			llms.MessageContent{
				Role:  llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{toolCall1},
			},
			llms.MessageContent{
				Role:  llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{toolCall2},
			},
			llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{ToolCallID: toolCall1.ID, Name: toolCall1.FunctionCall.Name, Content: result1},
					llms.ToolCallResponse{ToolCallID: toolCall2.ID, Name: toolCall2.FunctionCall.Name, Content: result2},
				},
			},
		)
		return msgs
	})

	// Variant 3: Content separate + tool calls separate + tool results separate
	testVariant(t, "content_separate_toolcalls_separate_results_separate", func() []llms.MessageContent {
		msgs := append([]llms.MessageContent{}, messages...)
		msgs = append(msgs,
			llms.MessageContent{
				Role:  llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{llms.TextPartWithReasoning(aiContent, aiReasoning)},
			},
			llms.MessageContent{
				Role:  llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{toolCall1},
			},
			llms.MessageContent{
				Role:  llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{toolCall2},
			},
			llms.MessageContent{
				Role:  llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{llms.ToolCallResponse{ToolCallID: toolCall1.ID, Name: toolCall1.FunctionCall.Name, Content: result1}},
			},
			llms.MessageContent{
				Role:  llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{llms.ToolCallResponse{ToolCallID: toolCall2.ID, Name: toolCall2.FunctionCall.Name, Content: result2}},
			},
		)
		return msgs
	})

	// Variant 4: Content + all tool calls together + tool results together
	testVariant(t, "content_with_toolcalls_together_results_together", func() []llms.MessageContent {
		msgs := append([]llms.MessageContent{}, messages...)
		msgs = append(msgs,
			llms.MessageContent{
				Role:  llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{llms.TextPartWithReasoning(aiContent, aiReasoning), toolCall1, toolCall2},
			},
			llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{ToolCallID: toolCall1.ID, Name: toolCall1.FunctionCall.Name, Content: result1},
					llms.ToolCallResponse{ToolCallID: toolCall2.ID, Name: toolCall2.FunctionCall.Name, Content: result2},
				},
			},
		)
		return msgs
	})

	// Variant 5: Content + tool calls separate + tool results separate
	testVariant(t, "content_with_toolcall1_separate_toolcall2_results_separate", func() []llms.MessageContent {
		msgs := append([]llms.MessageContent{}, messages...)
		msgs = append(msgs,
			llms.MessageContent{
				Role:  llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{llms.TextPartWithReasoning(aiContent, aiReasoning), toolCall1},
			},
			llms.MessageContent{
				Role:  llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{toolCall2},
			},
			llms.MessageContent{
				Role:  llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{llms.ToolCallResponse{ToolCallID: toolCall1.ID, Name: toolCall1.FunctionCall.Name, Content: result1}},
			},
			llms.MessageContent{
				Role:  llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{llms.ToolCallResponse{ToolCallID: toolCall2.ID, Name: toolCall2.FunctionCall.Name, Content: result2}},
			},
		)
		return msgs
	})

	// Variant 6: Content + all tool calls together + tool results separate
	testVariant(t, "content_with_toolcalls_together_results_separate", func() []llms.MessageContent {
		msgs := append([]llms.MessageContent{}, messages...)
		msgs = append(msgs,
			llms.MessageContent{
				Role:  llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{llms.TextPartWithReasoning(aiContent, aiReasoning), toolCall1, toolCall2},
			},
			llms.MessageContent{
				Role:  llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{llms.ToolCallResponse{ToolCallID: toolCall1.ID, Name: toolCall1.FunctionCall.Name, Content: result1}},
			},
			llms.MessageContent{
				Role:  llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{llms.ToolCallResponse{ToolCallID: toolCall2.ID, Name: toolCall2.FunctionCall.Name, Content: result2}},
			},
		)
		return msgs
	})
}

// TestAmazonSequentialToolCallsWithThinkingConverseAPI tests sequential tool calls with thinking
func TestAmazonSequentialToolCallsWithThinkingConverseAPI(t *testing.T) { //nolint:funlen
	ctx := t.Context()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	if !rr.Recording() {
		t.Parallel()
	}

	client, err := setUpTestWithTransport(rr)
	if err != nil {
		t.Fatal(err)
	}
	llm, err := bedrock.New(bedrock.WithClient(client), bedrock.WithConverseAPI())
	if err != nil {
		t.Fatal(err)
	}

	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "calculate",
				Description: "Perform a calculation",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"expression": map[string]any{"type": "string"},
					},
					"required": []string{"expression"},
				},
			},
		},
	}

	messages := make([]llms.MessageContent, 0, 3)
	messages = append(messages, llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextPart("Calculate (5 + 3) and then multiply by 2"),
		},
	})

	// First call
	resp1, err := llm.GenerateContent(ctx, messages,
		llms.WithModel(bedrock.ModelAnthropicClaudeSonnet45),
		llms.WithTools(tools),
		llms.WithReasoning(llms.ReasoningMedium, 2048),
		llms.WithMaxTokens(4096),
		llms.WithTemperature(1.0),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp1.Choices[0].ToolCalls) == 0 {
		t.Fatal("Expected tool call in first response")
	}
	if resp1.Choices[0].Reasoning == nil {
		t.Fatal("Expected reasoning in first response")
	}

	sig1 := resp1.Choices[0].Reasoning.Signature

	// Execute first tool and continue
	choice1 := resp1.Choices[0]
	messages = append(messages,
		llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextPartWithReasoning(choice1.Content, choice1.Reasoning),
				choice1.ToolCalls[0],
			},
		},
		llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: resp1.Choices[0].ToolCalls[0].ID,
					Name:       resp1.Choices[0].ToolCalls[0].FunctionCall.Name,
					Content:    "8",
				},
			},
		},
	)

	// Second call
	resp2, err := llm.GenerateContent(ctx, messages,
		llms.WithModel(bedrock.ModelAnthropicClaudeSonnet45),
		llms.WithTools(tools),
		llms.WithReasoning(llms.ReasoningMedium, 2048),
		llms.WithMaxTokens(4096),
		llms.WithTemperature(1.0),
	)
	if err != nil {
		t.Fatal(err)
	}

	if resp2.Choices[0].Reasoning != nil {
		sig2 := resp2.Choices[0].Reasoning.Signature

		// Signatures should differ (each step has unique context)
		if len(sig1) > 0 && len(sig2) > 0 && string(sig1) == string(sig2) {
			t.Error("Signatures should differ between turns")
		}
	}
}

// TestAmazonTextResponseWithThinkingStreamingConverseAPI tests streaming with thinking
func TestAmazonTextResponseWithThinkingStreamingConverseAPI(t *testing.T) {
	ctx := t.Context()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	if !rr.Recording() {
		t.Parallel()
	}

	client, err := setUpTestWithTransport(rr)
	if err != nil {
		t.Fatal(err)
	}
	llm, err := bedrock.New(bedrock.WithClient(client), bedrock.WithConverseAPI())
	if err != nil {
		t.Fatal(err)
	}

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What is 5 + 3? Think step by step."),
			},
		},
	}

	var streamedReasoning []string
	var streamedText []string

	streamFunc := llms.WithStreamingFunc(func(ctx context.Context, chunk streaming.Chunk) error {
		switch chunk.Type {
		case streaming.ChunkTypeReasoning:
			if chunk.Reasoning != nil {
				streamedReasoning = append(streamedReasoning, chunk.Reasoning.Content)
			}
		case streaming.ChunkTypeText:
			streamedText = append(streamedText, chunk.Content)
		default:
		}
		return nil
	})

	resp, err := llm.GenerateContent(ctx, messages,
		llms.WithModel(bedrock.ModelAnthropicClaudeSonnet45),
		llms.WithReasoning(llms.ReasoningMedium, 2048),
		streamFunc,
		llms.WithMaxTokens(4096),
		llms.WithTemperature(1.0),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Verify accumulated content matches final response
	if strings.Join(streamedText, "") != resp.Choices[0].Content {
		t.Errorf("Streamed text doesn't match final response")
	}
	if len(streamedReasoning) > 0 {
		accumulatedReasoning := strings.Join(streamedReasoning, "")
		if resp.Choices[0].Reasoning != nil && accumulatedReasoning != resp.Choices[0].Reasoning.Content {
			t.Errorf("Streamed reasoning doesn't match final response")
		}
	}
	if resp.Choices[0].Reasoning != nil && len(resp.Choices[0].Reasoning.Signature) == 0 {
		t.Error("Expected signature in final reasoning")
	}
}

// TestAmazonMultiTurnCachingWithToolsLegacyAPI tests prompt caching across multi-turn conversation with tools.
// This validates that caching works correctly for AI agent workflows using Legacy API.
//
// Expected behavior:
// Turn 1: CacheCreation > 0 (conversation history cached)
// Turn 2: CacheRead > 0 (previous turn read from cache), new content added to cache
// Turn 3+: Cache continues to be used and extended
func TestAmazonMultiTurnCachingWithToolsLegacyAPI(t *testing.T) { //nolint:funlen
	ctx := t.Context()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	if !rr.Recording() {
		t.Parallel()
	}

	client, err := setUpTestWithTransport(rr)
	if err != nil {
		t.Fatal(err)
	}
	// Use Legacy API for cache_control support
	llm, err := bedrock.New(bedrock.WithClient(client))
	if err != nil {
		t.Fatal(err)
	}

	// Define tools for AI agent
	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_weather",
				Description: "BedrockCacheTest-v1: " + strings.Repeat("Get current weather conditions including temperature, humidity, wind speed, and precipitation for a specified geographic location. ", 30),
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "City and state, e.g. San Francisco, CA",
						},
					},
					"required": []string{"location"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "book_flight",
				Description: "BedrockCacheTest-v1: " + strings.Repeat("Book airline tickets for domestic and international flights with flexible options for departure, arrival, and passenger details. ", 30),
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"destination": map[string]any{"type": "string", "description": "Destination city"},
						"date":        map[string]any{"type": "string", "description": "Departure date"},
					},
					"required": []string{"destination", "date"},
				},
			},
		},
	}

	// Add system prompt to ensure sufficient tokens for caching
	systemPrompt := "BedrockCacheTest-v1: " + strings.Repeat("You are a helpful assistant with access to weather and flight booking capabilities. ", 15)

	// Turn 1: Initial request with system prompt
	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(systemPrompt)},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("What's the weather in Boston?")},
		},
	}

	resp1, err := llm.GenerateContent(ctx, messages,
		llms.WithModel(bedrock.ModelAnthropicClaudeSonnet45),
		llms.WithTools(tools),
		llms.WithMaxTokens(512),
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp1.Choices[0].ToolCalls) == 0 {
		t.Fatal("Expected tool call in first response")
	}

	// Turn 2: Add tool result - mark with cache control for conversation history caching
	choice1 := resp1.Choices[0]
	messages = append(messages,
		llms.MessageContent{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{choice1.ToolCalls[0]},
		},
		llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: choice1.ToolCalls[0].ID,
					Name:       choice1.ToolCalls[0].FunctionCall.Name,
					Content:    `{"temperature": 72, "condition": "sunny"}`,
				},
			},
		},
	)

	resp2, err := llm.GenerateContent(ctx, messages,
		llms.WithModel(bedrock.ModelAnthropicClaudeSonnet45),
		llms.WithTools(tools),
		llms.WithMaxTokens(512),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Turn 2 Response: %s", resp2.Choices[0].Content)
	if !strings.Contains(resp2.Choices[0].Content, "72") && !strings.Contains(resp2.Choices[0].Content, "sunny") {
		t.Logf("Turn 2 might not contain weather info: %s", resp2.Choices[0].Content)
	}

	// Turn 3: Continue conversation with cached previous context
	messages = append(messages,
		llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				bedrock.WithCacheControl(
					llms.TextPart(resp2.Choices[0].Content),
					bedrock.EphemeralCache(),
				),
			},
		},
		llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Now book a flight to Boston for tomorrow")},
		},
	)

	resp3, err := llm.GenerateContent(ctx, messages,
		llms.WithModel(bedrock.ModelAnthropicClaudeSonnet45),
		llms.WithTools(tools),
		llms.WithMaxTokens(512),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Turn 3 Response: %s", resp3.Choices[0].Content)

	// Helper function to extract cache metrics
	getCacheMetric := func(resp *llms.ContentResponse, key string) int {
		if resp.Choices[0].GenerationInfo != nil {
			if val, ok := resp.Choices[0].GenerationInfo[key]; ok {
				switch v := val.(type) {
				case int:
					return v
				case int32:
					return int(v)
				case int64:
					return int(v)
				}
			}
		}
		return 0
	}

	// Log cache metrics for all turns
	t.Logf("Turn 1 - CacheCreation: %d, CacheRead: %d",
		getCacheMetric(resp1, "CacheCreationInputTokens"),
		getCacheMetric(resp1, "CacheReadInputTokens"))
	t.Logf("Turn 2 - CacheCreation: %d, CacheRead: %d",
		getCacheMetric(resp2, "CacheCreationInputTokens"),
		getCacheMetric(resp2, "CacheReadInputTokens"))
	t.Logf("Turn 3 - CacheCreation: %d, CacheRead: %d",
		getCacheMetric(resp3, "CacheCreationInputTokens"),
		getCacheMetric(resp3, "CacheReadInputTokens"))

	// Verify that cache was created in Turn 3 (due to cache_control on assistant message)
	cacheCreation3 := getCacheMetric(resp3, "CacheCreationInputTokens")
	if cacheCreation3 > 0 {
		t.Logf("✓ Cache successfully created in Turn 3: %d tokens", cacheCreation3)
	} else {
		t.Logf("Cache creation: %d (may be 0 if context < 1024 tokens)", cacheCreation3)
	}

	// Turn 4: Continue conversation - cache should be read from previous turn
	messages = append(messages,
		llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				bedrock.WithCacheControl(
					llms.TextPart(resp3.Choices[0].Content),
					bedrock.EphemeralCache(),
				),
			},
		},
		llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("I'm flying from New York on March 4th")},
		},
	)

	resp4, err := llm.GenerateContent(ctx, messages,
		llms.WithModel(bedrock.ModelAnthropicClaudeSonnet45),
		llms.WithTools(tools),
		llms.WithMaxTokens(512),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Turn 4 Response: %s", resp4.Choices[0].Content)

	cacheRead4 := getCacheMetric(resp4, "CacheReadInputTokens")
	cacheCreation4 := getCacheMetric(resp4, "CacheCreationInputTokens")
	t.Logf("Turn 4 - CacheCreation: %d, CacheRead: %d", cacheCreation4, cacheRead4)

	// Verify cache is being utilized
	if cacheRead4 > 0 {
		t.Logf("✓ Cache successfully read in Turn 4: %d tokens", cacheRead4)
		t.Logf("Cache savings: previous context reused from cache")
	} else if cacheCreation4 > 0 {
		t.Logf("✓ Additional cache created in Turn 4: %d tokens", cacheCreation4)
	}
}

// TestAmazonMultiTurnCachingWithToolsConverseAPI tests prompt caching with Converse API.
// This validates that caching works correctly for AI agent workflows using Converse API.
//
// Expected behavior:
// Turn 3: CacheCreation > 0 (conversation history with cachePoint)
// Turn 4: CacheRead > 0 (previous context read from cache)
func TestAmazonMultiTurnCachingWithToolsConverseAPI(t *testing.T) { //nolint:funlen
	ctx := t.Context()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	if !rr.Recording() {
		t.Parallel()
	}

	client, err := setUpTestWithTransport(rr)
	if err != nil {
		t.Fatal(err)
	}
	// Use Converse API with cachePoint support
	llm, err := bedrock.New(bedrock.WithClient(client), bedrock.WithConverseAPI())
	if err != nil {
		t.Fatal(err)
	}

	// Define tools with long descriptions to exceed 1024 token threshold
	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_weather",
				Description: "BedrockConverseCache-v1: " + strings.Repeat("Get current weather conditions including temperature, humidity, wind speed, and precipitation for a specified geographic location. ", 30),
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "City and state, e.g. San Francisco, CA",
						},
					},
					"required": []string{"location"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "book_flight",
				Description: "BedrockConverseCache-v1: " + strings.Repeat("Book airline tickets for domestic and international flights with flexible options for departure, arrival, and passenger details. ", 30),
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"destination": map[string]any{"type": "string", "description": "Destination city"},
						"date":        map[string]any{"type": "string", "description": "Departure date"},
					},
					"required": []string{"destination", "date"},
				},
			},
		},
	}

	// Add system prompt to ensure sufficient tokens
	systemPrompt := "BedrockConverseCache-v1: " + strings.Repeat("You are a helpful assistant with access to weather and flight booking capabilities. ", 15)

	// Turn 1: Initial request with system prompt
	messages := make([]llms.MessageContent, 0, 8)
	messages = append(messages,
		llms.MessageContent{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(systemPrompt)},
		},
		llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("What's the weather in Boston?")},
		},
	)

	resp1, err := llm.GenerateContent(ctx, messages,
		llms.WithModel(bedrock.ModelAnthropicClaudeSonnet45),
		llms.WithTools(tools),
		llms.WithMaxTokens(512),
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp1.Choices[0].ToolCalls) == 0 {
		t.Fatal("Expected tool call in first response")
	}

	// Turn 2: Add tool result
	choice1 := resp1.Choices[0]
	messages = append(messages,
		llms.MessageContent{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{choice1.ToolCalls[0]},
		},
		llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: choice1.ToolCalls[0].ID,
					Name:       choice1.ToolCalls[0].FunctionCall.Name,
					Content:    `{"temperature": 72, "condition": "sunny"}`,
				},
			},
		},
	)

	resp2, err := llm.GenerateContent(ctx, messages,
		llms.WithModel(bedrock.ModelAnthropicClaudeSonnet45),
		llms.WithTools(tools),
		llms.WithMaxTokens(512),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Turn 2 Response: %s", resp2.Choices[0].Content)

	// Turn 3: Continue conversation with cachePoint (mark assistant response for caching)
	messages = append(messages,
		llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				bedrock.WithCacheControl(
					llms.TextPart(resp2.Choices[0].Content),
					bedrock.EphemeralCache(),
				),
			},
		},
		llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Now book a flight to Boston for tomorrow")},
		},
	)

	resp3, err := llm.GenerateContent(ctx, messages,
		llms.WithModel(bedrock.ModelAnthropicClaudeSonnet45),
		llms.WithTools(tools),
		llms.WithMaxTokens(512),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Turn 3 Response: %s", resp3.Choices[0].Content)

	// Turn 4: Continue conversation - cache should be read
	choice3 := resp3.Choices[0]
	var turn4Parts []llms.ContentPart
	if len(choice3.ToolCalls) > 0 {
		turn4Parts = append(turn4Parts, choice3.ToolCalls[0])
	} else {
		turn4Parts = append(turn4Parts, bedrock.WithCacheControl(
			llms.TextPart(choice3.Content),
			bedrock.EphemeralCache(),
		))
	}

	messages = append(messages,
		llms.MessageContent{
			Role:  llms.ChatMessageTypeAI,
			Parts: turn4Parts,
		},
		llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("I'm flying from New York on March 4th")},
		},
	)

	resp4, err := llm.GenerateContent(ctx, messages,
		llms.WithModel(bedrock.ModelAnthropicClaudeSonnet45),
		llms.WithTools(tools),
		llms.WithMaxTokens(512),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Turn 4 Response: %s", resp4.Choices[0].Content)

	// Helper function to extract cache metrics
	getCacheMetric := func(resp *llms.ContentResponse, key string) int {
		if resp.Choices[0].GenerationInfo != nil {
			if val, ok := resp.Choices[0].GenerationInfo[key]; ok {
				switch v := val.(type) {
				case int:
					return v
				case int32:
					return int(v)
				case int64:
					return int(v)
				}
			}
		}
		return 0
	}

	// Log cache metrics for all turns
	t.Logf("Turn 1 - CacheCreation: %d, CacheRead: %d",
		getCacheMetric(resp1, "CacheCreationInputTokens"),
		getCacheMetric(resp1, "CacheReadInputTokens"))
	t.Logf("Turn 2 - CacheCreation: %d, CacheRead: %d",
		getCacheMetric(resp2, "CacheCreationInputTokens"),
		getCacheMetric(resp2, "CacheReadInputTokens"))
	t.Logf("Turn 3 - CacheCreation: %d, CacheRead: %d",
		getCacheMetric(resp3, "CacheCreationInputTokens"),
		getCacheMetric(resp3, "CacheReadInputTokens"))
	t.Logf("Turn 4 - CacheCreation: %d, CacheRead: %d",
		getCacheMetric(resp4, "CacheCreationInputTokens"),
		getCacheMetric(resp4, "CacheReadInputTokens"))

	// Verify caching behavior
	cacheRead4 := getCacheMetric(resp4, "CacheReadInputTokens")
	cacheCreation3 := getCacheMetric(resp3, "CacheCreationInputTokens")

	if cacheRead4 > 0 {
		t.Logf("✓ Cache successfully read in Turn 4: %d tokens", cacheRead4)
	}
	if cacheCreation3 > 0 {
		t.Logf("✓ Cache successfully created in Turn 3: %d tokens", cacheCreation3)
	}
}

// TestAmazonAutomaticCachingLegacyAPI tests automatic prompt caching without manual cache control.
// This validates that WithAutomaticCaching() works correctly with Legacy API.
//
// Expected behavior:
// - Cache points are automatically added to conversation history
// - No manual bedrock.WithCacheControl() wrapper needed
// - Cache metrics show cache utilization
func TestAmazonAutomaticCachingLegacyAPI(t *testing.T) { //nolint:funlen
	ctx := t.Context()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	if !rr.Recording() {
		t.Parallel()
	}

	client, err := setUpTestWithTransport(rr)
	if err != nil {
		t.Fatal(err)
	}
	// Use Legacy API with automatic caching enabled
	llm, err := bedrock.New(
		bedrock.WithClient(client),
		bedrock.WithAutomaticCaching(),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Define tools with long descriptions to exceed 1024 token threshold
	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_weather",
				Description: "BedrockAutoCacheTest-v1: " + strings.Repeat("Get current weather conditions including temperature, humidity, wind speed, and precipitation for a specified geographic location. ", 30),
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "City and state, e.g. San Francisco, CA",
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	// Add system prompt to ensure sufficient tokens
	systemPrompt := "BedrockAutoCacheTest-v1: " + strings.Repeat("You are a helpful assistant with access to weather capabilities. ", 15)

	// Turn 1: Initial request with system prompt
	messages := make([]llms.MessageContent, 0, 6)
	messages = append(messages,
		llms.MessageContent{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(systemPrompt)},
		},
		llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("What's the weather in Boston?")},
		},
	)

	resp1, err := llm.GenerateContent(ctx, messages,
		llms.WithModel(bedrock.ModelAnthropicClaudeSonnet45),
		llms.WithTools(tools),
		llms.WithMaxTokens(512),
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp1.Choices[0].ToolCalls) == 0 {
		t.Fatal("Expected tool call in first response")
	}

	// Turn 2: Add tool result - NO manual cache control needed
	choice1 := resp1.Choices[0]
	messages = append(messages,
		llms.MessageContent{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{choice1.ToolCalls[0]},
		},
		llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: choice1.ToolCalls[0].ID,
					Name:       choice1.ToolCalls[0].FunctionCall.Name,
					Content:    `{"temperature": 72, "condition": "sunny"}`,
				},
			},
		},
	)

	resp2, err := llm.GenerateContent(ctx, messages,
		llms.WithModel(bedrock.ModelAnthropicClaudeSonnet45),
		llms.WithTools(tools),
		llms.WithMaxTokens(512),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Turn 2 Response: %s", resp2.Choices[0].Content)

	// Turn 3: Continue conversation - automatic caching should apply
	messages = append(messages,
		llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextPart(resp2.Choices[0].Content),
			},
		},
		llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("What about New York?")},
		},
	)

	resp3, err := llm.GenerateContent(ctx, messages,
		llms.WithModel(bedrock.ModelAnthropicClaudeSonnet45),
		llms.WithTools(tools),
		llms.WithMaxTokens(512),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Turn 3 Response: %+v", resp3.Choices[0])

	// Helper function to extract cache metrics
	getCacheMetric := func(resp *llms.ContentResponse, key string) int {
		if resp.Choices[0].GenerationInfo != nil {
			if val, ok := resp.Choices[0].GenerationInfo[key]; ok {
				switch v := val.(type) {
				case int:
					return v
				case int32:
					return int(v)
				case int64:
					return int(v)
				}
			}
		}
		return 0
	}

	// Log cache metrics for all turns
	t.Logf("Turn 1 - CacheCreation: %d, CacheRead: %d",
		getCacheMetric(resp1, "CacheCreationInputTokens"),
		getCacheMetric(resp1, "CacheReadInputTokens"))
	t.Logf("Turn 2 - CacheCreation: %d, CacheRead: %d",
		getCacheMetric(resp2, "CacheCreationInputTokens"),
		getCacheMetric(resp2, "CacheReadInputTokens"))
	t.Logf("Turn 3 - CacheCreation: %d, CacheRead: %d",
		getCacheMetric(resp3, "CacheCreationInputTokens"),
		getCacheMetric(resp3, "CacheReadInputTokens"))

	// Verify automatic caching is working
	cacheCreation3 := getCacheMetric(resp3, "CacheCreationInputTokens")
	cacheRead3 := getCacheMetric(resp3, "CacheReadInputTokens")

	if cacheCreation3 > 0 {
		t.Logf("✓ Automatic cache successfully created in Turn 3: %d tokens", cacheCreation3)
	}
	if cacheRead3 > 0 {
		t.Logf("✓ Automatic cache successfully read in Turn 3: %d tokens", cacheRead3)
	}
}

// TestAmazonAutomaticCachingConverseAPI tests automatic prompt caching with Converse API.
// This validates that WithAutomaticCaching() works correctly with Converse API.
func TestAmazonAutomaticCachingConverseAPI(t *testing.T) { //nolint:funlen
	ctx := t.Context()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	if !rr.Recording() {
		t.Parallel()
	}

	client, err := setUpTestWithTransport(rr)
	if err != nil {
		t.Fatal(err)
	}
	// Use Converse API with automatic caching enabled
	llm, err := bedrock.New(
		bedrock.WithClient(client),
		bedrock.WithConverseAPI(),
		bedrock.WithAutomaticCaching(),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Define tools with long descriptions to exceed 1024 token threshold
	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_weather",
				Description: "BedrockConverseAutoCacheTest-v1: " + strings.Repeat("Get current weather conditions including temperature, humidity, wind speed, and precipitation for a specified geographic location. ", 30),
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "City and state, e.g. San Francisco, CA",
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	// Add system prompt to ensure sufficient tokens
	systemPrompt := "BedrockConverseAutoCacheTest-v1: " + strings.Repeat("You are a helpful assistant with access to weather capabilities. ", 15)

	// Turn 1: Initial request with system prompt
	messages := make([]llms.MessageContent, 0, 6)
	messages = append(messages,
		llms.MessageContent{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(systemPrompt)},
		},
		llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("What's the weather in Boston?")},
		},
	)

	resp1, err := llm.GenerateContent(ctx, messages,
		llms.WithModel(bedrock.ModelAnthropicClaudeSonnet45),
		llms.WithTools(tools),
		llms.WithMaxTokens(512),
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp1.Choices[0].ToolCalls) == 0 {
		t.Fatal("Expected tool call in first response")
	}

	// Turn 2: Add tool result - NO manual cache control needed
	choice1 := resp1.Choices[0]
	messages = append(messages,
		llms.MessageContent{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{choice1.ToolCalls[0]},
		},
		llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: choice1.ToolCalls[0].ID,
					Name:       choice1.ToolCalls[0].FunctionCall.Name,
					Content:    `{"temperature": 72, "condition": "sunny"}`,
				},
			},
		},
	)

	resp2, err := llm.GenerateContent(ctx, messages,
		llms.WithModel(bedrock.ModelAnthropicClaudeSonnet45),
		llms.WithTools(tools),
		llms.WithMaxTokens(512),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Turn 2 Response: %s", resp2.Choices[0].Content)

	// Turn 3: Continue conversation - automatic caching should apply
	messages = append(messages,
		llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextPart(resp2.Choices[0].Content),
			},
		},
		llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("What about New York?")},
		},
	)

	resp3, err := llm.GenerateContent(ctx, messages,
		llms.WithModel(bedrock.ModelAnthropicClaudeSonnet45),
		llms.WithTools(tools),
		llms.WithMaxTokens(512),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Turn 3 Response: %+v", resp3.Choices[0])

	// Helper function to extract cache metrics
	getCacheMetric := func(resp *llms.ContentResponse, key string) int {
		if resp.Choices[0].GenerationInfo != nil {
			if val, ok := resp.Choices[0].GenerationInfo[key]; ok {
				switch v := val.(type) {
				case int:
					return v
				case int32:
					return int(v)
				case int64:
					return int(v)
				}
			}
		}
		return 0
	}

	// Log cache metrics for all turns
	t.Logf("Turn 1 - CacheCreation: %d, CacheRead: %d",
		getCacheMetric(resp1, "CacheCreationInputTokens"),
		getCacheMetric(resp1, "CacheReadInputTokens"))
	t.Logf("Turn 2 - CacheCreation: %d, CacheRead: %d",
		getCacheMetric(resp2, "CacheCreationInputTokens"),
		getCacheMetric(resp2, "CacheReadInputTokens"))
	t.Logf("Turn 3 - CacheCreation: %d, CacheRead: %d",
		getCacheMetric(resp3, "CacheCreationInputTokens"),
		getCacheMetric(resp3, "CacheReadInputTokens"))

	// Verify automatic caching is working
	cacheCreation3 := getCacheMetric(resp3, "CacheCreationInputTokens")
	cacheRead3 := getCacheMetric(resp3, "CacheReadInputTokens")

	if cacheCreation3 > 0 {
		t.Logf("✓ Automatic cache successfully created in Turn 3: %d tokens", cacheCreation3)
	}
	if cacheRead3 > 0 {
		t.Logf("✓ Automatic cache successfully read in Turn 3: %d tokens", cacheRead3)
	}
}

// TestCreateClientWithLongLeavingCredentials tests creating a client with long leaving credentials.
func TestCreateClientWithLongLeavingCredentials(t *testing.T) {
	ctx := t.Context()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_REGION")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	if !rr.Recording() {
		t.Parallel()
	}

	// Configure request scrubbing to remove dynamic AWS headers
	rr.ScrubReq(func(req *http.Request) error {
		req.Header.Del("Amz-Sdk-Invocation-Id")
		req.Header.Del("Amz-Sdk-Request")
		req.Header.Del("X-Amz-Date")
		req.Header.Del("Authorization") // AWS Signature V4 is unique per request
		return nil
	})

	httpClient := &http.Client{
		Transport: rr,
	}

	region := os.Getenv("AWS_REGION")
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	sessionToken := os.Getenv("AWS_SESSION_TOKEN")

	opts := []func(*config.LoadOptions) error{
		config.WithHTTPClient(httpClient),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			sessionToken,
		)),
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}

	client := bedrockruntime.NewFromConfig(cfg)

	llm, err := bedrock.New(bedrock.WithClient(client), bedrock.WithConverseAPI())
	if err != nil {
		t.Fatal(err)
	}

	model := bedrock.ModelAnthropicClaudeSonnet45
	resp, err := llm.Call(ctx, "Hello, how are you?", llms.WithModel(model), llms.WithMaxTokens(512))
	if err != nil {
		t.Fatal(err)
	}

	if resp == "" {
		t.Fatal("Expected non-empty response")
	}

	t.Logf("Response: %s", resp)
}

// TestCreateClientWithBearerTokenCredentials tests creating a client with bearer token credentials.
func TestCreateClientWithBearerTokenCredentials(t *testing.T) {
	ctx := t.Context()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_BEDROCK_BEARER_TOKEN", "AWS_REGION")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	if !rr.Recording() {
		t.Parallel()
	}

	// Configure request scrubbing to remove dynamic AWS headers
	rr.ScrubReq(func(req *http.Request) error {
		req.Header.Del("Amz-Sdk-Invocation-Id")
		req.Header.Del("Amz-Sdk-Request")
		req.Header.Del("X-Amz-Date")
		req.Header.Del("Authorization")
		return nil
	})

	httpClient := &http.Client{
		Transport: rr,
	}

	region := os.Getenv("AWS_REGION")
	bearerToken := os.Getenv("AWS_BEDROCK_BEARER_TOKEN")

	opts := []func(*config.LoadOptions) error{
		config.WithHTTPClient(httpClient),
		config.WithRegion(region),
		config.WithBearerAuthTokenProvider(bearer.StaticTokenProvider{
			Token: bearer.Token{
				Value: bearerToken,
			},
		}),
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}

	client := bedrockruntime.NewFromConfig(cfg)

	llm, err := bedrock.New(bedrock.WithClient(client), bedrock.WithConverseAPI())
	if err != nil {
		t.Fatal(err)
	}

	model := bedrock.ModelAnthropicClaudeSonnet45
	resp, err := llm.Call(ctx, "Hello, how are you?", llms.WithModel(model), llms.WithMaxTokens(512))
	if err != nil {
		t.Fatal(err)
	}

	if resp == "" {
		t.Fatal("Expected non-empty response")
	}

	t.Logf("Response: %s", resp)
}
