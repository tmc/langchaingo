package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/prompts"
)

func main() {
	// This example demonstrates how LangChain Go handles context length limits
	// using the Enhanced Token Buffer Memory
	
	fmt.Println("=== Context Length Management Example ===")
	
	// Check for required environment variable
	if os.Getenv("OPENAI_API_KEY") == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}

	// Create LLM
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}

	// Example 1: Basic Enhanced Token Buffer
	fmt.Println("\n1. Basic Enhanced Token Buffer with Automatic Trimming")
	basicExample(llm)

	// Example 2: Custom Token Counter
	fmt.Println("\n2. Custom Token Counter Example")
	customCounterExample(llm)

	// Example 3: Different Trimming Strategies
	fmt.Println("\n3. Different Trimming Strategies")
	trimmingStrategiesExample(llm)

	// Example 4: Integration with Conversation Chain
	fmt.Println("\n4. Integration with Conversation Chain")
	conversationChainExample(llm)
}

func basicExample(llm llms.Model) {
	ctx := context.Background()

	// Create enhanced memory buffer with a small token limit for demonstration
	memoryBuffer := memory.NewEnhancedTokenBuffer(
		memory.WithTokenLimit(500),                    // Small limit for demo
		memory.WithEncodingModel("gpt-3.5-turbo"),     // Model for accurate token counting
		memory.WithTrimStrategy(memory.TrimOldest),     // Remove oldest messages first
		memory.WithPreservePairs(true),                 // Keep human-AI pairs together
		memory.WithMinMessages(2),                      // Always preserve at least 2 messages
	)

	// Simulate a long conversation
	conversations := []map[string]string{
		{"input": "Hello, I'm interested in learning about machine learning.", "output": "Hello! Machine learning is a fascinating field that involves training algorithms to learn patterns from data. There are three main types: supervised learning (with labeled data), unsupervised learning (finding patterns in unlabeled data), and reinforcement learning (learning through trial and error). Would you like to explore any specific area?"},
		{"input": "Tell me about supervised learning in detail.", "output": "Supervised learning is a type of machine learning where algorithms learn from labeled training data. The goal is to map inputs to correct outputs. Common algorithms include linear regression for continuous outcomes, logistic regression for binary classification, decision trees for interpretable models, random forests for ensemble methods, and neural networks for complex patterns. The process involves training on labeled examples, validating performance, and then making predictions on new, unseen data."},
		{"input": "What about neural networks specifically?", "output": "Neural networks are inspired by biological neurons and consist of interconnected nodes (neurons) organized in layers. They include an input layer, one or more hidden layers, and an output layer. Each connection has a weight that's adjusted during training through backpropagation. Deep learning uses neural networks with many hidden layers. Popular architectures include feedforward networks, convolutional neural networks (CNNs) for images, and recurrent neural networks (RNNs) for sequences."},
		{"input": "How do I get started with machine learning?", "output": "To get started with machine learning: 1) Learn programming (Python is popular), 2) Understand statistics and linear algebra basics, 3) Start with simple algorithms like linear regression, 4) Use libraries like scikit-learn, pandas, and numpy, 5) Practice with datasets from Kaggle or UCI ML Repository, 6) Take online courses (Coursera, edX), 7) Work on personal projects, and 8) Join communities like Stack Overflow and GitHub to learn from others."},
		{"input": "What programming languages are best for ML?", "output": "Python is the most popular language for machine learning due to its extensive libraries (scikit-learn, TensorFlow, PyTorch, pandas, numpy), readable syntax, and strong community support. R is excellent for statistical analysis and data visualization. Java is used in enterprise environments and big data processing. C++ is chosen for performance-critical applications. JavaScript is emerging for web-based ML with TensorFlow.js. For beginners, Python is highly recommended due to its ease of learning and comprehensive ecosystem."},
	}

	// Add conversations and watch how memory manages token limits
	for i, conv := range conversations {
		fmt.Printf("Adding conversation %d...\n", i+1)
		
		// Save the conversation to memory
		err := memoryBuffer.SaveContext(ctx, map[string]any{
			"input":  conv["input"],
			"output": conv["output"],
		})
		if err != nil {
			log.Printf("Error saving context: %v", err)
			continue
		}

		// Check current token count
		if tokenCount, err := memoryBuffer.GetTokenCount(); err == nil {
			fmt.Printf("Current token count: %d\n", tokenCount)
		}

		// Get memory variables to see what's retained
		vars, err := memoryBuffer.LoadMemoryVariables(ctx, map[string]any{})
		if err != nil {
			log.Printf("Error loading memory: %v", err)
			continue
		}

		if history, ok := vars["history"]; ok {
			fmt.Printf("Messages in memory: %d\n", len(memoryBuffer.ChatHistory.Messages()))
		} else {
			fmt.Println("No history in memory variables")
		}
		fmt.Println("---")
	}

	// Show final memory state
	vars, _ := memoryBuffer.LoadMemoryVariables(ctx, map[string]any{})
	if history, ok := vars["history"]; ok {
		fmt.Printf("Final memory content:\n%s\n", history)
	}
}

func customCounterExample(llm llms.Model) {
	// Custom token counter that provides different counting logic
	type CustomTokenCounter struct{}
	
	func (c *CustomTokenCounter) CountTokens(text string) (int, error) {
		// Simple word-based approximation (you might use a more sophisticated method)
		words := len(strings.Fields(text))
		return int(float64(words) * 1.3), nil // Assume 1.3 tokens per word on average
	}
	
	func (c *CustomTokenCounter) CountTokensFromMessages(messages []llms.ChatMessage) (int, error) {
		total := 0
		for _, msg := range messages {
			count, err := c.CountTokens(msg.GetContent())
			if err != nil {
				return 0, err
			}
			total += count + 4 // Add overhead for message formatting
		}
		return total, nil
	}

	customCounter := &CustomTokenCounter{}
	
	memoryBuffer := memory.NewEnhancedTokenBuffer(
		memory.WithTokenLimit(100),                     // Very small limit for demo
		memory.WithTokenCounter(customCounter),         // Use custom counter
		memory.WithTrimStrategy(memory.TrimOldest),
	)

	ctx := context.Background()
	
	// Add some messages
	conversations := []map[string]string{
		{"input": "What is AI?", "output": "AI stands for Artificial Intelligence."},
		{"input": "Tell me more.", "output": "AI is the simulation of human intelligence in machines."},
		{"input": "Give examples.", "output": "Examples include chatbots, recommendation systems, and autonomous vehicles."},
	}

	for i, conv := range conversations {
		memoryBuffer.SaveContext(ctx, map[string]any{
			"input":  conv["input"],
			"output": conv["output"],
		})

		if tokenCount, err := memoryBuffer.GetTokenCount(); err == nil {
			fmt.Printf("Conversation %d - Custom token count: %d\n", i+1, tokenCount)
		}
	}
}

func trimmingStrategiesExample(llm llms.Model) {
	ctx := context.Background()
	
	strategies := []struct {
		name     string
		strategy memory.TrimStrategy
	}{
		{"Trim Oldest", memory.TrimOldest},
		{"Trim Middle", memory.TrimMiddle},
	}

	for _, strat := range strategies {
		fmt.Printf("\nTesting %s strategy:\n", strat.name)
		
		memoryBuffer := memory.NewEnhancedTokenBuffer(
			memory.WithTokenLimit(300),                 // Small limit
			memory.WithEncodingModel("gpt-3.5-turbo"),
			memory.WithTrimStrategy(strat.strategy),
			memory.WithMinMessages(2),                   // Always keep at least 2
		)

		// Add multiple messages
		messages := []string{
			"First message",
			"Second message with more content to increase token count",
			"Third message that continues the conversation flow",
			"Fourth message adding even more context and information",
			"Fifth message that should trigger trimming behavior",
		}

		for i, msg := range messages {
			memoryBuffer.SaveContext(ctx, map[string]any{
				"input":  fmt.Sprintf("User: %s", msg),
				"output": fmt.Sprintf("Assistant: I understand your message: %s", msg),
			})

			tokenCount, _ := memoryBuffer.GetTokenCount()
			messageCount := len(memoryBuffer.ChatHistory.Messages())
			fmt.Printf("Step %d: %d tokens, %d messages\n", i+1, tokenCount, messageCount)
		}
	}
}

func conversationChainExample(llm llms.Model) {
	ctx := context.Background()
	
	// Create enhanced memory
	memoryBuffer := memory.NewEnhancedTokenBuffer(
		memory.WithTokenLimit(800),                     // Reasonable limit
		memory.WithEncodingModel("gpt-3.5-turbo"),
		memory.WithTrimStrategy(memory.TrimOldest),
		memory.WithPreservePairs(true),                 // Keep conversation flow
		memory.WithMinMessages(2),
	)

	// Create a conversation chain with the enhanced memory
	prompt := prompts.NewPromptTemplate(
		`The following is a friendly conversation between a human and an AI. The AI is talkative and provides lots of specific details from its context.

Current conversation:
{history}
Human: {input}
AI:`,
		[]string{"history", "input"},
	)

	chain := chains.NewLLMChain(llm, prompt)
	chain.Memory = memoryBuffer

	// Have a conversation that will exceed token limits
	questions := []string{
		"Tell me about the history of artificial intelligence.",
		"What were the key breakthroughs in the 1980s and 1990s?",
		"How has machine learning evolved in the 21st century?",
		"What are the current challenges in AI research?",
		"What does the future hold for AI development?",
	}

	for i, question := range questions {
		fmt.Printf("Question %d: %s\n", i+1, question)
		
		response, err := chain.Call(ctx, map[string]any{
			"input": question,
		})
		if err != nil {
			log.Printf("Error in conversation: %v", err)
			continue
		}

		if text, ok := response["text"].(string); ok {
			fmt.Printf("Answer: %s\n", text[:min(200, len(text))]+"...")
		}

		// Show memory stats
		if tokenCount, err := memoryBuffer.GetTokenCount(); err == nil {
			messageCount := len(memoryBuffer.ChatHistory.Messages())
			fmt.Printf("Memory: %d tokens, %d messages\n", tokenCount, messageCount)
		}
		fmt.Println("---")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}