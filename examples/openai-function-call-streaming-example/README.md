# OpenAI Function Call Streaming Example

Welcome to this exciting example of using OpenAI's function calling feature with streaming in Go! 🎉

## What does this example do?

This example demonstrates how to use the LangChain Go library to interact with OpenAI's GPT-4 model, specifically showcasing function calling capabilities and streaming responses. Here's a breakdown of the main features:

1. **OpenAI Model Initialization**: The code sets up a connection to the GPT-4 Turbo model.

2. **Function Definitions**: Three functions are defined as tools that the AI can potentially use:
   - `getCurrentWeather`: Get current weather for a location
   - `getTomorrowWeather`: Get predicted weather for a location
   - `getSuggestedPrompts`: Generate related prompts based on user input

3. **User Query**: The example asks the AI about the weather in Boston.

4. **Streaming Response**: As the AI generates its response, the code streams and prints each chunk of the response in real-time.

5. **Function Call Detection**: If the AI decides to call a function, the code will detect and display this information.

## How it works

1. The program initializes the OpenAI model and sets up the context.
2. It sends a user query about the weather in Boston.
3. As the AI generates its response, each chunk is printed to the console.
4. If the AI decides to call a function (like `getCurrentWeather`), this will be detected and displayed.

## Why is this cool?

- **Real-time Interaction**: You can see the AI's thought process as it generates the response chunk by chunk.
- **Function Calling**: This showcases how AI can be integrated with external tools or data sources.
- **Flexible Tools**: The example defines multiple tools, demonstrating how you can give the AI various capabilities.

Give it a try and watch as the AI decides whether to call a function or provide a direct response about the weather in Boston! ☀️🌦️
