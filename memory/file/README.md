# File Chat Message History

`FileChatMessageHistory` is a file-based chat history implementation, similar to the `FileChatMessageHistory` in the Python version of LangChain. It stores chat messages in a JSON file, supporting persistent storage and retrieval.

## Features

- Saves chat messages to a file
- Thread-safe
- Automatically creates directory structure

## Installation

```bash
go get github.com/tmc/langchaingo
```

## Usage

```go
import (
    "context"
    "fmt"
    "log"
    
    "github.com/tmc/langchaingo/memory/file"
)

func main() {
    // Create a new FileChatMessageHistory instance
    chatHistory, err := file.NewFileChatMessageHistory(
        file.WithFilePath("/path/to/chat_history.json"),
        file.WithCreateDirIfNotExist(true),
    )
    if err != nil {
        log.Fatalf("Failed to create file chat history: %v", err)
    }
    
    ctx := context.Background()
    
    // Add messages
    if err := chatHistory.AddUserMessage(ctx, "Hello"); err != nil {
        log.Fatalf("Failed to add user message: %v", err)
    }
    
    if err := chatHistory.AddAIMessage(ctx, "Hi there! How can I help you?"); err != nil {
        log.Fatalf("Failed to add AI message: %v", err)
    }
    
    // Get all messages
    messages, err := chatHistory.Messages(ctx)
    if err != nil {
        log.Fatalf("Failed to get messages: %v", err)
    }
    
    // Print messages
    for i, msg := range messages {
        fmt.Printf("[%d] %s: %s\n", i+1, msg.GetType(), msg.GetContent())
    }
    
    // Clear history
    if err := chatHistory.Clear(ctx); err != nil {
        log.Fatalf("Failed to clear history: %v", err)
    }
}
```

## Integration with LangChainGo Memory Components

`FileChatMessageHistory` can be integrated with LangChainGo memory components:

```go
import (
    "github.com/tmc/langchaingo/memory"
    "github.com/tmc/langchaingo/memory/file"
)

// Create file chat history
chatHistory, err := file.NewFileChatMessageHistory(
    file.WithFilePath("/path/to/chat_history.json"),
)
if err != nil {
    // Handle error
}

// Use file chat history to create conversation buffer memory
memoryWithFileHistory := memory.NewConversationBufferMemory(memory.WithChatHistory(chatHistory))

// Now you can use this memory component in chains or agents
```

## Configuration Options

`FileChatMessageHistory` supports the following configuration options:

- `WithFilePath(filePath string)`: Sets the path for the chat history file
- `WithCreateDirIfNotExist(create bool)`: Sets whether to create the directory if it doesn't exist (defaults to true)

## File Format

Chat history is saved as a JSON array, with each message containing the following fields:

- `type`: Message type ("human", "ai", "system", "generic")
- `content`: Message content
- `name`: Message name (only used for generic message types)

Example:

```json
[
  {
    "type": "human",
    "content": "Hello"
  },
  {
    "type": "ai",
    "content": "Hi there! How can I help you?"
  },
  {
    "type": "system",
    "content": "This is a system message"
  }
]
```
