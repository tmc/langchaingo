# Tutorial: Basic Chat Application

This is the complete example implementation for the [Basic Chat Application Tutorial](../../docs/docs/tutorials/basic-chat-app.md). It demonstrates how to build a chat application using LangChainGo, progressing from a simple implementation to an advanced one with conversation memory and chains.

## Prerequisites

- Go 1.21 or later
- OpenAI API key

## Setup

Set your OpenAI API key as an environment variable:

```bash
export OPENAI_API_KEY="your-api-key-here"
```

## Running the Examples

This example includes implementations of all steps from the tutorial:

### Step 3: Basic Chat (One-shot)
Sends a single message to the LLM and prints the response.

```bash
go run . step3
# or
go run . basic
```

### Step 4: Interactive Chat
Interactive chat session without memory (each message is independent).

```bash
go run . step4
# or
go run . interactive
```

### Step 5: Chat with Memory
Interactive chat that remembers the conversation history.

```bash
go run . step5
# or
go run . memory
```

### Step 6: Advanced Chat with Chains
Full-featured chat using chains with automatic memory management.

```bash
go run . step6
# or
go run . advanced
# or just run without arguments (default)
go run .
```

## Features by Step

### Step 3 - Basic Features:
- Simple LLM initialization
- Single prompt/response interaction
- Basic error handling

### Step 4 - Interactive Features:
- Interactive input loop
- Graceful exit with 'quit' command
- Error handling for failed requests

### Step 5 - Memory Features:
- Conversation buffer memory
- Context preservation across messages
- Manual prompt construction with history

### Step 6 - Advanced Features:
- Conversation chains with built-in templates
- Automatic memory management
- Structured conversation flow
- Clean separation of concerns

## Code Structure

- `main.go` - Complete implementation with all steps
- `step3_basic.go` - Basic chat implementation (build tag: example)
- `step4_interactive.go` - Interactive chat implementation (build tag: example)
- `step5_memory.go` - Chat with memory implementation (build tag: example)
- `step6_advanced.go` - Advanced chat with chains (build tag: example)

## Tutorial Reference

This example implements the tutorial from:
[Building a Basic Chat Application](../../docs/docs/tutorials/basic-chat-app.md)

## Customization

You can customize the behavior by modifying:

1. **Prompt Template** (Step 6): Edit the template string to change the AI's personality
2. **Model Selection**: Pass options to `openai.New()` to use different models
3. **Memory Type**: Replace `memory.NewConversationBuffer()` with other memory types
4. **Error Handling**: Add retry logic or custom error messages

## Example Session

```
$ go run .
=== Step 6: Advanced Chat with Chains ===
Advanced Chat Application (type 'quit' to exit)
----------------------------------------
You: Hello! Who are you?
AI: Hello! I'm an AI assistant created to help answer questions and have conversations. I'm here to provide helpful, accurate, and friendly responses to whatever you'd like to discuss. How can I assist you today?

You: Can you remember what I just asked?
AI: Yes, I can remember our conversation! You just asked me "Hello! Who are you?" and I introduced myself as an AI assistant who is here to help answer questions and have conversations with you.

You: quit
Goodbye!
```