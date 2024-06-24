package bedrock_test

import (
	"context"
	"os"
	"testing"
	"unicode/utf8"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/bedrock"
)

func setUpTest() (*bedrockruntime.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}
	client := bedrockruntime.NewFromConfig(cfg)
	return client, nil
}

func TestAmazonOutput(t *testing.T) {
	if os.Getenv("TEST_AWS") != "true" {
		t.Skip("Skipping test, requires AWS access")
	}

	client, err := setUpTest()
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

	type testcase struct {
		model string
	}

	// All the test models.
	tests := []testcase{
		{model: bedrock.ModelAi21J2MidV1},
		{model: bedrock.ModelAi21J2UltraV1},
		{model: bedrock.ModelAmazonTitanTextLiteV1},
		{model: bedrock.ModelAmazonTitanTextExpressV1},
		{model: bedrock.ModelAnthropicClaudeV3Sonnet},
		{model: bedrock.ModelAnthropicClaudeV3Haiku},
		{model: bedrock.ModelAnthropicClaudeV21},
		{model: bedrock.ModelAnthropicClaudeV2},
		{model: bedrock.ModelAnthropicClaudeV35Sonnet},
		{model: bedrock.ModelAnthropicClaudeInstantV1},
		{model: bedrock.ModelCohereCommandTextV14},
		{model: bedrock.ModelCohereCommandLightTextV14},
		{model: bedrock.ModelMetaLlama213bChatV1},
		{model: bedrock.ModelMetaLlama270bChatV1},
		{model: bedrock.ModelMetaLlama38bInstructV1},
		{model: bedrock.ModelMetaLlama370bInstructV1},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			t.Parallel()
			t.Logf("Model output for %s:-", tt.model)
			resp, err := llm.GenerateContent(ctx, msgs, llms.WithModel(tt.model), llms.WithMaxTokens(512))
			if err != nil {
				t.Fatal(err)
			}
			assert.NotEmpty(t, resp.Choices)
			for i, choice := range resp.Choices {
				t.Logf("Choice %d: %s", i, choice.Content)
			}
			assert.Greater(t, utf8.RuneCountInString(resp.Choices[0].Content), 5)
		})
	}
}

func TestBedrockConverseStream(t *testing.T) {
	t.Parallel()
	if os.Getenv("TEST_AWS") != "true" {
		t.Skip("Skipping test, requires AWS access")
	}

	client, err := setUpTest()
	if err != nil {
		t.Fatal(err)
	}
	llm, err := bedrock.New(bedrock.WithClient(client))
	if err != nil {
		t.Fatalf("%v", err)
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

	type testcase struct {
		model string
	}

	// All the test models.
	tests := []testcase{
		{model: bedrock.ModelAi21J2MidV1},
		{model: bedrock.ModelAi21J2UltraV1},
		{model: bedrock.ModelAmazonTitanTextLiteV1},
		{model: bedrock.ModelAmazonTitanTextExpressV1},
		{model: bedrock.ModelAnthropicClaudeV3Sonnet},
		{model: bedrock.ModelAnthropicClaudeV3Haiku},
		{model: bedrock.ModelAnthropicClaudeV21},
		{model: bedrock.ModelAnthropicClaudeV2},
		{model: bedrock.ModelAnthropicClaudeV35Sonnet},
		{model: bedrock.ModelAnthropicClaudeInstantV1},
		{model: bedrock.ModelCohereCommandTextV14},
		{model: bedrock.ModelCohereCommandLightTextV14},
		{model: bedrock.ModelMetaLlama213bChatV1},
		{model: bedrock.ModelMetaLlama270bChatV1},
		{model: bedrock.ModelMetaLlama38bInstructV1},
		{model: bedrock.ModelMetaLlama370bInstructV1},
	}

	ctx := context.Background()
	streamFunc := func(_ context.Context, _ []byte) error {
		// t.Logf("Stream chunk: %s", string(chunk))
		return nil
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			t.Parallel()
			t.Logf("Model output for %s:-", tt.model)
			resp, err := llm.GenerateContent(
				ctx,
				msgs,
				llms.WithModel(tt.model),
				llms.WithMaxTokens(512),
				llms.WithStreamingFunc(streamFunc),
			)
			if err != nil {
				t.Fatal(err)
			}
			assert.NotEmpty(t, resp.Choices)
			for i, choice := range resp.Choices {
				t.Logf("Choice %d: %s", i, choice.Content)
			}
			assert.Greater(t, utf8.RuneCountInString(resp.Choices[0].Content), 5)
		})
	}
}
