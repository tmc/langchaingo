# Google AI Tool Call Example 🌞🔧

Hello there! Welcome to this exciting example that demonstrates how to use Google AI's language model with tool calling capabilities. Let's dive in and see what this cool code does!

## What's This All About? 🤔

This example shows you how to:

1. Set up a conversation with Google AI's language model
2. Define and use custom tools (functions) that the AI can call
3. Handle the AI's responses and tool calls
4. Provide tool call results back to the AI

## The Weather Inquiry Adventure 🌦️

Here's what happens in this fun little program:

1. We start by asking the AI about the weather in Chicago.
2. We give the AI a special tool called `getCurrentWeather` that it can use to check the weather.
3. The AI recognizes it needs more info and calls the `getCurrentWeather` function.
4. Our code "executes" this function call (it's just pretend data for this example).
5. We send the weather info back to the AI.
6. The AI then gives us a final response about the weather in Chicago!

## Cool Features 🌟

- Uses the `github.com/vendasta/langchaingo` library to interact with Google AI
- Demonstrates how to set up and use custom tools with the AI
- Shows how to maintain a conversation history for context
- Handles tool calls and responses in a neat way

## How to Run 🏃‍♂️

1. Make sure you have Go installed on your machine.
2. Set the `GENAI_API_KEY` environment variable with your Google AI API key.
3. Run the program with `go run googleai-tool-call-example.go`

## What You'll See 👀

When you run the program, you'll see the AI's final response after the tool call. It should include a friendly message about the weather in Chicago, based on the information our mock `getCurrentWeather` function provided!

Have fun exploring this example and learning about AI tool calls! 🎉🤖
