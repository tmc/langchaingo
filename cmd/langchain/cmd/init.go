package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init <project-name>",
	Short: "Initialize a new LangChain Go project",
	Long: `Initialize a new LangChain Go project from a template.

Available templates:
  basic-llm      - Simple LLM completion example
  chat-bot       - Interactive chat bot
  rag-system     - Retrieval-augmented generation
  agent          - Tool-calling agent
  vectorstore    - Document embedding and search

Example:
  langchain init my-chat-bot --template chat-bot
  langchain init rag-app --template rag-system`,
	Args: cobra.ExactArgs(1),
	RunE: initProject,
}

var (
	templateName string
	forceCreate  bool
)

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVar(&templateName, "template", "basic-llm", "Template to use")
	initCmd.Flags().BoolVar(&forceCreate, "force", false, "Create directory even if it exists")
}

func initProject(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	// Create project directory
	if err := createProjectDirectory(projectName); err != nil {
		return err
	}

	// Generate project files based on template
	return generateFromTemplate(projectName, templateName)
}

func createProjectDirectory(projectName string) error {
	if _, err := os.Stat(projectName); err == nil {
		if !forceCreate {
			return fmt.Errorf("directory '%s' already exists. Use --force to overwrite", projectName)
		}
	}

	return os.MkdirAll(projectName, 0755)
}

func generateFromTemplate(projectName, templateName string) error {
	projectTemplate, exists := templates[templateName]
	if !exists {
		return fmt.Errorf("template '%s' not found. Available templates: %s",
			templateName, strings.Join(getTemplateNames(), ", "))
	}

	fmt.Printf("Generating project '%s' with template '%s'...\n", projectName, templateName)

	templateData := TemplateData{
		ProjectName: projectName,
		ModuleName:  projectName,
	}

	for filename, content := range projectTemplate.Files {
		filePath := filepath.Join(projectName, filename)

		// Create directory if needed
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Execute template
		tmpl, err := template.New(filename).Parse(content)
		if err != nil {
			return fmt.Errorf("failed to parse template: %w", err)
		}

		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer file.Close()

		if err := tmpl.Execute(file, templateData); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

		fmt.Printf("Created: %s\n", filePath)
	}

	fmt.Printf("\n✅ Project '%s' created successfully!\n", projectName)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Printf("  go mod tidy\n")
	fmt.Printf("  go run main.go\n")

	if projectTemplate.Instructions != "" {
		fmt.Printf("\n%s\n", projectTemplate.Instructions)
	}

	return nil
}

type TemplateData struct {
	ProjectName string
	ModuleName  string
}

type ProjectTemplate struct {
	Files        map[string]string
	Instructions string
}

var templates = map[string]ProjectTemplate{
	"basic-llm": {
		Files: map[string]string{
			"main.go":      basicLLMMain,
			"main_test.go": basicLLMTest,
			"go.mod":       goModTemplate,
			"README.md":    basicLLMReadme,
			".env.example": envExample,
		},
		Instructions: "Set your API keys in .env file (copy from .env.example)",
	},
	"chat-bot": {
		Files: map[string]string{
			"main.go":      chatBotMain,
			"go.mod":       goModTemplate,
			"README.md":    chatBotReadme,
			".env.example": envExample,
		},
		Instructions: "Interactive chat bot. Set API keys and run to start chatting!",
	},
	"rag-system": {
		Files: map[string]string{
			"main.go":              ragMain,
			"go.mod":               goModTemplate,
			"README.md":            ragReadme,
			".env.example":         envExample,
			"documents/sample.txt": sampleDocument,
		},
		Instructions: "Add your documents to the 'documents' directory before running.",
	},
	"gemini-llm": {
		Files: map[string]string{
			"main.go":      geminiLLMMain,
			"main_test.go": geminiLLMTest,
			"go.mod":       goModTemplate,
			"README.md":    geminiLLMReadme,
			".env.example": geminiEnvExample,
		},
		Instructions: "Set your GOOGLE_API_KEY in .env file (copy from .env.example)",
	},
}

func getTemplateNames() []string {
	names := make([]string, 0, len(templates))
	for name := range templates {
		names = append(names, name)
	}
	return names
}

// Template content strings
const goModTemplate = `module {{.ModuleName}}

go 1.23

require github.com/tmc/langchaingo v0.1.12
`

const basicLLMMain = `package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	ctx := context.Background()
	
	// Check for API key
	if os.Getenv("OPENAI_API_KEY") == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable is not set")
		fmt.Println("\nTo set it:")
		fmt.Println("  export OPENAI_API_KEY='your-api-key-here'")
		fmt.Println("\nOr create a .env file (see .env.example)")
		os.Exit(1)
	}
	
	// Initialize the LLM
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}

	// Simple completion
	prompt := "What are the key benefits of using Go for AI applications?"
	
	fmt.Printf("Prompt: %s\n\n", prompt)
	
	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Response: %s\n", completion)
}
`

const chatBotMain = `package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
)

func main() {
	ctx := context.Background()
	
	// Initialize the LLM
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("🤖 LangChain Go Chat Bot")
	fmt.Println("Type 'quit' to exit\n")

	scanner := bufio.NewScanner(os.Stdin)
	var messages []schema.ChatMessage

	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "quit" {
			break
		}

		// Add user message
		messages = append(messages, schema.HumanChatMessage{Content: input})

		// Get response
		response, err := llm.GenerateContent(ctx, messages)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		botResponse := response.Choices[0].Content
		fmt.Printf("Bot: %s\n\n", botResponse)

		// Add bot response to conversation
		messages = append(messages, schema.AIChatMessage{Content: botResponse})
	}

	fmt.Println("Goodbye! 👋")
}
`

const ragMain = `package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/chroma"
)

func main() {
	ctx := context.Background()
	
	// Initialize components
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}

	embedder, err := embeddings.NewEmbedder(llm)
	if err != nil {
		log.Fatal(err)
	}

	// Load documents
	fmt.Println("Loading documents...")
	docs, err := loadDocuments("documents")
	if err != nil {
		log.Fatal(err)
	}

	// Split documents into chunks
	splitter := textsplitter.NewRecursiveCharacter()
	chunks, err := textsplitter.SplitDocuments(splitter, docs)
	if err != nil {
		log.Fatal(err)
	}

	// Create vector store and add documents
	store, err := chroma.New(
		chroma.WithEmbedder(embedder),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Adding %d document chunks to vector store...\n", len(chunks))
	_, err = store.AddDocuments(ctx, chunks)
	if err != nil {
		log.Fatal(err)
	}

	// Query the system
	query := "What is this document about?"
	fmt.Printf("\nQuery: %s\n", query)
	
	docs, err = store.SimilaritySearch(ctx, query, 3)
	if err != nil {
		log.Fatal(err)
	}

	// Generate response based on retrieved documents
	context_docs := ""
	for _, doc := range docs {
		context_docs += doc.PageContent + "\n\n"
	}

	prompt := fmt.Sprintf("Based on the following context, answer the question:\n\nContext:\n%s\n\nQuestion: %s\n\nAnswer:", context_docs, query)

	response, err := llm.GenerateContent(ctx, []schema.ChatMessage{
		schema.HumanChatMessage{Content: prompt},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nAnswer: %s\n", response.Choices[0].Content)
}

func loadDocuments(dir string) ([]schema.Document, error) {
	var docs []schema.Document
	
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && filepath.Ext(path) == ".txt" {
			loader := documentloaders.NewText(path)
			fileDocs, err := loader.Load(context.Background())
			if err != nil {
				return err
			}
			docs = append(docs, fileDocs...)
		}
		
		return nil
	})
	
	return docs, err
}
`

const geminiLLMMain = `package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
)

func main() {
	ctx := context.Background()
	
	// Check for API key
	if os.Getenv("GOOGLE_API_KEY") == "" {
		fmt.Println("Error: GOOGLE_API_KEY environment variable is not set")
		fmt.Println("\nTo set it:")
		fmt.Println("  export GOOGLE_API_KEY='your-api-key-here'")
		fmt.Println("\nOr create a .env file (see .env.example)")
		os.Exit(1)
	}
	
	// Initialize Google AI (Gemini)
	llm, err := googleai.New(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Simple completion
	prompt := "What are the key benefits of using Go for AI applications?"
	
	fmt.Printf("Using Google AI (Gemini)\n")
	fmt.Printf("Prompt: %s\n\n", prompt)
	
	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Response: %s\n", completion)
}
`

const geminiLLMTest = `package main

import (
	"context"
	"testing"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/fake"
)

func TestGeminiPromptGeneration(t *testing.T) {
	ctx := context.Background()
	
	// Use fake LLM for testing with predefined responses
	responses := []string{
		"Go offers excellent performance, strong typing, and built-in concurrency for AI applications.",
		"Another test response",
	}
	llm := fake.NewFakeLLM(responses)

	// Test completion
	prompt := "Test prompt"
	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		t.Fatalf("Failed to generate completion: %v", err)
	}

	if completion == "" {
		t.Error("Expected non-empty completion")
	}
	
	// Verify we get the expected response
	if completion != responses[0] {
		t.Errorf("Expected %q, got %q", responses[0], completion)
	}
}
`

const geminiLLMReadme = `# {{.ProjectName}}

A LangChain Go application using Google AI (Gemini) for language model completions.

## Setup

1. Get a Google AI API key from [Google AI Studio](https://makersuite.google.com/app/apikey)
2. Copy .env.example to .env and add your API key
3. Run go mod tidy to install dependencies
4. Run the application: go run main.go

## Testing

Run tests with: go test -v
`

const geminiEnvExample = `# Google AI (Gemini) API Key
# Get yours at: https://makersuite.google.com/app/apikey
GOOGLE_API_KEY=your-api-key-here
`

const basicLLMTest = `package main

import (
	"context"
	"testing"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/fake"
)

func TestPromptGeneration(t *testing.T) {
	ctx := context.Background()
	
	// Use fake LLM for testing with predefined responses
	responses := []string{
		"Go offers excellent performance, strong typing, and built-in concurrency for AI applications.",
		"Another test response",
	}
	llm := fake.NewFakeLLM(responses)

	// Test completion
	prompt := "Test prompt"
	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		t.Fatalf("Failed to generate completion: %v", err)
	}

	if completion == "" {
		t.Error("Expected non-empty completion")
	}
	
	// Verify we get the expected response
	if completion != responses[0] {
		t.Errorf("Expected %q, got %q", responses[0], completion)
	}
}
`

const basicLLMReadme = `# {{.ProjectName}}

A basic LangChain Go application demonstrating simple LLM completion.

## Setup

1. Copy the environment file:
   ` + "```bash\n   cp .env.example .env\n   ```" + `

2. Add your API keys to the .env file

3. Install dependencies:
   ` + "```bash\n   go mod tidy\n   ```" + `

4. Run the application:
   ` + "```bash\n   go run main.go\n   ```" + `

## Features

- Simple LLM completion using OpenAI
- Clean error handling
- Environment variable configuration
`

const chatBotReadme = `# {{.ProjectName}}

An interactive chat bot built with LangChain Go.

## Setup

1. Copy the environment file:
   ` + "```bash\n   cp .env.example .env\n   ```" + `

2. Add your API keys to the .env file

3. Install dependencies:
   ` + "```bash\n   go mod tidy\n   ```" + `

4. Run the chat bot:
   ` + "```bash\n   go run main.go\n   ```" + `

## Features

- Interactive conversation
- Message history maintenance
- Clean exit with 'quit'
`

const ragReadme = `# {{.ProjectName}}

A Retrieval-Augmented Generation (RAG) system using LangChain Go.

## Setup

1. Copy the environment file:
   ` + "```bash\n   cp .env.example .env\n   ```" + `

2. Add your API keys to the .env file

3. Add documents to the documents/ directory

4. Install dependencies:
   ` + "```bash\n   go mod tidy\n   ```" + `

5. Run the RAG system:
   ` + "```bash\n   go run main.go\n   ```" + `

## Features

- Document loading and chunking
- Vector embeddings with Chroma
- Similarity search
- Context-aware responses
`

const envExample = `# Copy this file to .env and add your API keys

# OpenAI API Key
OPENAI_API_KEY=your_openai_api_key_here

# Other provider keys (uncomment as needed)
# ANTHROPIC_API_KEY=your_anthropic_key_here
# COHERE_API_KEY=your_cohere_key_here
# GOOGLE_API_KEY=your_google_key_here
`

const sampleDocument = `This is a sample document for the RAG system.

LangChain Go is a powerful framework for building AI applications in Go.
It provides abstractions for working with language models, vector stores,
document loaders, and more.

Key features include:
- Multiple LLM provider integrations
- Vector store support for embeddings
- Document processing capabilities  
- Chain and agent frameworks
- Memory management systems

You can replace this document with your own content to test the RAG system.
`
