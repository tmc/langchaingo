# NVIDIA Chat Completion Example

Welcome to this exciting example of using NVIDIA's AI services with Go! 🚀

This project demonstrates how to leverage NVIDIA's API to perform chat completions using a powerful language model. It's a great way to see how easily you can integrate advanced AI capabilities into your Go applications.

## What This Example Does

1. **Connects to NVIDIA's API**: The code sets up a connection to NVIDIA's API using your API key.

2. **Uses a Specific Model**: It's configured to use the "mistralai/mixtral-8x7b-instruct-v0.1" model, which is a large language model capable of generating human-like text.

3. **Sets Up a Chat Scenario**: The example creates a chat scenario where the AI is instructed to be a Golang expert.

4. **Generates Content**: It then asks the AI to explain why Go is a great fit for AI-based products.

5. **Streams the Response**: The AI's response is streamed in real-time, printing each chunk of the answer as it's generated.

## How to Run

1. Make sure you have Go installed on your system.
2. Set the `NVIDIA_API_KEY` environment variable with your NVIDIA API key.
3. Run the example using `go run nvidia_chat_completion_example.go`.

## Key Features

- **Easy Integration**: Shows how simple it is to integrate NVIDIA's AI services into a Go application.
- **Streaming Responses**: Demonstrates real-time streaming of AI-generated content.
- **Customizable Prompts**: You can easily modify the system message and user prompt to explore different scenarios.

This example is perfect for developers looking to add AI-powered chat capabilities to their Go projects or those curious about how to interact with large language models programmatically.

Have fun exploring the world of AI with Go and NVIDIA! 🎉🤖
