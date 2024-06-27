# Ollama Functions Example

This example demonstrates how to use function calling capabilities with the Ollama language model using the langchaingo library. It showcases a simple weather information retrieval system.

## What it does

1. Sets up an Ollama language model client with JSON output format.
2. Defines a set of tools (functions) that the model can use:
   - `getCurrentWeather`: Retrieves weather information for a given location.
   - `finalResponse`: Provides the final response to the user query.
3. Sends a user query about the weather in Beijing.
4. Processes the model's responses, which may include function calls.
5. Handles function calls by dispatching them to the appropriate logic.
6. Continues the conversation until a final response is generated or the maximum number of retries is reached.

## Key Features

- **Function Calling**: Demonstrates how to define and use custom functions with Ollama.
- **Conversation Flow**: Manages a multi-turn conversation between the user and the model.
- **Error Handling**: Includes retry logic and validation of function calls.
- **Customization**: Allows specifying a custom Ollama model via the `OLLAMA_TEST_MODEL` environment variable.

## How to Run

1. Ensure you have Ollama set up and running on your system.
2. Run the example with: `go run ollama_functions_example.go`
3. Use the `-v` flag for verbose output: `go run ollama_functions_example.go -v`

## Note

This example is a great starting point for understanding how to implement function calling with Ollama and manage more complex conversations with AI models. It can be extended to include more tools and handle various types of queries beyond weather information.
