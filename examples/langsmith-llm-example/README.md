# LangSmith Tracing Example with OpenAI

Welcome to this example of using LangSmith tracing with OpenAI and Go! 🎉

This project demonstrates how to integrate LangSmith tracing into your LangChain Go applications, allowing you to monitor and debug your LLM interactions.

## What This Example Does

This example showcases several key features:

1. 🤖 Sets up a connection to OpenAI's GPT-4 model
2. 📊 Configures LangSmith tracing for monitoring LLM interactions
3. 🌐 Creates a translation chain that converts text between languages
4. 📝 Logs all langchain interactions using a custom logger

## How It Works

1. Creates an OpenAI client with the GPT-4 model
2. Sets up LangSmith client and tracer with:
   - API key configuration
   - Custom logging
   - Project name tracking
3. Creates a translation chain with:
   - System prompt defining the AI as a translation expert
   - Human prompt template for translation requests
4. Executes the chain with tracing enabled

## Running the Example

To run this example, you'll need:

1. Go installed on your system
2. Environment variables set up:
   - `OPENAI_API_KEY` - Your OpenAI API key
   - `LANGCHAIN_API_KEY` - Your LangSmith API key
   - `LANGCHAIN_PROJECT` - Your LangSmith project name (optional)

You can also provide the LangSmith configuration via flags:
```bash
go run . --langchain-api-key=your_key --langchain-project=your_project
```

## Example Output

The program will output the translation results in JSON format, and all interactions will be traced in your LangSmith dashboard.