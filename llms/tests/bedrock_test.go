package tests

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"

	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/bedrock"
)

// nolint:gochecknoglobals
var (
	defaultBedrockClient llms.Model
	usEast1BedrockClient llms.Model
)

func getDefaultBedrockClient(t *testing.T) llms.Model {
	t.Helper()
	if os.Getenv("TEST_AWS") != "true" {
		t.Skip("Skipping test, requires AWS access")
	}
	if defaultBedrockClient != nil {
		return defaultBedrockClient
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	client := bedrockruntime.NewFromConfig(cfg)
	llm, err := bedrock.New(bedrock.WithClient(client))
	if err != nil {
		t.Fatal(err)
	}
	return llm
}

// As of June 2024, only the us-east-1 region supports bedrock.ModelAnthropicClaudeV35Sonnet model.
func getUsEast1BedrockClient(t *testing.T) llms.Model {
	t.Helper()
	if os.Getenv("TEST_AWS") != "true" {
		t.Skip("Skipping test, requires AWS access")
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if usEast1BedrockClient != nil {
		return usEast1BedrockClient
	}

	cfg.Region = "us-east-1"
	client := bedrockruntime.NewFromConfig(cfg)
	llm, err := bedrock.New(bedrock.WithClient(client))
	if err != nil {
		t.Fatal(err)
	}
	return llm
}

type bedrockTestCase struct {
	model  string
	client llms.Model
}

func runBedrockTestCase(t *testing.T, test *bedrockTestCase, arg testArgs) {
	t.Helper()
	var client llms.Model
	if test.client == nil {
		client = getDefaultBedrockClient(t)
	} else {
		client = test.client
	}
	recv := newStreamRecv()
	resp, err := client.GenerateContent(
		context.Background(),
		arg.messages,
		llms.WithModel(test.model),
		llms.WithStreamingFunc(recv.streamFunc),
		llms.WithTools(arg.tools),
		llms.WithToolChoice(arg.toolChoice),
	)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	assert.NotEmpty(t, resp.Choices)
	t.Logf("content: %s", resp.Choices[0].Content)
	t.Logf("tool_call: %v", resp.Choices[0].ToolCalls)
	logTools(t, resp.Choices[0].ToolCalls)
	if len(resp.Choices[0].ToolCalls) > 0 {
		messages := make([]llms.MessageContent, 0)
		messages = append(messages, arg.messages...)
		if resp.Choices[0].Content != "" {
			messages = append(messages, llms.MessageContent{
				Role:  llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{llms.TextPart(resp.Choices[0].Content)},
			})
		}
		for _, toolCall := range resp.Choices[0].ToolCalls {
			messages = append(messages, llms.MessageContent{
				Role:  llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{toolCall},
			})
			part, err := callFunction(toolCall)
			if part == nil || err != nil {
				t.Fatalf("%+v", err)
			}
			messages = append(messages, llms.MessageContent{
				Role:  llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{*part},
			})
		}
		resp, err = client.GenerateContent(
			context.Background(),
			messages,
			llms.WithModel(test.model),
			llms.WithStreamingFunc(recv.streamFunc),
			llms.WithTools(arg.tools),
			llms.WithToolChoice(arg.toolChoice),
		)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		assert.NotEmpty(t, resp.Choices)
		t.Logf("content: %s", resp.Choices[0].Content)
	}
}

func getDefaultBedrockClaude3TestCases(t *testing.T) []bedrockTestCase {
	t.Helper()
	return []bedrockTestCase{
		{model: bedrock.ModelAnthropicClaudeV3Haiku},
		{model: bedrock.ModelAnthropicClaudeV35Sonnet, client: getUsEast1BedrockClient(t)},
		{model: bedrock.ModelAnthropicClaudeV3Sonnet},
		{model: bedrock.ModelAnthropicClaudeV3Opus},
	}
}

func TestBedrockStreamMessages(t *testing.T) {
	t.Parallel()
	modelList := getDefaultBedrockClaude3TestCases(t)
	for _, test := range modelList {
		t.Parallel()
		t.Run(test.model, func(t *testing.T) {
			runBedrockTestCase(t, &test, converseWithSystem)
		})
	}
}

func TestBedrockStreamMessagesWithImage(t *testing.T) {
	t.Parallel()
	modelList := getDefaultBedrockClaude3TestCases(t)
	for _, test := range modelList {
		t.Parallel()
		t.Run(test.model, func(t *testing.T) {
			runBedrockTestCase(t, &test, converseWithImage)
		})
	}
}

func TestBedrockStreamMessagesWithImageAndTools(t *testing.T) {
	modelList := getDefaultBedrockClaude3TestCases(t)
	t.Parallel()
	for _, test := range modelList {
		t.Run(test.model, func(t *testing.T) {
			t.Parallel()
			runBedrockTestCase(t, &test, converseImageWithTools)
		})
	}
}
