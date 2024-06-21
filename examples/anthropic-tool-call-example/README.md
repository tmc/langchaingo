# Anthropic Tool Call Example 🌟

Welcome to the Anthropic Tool Call Example! This fun little Go program demonstrates how to use the Anthropic API to create an AI assistant that can answer questions about the weather using function calling. Let's dive in and see what it does!

## What Does This Example Do? 🤔

This example showcases the following cool features:

1. **AI-Powered Weather Assistant**: It creates an AI assistant using Anthropic's Claude model that can answer questions about the weather in different cities.

2. **Function Calling**: The assistant can use a special tool (function) called `getCurrentWeather` to fetch weather information for specific locations.

3. **Conversation Flow**: It demonstrates a back-and-forth conversation between a human and the AI assistant, including multiple queries about weather in different cities.

4. **Tool Execution**: When the AI assistant needs to use the weather tool, the program executes it and provides the results back to the assistant.

## How It Works 🛠️

1. The program starts by creating an Anthropic client using the Claude 3 Haiku model.

2. It then initiates a conversation by asking about the weather in Boston.

3. The AI assistant recognizes the need for weather information and calls the `getCurrentWeather` function.

4. The program executes the function call, fetching mock weather data for Boston.

5. The AI assistant receives the weather data and formulates a response.

6. The conversation continues with additional questions about weather in Chicago, demonstrating the assistant's ability to handle multiple queries and retain context.

## Fun Features 🎉

- **Mock Weather Data**: The example uses a simple map to provide mock weather data for Boston and Chicago. It's not real-time data, but it's perfect for demonstrating how the system works!

- **Flexible Conversations**: You can easily modify the conversation flow by adding more questions or changing the cities mentioned.

- **Tool Definition**: The `availableTools` slice defines the `getCurrentWeather` function, which the AI can use to fetch weather information.

## Try It Out! 🚀

Run the example and watch as the AI assistant cheerfully answers questions about the weather in different cities. Feel free to modify the code to add more cities or even create your own tools for the AI to use!

Happy coding, and may your weather always be sunny! ☀️
