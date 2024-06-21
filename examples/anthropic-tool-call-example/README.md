# Anthropic Tool Call Example 🛠️🤖

Welcome to the Anthropic Tool Call Example! This fun little project demonstrates how to use the Anthropic API with tool calling capabilities in Go. It's a great way to see how AI models can interact with external tools and functions!

## What's Inside? 📦

This directory contains two main files:

1. `anthropic-tool-call-example.go`: The star of the show! 🌟 This Go file contains a complete example of how to:
   - Set up an Anthropic LLM client
   - Define available tools (in this case, a weather function)
   - Send queries to the model
   - Handle tool calls and responses
   - Maintain a conversation history

2. `go.mod`: The module definition file for this project. It lists the required dependencies, including the awesome `# Anthropic Tool Call Example 🛠️🤖

Welcome to the Anthropic Tool Call Example! This fun little project demonstrates how to use the Anthropic API with tool calling capabilities in Go. It's a great way to see how AI models can interact with external tools and functions!

## What's Inside? 📦

This directory contains a main Go file:

`anthropic-tool-call-example.go`: The star of the show! 🌟 This Go file contains a complete example of how to:
   - Set up an Anthropic LLM client
   - Define available tools (in this case, a weather function)
   - Send queries to the model
   - Handle tool calls and responses
   - Maintain a conversation history

## What Does It Do? 🤔

This example showcases a conversation with an AI model about the weather in different cities. Here's what happens:

1. It sets up an Anthropic LLM client using the Claude 3 Haiku model.
2. Defines a `getCurrentWeather` function as an available tool.
3. Sends an initial query about the weather in Boston.
4. The AI model calls the weather function to get information.
5. The program executes the tool call and sends the result back to the model.
6. The conversation continues with questions about weather in Chicago.
7. The program demonstrates how to maintain context and use tool calls throughout a multi-turn conversation.

## Cool Features 😎

- **Tool Calling**: Shows how to define and use external tools with the AI model.
- **Conversation History**: Demonstrates maintaining context across multiple interactions.
- **Error Handling**: Includes proper error checking and logging.
- **Flexible Weather Info**: Uses a simple map to simulate weather data for different cities.

## How to Run 🏃‍♂️

1. Make sure you have Go installed on your system.
2. Set up your Anthropic API key as an environment variable.
3. Run the example with: `go run anthropic-tool-call-example.go`

Enjoy exploring the world of AI and tool calling with this fun example! 🎉🤖🌦️
