# OpenAI O1 Example

This example demonstrates how to use the OpenAI O1 model with the LangChain Go library to generate content based on a prompt.

## What This Example Does

- Initializes an OpenAI language model client, specifying the "o1-preview" model
- Sets up a prompt asking for ideas to build a Go app for question answering using a database
- Generates content from the model based on the prompt
- Prints the generated content and some metadata about the generation

## Key Features

- Uses the OpenAI O1 preview model
- Demonstrates setting custom parameters like max tokens and temperature
- Shows how to extract and print generation metadata

## How to Run

1. Ensure you have Go installed and your OpenAI API credentials set up
2. Run the example:

```
go run openai_o1_chat_example.go
```

3. Optionally use the `-model` flag to specify a different model, e.g.:

```
go run openai_o1_chat_example.go -model o1-mini
```

## Learn More

- [LangChain Go Documentation](https://github.com/vendasta/langchaingo)
- [OpenAI API Documentation](https://platform.openai.com/docs/api-reference)