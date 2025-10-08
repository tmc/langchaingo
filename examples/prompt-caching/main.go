package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/vendasta/langchaingo/llms"
	"github.com/vendasta/langchaingo/llms/anthropic"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== Prompt Caching Demo ===")
	fmt.Println("Demonstrating cost savings with Anthropic's prompt caching")
	fmt.Println()

	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey == "" {
		fmt.Println("Error: ANTHROPIC_API_KEY environment variable not set")
		fmt.Println("\nTo run this demo:")
		fmt.Println("  export ANTHROPIC_API_KEY='your-api-key'")
		fmt.Println("  go run main.go")
		return
	}

	// Initialize Anthropic client  
	llm, err := anthropic.New(anthropic.WithModel("claude-3-5-sonnet-20241022"))
	if err != nil {
		fmt.Printf("Error initializing Anthropic: %v\n", err)
		return
	}

	// Large context that will be cached (minimum 1024 tokens for caching)
	largeContext := `You are an expert software architect with deep knowledge of system design patterns.

## System Design Patterns Reference

### 1. Microservices Architecture
- Service decomposition based on business capabilities
- Independent deployment and scaling
- Service discovery and registration
- API Gateway pattern for unified entry point
- Circuit breaker for fault tolerance
- Event-driven communication via message queues
- Database per service for data isolation
- Saga pattern for distributed transactions

### 2. Event-Driven Architecture
- Event sourcing for audit trails
- CQRS (Command Query Responsibility Segregation)
- Event streaming with Apache Kafka or similar
- Event store for event persistence
- Projections for read models
- Eventual consistency considerations

### 3. Caching Strategies
- Cache-aside (lazy loading)
- Write-through caching
- Write-behind caching
- Distributed caching with Redis/Memcached
- CDN for static content
- Application-level caching
- Database query result caching

### 4. Load Balancing
- Round-robin distribution
- Least connections algorithm
- IP hash for session affinity
- Weighted distribution
- Health checks and failover
- Geographic load balancing

### 5. Data Storage Patterns
- SQL vs NoSQL selection criteria
- Sharding for horizontal scaling
- Read replicas for read-heavy workloads
- Master-slave replication
- Multi-master replication
- Time-series databases for metrics
- Object storage for large files

### 6. Security Patterns
- Authentication vs Authorization
- OAuth 2.0 and OpenID Connect
- JWT tokens for stateless auth
- API key management
- Rate limiting and throttling
- WAF (Web Application Firewall)
- Encryption at rest and in transit

### 7. Monitoring and Observability
- Distributed tracing (OpenTelemetry)
- Centralized logging (ELK stack)
- Metrics collection (Prometheus)
- Alerting and incident management
- Performance monitoring
- Error tracking and reporting

### 8. Deployment Patterns
- Blue-green deployments
- Canary releases
- Feature flags
- Rolling updates
- Immutable infrastructure
- Infrastructure as Code (Terraform)
- Container orchestration (Kubernetes)

When answering questions, consider these patterns and provide specific, actionable recommendations.`

	fmt.Println("Context Size:", len(largeContext), "characters")
	fmt.Println("(Approximately", len(strings.Fields(largeContext)), "words)")
	fmt.Println()

	// Series of questions using the same cached context
	questions := []string{
		"What caching strategy would you recommend for a read-heavy e-commerce product catalog?",
		"How should I implement authentication for a microservices architecture?",
		"What's the best approach for handling distributed transactions across services?",
		"How can I ensure high availability for a global application?",
	}

	var totalCachedTokens, totalSavedTokens int

	for i, question := range questions {
		fmt.Printf("%s\n", strings.Repeat("=", 60))
		fmt.Printf("Request %d: %s\n", i+1, question)
		fmt.Printf("%s\n", strings.Repeat("-", 60))

		messages := []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeSystem,
				Parts: []llms.ContentPart{
					// Mark the large context for caching
					llms.WithCacheControl(llms.TextPart(largeContext), anthropic.EphemeralCache()),
				},
			},
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.TextPart(question),
				},
			},
		}

		resp, err := llm.GenerateContent(ctx, messages,
			llms.WithMaxTokens(200),
			anthropic.WithPromptCaching(), // Enable prompt caching beta feature
		)

		if err != nil {
			fmt.Printf("Error: %v\n\n", err)
			continue
		}

		// Display response (truncated)
		content := resp.Choices[0].Content
		if len(content) > 250 {
			content = content[:250] + "..."
		}
		fmt.Printf("\nResponse: %s\n", content)

		// Display caching metrics
		if genInfo := resp.Choices[0].GenerationInfo; genInfo != nil {
			// Extract cache token information manually from generation info
			usage := extractCacheUsage(genInfo)

			fmt.Printf("\nToken Usage:\n")
			fmt.Printf("  Input Tokens:  %d\n", usage.InputTokens)
			fmt.Printf("  Output Tokens: %d\n", usage.OutputTokens)

			if usage.CacheCreationInputTokens > 0 {
				fmt.Printf("  Cache Creation: %d tokens (25%% premium for initial caching)\n",
					usage.CacheCreationInputTokens)
			}

			if usage.CachedInputTokens > 0 {
				fmt.Printf("  Cached Tokens: %d (%.0f%% discount applied) ✓\n",
					usage.CachedInputTokens, usage.CacheDiscountPercent)

				savedTokens := int(float64(usage.CachedInputTokens) * (usage.CacheDiscountPercent / 100.0))
				totalCachedTokens += usage.CachedInputTokens
				totalSavedTokens += savedTokens

				fmt.Printf("  Token Savings: %d tokens\n", savedTokens)
			} else if i > 0 {
				fmt.Println("  Cache Status: MISS (context not cached)")
			} else {
				fmt.Println("  Cache Status: CREATING (first request)")
			}
		}
		fmt.Println()
	}

	// Display summary
	fmt.Printf("%s\n", strings.Repeat("=", 60))
	fmt.Println("CACHING SUMMARY")
	fmt.Printf("%s\n", strings.Repeat("=", 60))
	fmt.Printf("Total Requests:       %d\n", len(questions))
	fmt.Printf("Total Cached Tokens:  %d\n", totalCachedTokens)
	fmt.Printf("Total Token Savings:  %d\n", totalSavedTokens)
	if totalCachedTokens > 0 {
		fmt.Printf("Average Discount:     90%%\n")
		fmt.Printf("\nCost Reduction:       ~%.0f%% on input tokens after first request\n",
			90.0) // Anthropic provides 90% discount on cached tokens
	}

	fmt.Println("\nKey Benefits:")
	fmt.Println("✓ Significant cost reduction for repeated context")
	fmt.Println("✓ Faster response times (pre-processed context)")
	fmt.Println("✓ Consistent context across multiple queries")
	fmt.Println("✓ Ideal for chatbots, Q&A systems, and analysis tools")
}

// CacheUsage represents token usage with caching information
type CacheUsage struct {
	InputTokens              int
	OutputTokens             int
	CacheCreationInputTokens int
	CachedInputTokens        int
	CacheDiscountPercent     float64
}

// extractCacheUsage extracts cache-related token information from generation info
func extractCacheUsage(genInfo map[string]any) *CacheUsage {
	usage := &CacheUsage{}

	// Standard token fields
	if v, ok := genInfo["InputTokens"].(int); ok {
		usage.InputTokens = v
	}
	if v, ok := genInfo["OutputTokens"].(int); ok {
		usage.OutputTokens = v
	}

	// Cache-specific fields (Anthropic)
	if v, ok := genInfo["CacheCreationInputTokens"].(int); ok {
		usage.CacheCreationInputTokens = v
	}
	if v, ok := genInfo["CacheReadInputTokens"].(int); ok {
		usage.CachedInputTokens = v
	}

	// Calculate discount (Anthropic provides 90% discount on cached tokens)
	if usage.CachedInputTokens > 0 {
		usage.CacheDiscountPercent = 90.0
	}

	return usage
}
