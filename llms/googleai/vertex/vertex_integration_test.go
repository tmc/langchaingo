//go:build integration
// +build integration

package vertex

import (
	"context"
	"os"
	"testing"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
)

// Integration tests for Vertex AI with google.golang.org/genai library.
//
// These tests require a real API key and will make actual API calls.
// They are excluded from normal test runs by the //+build integration tag.
//
// To run these tests:
//
//   1. Set your API key as an environment variable:
//      export VERTEX_API_KEY="your-api-key-here"
//      export VERTEX_PROJECT_ID="your-gcp-project-id"
//
//   2. Run with integration tag:
//      go test -tags=integration -v
//
//   3. Or run a specific test:
//      go test -tags=integration -v -run TestVertexWithAPIKey
//
// Note: You can also use GOOGLE_API_KEY instead of VERTEX_API_KEY.

func TestVertexWithAPIKey(t *testing.T) {
	// Check for API key - can be set via VERTEX_API_KEY or GOOGLE_API_KEY env vars
	apiKey := os.Getenv("VERTEX_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey == "" {
		t.Skip("SKIP: VERTEX_API_KEY or GOOGLE_API_KEY not set - skipping integration test")
	}

	ctx := context.Background()

	// Create a Vertex client - API key will be picked up from environment
	// When using API key, project/location are not required (in fact, they're mutually exclusive)
	client, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create Vertex client: %v (make sure VERTEX_API_KEY is set)", err)
	}
	defer client.Close()

	// Test basic completion
	t.Run("BasicCompletion", func(t *testing.T) {
		response, err := llms.GenerateFromSinglePrompt(
			ctx,
			client,
			"Say 'Hello, Vertex AI!' in one sentence.",
			llms.WithModel("gemini-2.5-flash"),
		)
		if err != nil {
			t.Fatalf("GenerateFromSinglePrompt failed: %v", err)
		}

		t.Logf("Response: %s", response)
		if len(response) == 0 {
			t.Error("Empty response content")
		}
	})

	// Test embeddings
	t.Run("Embeddings", func(t *testing.T) {
		embeddings, err := client.CreateEmbedding(ctx, []string{"Hello, Vertex AI!"})
		if err != nil {
			t.Fatalf("CreateEmbedding failed: %v", err)
		}

		if len(embeddings) != 1 {
			t.Fatalf("Expected 1 embedding, got %d", len(embeddings))
		}

		if len(embeddings[0]) == 0 {
			t.Error("Empty embedding vector")
		}

		t.Logf("Generated embedding with %d dimensions", len(embeddings[0]))
	})

	// Test multi-turn conversation
	t.Run("Conversation", func(t *testing.T) {
		messages := []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, "What is 2+2?"),
			llms.TextParts(llms.ChatMessageTypeAI, "2+2 equals 4."),
			llms.TextParts(llms.ChatMessageTypeHuman, "What about 3+3?"),
		}

		response, err := client.GenerateContent(ctx, messages, llms.WithModel("gemini-2.5-flash"))
		if err != nil {
			t.Fatalf("GenerateContent failed: %v", err)
		}

		if len(response.Choices) == 0 {
			t.Fatal("No choices returned")
		}

		t.Logf("Conversation response: %s", response.Choices[0].Content)
		if len(response.Choices[0].Content) == 0 {
			t.Error("Empty response content")
		}
	})
}

func TestVertexEmbeddingsBatch(t *testing.T) {
	apiKey := os.Getenv("VERTEX_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey == "" {
		t.Skip("SKIP: VERTEX_API_KEY or GOOGLE_API_KEY not set - skipping integration test")
	}

	ctx := context.Background()

	client, err := New(ctx,
		googleai.WithCloudProject(os.Getenv("VERTEX_PROJECT_ID")),
		googleai.WithCloudLocation("us-central1"),
	)
	if err != nil {
		t.Fatalf("Failed to create Vertex client: %v", err)
	}
	defer client.Close()

	// Test batch embeddings
	texts := []string{
		"The capital of France is Paris.",
		"Python is a programming language.",
		"The sky is blue during the day.",
	}

	embeddings, err := client.CreateEmbedding(ctx, texts)
	if err != nil {
		t.Fatalf("CreateEmbedding failed: %v", err)
	}

	if len(embeddings) != len(texts) {
		t.Fatalf("Expected %d embeddings, got %d", len(texts), len(embeddings))
	}

	// Verify all embeddings have same dimensions
	dim := len(embeddings[0])
	if dim == 0 {
		t.Error("Empty embedding dimensions")
	}

	for i, emb := range embeddings {
		if len(emb) != dim {
			t.Errorf("Embedding %d has different dimensions: got %d, expected %d", i, len(emb), dim)
		}
	}

	t.Logf("Generated %d embeddings, each with %d dimensions", len(embeddings), dim)
}

// Test with minimal configuration - just project ID
func TestVertexMinimalConfig(t *testing.T) {
	if os.Getenv("VERTEX_API_KEY") == "" && os.Getenv("GOOGLE_API_KEY") == "" {
		t.Skip("SKIP: VERTEX_API_KEY or GOOGLE_API_KEY not set")
	}

	ctx := context.Background()

	projectID := os.Getenv("VERTEX_PROJECT_ID")
	if projectID == "" {
		t.Skip("SKIP: VERTEX_PROJECT_ID not set")
	}

	// Create client with minimal options - API key from environment
	client, err := New(ctx,
		googleai.WithCloudProject(projectID),
		googleai.WithCloudLocation("us-central1"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	response, err := llms.GenerateFromSinglePrompt(ctx, client, "Say hello", llms.WithModel("gemini-2.5-flash"))
	if err != nil {
		t.Fatalf("Failed to generate content: %v", err)
	}

	if len(response) == 0 {
		t.Fatal("Empty response")
	}

	t.Logf("Successfully got response: %s", response)
}

func TestVertexWithToolCalls(t *testing.T) {
	// Check for API key - can be set via VERTEX_API_KEY or GOOGLE_API_KEY env vars
	apiKey := os.Getenv("VERTEX_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey == "" {
		t.Skip("SKIP: VERTEX_API_KEY or GOOGLE_API_KEY not set - skipping integration test")
	}

	ctx := context.Background()

	// Create a Vertex client - API key will be picked up from environment
	client, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create Vertex client: %v", err)
	}
	defer client.Close()

	// Define a simple function tool for testing
	testTools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_weather",
				Description: "Get the current weather in a given location",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "The city and state, e.g. San Francisco, CA",
						},
						"unit": map[string]any{
							"type":        "string",
							"enum":        []string{"celsius", "fahrenheit"},
							"description": "The unit of temperature to return",
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	// Test: The model should try to call the tool
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "What's the weather in San Francisco?"),
	}

	response, err := client.GenerateContent(ctx, messages, llms.WithModel("gemini-2.5-flash"), llms.WithTools(testTools))
	if err != nil {
		t.Fatalf("GenerateContent failed: %v", err)
	}

	if len(response.Choices) == 0 {
		t.Fatal("No choices returned")
	}

	choice := response.Choices[0]
	t.Logf("Response: %s", choice.Content)
	t.Logf("Stop reason: %s", choice.StopReason)

	// Check if tool calls were generated
	if len(choice.ToolCalls) > 0 {
		t.Logf("Tool calls detected: %d", len(choice.ToolCalls))
		for i, toolCall := range choice.ToolCalls {
			t.Logf("Tool Call %d: %s with args: %s", i+1, toolCall.FunctionCall.Name, toolCall.FunctionCall.Arguments)
		}

		// Simulate a tool response and make a second call
		// Add the model's tool call to the conversation
		if len(choice.ToolCalls) > 0 && choice.ToolCalls[0].FunctionCall.Name == "get_weather" {
			// Add the assistant's tool call
			messages = append(messages, llms.MessageContent{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.ToolCall{
						FunctionCall: choice.ToolCalls[0].FunctionCall,
					},
				},
			})

			// Add the tool response
			messages = append(messages, llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						Name:    choice.ToolCalls[0].FunctionCall.Name,
						Content: `{"location": "San Francisco, CA", "temperature": "72", "unit": "fahrenheit", "condition": "Partly cloudy"}`,
					},
				},
			})

			// Get the final response
			finalResponse, err := client.GenerateContent(ctx, messages, llms.WithModel("gemini-2.5-flash"))
			if err != nil {
				t.Fatalf("Second GenerateContent failed: %v", err)
			}

			if len(finalResponse.Choices) == 0 {
				t.Fatal("No choices in final response")
			}

			t.Logf("Final response: %s", finalResponse.Choices[0].Content)
		}
	} else {
		t.Log("No tool calls generated - this is okay, model may have decided to respond directly")
	}
}

func TestVertexMultipleToolCalls(t *testing.T) {
	// Check for API key
	apiKey := os.Getenv("VERTEX_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey == "" {
		t.Skip("SKIP: VERTEX_API_KEY or GOOGLE_API_KEY not set")
	}

	ctx := context.Background()

	client, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create Vertex client: %v", err)
	}
	defer client.Close()

	// Define multiple tools
	testTools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_weather",
				Description: "Get the current weather in a given location",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "The city and state, e.g. San Francisco, CA",
						},
						"unit": map[string]any{
							"type":        "string",
							"enum":        []string{"celsius", "fahrenheit"},
							"description": "The unit of temperature to return",
						},
					},
					"required": []string{"location"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_current_time",
				Description: "Get the current time in a specific timezone",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"timezone": map[string]any{
							"type":        "string",
							"description": "The timezone, e.g. America/Los_Angeles",
						},
					},
					"required": []string{"timezone"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "calculate",
				Description: "Perform a mathematical calculation",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"expression": map[string]any{
							"type":        "string",
							"description": "The mathematical expression to evaluate, e.g. '23 * 5'",
						},
					},
					"required": []string{"expression"},
				},
			},
		},
	}

	// Test 1: Simple request that should trigger weather tool call
	t.Run("SingleToolCall", func(t *testing.T) {
		messages := []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, "What's the weather in Tokyo?"),
		}

		response, err := client.GenerateContent(ctx, messages, llms.WithModel("gemini-2.5-flash"), llms.WithTools(testTools))
		if err != nil {
			t.Fatalf("GenerateContent failed: %v", err)
		}

		if len(response.Choices) == 0 {
			t.Fatal("No choices returned")
		}

		choice := response.Choices[0]
		t.Logf("Stop reason: %s", choice.StopReason)
		t.Logf("Content: %s", choice.Content)

		if len(choice.ToolCalls) > 0 {
			t.Logf("Tool calls: %d", len(choice.ToolCalls))
			for i, toolCall := range choice.ToolCalls {
				t.Logf("Tool %d: %s with args: %s", i+1, toolCall.FunctionCall.Name, toolCall.FunctionCall.Arguments)
			}
		}
	})

	// Test 2: Request that might trigger multiple tool calls in sequence
	t.Run("MultipleSequentialToolCalls", func(t *testing.T) {
		messages := []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, "What's the weather in Paris, what time is it in Europe/Paris, and calculate 25 * 8?"),
		}

		toolCallCount := 0
		maxIterations := 3

		for iteration := 0; iteration < maxIterations; iteration++ {
			response, err := client.GenerateContent(ctx, messages, llms.WithModel("gemini-2.5-flash"), llms.WithTools(testTools))
			if err != nil {
				t.Fatalf("GenerateContent (iteration %d) failed: %v", iteration, err)
			}

			if len(response.Choices) == 0 {
				t.Fatalf("No choices in iteration %d", iteration)
			}

			choice := response.Choices[0]
			t.Logf("Iteration %d - Stop reason: %s, Content: %s", iteration, choice.StopReason, choice.Content)

			// Check if model wants to call tools
			if len(choice.ToolCalls) > 0 {
				toolCallCount += len(choice.ToolCalls)
				t.Logf("Iteration %d - Tool calls: %d", iteration, len(choice.ToolCalls))

				for i, toolCall := range choice.ToolCalls {
					t.Logf("  Tool %d: %s with args: %s", i+1, toolCall.FunctionCall.Name, toolCall.FunctionCall.Arguments)

					// Add assistant's tool call to conversation
					messages = append(messages, llms.MessageContent{
						Role: llms.ChatMessageTypeAI,
						Parts: []llms.ContentPart{
							llms.ToolCall{
								FunctionCall: toolCall.FunctionCall,
							},
						},
					})

					// Provide tool response
					var toolResponse string
					switch toolCall.FunctionCall.Name {
					case "get_weather":
						toolResponse = `{"location": "Paris, France", "temperature": "18", "unit": "celsius", "condition": "Partly cloudy"}`
					case "get_current_time":
						toolResponse = `{"time": "15:45:00", "timezone": "Europe/Paris", "date": "2024-01-15"}`
					case "calculate":
						// Parse the expression and calculate
						toolResponse = `{"result": 200}`
					default:
						toolResponse = `{"error": "Unknown tool"}`
					}

					messages = append(messages, llms.MessageContent{
						Role: llms.ChatMessageTypeTool,
						Parts: []llms.ContentPart{
							llms.ToolCallResponse{
								Name:    toolCall.FunctionCall.Name,
								Content: toolResponse,
							},
						},
					})
				}
			} else {
				// No more tool calls, conversation complete
				t.Logf("Conversation complete after %d iterations with %d tool calls total", iteration+1, toolCallCount)
				break
			}
		}

		if toolCallCount == 0 {
			t.Log("No tool calls were generated - model may have responded directly")
		} else {
			t.Logf("Successfully handled %d total tool calls across multiple iterations", toolCallCount)
		}
	})
}

func TestVertexUsageInformation(t *testing.T) {
	// Check for API key
	apiKey := os.Getenv("VERTEX_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey == "" {
		t.Skip("SKIP: VERTEX_API_KEY or GOOGLE_API_KEY not set")
	}

	ctx := context.Background()

	client, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create Vertex client: %v", err)
	}
	defer client.Close()

	// Test 1: Simple completion with usage info
	t.Run("SimpleCompletionUsage", func(t *testing.T) {
		response, err := client.GenerateContent(ctx, []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, "Say hello in one word"),
		}, llms.WithModel("gemini-2.5-flash"))

		if err != nil {
			t.Fatalf("GenerateContent failed: %v", err)
		}

		if len(response.Choices) == 0 {
			t.Fatal("No choices returned")
		}

		choice := response.Choices[0]
		if len(choice.Content) == 0 {
			t.Error("Empty response content")
		}

		// Check usage metadata
		if choice.GenerationInfo == nil {
			t.Fatal("GenerationInfo is nil")
		}

		// Verify usage fields exist (can be int32 or int64 depending on API)
		var inputTokensVal, outputTokensVal, totalTokensVal int64

		inputTokens, ok := choice.GenerationInfo["input_tokens"]
		if !ok {
			t.Error("input_tokens not found in GenerationInfo")
		} else {
			switch v := inputTokens.(type) {
			case int32:
				inputTokensVal = int64(v)
			case int64:
				inputTokensVal = v
			default:
				t.Errorf("input_tokens has unexpected type: %T", inputTokens)
			}
			if inputTokensVal <= 0 {
				t.Errorf("input_tokens should be > 0, got %d", inputTokensVal)
			} else {
				t.Logf("Input tokens: %d", inputTokensVal)
			}
		}

		outputTokens, ok := choice.GenerationInfo["output_tokens"]
		if !ok {
			t.Error("output_tokens not found in GenerationInfo")
		} else {
			switch v := outputTokens.(type) {
			case int32:
				outputTokensVal = int64(v)
			case int64:
				outputTokensVal = v
			default:
				t.Errorf("output_tokens has unexpected type: %T", outputTokens)
			}
			if outputTokensVal <= 0 {
				t.Errorf("output_tokens should be > 0, got %d", outputTokensVal)
			} else {
				t.Logf("Output tokens: %d", outputTokensVal)
			}
		}

		totalTokens, ok := choice.GenerationInfo["total_tokens"]
		if !ok {
			t.Error("total_tokens not found in GenerationInfo")
		} else {
			switch v := totalTokens.(type) {
			case int32:
				totalTokensVal = int64(v)
			case int64:
				totalTokensVal = v
			default:
				t.Errorf("total_tokens has unexpected type: %T", totalTokens)
			}
			if totalTokensVal <= 0 {
				t.Errorf("total_tokens should be > 0, got %d", totalTokensVal)
			} else {
				t.Logf("Total tokens: %d", totalTokensVal)
			}
		}

		// Verify total is reasonable (may include overhead from system prompts, etc.)
		if inputTokensVal > 0 && outputTokensVal > 0 && totalTokensVal > 0 {
			if totalTokensVal < inputTokensVal+outputTokensVal {
				t.Errorf("total_tokens (%d) should be >= input_tokens (%d) + output_tokens (%d)",
					totalTokensVal, inputTokensVal, outputTokensVal)
			}
		}
	})

	// Test 2: Conversation with usage tracking across multiple turns
	t.Run("ConversationUsageTracking", func(t *testing.T) {
		var totalInputTokens int64
		var totalOutputTokens int64

		messages := []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, "What is 2+2?"),
		}

		// First turn
		response1, err := client.GenerateContent(ctx, messages, llms.WithModel("gemini-2.5-flash"))
		if err != nil {
			t.Fatalf("First GenerateContent failed: %v", err)
		}

		if len(response1.Choices) > 0 && response1.Choices[0].GenerationInfo != nil {
			if inputTokens := response1.Choices[0].GenerationInfo["input_tokens"]; inputTokens != nil {
				switch v := inputTokens.(type) {
				case int32:
					totalInputTokens += int64(v)
					t.Logf("Turn 1 input tokens: %d", v)
				case int64:
					totalInputTokens += v
					t.Logf("Turn 1 input tokens: %d", v)
				}
			}
			if outputTokens := response1.Choices[0].GenerationInfo["output_tokens"]; outputTokens != nil {
				switch v := outputTokens.(type) {
				case int32:
					totalOutputTokens += int64(v)
					t.Logf("Turn 1 output tokens: %d", v)
				case int64:
					totalOutputTokens += v
					t.Logf("Turn 1 output tokens: %d", v)
				}
			}
		}

		// Add AI response to conversation
		messages = append(messages, llms.TextParts(llms.ChatMessageTypeAI, response1.Choices[0].Content))
		messages = append(messages, llms.TextParts(llms.ChatMessageTypeHuman, "What about 3+3?"))

		// Second turn
		response2, err := client.GenerateContent(ctx, messages, llms.WithModel("gemini-2.5-flash"))
		if err != nil {
			t.Fatalf("Second GenerateContent failed: %v", err)
		}

		if len(response2.Choices) > 0 && response2.Choices[0].GenerationInfo != nil {
			if inputTokens := response2.Choices[0].GenerationInfo["input_tokens"]; inputTokens != nil {
				switch v := inputTokens.(type) {
				case int32:
					totalInputTokens += int64(v)
					t.Logf("Turn 2 input tokens: %d", v)
				case int64:
					totalInputTokens += v
					t.Logf("Turn 2 input tokens: %d", v)
				}
			}
			if outputTokens := response2.Choices[0].GenerationInfo["output_tokens"]; outputTokens != nil {
				switch v := outputTokens.(type) {
				case int32:
					totalOutputTokens += int64(v)
					t.Logf("Turn 2 output tokens: %d", v)
				case int64:
					totalOutputTokens += v
					t.Logf("Turn 2 output tokens: %d", v)
				}
			}
		}

		t.Logf("Total input tokens across conversation: %d", totalInputTokens)
		t.Logf("Total output tokens across conversation: %d", totalOutputTokens)
	})

	// Test 3: Tool call usage tracking
	t.Run("ToolCallUsageTracking", func(t *testing.T) {
		testTools := []llms.Tool{
			{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "get_weather",
					Description: "Get the current weather in a given location",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"location": map[string]any{
								"type":        "string",
								"description": "The city and state",
							},
						},
						"required": []string{"location"},
					},
				},
			},
		}

		response, err := client.GenerateContent(ctx, []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, "What's the weather in Paris?"),
		}, llms.WithModel("gemini-2.5-flash"), llms.WithTools(testTools))

		if err != nil {
			t.Fatalf("GenerateContent failed: %v", err)
		}

		if len(response.Choices) == 0 {
			t.Fatal("No choices returned")
		}

		choice := response.Choices[0]

		// Check usage metadata exists even with tool calls
		if choice.GenerationInfo == nil {
			t.Fatal("GenerationInfo is nil")
		}

		if inputTokens := choice.GenerationInfo["input_tokens"]; inputTokens != nil {
			switch v := inputTokens.(type) {
			case int32:
				if v > 0 {
					t.Logf("Tool call input tokens: %d", v)
				} else {
					t.Error("input_tokens should be > 0")
				}
			case int64:
				if v > 0 {
					t.Logf("Tool call input tokens: %d", v)
				} else {
					t.Error("input_tokens should be > 0")
				}
			default:
				t.Errorf("input_tokens has unexpected type: %T", inputTokens)
			}
		} else {
			t.Error("input_tokens missing in tool call response")
		}

		if outputTokens := choice.GenerationInfo["output_tokens"]; outputTokens != nil {
			switch v := outputTokens.(type) {
			case int32:
				// Output tokens might be 0 if model requested a tool call
				t.Logf("Tool call output tokens: %d", v)
			case int64:
				t.Logf("Tool call output tokens: %d", v)
			}
		}

		t.Logf("Stop reason: %s", choice.StopReason)
		t.Logf("Has tool calls: %v", len(choice.ToolCalls) > 0)
	})
}
