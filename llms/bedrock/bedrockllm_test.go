package bedrock_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/bedrock"
)

func setUpTestWithTransport(transport http.RoundTripper) (*bedrockruntime.Client, error) {
	httpClient := &http.Client{
		Transport: transport,
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithHTTPClient(httpClient))
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
		bedrock.ModelAmazonNovaMicroV1,
		bedrock.ModelAmazonNovaLiteV1,
		bedrock.ModelAmazonNovaProV1,
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
			t.Fatal(err)
		}
		for i, choice := range resp.Choices {
			t.Logf("Choice %d: %s", i, choice.Content)
		}
	}
}

func TestBedrockWithTools(t *testing.T) {
	ctx := context.Background()

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

	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "getCurrentWeather",
				Description: "Get the current weather in a given location",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{
							"type":        "string",
							"description": "The city and state, e.g. San Francisco, CA",
						},
						"unit": map[string]interface{}{
							"type":        "string",
							"description": "Temperature unit",
							"enum":        []string{"celsius", "fahrenheit"},
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "What is the weather like in Chicago?"),
	}

	resp, err := llm.GenerateContent(ctx, content,
		llms.WithTools(tools),
		llms.WithModel(bedrock.ModelAnthropicClaudeV3Sonnet))
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("No response choices returned")
	}

	c1 := resp.Choices[0]

	// Check if tool call was made
	if len(c1.ToolCalls) > 0 {
		t.Logf("Tool call made: %v", c1.ToolCalls[0].FunctionCall.Name)

		// Update chat history with assistant's response, with its tool calls.
		assistantResp := llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
		}
		for _, tc := range c1.ToolCalls {
			assistantResp.Parts = append(assistantResp.Parts, tc)
		}
		content = append(content, assistantResp)

		// "Execute" tool call
		for _, tc := range c1.ToolCalls {
			switch tc.FunctionCall.Name {
			case "getCurrentWeather":
				var args struct {
					Location string `json:"location"`
				}
				if err := json.Unmarshal([]byte(tc.FunctionCall.Arguments), &args); err != nil {
					t.Fatal(err)
				}
				if strings.Contains(strings.ToLower(args.Location), "chicago") {
					toolResponse := llms.MessageContent{
						Role: llms.ChatMessageTypeTool,
						Parts: []llms.ContentPart{
							llms.ToolCallResponse{
								ToolCallID: tc.ID,
								Name:       tc.FunctionCall.Name,
								Content:    "64 and sunny",
							},
						},
					}
					content = append(content, toolResponse)
				}
			default:
				t.Errorf("got unexpected function call: %v", tc.FunctionCall.Name)
			}
		}

		// Send follow-up request with tool response
		resp, err = llm.GenerateContent(ctx, content,
			llms.WithTools(tools),
			llms.WithModel(bedrock.ModelAnthropicClaudeV3Sonnet))
		if err != nil {
			t.Fatal(err)
		}

		if len(resp.Choices) == 0 {
			t.Fatal("No response choices returned after tool call")
		}

		c1 = resp.Choices[0]
		t.Logf("Final response: %s", c1.Content)
		if !strings.Contains(strings.ToLower(c1.Content), "64") {
			t.Logf("Warning: Expected weather data '64' not found in response: %s", c1.Content)
		}
	}
}

func TestBedrockToolCallMultipleIterations(t *testing.T) {
	ctx := context.Background()

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

	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "searchWeb",
				Description: "Search the web for information",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "Search query",
						},
					},
					"required": []string{"query"},
				},
			},
		},
	}

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "What is the capital of France?"),
	}

	// First iteration
	resp, err := llm.GenerateContent(ctx, content,
		llms.WithTools(tools),
		llms.WithModel(bedrock.ModelAnthropicClaudeV3Sonnet))
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("No response choices returned")
	}

	c1 := resp.Choices[0]
	if len(c1.ToolCalls) > 0 {
		// Add assistant response
		assistantResp := llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
		}
		for _, tc := range c1.ToolCalls {
			assistantResp.Parts = append(assistantResp.Parts, tc)
		}
		content = append(content, assistantResp)

		// Add tool response
		for _, tc := range c1.ToolCalls {
			toolResponse := llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						ToolCallID: tc.ID,
						Name:       tc.FunctionCall.Name,
						Content:    "Paris is the capital of France.",
					},
				},
			}
			content = append(content, toolResponse)
		}

		// Second iteration
		resp, err = llm.GenerateContent(ctx, content,
			llms.WithTools(tools),
			llms.WithModel(bedrock.ModelAnthropicClaudeV3Sonnet))
		if err != nil {
			t.Fatal(err)
		}

		if len(resp.Choices) == 0 {
			t.Fatal("No response choices returned after tool call")
		}

		// Should handle multiple iterations without errors
		if resp.Choices[0].Content == "" {
			t.Fatal("Empty response content after tool call iteration")
		}
		t.Logf("Multi-iteration response: %s", resp.Choices[0].Content)
	}
}
