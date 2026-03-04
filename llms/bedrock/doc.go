// Package bedrock provides AWS Bedrock integration for LangChainGo.
//
// # Overview
//
// This package implements LLM client for AWS Bedrock, supporting multiple model providers
// including Anthropic Claude, Amazon Nova, Meta Llama, Cohere, AI21, and DeepSeek.
//
// # Architecture
//
// The package consists of three layers:
//
//  1. Public API Layer (bedrockllm.go): Exposes bedrock.LLM and bedrock.New() constructor
//  2. Message Processing Layer: Converts llms.MessageContent to provider-specific formats
//  3. Internal Client Layer (internal/bedrockclient): Handles AWS SDK interactions
//
// Two API modes are supported:
//
//   - Legacy API: Model-specific implementations via InvokeModel/InvokeModelWithResponseStream
//   - Converse API: Unified implementation via Converse/ConverseStream (recommended)
//
// # Basic Usage
//
// Create a Bedrock client:
//
//	import "github.com/vxcontrol/langchaingo/llms/bedrock"
//
//	llm, err := bedrock.New(
//	    bedrock.WithModel(bedrock.ModelAnthropicClaudeSonnet45),
//	    bedrock.WithConverseAPI(),
//	)
//
// Generate content:
//
//	messages := []llms.MessageContent{
//	    llms.TextParts(llms.ChatMessageTypeHuman, "Hello!"),
//	}
//	resp, err := llm.GenerateContent(ctx, messages,
//	    llms.WithMaxTokens(1024),
//	)
//
// # Automatic Prompt Caching
//
// For Claude 4.x models (Opus 4, Sonnet 4, Haiku 4), automatic caching is available:
//
//	llm, err := bedrock.New(
//	    bedrock.WithModel(bedrock.ModelAnthropicClaudeSonnet45),
//	    bedrock.WithConverseAPI(),
//	    bedrock.WithAutomaticCaching(),  // Enable automatic caching
//	)
//
// When enabled, the client automatically:
//   - Detects Claude 4.x models by model ID patterns (anthropic.claude-opus-4, sonnet-4, haiku-4)
//   - Adds cache points to the last assistant or tool message before new user input
//   - Uses ephemeral 5-minute TTL by default
//   - Works transparently without modifying client code
//
// Benefits:
//   - 90% cost reduction on cached input tokens
//   - No manual cache control wrappers needed
//   - Automatic conversation history caching
//
// Manual caching (for fine-grained control):
//
//	messages := []llms.MessageContent{
//	    {
//	        Role: llms.ChatMessageTypeAI,
//	        Parts: []llms.ContentPart{
//	            bedrock.WithCacheControl(
//	                llms.TextPart("long context..."),
//	                bedrock.EphemeralCache(),
//	            ),
//	        },
//	    },
//	}
//
// # Tool Calling
//
// Both APIs support tool calling for compatible models:
//
//	tools := []llms.Tool{
//	    {
//	        Type: "function",
//	        Function: &llms.FunctionDefinition{
//	            Name: "get_weather",
//	            Description: "Get weather for location",
//	            Parameters: map[string]any{...},
//	        },
//	    },
//	}
//
//	resp, err := llm.GenerateContent(ctx, messages,
//	    llms.WithTools(tools),
//	)
//
// # Reasoning Support
//
// Claude 4.x and 3.7 models support reasoning (thinking) mode:
//
//	resp, err := llm.GenerateContent(ctx, messages,
//	    llms.WithReasoning(llms.ReasoningMedium, 2048),
//	)
//
//	// Access reasoning content
//	if resp.Choices[0].Reasoning != nil {
//	    fmt.Println(resp.Choices[0].Reasoning.Content)
//	}
//
// # Streaming
//
// Both APIs support streaming responses:
//
//	streamFunc := func(ctx context.Context, chunk streaming.Chunk) error {
//	    switch chunk.Type {
//	    case streaming.ChunkTypeText:
//	        fmt.Print(chunk.Content)
//	    case streaming.ChunkTypeReasoning:
//	        fmt.Println("Thinking:", chunk.Reasoning.Content)
//	    case streaming.ChunkTypeToolCall:
//	        fmt.Println("Tool:", chunk.ToolCall.Name)
//	    }
//	    return nil
//	}
//
//	resp, err := llm.GenerateContent(ctx, messages,
//	    llms.WithStreamingFunc(streamFunc),
//	)
//
// # Supported Models
//
// See models_list.go for complete list. Major providers:
//
//   - Anthropic: Claude 4.6 (Opus, Sonnet), Claude 4.5, 4.1, 4, 3.7, 3.5
//   - Amazon: Nova 2 Lite, Nova Premier, Nova Pro, Nova Lite, Nova Micro
//   - Meta: Llama 4, Llama 3.3, 3.2, 3.1, 3
//   - Cohere: Command R, Command R+
//   - AI21: Jamba 1.5 Large, Mini
//   - DeepSeek: R1
//   - OpenAI: GPT-OSS-120B, GPT-OSS-20B
//   - Qwen: Qwen3 Next, Qwen3 VL, Qwen3 32B, Qwen3 Coder (30B, Next)
//   - Mistral: Large 3, Magistral Small
//   - Moonshot: Kimi K2.5, Kimi K2 Thinking
//   - Z.AI: GLM-4.7, GLM-4.7-Flash
//
// # Error Handling
//
// Provider-specific errors are mapped to standardized error codes:
//
//	resp, err := llm.GenerateContent(ctx, messages)
//	if err != nil {
//	    if llmErr, ok := err.(*llms.Error); ok {
//	        switch llmErr.Code {
//	        case llms.ErrCodeRateLimit:
//	            // Handle rate limiting
//	        case llms.ErrCodeAuthentication:
//	            // Handle auth errors
//	        }
//	    }
//	}
//
// See errors.go for complete error mapping.
//
// # AWS Configuration
//
// The client uses AWS SDK v2 configuration:
//
//   - Credentials: From environment (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY) or AWS config
//   - Region: From environment (AWS_REGION) or default config
//   - Custom configuration: Use bedrock.WithClient() with pre-configured bedrockruntime.Client
//
// # Performance Considerations
//
//   - Converse API is recommended for new applications (unified, better error handling)
//   - Automatic caching reduces costs by 90% for cached tokens (Claude 4.x only)
//   - Streaming reduces latency for interactive applications
//   - Minimum cache checkpoint: 1024 tokens (Sonnet 4.5), 4096 tokens (Haiku 4.5)
//
// # Maintenance
//
// When adding new models:
//  1. Add model constant to models_list.go with documentation
//  2. Update provider detection in internal/bedrockclient/bedrockclient.go if needed
//  3. Add provider-specific implementation in internal/bedrockclient/provider_*.go
//  4. Update tests in bedrockllm_test.go to include new model
//  5. For caching support, add pattern to supportsCaching() method
//
// When updating API:
//  1. Converse API changes go to internal/bedrockclient/bedrockclient_converse.go
//  2. Legacy API changes go to internal/bedrockclient/provider_*.go
//  3. Message processing changes go to bedrockllm.go (processMessages, processMessagesWithCaching)
//  4. Always maintain backward compatibility
//  5. Add integration tests with httprr recording
//
// # Testing
//
// Tests use httprr for HTTP recording/replay:
//
//   - Integration tests: bedrockllm_test.go (requires AWS credentials)
//   - Unit tests: bedrockllm_unit_test.go (no credentials needed)
//   - Tool calling: bedrock_tool_integration_test.go
//
// Recording new HTTP interactions:
//
//	HTTPRR_RECORD=. go test -v -run TestName ./llms/bedrock/
//
// Debug HTTP interactions:
//
//	HTTPRR_RECORD=. HTTPRR_DEBUG=true go test -v -run TestName ./llms/bedrock/
package bedrock
