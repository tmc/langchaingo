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
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
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

func TestAmazonOutputConverseAPI(t *testing.T) {
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
		bedrock.ModelAnthropicClaudeOpus45,
		bedrock.ModelAnthropicClaudeHaiku45,
		bedrock.ModelAnthropicClaudeSonnet45,
		bedrock.ModelAnthropicClaudeOpus41,
		bedrock.ModelAnthropicClaudeOpus4,
		bedrock.ModelAnthropicClaudeSonnet4,
		bedrock.ModelAnthropicClaude37Sonnet,
		bedrock.ModelAnthropicClaude35Haiku,

		// Cohere models (only Command-R supports streaming and Converse API)
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

func TestAnthropicNovaImage(t *testing.T) {
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
		bedrock.ModelAnthropicClaudeOpus45,
		bedrock.ModelAnthropicClaudeHaiku45,
		bedrock.ModelAnthropicClaudeSonnet45,
		// bedrock.ModelAnthropicClaudeOpus41,        // Has very hard rate limits
		bedrock.ModelAnthropicClaudeOpus4,
		bedrock.ModelAnthropicClaudeSonnet4,
		bedrock.ModelAnthropicClaude37Sonnet,
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
		bedrock.ModelAnthropicClaudeOpus45,
		bedrock.ModelAnthropicClaudeHaiku45,
		bedrock.ModelAnthropicClaudeSonnet45,
		// bedrock.ModelAnthropicClaudeOpus41,        // Has very hard rate limits
		bedrock.ModelAnthropicClaudeOpus4,
		bedrock.ModelAnthropicClaudeSonnet4,
		bedrock.ModelAnthropicClaude37Sonnet,
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

func testReasoningWorkflow(
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

	contents := []llms.MessageContent{
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
