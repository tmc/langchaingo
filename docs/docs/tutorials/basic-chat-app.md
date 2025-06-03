# Building a Simple Chat Application

In this tutorial, you'll build a basic chat application using LangChainGo. This application will demonstrate core concepts like LLM integration, conversation memory, and basic prompt templates.

## Prerequisites

- Go 1.21+
- OpenAI API key
- Basic Go programming knowledge

## Step 1: Project Setup

Create a new Go module for your chat application:

```bash
mkdir langchain-chat-app
cd langchain-chat-app
go mod init chat-app
```

Add LangChainGo as a dependency:

```bash
go get github.com/tmc/langchaingo
```

## Step 2: Basic Chat Implementation

Create `main.go` with a simple chat loop:

```go
package main

import (
    "bufio"
    "context"
    "fmt"
    "log"
    "os"
    "strings"

    "github.com/tmc/langchaingo/llms"
    "github.com/tmc/langchaingo/llms/openai"
    "github.com/tmc/langchaingo/memory"
)

func main() {
    // Initialize LLM
    llm, err := openai.New()
    if err != nil {
        log.Fatal(err)
    }

    // Create conversation memory
    chatMemory := memory.NewConversationBuffer()

    fmt.Println("Chat Application Started! Type 'quit' to exit.")
    
    scanner := bufio.NewScanner(os.Stdin)
    ctx := context.Background()

    for {
        fmt.Print("You: ")
        if !scanner.Scan() {
            break
        }

        input := strings.TrimSpace(scanner.Text())
        if input == "quit" {
            break
        }

        // Get response from LLM
        response, err := llm.GenerateContent(ctx, []llms.MessageContent{
            llms.TextParts(llms.ChatMessageTypeHuman, input),
        })
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            continue
        }

        aiResponse := response.Choices[0].Content
        fmt.Printf("AI: %s\n\n", aiResponse)

        // Store conversation in memory
        chatMemory.ChatHistory.AddUserMessage(ctx, input)
        chatMemory.ChatHistory.AddAIMessage(ctx, aiResponse)
    }
}
```

## Step 3: Add Environment Variable Setup

Set your OpenAI API key:

```bash
export OPENAI_API_KEY="your-api-key-here"
```

## Step 4: Run Your Chat App

```bash
go run main.go
```

You should see:
```
Chat Application Started! Type 'quit' to exit.
You: Hello!
AI: Hello! How can I help you today?

You: quit
```

## Step 5: Enhanced Chat with Memory

Let's improve the chat app to use conversation memory:

```go
package main

import (
    "bufio"
    "context"
    "fmt"
    "log"
    "os"
    "strings"

    "github.com/tmc/langchaingo/chains"
    "github.com/tmc/langchaingo/llms/openai"
    "github.com/tmc/langchaingo/memory"
)

func main() {
    // Initialize LLM
    llm, err := openai.New()
    if err != nil {
        log.Fatal(err)
    }

    // Create conversation memory
    chatMemory := memory.NewConversationBuffer()

    // Create conversation chain
    chain := chains.NewConversationChain(llm, chatMemory)

    fmt.Println("Enhanced Chat Application Started! Type 'quit' to exit.")
    
    scanner := bufio.NewScanner(os.Stdin)
    ctx := context.Background()

    for {
        fmt.Print("You: ")
        if !scanner.Scan() {
            break
        }

        input := strings.TrimSpace(scanner.Text())
        if input == "quit" {
            break
        }

        // Use chain for stateful conversation
        result, err := chains.Run(ctx, chain, input)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            continue
        }

        fmt.Printf("AI: %s\n\n", result)
    }
}
```

## Step 6: Add Custom Prompt Template

Create a more sophisticated chat experience with custom prompts:

```go
package main

import (
    "bufio"
    "context"
    "fmt"
    "log"
    "os"
    "strings"

    "github.com/tmc/langchaingo/chains"
    "github.com/tmc/langchaingo/llms/openai"
    "github.com/tmc/langchaingo/memory"
    "github.com/tmc/langchaingo/prompts"
)

func main() {
    // Initialize LLM
    llm, err := openai.New()
    if err != nil {
        log.Fatal(err)
    }

    // Create conversation memory
    chatMemory := memory.NewConversationBuffer()

    // Create custom prompt template
    template := `You are a helpful AI assistant. You are having a conversation with a human.

Current conversation:
{history}