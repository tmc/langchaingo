// Package llmtest provides utilities for testing LLM implementations.
//
// Inspired by Go's testing/fstest package, llmtest offers a simple,
// backend-independent way to verify that LLM implementations conform
// to the expected interfaces and behaviors.
//
// # Design Philosophy
//
// Following the principles of testing/fstest:
//   - Minimal API surface - one main function (TestLLM)
//   - Automatic capability discovery - no configuration required
//   - Comprehensive by default - tests all detected capabilities
//   - Interface testing - works with any llms.Model implementation
//   - Simple usage pattern - just pass the model to test
//
// # Usage
//
// Testing an LLM implementation is straightforward:
//
//	func TestMyLLM(t *testing.T) {
//	    llm, err := mylllm.New()
//	    if err != nil {
//	        t.Fatal(err)
//	    }
//	    llmtest.TestLLM(t, llm)
//	}
//
// # Automatic Capability Discovery
//
// The package automatically detects and tests supported capabilities:
//   - Basic operations (Call, GenerateContent)
//   - Streaming (if model implements streaming interface)
//   - Tool/Function calling (probed with test tool)
//   - Reasoning/Thinking mode (if supported)
//   - Token counting (if usage information provided)
//   - Context caching (if implemented)
//
// # Mock Implementation
//
// A MockLLM is provided for testing without making actual API calls:
//
//	mock := &llmtest.MockLLM{
//	    CallFunc: func(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
//	        return "mocked response", nil
//	    },
//	}
//	llmtest.TestLLM(t, mock)
//
// # Parallel Testing
//
// All tests run in parallel by default for better performance:
//   - Core tests (Call, GenerateContent) run concurrently
//   - Capability tests run in parallel when detected
//   - Safe for concurrent execution with independent contexts
//
// # Provider Coverage
//
// The package is used to test all LangChain Go providers:
// anthropic, bedrock, cloudflare, cohere, ernie, fake, googleai,
// huggingface, llamafile, local, maritaca, mistral, ollama, openai,
// watsonx, and more.
package llmtest
