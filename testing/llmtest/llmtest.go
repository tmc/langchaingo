// Package llmtest provides support for testing LLM implementations.
//
// Following the design of testing/fstest, this package provides a simple
// TestLLM function that verifies an LLM implementation behaves correctly.
package llmtest

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/tmc/langchaingo/llms"
)

// TestLLM tests an LLM implementation.
// It performs basic operations and checks that the model behaves correctly.
// It automatically discovers and tests capabilities by probing the model.
//
// If TestLLM finds any misbehaviors, it reports them via t.Error/t.Fatal.
//
// Typical usage inside a test:
//
//	func TestLLM(t *testing.T) {
//	    llm, err := mylllm.New(...)
//	    if err != nil {
//	        t.Fatal(err)
//	    }
//	    llmtest.TestLLM(t, llm)
//	}
func TestLLM(t *testing.T, model llms.Model) {
	t.Helper()
	t.Parallel()

	// Run core tests as subtests - these should always work
	t.Run("Core", func(t *testing.T) {
		t.Parallel()

		t.Run("Call", func(t *testing.T) {
			t.Parallel()
			testCall(t, model)
		})

		t.Run("GenerateContent", func(t *testing.T) {
			t.Parallel()
			testGenerateContent(t, model)
		})
	})

	// Discover and test capabilities
	t.Run("Capabilities", func(t *testing.T) {
		t.Parallel()

		// Test streaming if supported
		if supportsStreaming(model) {
			t.Run("Streaming", func(t *testing.T) {
				t.Parallel()
				testStreaming(t, model)
			})
		}

		// Test tool calls if supported
		if supportsTools(model) {
			t.Run("ToolCalls", func(t *testing.T) {
				t.Parallel()
				testToolCalls(t, model)
			})
		}

		// Test reasoning if supported
		if supportsReasoning(model) {
			t.Run("Reasoning", func(t *testing.T) {
				t.Parallel()
				testReasoning(t, model)
			})
		}

		// Test structured output if supported
		if supportsStructuredOutput(model) {
			t.Run("StructuredOutput", func(t *testing.T) {
				t.Parallel()
				testStructuredOutput(t, model)
			})
		}

		// Test multimodal/vision if supported
		if supportsMultimodal(model) {
			t.Run("Multimodal", func(t *testing.T) {
				t.Parallel()
				testMultimodal(t, model)
			})
		}

		// Test caching by trying it - if it works, great
		t.Run("Caching", func(t *testing.T) {
			t.Parallel()
			testCaching(t, model)
		})

		// Test token counting - always run but don't fail if not supported
		t.Run("TokenCounting", func(t *testing.T) {
			t.Parallel()
			testTokenCounting(t, model)
		})

		// Test error handling - always run
		t.Run("ErrorHandling", func(t *testing.T) {
			t.Parallel()
			testErrorHandling(t, model)
		})
	})
}

// Capability detection functions

// supportsStreaming checks if the model supports streaming
func supportsStreaming(model llms.Model) bool {
	// Check if model implements the streaming interface
	_, ok := model.(interface {
		GenerateContentStream(context.Context, []llms.MessageContent, ...llms.CallOption) (<-chan llms.ContentResponse, error)
	})
	return ok
}

// supportsTools probes if the model supports tool calls
func supportsTools(model llms.Model) bool {
	// Try a simple tool call with a dummy tool
	ctx := context.Background()
	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "test_tool",
				Description: "Test tool",
				Parameters:  map[string]any{"type": "object"},
			},
		},
	}

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("test"),
			},
		},
	}

	// Try with tools - if it doesn't error out, it's supported
	_, err := model.GenerateContent(ctx, messages,
		llms.WithTools(tools),
		llms.WithMaxTokens(1),
	)

	// If we get a specific "tools not supported" error, return false
	// Otherwise assume it's supported (even if other errors occur)
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "not support") {
		return false
	}
	return err == nil || !strings.Contains(strings.ToLower(err.Error()), "tool")
}

// supportsReasoning checks if the model supports reasoning/thinking
func supportsReasoning(model llms.Model) bool {
	// Check if model implements reasoning interface
	if reasoner, ok := model.(interface {
		SupportsReasoning() bool
	}); ok {
		return reasoner.SupportsReasoning()
	}

	// Try using thinking mode and see if it works
	ctx := context.Background()
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("test"),
			},
		},
	}

	// Try with thinking mode
	resp, err := model.GenerateContent(ctx, messages,
		llms.WithMaxTokens(10),
		llms.WithThinkingMode(llms.ThinkingModeLow),
	)

	// Check if thinking tokens are reported
	if err == nil && resp != nil && len(resp.Choices) > 0 {
		if genInfo := resp.Choices[0].GenerationInfo; genInfo != nil {
			if _, ok := genInfo["ThinkingTokens"]; ok {
				return true
			}
		}
	}

	return false
}

// supportsStructuredOutput checks if the model supports structured/JSON output
func supportsStructuredOutput(model llms.Model) bool {
	// Try a simple structured output request
	ctx := context.Background()
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Return a JSON object with a single field 'test' set to true"),
			},
		},
	}

	// Try with JSON mode or schema
	_, err := model.GenerateContent(ctx, messages,
		llms.WithMaxTokens(50),
		llms.WithJSONMode(),
	)

	// If no error or doesn't contain "not support", assume it's supported
	return err == nil || !strings.Contains(strings.ToLower(err.Error()), "not support")
}

// supportsMultimodal checks if the model supports multimodal content (images, etc)
func supportsMultimodal(model llms.Model) bool {
	// Check if we can create image content parts
	// This is a heuristic - we don't actually send an image in the probe
	// Just check if the API would accept it
	ctx := context.Background()

	// Create a minimal base64 encoded 1x1 pixel transparent PNG
	testImage := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.ImageURLPart(fmt.Sprintf("data:image/png;base64,%s", testImage)),
				llms.TextPart("What do you see?"),
			},
		},
	}

	_, err := model.GenerateContent(ctx, messages, llms.WithMaxTokens(10))

	// If no error or doesn't explicitly say images not supported, assume multimodal
	if err != nil && (strings.Contains(strings.ToLower(err.Error()), "image") ||
		strings.Contains(strings.ToLower(err.Error()), "multimodal") ||
		strings.Contains(strings.ToLower(err.Error()), "vision")) {
		return false
	}
	return err == nil
}

// TestLLMWithOptions tests an LLM with specific test options.
func TestLLMWithOptions(t *testing.T, model llms.Model, opts TestOptions, expected ...string) {
	t.Helper()

	// Store options for test functions to use
	testCtx := &testContext{
		model:    model,
		options:  opts,
		expected: expected,
	}

	// Run tests with context
	runTestsWithContext(t, testCtx)
}

// TestOptions configures test execution.
type TestOptions struct {
	// Timeout for each test operation
	Timeout time.Duration

	// Skip specific test categories
	SkipCall            bool
	SkipGenerateContent bool
	SkipStreaming       bool

	// Custom test prompts
	TestPrompt   string
	TestMessages []llms.MessageContent

	// For providers that need special options
	CallOptions []llms.CallOption
}

// Internal test context
type testContext struct {
	model    llms.Model
	options  TestOptions
	expected []string
}

func runTestsWithContext(t *testing.T, ctx *testContext) {
	behaviors := make(map[string]bool)
	for _, exp := range ctx.expected {
		behaviors[exp] = true
	}

	if !ctx.options.SkipCall {
		t.Run("Call", func(t *testing.T) {
			testCallWithContext(t, ctx)
		})
	}

	if !ctx.options.SkipGenerateContent {
		t.Run("GenerateContent", func(t *testing.T) {
			testGenerateContentWithContext(t, ctx)
		})
	}

	if behaviors["supports-streaming"] && !ctx.options.SkipStreaming {
		t.Run("Streaming", func(t *testing.T) {
			testStreamingWithContext(t, ctx)
		})
	}
}

// Core test implementations

func testCall(t *testing.T, model llms.Model) {
	t.Helper()
	ctx := context.Background()

	result, err := llms.GenerateFromSinglePrompt(ctx, model, "Reply with 'OK' and nothing else", llms.WithMaxTokens(10))
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	if result == "" {
		t.Error("Call returned empty result")
	}
}

func testCallWithContext(t *testing.T, tctx *testContext) {
	t.Helper()
	ctx := context.Background()
	if tctx.options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, tctx.options.Timeout)
		defer cancel()
	}

	prompt := "Reply with 'OK' and nothing else"
	if tctx.options.TestPrompt != "" {
		prompt = tctx.options.TestPrompt
	}

	opts := append([]llms.CallOption{llms.WithMaxTokens(10)}, tctx.options.CallOptions...)
	result, err := llms.GenerateFromSinglePrompt(ctx, tctx.model, prompt, opts...)
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	if result == "" {
		t.Error("Call returned empty result")
	}
}

func testGenerateContent(t *testing.T, model llms.Model) {
	t.Helper()
	ctx := context.Background()

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Reply with 'Hello' and nothing else"),
			},
		},
	}

	resp, err := model.GenerateContent(ctx, messages, llms.WithMaxTokens(10))
	if err != nil {
		t.Fatalf("GenerateContent failed: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("No choices in response")
	}

	content := resp.Choices[0].Content
	if content == "" {
		t.Error("GenerateContent returned empty response")
	}
}

func testGenerateContentWithContext(t *testing.T, tctx *testContext) {
	t.Helper()
	ctx := context.Background()
	if tctx.options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, tctx.options.Timeout)
		defer cancel()
	}

	messages := tctx.options.TestMessages
	if len(messages) == 0 {
		messages = []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.TextPart("Reply with 'Hello' and nothing else"),
				},
			},
		}
	}

	opts := append([]llms.CallOption{llms.WithMaxTokens(10)}, tctx.options.CallOptions...)
	resp, err := tctx.model.GenerateContent(ctx, messages, opts...)
	if err != nil {
		t.Fatalf("GenerateContent failed: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("No choices in response")
	}
}

func testStreaming(t *testing.T, model llms.Model) {
	t.Helper()
	ctx := context.Background()

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Count from 1 to 3"),
			},
		},
	}

	// Skip if model doesn't support streaming
	streamer, ok := model.(interface {
		GenerateContentStream(context.Context, []llms.MessageContent, ...llms.CallOption) (<-chan llms.ContentResponse, error)
	})
	if !ok {
		t.Skip("Model doesn't support streaming")
	}

	stream, err := streamer.GenerateContentStream(ctx, messages, llms.WithMaxTokens(50))
	if err != nil {
		t.Fatalf("GenerateContentStream failed: %v", err)
	}

	var chunks []string
	for chunk := range stream {
		if len(chunk.Choices) > 0 {
			chunks = append(chunks, chunk.Choices[0].Content)
		}
	}

	if len(chunks) == 0 {
		t.Error("No chunks received from stream")
	}

	fullContent := strings.Join(chunks, "")
	if fullContent == "" {
		t.Error("Stream produced no content")
	}
}

func testStreamingWithContext(t *testing.T, tctx *testContext) {
	t.Helper()
	ctx := context.Background()
	if tctx.options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, tctx.options.Timeout)
		defer cancel()
	}

	messages := tctx.options.TestMessages
	if len(messages) == 0 {
		messages = []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.TextPart("Count from 1 to 3"),
				},
			},
		}
	}

	// Skip if model doesn't support streaming
	streamer, ok := tctx.model.(interface {
		GenerateContentStream(context.Context, []llms.MessageContent, ...llms.CallOption) (<-chan llms.ContentResponse, error)
	})
	if !ok {
		t.Skip("Model doesn't support streaming")
	}

	opts := append([]llms.CallOption{llms.WithMaxTokens(50)}, tctx.options.CallOptions...)
	stream, err := streamer.GenerateContentStream(ctx, messages, opts...)
	if err != nil {
		t.Fatalf("GenerateContentStream failed: %v", err)
	}

	var chunks []string
	for chunk := range stream {
		if len(chunk.Choices) > 0 {
			chunks = append(chunks, chunk.Choices[0].Content)
		}
	}

	if len(chunks) == 0 {
		t.Error("No chunks received from stream")
	}
}

func testToolCalls(t *testing.T, model llms.Model) {
	t.Helper()
	ctx := context.Background()

	// Define a simple tool
	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_weather",
				Description: "Get the weather for a location",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "The city and country",
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What's the weather in San Francisco?"),
			},
		},
	}

	resp, err := model.GenerateContent(ctx, messages,
		llms.WithTools(tools),
		llms.WithMaxTokens(100),
	)
	if err != nil {
		t.Fatalf("GenerateContent with tools failed: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("No choices in response")
	}

	// Check if tool was called
	choice := resp.Choices[0]
	if len(choice.ToolCalls) == 0 {
		t.Log("No tool calls in response (model may not support tools)")
	} else {
		toolCall := choice.ToolCalls[0]
		if toolCall.FunctionCall.Name != "get_weather" {
			t.Errorf("Expected get_weather tool call, got: %s", toolCall.FunctionCall.Name)
		}
	}
}

func testReasoning(t *testing.T, model llms.Model) {
	t.Helper()

	// Check if model supports reasoning
	if reasoner, ok := model.(interface {
		SupportsReasoning() bool
	}); ok && !reasoner.SupportsReasoning() {
		t.Skip("Model doesn't support reasoning")
	}

	ctx := context.Background()
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What is 25 + 17? Think step by step."),
			},
		},
	}

	// Try with thinking mode if available
	var opts []llms.CallOption
	opts = append(opts, llms.WithMaxTokens(200))

	// Try to use thinking mode (may not be supported)
	if thinkingMode := llms.ThinkingModeMedium; true {
		opts = append(opts, llms.WithThinkingMode(thinkingMode))
	}

	resp, err := model.GenerateContent(ctx, messages, opts...)
	if err != nil {
		t.Fatalf("GenerateContent failed: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("No choices in response")
	}

	content := resp.Choices[0].Content
	if !strings.Contains(content, "42") {
		t.Log("Answer might be incorrect (expected 42)")
	}

	// Check for reasoning tokens if available
	if genInfo := resp.Choices[0].GenerationInfo; genInfo != nil {
		if thinkingTokens, ok := genInfo["ThinkingTokens"].(int); ok {
			t.Logf("Used %d thinking tokens", thinkingTokens)
		}
	}
}

func testCaching(t *testing.T, model llms.Model) {
	t.Helper()
	ctx := context.Background()

	// Long context that benefits from caching
	longContext := strings.Repeat("This is cached context. ", 50)

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextPart(longContext),
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Say 'OK'"),
			},
		},
	}

	// First call (cache miss)
	_, err := model.GenerateContent(ctx, messages, llms.WithMaxTokens(10))
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}

	// Second call (potential cache hit)
	resp2, err := model.GenerateContent(ctx, messages, llms.WithMaxTokens(10))
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}

	// Check if caching info is available
	if genInfo := resp2.Choices[0].GenerationInfo; genInfo != nil {
		if cached, ok := genInfo["CachedTokens"].(int); ok && cached > 0 {
			t.Logf("Cached %d tokens", cached)
		}
	}
}

func testTokenCounting(t *testing.T, model llms.Model) {
	t.Helper()
	ctx := context.Background()

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Count to 5"),
			},
		},
	}

	resp, err := model.GenerateContent(ctx, messages, llms.WithMaxTokens(50))
	if err != nil {
		t.Fatalf("GenerateContent failed: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("No choices in response")
	}

	genInfo := resp.Choices[0].GenerationInfo
	if genInfo == nil {
		t.Skip("No generation info provided")
	}

	// Check for token counts
	var hasTokenInfo bool
	for _, field := range []string{"TotalTokens", "PromptTokens", "CompletionTokens"} {
		if v, ok := genInfo[field].(int); ok && v > 0 {
			hasTokenInfo = true
			t.Logf("%s: %d", field, v)
		}
	}

	if !hasTokenInfo {
		t.Log("No token counting information provided")
	}
}

func testStructuredOutput(t *testing.T, model llms.Model) {
	t.Helper()
	ctx := context.Background()

	// Test JSON mode
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart(`Return a JSON object with two fields: "name" set to "test" and "value" set to 42`),
			},
		},
	}

	resp, err := model.GenerateContent(ctx, messages,
		llms.WithMaxTokens(100),
		llms.WithJSONMode(),
	)
	if err != nil {
		t.Logf("JSON mode failed (may not be supported): %v", err)
		return
	}

	if len(resp.Choices) == 0 {
		t.Fatal("No choices in response")
	}

	content := resp.Choices[0].Content
	if content == "" {
		t.Error("Empty response from structured output")
	}

	// Check if response looks like JSON
	if !strings.Contains(content, "{") || !strings.Contains(content, "}") {
		t.Logf("Response doesn't look like JSON: %s", content)
	}
}

func testMultimodal(t *testing.T, model llms.Model) {
	t.Helper()
	ctx := context.Background()

	// Create a minimal base64 encoded 1x1 pixel red PNG
	testImage := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8DwHwAFBQIAX8jx0gAAAABJRU5ErkJggg=="

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.ImageURLPart(fmt.Sprintf("data:image/png;base64,%s", testImage)),
				llms.TextPart("What color is this image?"),
			},
		},
	}

	resp, err := model.GenerateContent(ctx, messages, llms.WithMaxTokens(50))
	if err != nil {
		t.Logf("Multimodal request failed (may not be supported): %v", err)
		return
	}

	if len(resp.Choices) == 0 {
		t.Fatal("No choices in response")
	}

	content := strings.ToLower(resp.Choices[0].Content)
	if content == "" {
		t.Error("Empty response from multimodal request")
	}
}

func testErrorHandling(t *testing.T, model llms.Model) {
	t.Helper()

	t.Run("InvalidMaxTokens", func(t *testing.T) {
		ctx := context.Background()
		messages := []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.TextPart("test"),
				},
			},
		}

		// Try with invalid max tokens (too large)
		_, err := model.GenerateContent(ctx, messages, llms.WithMaxTokens(999999999))
		if err != nil {
			t.Logf("Invalid max tokens correctly rejected: %v", err)
		}
	})

	t.Run("EmptyPrompt", func(t *testing.T) {
		ctx := context.Background()
		messages := []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.TextPart(""),
				},
			},
		}

		// Try with empty prompt
		resp, err := model.GenerateContent(ctx, messages, llms.WithMaxTokens(10))
		if err != nil {
			t.Logf("Empty prompt handling: %v", err)
		} else if len(resp.Choices) > 0 {
			t.Logf("Empty prompt returned: %s", resp.Choices[0].Content)
		}
	})

	t.Run("CancelledContext", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		messages := []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.TextPart("test"),
				},
			},
		}

		_, err := model.GenerateContent(ctx, messages, llms.WithMaxTokens(10))
		if err == nil {
			t.Error("Expected error from cancelled context")
		} else if !errors.Is(err, context.Canceled) {
			t.Logf("Cancelled context error: %v", err)
		}
	})
}

// ValidateLLM checks if a model satisfies basic requirements without running tests.
// It returns an error describing what's wrong, or nil if the model is valid.
func ValidateLLM(model llms.Model) error {
	if model == nil {
		return errors.New("model is nil")
	}

	// Check if required methods are implemented
	ctx := context.Background()

	// Try a simple call
	_, err := llms.GenerateFromSinglePrompt(ctx, model, "test", llms.WithMaxTokens(1))
	if err != nil {
		return fmt.Errorf("Call method failed: %w", err)
	}

	// Try GenerateContent
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("test"),
			},
		},
	}
	_, err = model.GenerateContent(ctx, messages, llms.WithMaxTokens(1))
	if err != nil {
		return fmt.Errorf("GenerateContent method failed: %w", err)
	}

	return nil
}

// MockLLM provides a simple mock implementation for testing.
type MockLLM struct {
	// Response to return from Call
	CallResponse string
	CallError    error

	// Response to return from GenerateContent
	GenerateResponse *llms.ContentResponse
	GenerateError    error

	// Custom response functions for dynamic behavior
	CallFunc     func(ctx context.Context, prompt string, options ...llms.CallOption) (string, error)
	GenerateFunc func(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error)

	// Feature flags
	SupportsToolCalls        bool
	SupportsReasoningMode    bool
	SupportsStructuredOutput bool
	SupportsMultimodalInput  bool

	// Track calls for verification
	CallCount     int
	GenerateCount int
	StreamCount   int
	LastPrompt    string
	LastMessages  []llms.MessageContent
	LastOptions   []llms.CallOption
}

// Call implements llms.Model
func (m *MockLLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	m.CallCount++
	m.LastPrompt = prompt
	m.LastOptions = options

	// Use custom function if provided
	if m.CallFunc != nil {
		return m.CallFunc(ctx, prompt, options...)
	}

	return m.CallResponse, m.CallError
}

// GenerateContent implements llms.Model
func (m *MockLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	m.GenerateCount++
	m.LastMessages = messages
	m.LastOptions = options

	// Check for cancelled context
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Use custom function if provided
	if m.GenerateFunc != nil {
		return m.GenerateFunc(ctx, messages, options...)
	}

	if m.GenerateResponse != nil {
		return m.GenerateResponse, m.GenerateError
	}

	// Default response with token counting
	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: "mock response",
				GenerationInfo: map[string]any{
					"PromptTokens":     10,
					"CompletionTokens": 5,
					"TotalTokens":      15,
				},
			},
		},
	}, m.GenerateError
}

// GenerateContentStream implements streaming
func (m *MockLLM) GenerateContentStream(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (<-chan llms.ContentResponse, error) {
	m.StreamCount++
	m.LastMessages = messages
	m.LastOptions = options

	// Create a channel and send the mock response
	ch := make(chan llms.ContentResponse, 3)

	// Send the response in chunks
	go func() {
		defer close(ch)

		// Check for cancelled context
		if ctx.Err() != nil {
			return
		}

		// Simulate streaming by sending the response in parts
		if m.GenerateResponse != nil {
			ch <- *m.GenerateResponse
		} else {
			// Default streaming response - send in multiple chunks
			ch <- llms.ContentResponse{
				Choices: []*llms.ContentChoice{
					{
						Content: "mock",
					},
				},
			}
			ch <- llms.ContentResponse{
				Choices: []*llms.ContentChoice{
					{
						Content: " response",
					},
				},
			}
			ch <- llms.ContentResponse{
				Choices: []*llms.ContentChoice{
					{
						Content: "",
						GenerationInfo: map[string]any{
							"PromptTokens":     10,
							"CompletionTokens": 5,
							"TotalTokens":      15,
						},
					},
				},
			}
		}
	}()

	return ch, nil
}

// SupportsReasoning returns whether the mock supports reasoning mode
func (m *MockLLM) SupportsReasoning() bool {
	return m.SupportsReasoningMode
}

// Verify MockLLM implements llms.Model
var _ llms.Model = (*MockLLM)(nil)

// BenchmarkLLM runs performance benchmarks for an LLM implementation.
// It measures response time, throughput, and token efficiency.
func BenchmarkLLM(b *testing.B, model llms.Model) {
	b.Helper()

	b.Run("Call", func(b *testing.B) {
		ctx := context.Background()
		prompt := "Reply with 'OK'"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := llms.GenerateFromSinglePrompt(ctx, model, prompt, llms.WithMaxTokens(10))
			if err != nil {
				b.Fatalf("Call failed: %v", err)
			}
		}
	})

	b.Run("GenerateContent", func(b *testing.B) {
		ctx := context.Background()
		messages := []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.TextPart("Reply with 'Hello'"),
				},
			},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := model.GenerateContent(ctx, messages, llms.WithMaxTokens(10))
			if err != nil {
				b.Fatalf("GenerateContent failed: %v", err)
			}
		}
	})

	if supportsStreaming(model) {
		b.Run("Streaming", func(b *testing.B) {
			ctx := context.Background()
			messages := []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextPart("Count to 5"),
					},
				},
			}

			streamer := model.(interface {
				GenerateContentStream(context.Context, []llms.MessageContent, ...llms.CallOption) (<-chan llms.ContentResponse, error)
			})

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				stream, err := streamer.GenerateContentStream(ctx, messages, llms.WithMaxTokens(50))
				if err != nil {
					b.Fatalf("GenerateContentStream failed: %v", err)
				}

				// Consume the stream
				for range stream {
				}
			}
		})
	}
}

// BenchmarkOptions configures benchmark execution
type BenchmarkOptions struct {
	// Number of iterations
	Iterations int

	// Prompt to use for benchmarking
	Prompt string

	// Max tokens for each request
	MaxTokens int
}

// BenchmarkLLMWithOptions runs benchmarks with custom options
func BenchmarkLLMWithOptions(b *testing.B, model llms.Model, opts BenchmarkOptions) {
	b.Helper()

	if opts.Iterations > 0 {
		b.N = opts.Iterations
	}

	if opts.Prompt == "" {
		opts.Prompt = "Reply with 'OK'"
	}

	if opts.MaxTokens == 0 {
		opts.MaxTokens = 10
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := llms.GenerateFromSinglePrompt(ctx, model, opts.Prompt, llms.WithMaxTokens(opts.MaxTokens))
		if err != nil {
			b.Fatalf("Call failed: %v", err)
		}
	}
}
