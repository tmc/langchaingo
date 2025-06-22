package bedrock_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/tmc/langchaingo/httputil"
	"github.com/tmc/langchaingo/internal/httprr"
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
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	// Only run tests in parallel when not recording (to avoid rate limits)
	if !rr.Recording() {
		t.Parallel()
	}

	// Replace httputil.DefaultClient with httprr client
	oldClient := httputil.DefaultClient
	httputil.DefaultClient = rr.Client()
	defer func() { httputil.DefaultClient = oldClient }()

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

	// All the test models.
	models := []string{
		bedrock.ModelAi21J2MidV1,
		bedrock.ModelAi21J2UltraV1,
		bedrock.ModelAmazonTitanTextLiteV1,
		bedrock.ModelAmazonTitanTextExpressV1,
		bedrock.ModelAnthropicClaudeV3Sonnet,
		bedrock.ModelAnthropicClaudeV3Haiku,
		bedrock.ModelAnthropicClaudeV21,
		bedrock.ModelAnthropicClaudeV2,
		bedrock.ModelAnthropicClaudeInstantV1,
		bedrock.ModelCohereCommandTextV14,
		bedrock.ModelCohereCommandLightTextV14,
		bedrock.ModelMetaLlama213bChatV1,
		bedrock.ModelMetaLlama270bChatV1,
		bedrock.ModelMetaLlama38bInstructV1,
		bedrock.ModelMetaLlama370bInstructV1,
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
