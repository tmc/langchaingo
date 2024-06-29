# JSON Mode Example

Hello there! 👋 Welcome to this exciting JSON Mode example using the LangChain Go library!

## What does this example do?

This nifty little program demonstrates how to use different language model backends to generate responses in JSON format. It's a great way to see how you can get structured data from various AI models!

Here's what it does in a nutshell:

1. It sets up a command-line flag to choose which AI backend you want to use.
2. It initializes the chosen AI backend (OpenAI, Ollama, Anthropic, or Google AI).
3. It sends a prompt asking "Who was the first man to walk on the moon?" and requests the response in JSON format.
4. It prints out the JSON response from the AI model.

## Cool features:

- 🔀 Supports multiple AI backends (OpenAI, Ollama, Anthropic, Google AI)
- 🌡️ Sets the temperature to 0 for more deterministic responses
- 🧠 Uses JSON mode for structured output
- 🚀 Easy to run and experiment with!

## How to run:

1. Make sure you have the necessary API keys set up for the backend you want to use.
2. Run the program with the desired backend flag:

```
go run json_mode_example.go -backend=openai
```

Replace `openai` with `ollama`, `anthropic`, or `googleai` to try different backends!

## Have fun!

This example is a great starting point for exploring how to get structured data from AI models. Feel free to modify the prompt or try different settings to see how the responses change. Happy coding! 🎉
