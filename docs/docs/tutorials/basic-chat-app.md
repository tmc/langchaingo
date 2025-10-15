# Building a Basic Chat Application

This tutorial will guide you through building a simple chat application using LangChainGo.

## Step 1: Set Up Your Environment

First, create a new Go project:

```bash
mkdir langchain-chat-app
cd langchain-chat-app
go mod init chat-app
```

Install LangChainGo:

```bash
go get github.com/vendasta/langchaingo
```

## Step 2: Configure Your API Key

Set your OpenAI API key as an environment variable:

```bash
export OPENAI_API_KEY="your-api-key-here"
```

## Step 3: Create the Basic Chat Application

Let's start with a simple chat application:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/vendasta/langchaingo/llms"
    "github.com/vendasta/langchaingo/llms/openai"
)

func main() {
    // Initialize the OpenAI LLM
    llm, err := openai.New()
    if err != nil {
        log.Fatal(err)
    }

    // Create a context
    ctx := context.Background()

    // Send a message to the LLM
    response, err := llms.GenerateFromSinglePrompt(
        ctx,
        llm,
        "Hello! How can you help me today?",
    )
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("AI:", response)
}
```

## Step 4: Add Interactive Chat

Now let's make it interactive:

```go
package main

import (
    "bufio"
    "context"
    "fmt"
    "log"
    "os"
    "strings"

    "github.com/vendasta/langchaingo/llms"
    "github.com/vendasta/langchaingo/llms/openai"
)

func main() {
    // Initialize LLM
    llm, err := openai.New()
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    reader := bufio.NewReader(os.Stdin)

    fmt.Println("Chat Application Started (type 'quit' to exit)")
    fmt.Println("----------------------------------------")

    for {
        fmt.Print("You: ")
        input, _ := reader.ReadString('\n')
        input = strings.TrimSpace(input)

        if input == "quit" {
            break
        }

        response, err := llms.GenerateFromSinglePrompt(ctx, llm, input)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            continue
        }

        fmt.Printf("AI: %s\n\n", response)
    }
}
```

## Step 5: Add Conversation Memory

To make the chat remember previous messages:

```go
package main

import (
    "bufio"
    "context"
    "fmt"
    "log"
    "os"
    "strings"

    "github.com/vendasta/langchaingo/llms"
    "github.com/vendasta/langchaingo/llms/openai"
    "github.com/vendasta/langchaingo/memory"
)

func main() {
    // Initialize LLM
    llm, err := openai.New()
    if err != nil {
        log.Fatal(err)
    }

    // Create conversation memory
    chatMemory := memory.NewConversationBuffer()
    ctx := context.Background()
    reader := bufio.NewReader(os.Stdin)

    fmt.Println("Chat with Memory (type 'quit' to exit)")
    fmt.Println("----------------------------------------")

    for {
        fmt.Print("You: ")
        input, _ := reader.ReadString('\n')
        input = strings.TrimSpace(input)

        if input == "quit" {
            break
        }

        // Get conversation history
        messages, _ := chatMemory.ChatHistory.Messages(ctx)
        
        // Format the conversation
        var conversation string
        for _, msg := range messages {
            conversation += msg.GetContent() + "\n"
        }
        
        // Add current input to the conversation
        fullPrompt := conversation + "Human: " + input + "\nAssistant:"

        // Generate response
        response, err := llms.GenerateFromSinglePrompt(ctx, llm, fullPrompt)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            continue
        }

        // Save to memory
        chatMemory.ChatHistory.AddUserMessage(ctx, input)
        chatMemory.ChatHistory.AddAIMessage(ctx, response)

        fmt.Printf("AI: %s\n\n", response)
    }
}
```

## Step 6: Add a Conversation Chain

For a more sophisticated approach using chains that automatically manage memory:

```go
package main

import (
    "bufio"
    "context"
    "fmt"
    "log"
    "os"
    "strings"

    "github.com/vendasta/langchaingo/chains"
    "github.com/vendasta/langchaingo/llms/openai"
    "github.com/vendasta/langchaingo/memory"
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
    // The built-in conversation chain includes a default prompt template
    // and handles memory automatically
    conversationChain := chains.NewConversation(llm, chatMemory)

    ctx := context.Background()
    reader := bufio.NewReader(os.Stdin)

    fmt.Println("Advanced Chat Application (type 'quit' to exit)")
    fmt.Println("----------------------------------------")

    for {
        fmt.Print("You: ")
        input, _ := reader.ReadString('\n')
        input = strings.TrimSpace(input)

        if input == "quit" {
            break
        }

        // Run the chain with the input
        result, err := chains.Run(ctx, conversationChain, input)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            continue
        }

        fmt.Printf("AI: %s\n\n", result)
    }

    fmt.Println("Goodbye!")
}
```

## Step 7: Running Your Application

Save any of the above examples to `main.go` and run:

```bash
go run main.go
```

## Complete Example

You can find the complete working example with all steps in the [tutorial-basic-chat-app](https://github.com/vendasta/langchaingo/tree/main/examples/tutorial-basic-chat-app) directory.

## Conclusion

You've now built a fully functional chat application with LangChainGo! This foundation can be extended with additional features like tool calling, RAG (Retrieval Augmented Generation), and more sophisticated conversation management.