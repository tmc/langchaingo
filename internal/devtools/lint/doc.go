// Package lint provides architectural linting for the LangChain Go codebase.
//
// This linter enforces architectural patterns and best practices specific to
// building robust, maintainable Go libraries, with special focus on AI/ML
// service integration patterns common in LangChain Go.
//
// # Implemented Rules
//
// The following architectural rules are currently implemented:
//
//   - HTTP Client Usage: Enforce httputil.DefaultClient over http.DefaultClient
//   - Provider Isolation: Prevent cross-provider dependencies
//   - Options Pattern: Ensure constructors use functional options
//   - Test HTTP Usage: Enforce httprr usage in tests for deterministic HTTP behavior
//   - Internal Package Access: Validate internal package import rules
//
// # TODO: Additional Architectural Patterns to Implement
//
// ## High Priority - Core Go Library Design Patterns
//
// TODO: Context Propagation Rule
// All public functions that make network calls, file I/O, or could be long-running
// should accept context.Context as the first parameter. This enables cancellation,
// timeouts, and request tracing.
//
//	❌ func (llm *OpenAI) Generate(prompt string) (string, error)
//	✅ func (llm *OpenAI) Generate(ctx context.Context, prompt string) (string, error)
//
// TODO: Error Wrapping Consistency
// All errors should be wrapped with context using fmt.Errorf with %w verb to
// maintain error chains and provide debugging context.
//
//	❌ return fmt.Errorf("failed to call API")
//	✅ return fmt.Errorf("failed to call OpenAI API: %w", err)
//
// TODO: Resource Cleanup Patterns
// Functions that acquire resources (connections, files, etc.) should either use
// defer for cleanup or return a cleanup function.
//
//	❌ func Connect() *Client
//	✅ func Connect() (*Client, func() error, error)
//
// TODO: Interface Segregation
// Interfaces should be small and focused (1-3 methods max). Large interfaces
// should be composed of smaller ones.
//
//	❌ type LLMProvider interface { Generate(...); Embed(...); Complete(...); Stream(...); Validate(...) }
//	✅ type Generator interface { Generate(...) }
//	✅ type Embedder interface { Embed(...) }
//
// TODO: Package Dependency Direction
// Core packages (chains, agents, memory) should not depend on specific provider
// implementations. Dependencies should flow from specific to general.
//
//	❌ import "github.com/vendasta/langchaingo/llms/openai" in chains package
//	✅ import "github.com/vendasta/langchaingo/llms" in chains package
//
// TODO: Global State Avoidance
// Avoid global mutable state. Prefer dependency injection and explicit configuration.
//
//	❌ var DefaultLLM = openai.New()
//	✅ func NewChain(llm llms.Model) *Chain
//
// TODO: Configuration Validation
// All configuration should be validated at creation time, not at usage time,
// to fail fast and provide clear error messages.
//
//	❌ Validate during Generate() call
//	✅ Validate during New() constructor
//
// ## Medium Priority - Performance & Reliability Patterns
//
// TODO: Streaming Interface Consistency
// Operations that return large datasets should offer streaming variants to
// handle memory-constrained environments.
//
//	✅ func (emb *Embedder) EmbedDocuments(ctx context.Context, docs []string) ([][]float32, error)
//	✅ func (emb *Embedder) EmbedDocumentsStream(ctx context.Context, docs []string) (<-chan EmbeddingResult, error)
//
// TODO: Memory Pool Usage
// Use sync.Pool for frequently allocated/deallocated objects to reduce GC pressure.
//
//	✅ var bufferPool = sync.Pool{New: func() interface{} { return &bytes.Buffer{} }}
//
// TODO: Rate Limiting Integration
// Network-bound operations should support rate limiting through context or options.
//
//	✅ func WithRateLimit(rps int) Option
//	✅ Check context.Context for rate limiting signals
//
// TODO: Timeout Patterns
// All network operations should have reasonable default timeouts and respect
// context deadlines.
//
//	✅ ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
//
// TODO: Retry Logic Consistency
// Implement consistent retry patterns with exponential backoff for transient failures.
//
//	✅ func WithRetryPolicy(maxRetries int, backoff BackoffStrategy) Option
//
// TODO: Circuit Breaker Pattern
// Protect against cascading failures in external service calls.
//
//	✅ Implement circuit breaker for provider failures
//
// ## AI/ML Domain-Specific Patterns
//
// TODO: Token Counting Consistency
// All LLM providers should implement consistent token counting for cost estimation
// and request validation.
//
//	✅ type TokenCounter interface { CountTokens(text string) int }
//	✅ All LLM providers implement TokenCounter
//
// TODO: Model Capability Abstraction
// Model-specific features should be behind feature interfaces rather than
// provider-specific methods.
//
//	✅ type ImageCapable interface { GenerateWithImages(...) }
//	✅ type FunctionCallCapable interface { CallFunction(...) }
//
// TODO: Prompt Template Safety
// Implement prompt injection protection and template validation.
//
//	✅ func ValidatePrompt(template string) error
//	✅ func SanitizeInput(userInput string) string
//
// TODO: Response Validation
// LLM responses should be validated for completeness and safety before returning.
//
//	✅ func ValidateResponse(response string) error
//	✅ Check for truncated responses, safety violations
//
// TODO: Embedding Dimension Consistency
// Embedding operations should validate and document vector dimensions.
//
//	✅ type EmbeddingModel interface { Dimensions() int }
//	✅ Validate dimension consistency across operations
//
// TODO: Chain Composability
// Chains should be composable and testable in isolation without external dependencies.
//
//	✅ type Chain interface { Run(ctx context.Context, input ChainInput) (ChainOutput, error) }
//	✅ Chains accept other chains as dependencies
//
// TODO: Agent State Immutability
// Agent state should be immutable or clearly documented as mutable with thread-safety guarantees.
//
//	✅ func (a *Agent) WithMemory(mem Memory) *Agent  // returns new agent
//	❌ func (a *Agent) SetMemory(mem Memory)         // mutates existing agent
//
// TODO: Memory Retention Policies
// Conversation memory should have clear retention and cleanup policies.
//
//	✅ type Memory interface { Prune(maxItems int) error }
//	✅ func WithRetentionPolicy(policy RetentionPolicy) Option
//
// TODO: Tool Input/Output Validation
// Tools should validate their inputs and outputs for type safety and security.
//
//	✅ type Tool interface {
//		Validate(input string) error
//		Execute(ctx context.Context, input string) (string, error)
//	}
//
// TODO: Streaming Response Handling
// LLM streaming should handle partial responses, reconnections, and error recovery gracefully.
//
//	✅ Handle partial JSON responses
//	✅ Implement connection recovery
//	✅ Provide progress callbacks
//
// ## Testing & Quality Patterns
//
// TODO: Test Isolation
// Tests should not depend on external services or shared state between test runs.
//
//	✅ Use httprr for HTTP mocking
//	✅ Use testcontainers for database/service dependencies
//	❌ Direct calls to external APIs in tests
//
// TODO: Provider Compliance Testing
// All providers implementing the same interface should pass the same compliance test suite.
//
//	✅ func TestProviderCompliance(t *testing.T, provider Provider)
//
// TODO: Benchmark Coverage
// Performance-critical paths should have benchmark tests to prevent regressions.
//
//	✅ func BenchmarkEmbedding(b *testing.B)
//	✅ func BenchmarkTokenCounting(b *testing.B)
//
// TODO: Property-Based Testing
// Use property-based testing for validating invariants in chain composition and data processing.
//
//	✅ Test that chain composition is associative
//	✅ Test that embedding operations are deterministic
//
// ## Security & Observability Patterns
//
// TODO: Secret Handling
// API keys and secrets should never be logged or exposed in error messages.
//
//	✅ Redact secrets in logs and errors
//	✅ Use secure environment variable patterns
//
// TODO: Input Sanitization
// All user inputs should be sanitized to prevent injection attacks.
//
//	✅ func SanitizeUserInput(input string) string
//	✅ Validate prompt templates for safety
//
// TODO: Observability Integration
// Operations should emit structured logs, metrics, and traces for monitoring.
//
//	✅ Use structured logging (slog)
//	✅ Emit metrics for token usage, latency, errors
//	✅ Support OpenTelemetry tracing
//
// TODO: TLS Configuration
// All external HTTP calls should use proper TLS configuration.
//
//	✅ Enforce minimum TLS version
//	✅ Certificate validation
//	✅ Configurable TLS settings
//
// ## Documentation & API Design
//
// TODO: Example Test Coverage
// All public APIs should have example tests demonstrating usage.
//
//	✅ func ExampleNewOpenAI()
//	✅ func ExampleChain_Run()
//
// TODO: Godoc Completeness
// All public types, functions, and methods should have comprehensive godoc.
//
//	✅ Document expected behavior
//	✅ Document error conditions
//	✅ Document thread safety guarantees
//
// TODO: Backward Compatibility
// API changes should maintain backward compatibility or follow proper deprecation cycles.
//
//	✅ Use interface evolution patterns
//	✅ Proper deprecation notices
//	✅ Semantic versioning compliance
//
// # Implementation Priority
//
// These rules should be implemented in order of impact on library quality and user experience:
//
// 1. Context propagation and error handling (critical for reliability)
// 2. Test isolation and HTTP mocking (critical for CI/CD stability)
// 3. Provider compliance and interface consistency (critical for maintainability)
// 4. Performance patterns and resource management (important for production use)
// 5. Security and observability patterns (important for enterprise adoption)
// 6. Documentation and API design patterns (important for developer experience)
//
// Each rule should include:
// - AST-based detection logic
// - Clear error messages with examples
// - Automatic fixes where possible
// - Comprehensive test coverage
// - Documentation of the architectural reasoning
package main
