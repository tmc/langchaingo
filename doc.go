// Package langchaingo provides a Go implementation of LangChain, a framework for building applications with Large Language Models (LLMs) through composability.
//
// LangchainGo enables developers to create powerful AI-driven applications by providing a unified interface to various LLM providers, vector databases, and other AI services.
// The framework emphasizes modularity, extensibility, and ease of use.
//
// # Core Components
//
// The framework is organized around several key packages:
//
//   - [github.com/tmc/langchaingo/llms]: Interfaces and implementations for various language models (OpenAI, Anthropic, Google, etc.)
//   - [github.com/tmc/langchaingo/chains]: Composable operations that can be linked together to create complex workflows
//   - [github.com/tmc/langchaingo/agents]: Autonomous entities that can use tools to accomplish tasks
//   - [github.com/tmc/langchaingo/embeddings]: Text embedding functionality for semantic search and similarity
//   - [github.com/tmc/langchaingo/vectorstores]: Interfaces to vector databases for storing and querying embeddings
//   - [github.com/tmc/langchaingo/memory]: Conversation history and context management
//   - [github.com/tmc/langchaingo/tools]: External tool integrations (web search, calculators, databases, etc.)
//
// # Quick Start
//
// Basic text generation with OpenAI:
//
//	import (
//		"context"
//		"log"
//
//		"github.com/tmc/langchaingo/llms"
//		"github.com/tmc/langchaingo/llms/openai"
//	)
//
//	ctx := context.Background()
//	llm, err := openai.New()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	completion, err := llm.GenerateContent(ctx, []llms.MessageContent{
//		llms.TextParts(llms.ChatMessageTypeHuman, "What is the capital of France?"),
//	})
//
// Creating embeddings and using vector search:
//
//	import (
//		"github.com/tmc/langchaingo/embeddings"
//		"github.com/tmc/langchaingo/schema"
//		"github.com/tmc/langchaingo/vectorstores/chroma"
//	)
//
//	// Create an embedder
//	embedder, err := embeddings.NewEmbedder(llm)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Create a vector store
//	store, err := chroma.New(
//		chroma.WithChromaURL("http://localhost:8000"),
//		chroma.WithEmbedder(embedder),
//	)
//
//	// Add documents
//	docs := []schema.Document{
//		{PageContent: "Paris is the capital of France"},
//		{PageContent: "London is the capital of England"},
//	}
//	store.AddDocuments(ctx, docs)
//
//	// Search for similar documents
//	results, err := store.SimilaritySearch(ctx, "French capital", 1)
//
// Building a chain for question answering:
//
//	import (
//		"github.com/tmc/langchaingo/chains"
//		"github.com/tmc/langchaingo/vectorstores"
//	)
//
//	chain := chains.NewRetrievalQAFromLLM(
//		llm,
//		vectorstores.ToRetriever(store, 3),
//	)
//
//	answer, err := chains.Run(ctx, chain, "What is the capital of France?")
//
// # Provider Support
//
// LangchainGo supports numerous providers:
//
// LLM Providers:
//   - OpenAI (GPT-3.5, GPT-4, GPT-4 Turbo)
//   - Anthropic (Claude family)
//   - Google AI (Gemini, PaLM)
//   - AWS Bedrock (Claude, Llama, Titan)
//   - Cohere
//   - Mistral AI
//   - Ollama (local models)
//   - Hugging Face Inference
//   - And many more...
//
// Embedding Providers:
//   - OpenAI
//   - Hugging Face
//   - Jina AI
//   - Voyage AI
//   - Google Vertex AI
//   - AWS Bedrock
//
// Vector Stores:
//   - Chroma
//   - Pinecone
//   - Weaviate
//   - Qdrant
//   - PostgreSQL with pgvector
//   - Redis
//   - Milvus
//   - MongoDB Atlas Vector Search
//   - OpenSearch
//   - Azure AI Search
//
// # Agents and Tools
//
// Create agents that can use tools to accomplish complex tasks:
//
//	import (
//		"github.com/tmc/langchaingo/agents"
//		"github.com/tmc/langchaingo/tools/serpapi"
//		"github.com/tmc/langchaingo/tools/calculator"
//	)
//
//	// Create tools
//	searchTool := serpapi.New("your-api-key")
//	calcTool := calculator.New()
//
//	// Create an agent
//	agent := agents.NewMRKLAgent(llm, []tools.Tool{searchTool, calcTool})
//	executor := agents.NewExecutor(agent)
//
//	// Run the agent
//	result, err := executor.Call(ctx, map[string]any{
//		"input": "What's the current population of Tokyo multiplied by 2?",
//	})
//
// # Memory and Conversation
//
// Maintain conversation context across multiple interactions:
//
//	import (
//		"github.com/tmc/langchaingo/memory"
//		"github.com/tmc/langchaingo/chains"
//	)
//
//	// Create memory
//	memory := memory.NewConversationBuffer()
//
//	// Create a conversation chain
//	chain := chains.NewConversation(llm, memory)
//
//	// Have a conversation
//	chains.Run(ctx, chain, "Hello, my name is Alice")
//	chains.Run(ctx, chain, "What's my name?") // Will remember "Alice"
//
// # Advanced Features
//
// Streaming responses:
//
//	stream, err := llm.GenerateContentStream(ctx, messages)
//	for stream.Next() {
//		chunk := stream.Value()
//		fmt.Print(chunk.Choices[0].Content)
//	}
//
// Function calling:
//
//	tools := []llms.Tool{
//		{
//			Type: "function",
//			Function: &llms.FunctionDefinition{
//				Name: "get_weather",
//				Parameters: map[string]any{
//					"type": "object",
//					"properties": map[string]any{
//						"location": map[string]any{"type": "string"},
//					},
//				},
//			},
//		},
//	}
//
//	content, err := llm.GenerateContent(ctx, messages, llms.WithTools(tools))
//
// Multi-modal inputs (text and images):
//
//	parts := []llms.ContentPart{
//		llms.TextPart("What's in this image?"),
//		llms.ImagePart("data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQ..."),
//	}
//	content, err := llm.GenerateContent(ctx, []llms.MessageContent{
//		{Role: llms.ChatMessageTypeHuman, Parts: parts},
//	})
//
// # Configuration and Environment
//
// Most providers require API keys set as environment variables:
//
//	export OPENAI_API_KEY="your-openai-key"
//	export ANTHROPIC_API_KEY="your-anthropic-key"
//	export GOOGLE_API_KEY="your-google-key"
//	export HUGGINGFACEHUB_API_TOKEN="your-hf-token"
//
// # Error Handling
//
// LangchainGo provides standardized error handling:
//
//	import "github.com/tmc/langchaingo/llms"
//
//	if err != nil {
//		if llms.IsAuthenticationError(err) {
//			log.Fatal("Invalid API key")
//		}
//		if llms.IsRateLimitError(err) {
//			log.Println("Rate limited, retrying...")
//		}
//	}
//
// # Testing
//
// LangchainGo includes comprehensive testing utilities including HTTP record/replay for internal tests.
// The httprr package provides deterministic testing of HTTP interactions:
//
//	import "github.com/tmc/langchaingo/internal/httprr"
//
//	func TestMyFunction(t *testing.T) {
//		rr := httprr.OpenForTest(t, http.DefaultTransport)
//		defer rr.Close()
//
//		client := rr.Client()
//		// Use client for HTTP requests - they'll be recorded/replayed for deterministic testing
//	}
//
// # Examples
//
// See the examples/ directory for complete working examples including:
//   - Basic LLM usage
//   - RAG (Retrieval Augmented Generation)
//   - Agent workflows
//   - Vector database integration
//   - Multi-modal applications
//   - Streaming responses
//   - Function calling
//
// # Contributing
//
// LangchainGo welcomes contributions! The project follows Go best practices
// and includes comprehensive testing, linting, and documentation standards.
//
// See CONTRIBUTING.md for detailed guidelines.
package langchaingo
