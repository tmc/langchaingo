// Package perplexity provides integration with Perplexity AI's API for AI agents.
//
// Perplexity AI functions as an AI-powered search engine that indexes, analyzes,
// and summarizes content from across the internet. This package allows you to
// integrate Perplexity's capabilities into your AI agents to enrich them with
// up-to-date web data.
//
// Example usage:
//
//	llm, err := openai.New(
//		openai.WithModel("gpt-4-mini"),
//		openai.WithCallback(callbacks.LogHandler{}),
//	)
//	if err != nil {
//		return err
//	}
//
//	// Create a new Perplexity instance
//	perpl, err := perplexity.New(
//		perplexity.WithModel(perplexity.ModelSonar),
//		perplexity.WithAPIKey("your-api-key"), // Optional: defaults to PERPLEXITY_API_KEY env var
//	)
//	if err != nil {
//		return err
//	}
//
//	// Add Perplexity as a tool for your agent
//	agentTools := []tools.Tool{
//		perpl,
//	}
//
//	// Create and use the agent
//	toolAgent := agents.NewOneShotAgent(llm,
//		agentTools,
//		agents.WithMaxIterations(2),
//	)
//	executor := agents.NewExecutor(toolAgent)
//
//	answer, err := chains.Run(context.Background(), executor, "your question here")
package perplexity
