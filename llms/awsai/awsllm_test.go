package awsai_test

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/awsai"
	"github.com/tmc/langchaingo/schema"
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
	t.Parallel()

	if os.Getenv("TEST_AWS") != "true" {
		t.Skip("Skipping test, requires AWS access")
	}

	client, err := setUpTest()
	if err != nil {
		t.Fatal(err)
	}
	llm, err := awsai.New(awsai.WithClient(client), awsai.WithApiType(awsai.ApiTypeBedrock))
	if err != nil {
		t.Fatal(err)
	}

	msgs := []llms.MessageContent{
		{
			Role: schema.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextPart("You know all about AI."),
			},
		},
		{
			Role: schema.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Explain AI in 10 words or less."),
			},
		},
	}

	// All the test models.
	models := []string{
		awsai.ModelAi21J2MidV1,
		awsai.ModelAi21J2UltraV1,
		awsai.ModelAmazonTitanTextLiteV1,
		awsai.ModelAmazonTitanTextExpressV1,
		awsai.ModelAnthropicClaudeV3Sonnet,
		awsai.ModelAnthropicClaudeV3Haiku,
		awsai.ModelAnthropicClaudeV21,
		awsai.ModelAnthropicClaudeV2,
		awsai.ModelAnthropicClaudeInstantV1,
		awsai.ModelCohereCommandTextV14,
		awsai.ModelCohereCommandLightTextV14,
		awsai.ModelMetaLlama213bChatV1,
		awsai.ModelMetaLlama270bChatV1,
	}

	ctx := context.Background()

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
