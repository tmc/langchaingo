# DeepSeek Models Example

This example demonstrates how to use different DeepSeek models with the dedicated DeepSeek package in langchaingo.

## Prerequisites

1. Get a DeepSeek API key from [DeepSeek Platform](https://platform.deepseek.com/)
2. Set your API key as an environment variable:
   ```bash
   export DEEPSEEK_API_KEY="your-api-key-here"
   ```

## Available Models

- **deepseek-chat**: General purpose conversational model
- **deepseek-coder**: Specialized for coding and programming tasks
- **deepseek-reasoner**: Advanced reasoning model with step-by-step thinking
- **deepseek-v3**: Latest general purpose model

## Running the Example

```bash
go run deepseek-models-example.go
```

## What This Example Shows

1. **DeepSeek Chat**: General conversation and explanation
2. **DeepSeek Coder**: Code generation and programming help
3. **DeepSeek Reasoner**: Complex reasoning with visible thought process
4. **Streaming**: Real-time reasoning and response streaming

## Key Features

- **Dedicated Package**: No need to use the OpenAI client directly
- **Model Constants**: Pre-defined model names for type safety
- **Reasoning Support**: Access to the model's reasoning process (for reasoner model)
- **Streaming**: Real-time responses with reasoning chunks
- **Easy Configuration**: Simple API key and model selection

## Benefits Over Direct OpenAI Client Usage

1. **Clarity**: Clear intent that you're using DeepSeek
2. **Defaults**: Proper DeepSeek API endpoint pre-configured
3. **Model Names**: Type-safe model constants
4. **Documentation**: DeepSeek-specific docs and examples
5. **Future Features**: DeepSeek-specific features can be added easily