# Building a Basic Chat Application with LangChainGo

## Overview
Learn how to build a conversational AI chat application using LangChainGo with OpenAI's GPT models.

## Step 1: Basic setup
Create a simple chat loop that takes user input and generates responses:

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
)

func main() {
    // Initialize the OpenAI LLM
    llm, err := openai.New()
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    reader := bufio.NewReader(os.Stdin)

    fmt.Println("Chat started. Type 'quit' to exit.")

    for {
        fmt.Print("You: ")
        input, _ := reader.ReadString('\n')
        input = strings.TrimSpace(input)

        if input == "quit" {
            break
        }

        // Generate response
        response, err := llm.Call(ctx, input,
            llms.WithTemperature(0.8),
        )
        if err != nil {
            log.Printf("Error: %v\n", err)
            continue
        }

        fmt.Printf("AI: %s\n", response)
    }
}
```

## Step 2: Add conversation history
Maintain context across multiple messages:

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
)

func main() {
    llm, err := openai.New()
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    reader := bufio.NewReader(os.Stdin)

    // Store conversation history
    var messages []llms.MessageContent

    fmt.Println("Chat with memory started. Type 'quit' to exit.")

    for {
        fmt.Print("You: ")
        input, _ := reader.ReadString('\n')
        input = strings.TrimSpace(input)

        if input == "quit" {
            break
        }

        // Add user message to history
        messages = append(messages, llms.MessageContent{
            Role: llms.ChatMessageTypeHuman,
            Parts: []llms.ContentPart{
                llms.TextContent{Text: input},
            },
        })

        // Generate response with full conversation history
        response, err := llm.GenerateContent(ctx, messages,
            llms.WithTemperature(0.8),
        )
        if err != nil {
            log.Printf("Error: %v\n", err)
            continue
        }

        // Add AI response to history
        if len(response.Choices) > 0 {
            aiMessage := response.Choices[0].Content
            messages = append(messages, llms.MessageContent{
                Role: llms.ChatMessageTypeAI,
                Parts: []llms.ContentPart{
                    llms.TextContent{Text: aiMessage},
                },
            })
            fmt.Printf("AI: %s\n", aiMessage)
        }
    }
}
```

## Step 3: Use conversation chain
Simplify with LangChainGo's conversation chain:

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
    llm, err := openai.New()
    if err != nil {
        log.Fatal(err)
    }

    // Create conversation chain with memory
    conversationChain := chains.NewConversation(
        llm,
        memory.NewConversationBuffer(),
    )

    ctx := context.Background()
    reader := bufio.NewReader(os.Stdin)

    fmt.Println("Conversation chain started. Type 'quit' to exit.")

    for {
        fmt.Print("You: ")
        input, _ := reader.ReadString('\n')
        input = strings.TrimSpace(input)

        if input == "quit" {
            break
        }

        // Run the chain
        result, err := chains.Run(
            ctx,
            conversationChain,
            input,
        )
        if err != nil {
            log.Printf("Error: %v\n", err)
            continue
        }

        fmt.Printf("AI: %s\n", result)
    }
}
```

## Step 4: Add streaming responses
Show responses as they're generated:

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
)

func main() {
    llm, err := openai.New()
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    reader := bufio.NewReader(os.Stdin)

    var messages []llms.MessageContent

    fmt.Println("Streaming chat started. Type 'quit' to exit.")

    for {
        fmt.Print("You: ")
        input, _ := reader.ReadString('\n')
        input = strings.TrimSpace(input)

        if input == "quit" {
            break
        }

        messages = append(messages, llms.MessageContent{
            Role: llms.ChatMessageTypeHuman,
            Parts: []llms.ContentPart{
                llms.TextContent{Text: input},
            },
        })

        fmt.Print("AI: ")
        var fullResponse string

        // Stream the response
        _, err := llm.GenerateContent(ctx, messages,
            llms.WithTemperature(0.8),
            llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
                content := string(chunk)
                fmt.Print(content)
                fullResponse += content
                return nil
            }),
        )

        if err != nil {
            log.Printf("\nError: %v\n", err)
            continue
        }

        fmt.Println() // New line after response

        // Add complete response to history
        messages = append(messages, llms.MessageContent{
            Role: llms.ChatMessageTypeAI,
            Parts: []llms.ContentPart{
                llms.TextContent{Text: fullResponse},
            },
        })
    }
}
```

## Step 5: Add system prompts
Configure the AI's behavior with system prompts:

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
)

func main() {
    llm, err := openai.New()
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    reader := bufio.NewReader(os.Stdin)

    // Start with a system message
    messages := []llms.MessageContent{
        {
            Role: llms.ChatMessageTypeSystem,
            Parts: []llms.ContentPart{
                llms.TextContent{
                    Text: "You are a helpful AI assistant who speaks like a pirate. Always maintain this character while being helpful and informative.",
                },
            },
        },
    }

    fmt.Println("Pirate AI Chat started. Type 'quit' to exit.")

    for {
        fmt.Print("You: ")
        input, _ := reader.ReadString('\n')
        input = strings.TrimSpace(input)

        if input == "quit" {
            break
        }

        messages = append(messages, llms.MessageContent{
            Role: llms.ChatMessageTypeHuman,
            Parts: []llms.ContentPart{
                llms.TextContent{Text: input},
            },
        })

        response, err := llm.GenerateContent(ctx, messages,
            llms.WithTemperature(0.8),
        )
        if err != nil {
            log.Printf("Error: %v\n", err)
            continue
        }

        if len(response.Choices) > 0 {
            aiMessage := response.Choices[0].Content
            messages = append(messages, llms.MessageContent{
                Role: llms.ChatMessageTypeAI,
                Parts: []llms.ContentPart{
                    llms.TextContent{Text: aiMessage},
                },
            })
            fmt.Printf("AI: %s\n", aiMessage)
        }
    }
}
```

## Step 6: Add custom prompt template
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