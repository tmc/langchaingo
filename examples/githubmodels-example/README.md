# GitHub Models Example

This example demonstrates how to use the GitHub Models API with LangChain Go library. The example shows both simple prompt completion and chat-based interactions with the GitHub Models API.

## What This Example Does

1. **Sets up a GitHub Models LLM**: 
   - Initializes the GitHub Models LLM client with your GitHub token.
   - Configures the model to use (default is "openai/gpt-4.1").

2. **Performs a Simple Query**:
   - Sends a basic prompt asking about the capital of France.
   - Prints the response from the model.

3. **Performs a Chat Completion**:
   - Creates a conversation with system and user messages.
   - Sends the conversation to the GitHub Models API.
   - Prints the assistant's response.

## Prerequisites

- A GitHub token with appropriate permissions to access the GitHub Models API.
- Go installed on your system.

## Running the Example

1. Set your GitHub token as an environment variable:

```bash
# For Linux/macOS
export GITHUB_TOKEN=your_github_token_here

# For Windows PowerShell
$env:GITHUB_TOKEN = "your_github_token_here"
```

2. Run the example:

```bash
go run githubmodels_example.go
```

## Available Models

GitHub Models supports various models, including:

- openai/gpt-4.1
- anthropic/claude-3-sonnet
- anthropic/claude-3-haiku
- mistral/mistral-large
- mistral/mistral-small

You can specify the model to use with the `WithModel` option when creating the LLM client.

## API Reference

For more details on the GitHub Models API, see the [GitHub Models documentation](https://docs.github.com/en/models).

## Code Implementation

To see the full implementation, check the `githubmodels_example.go` file in this directory.

The example demonstrates:
- How to initialize the GitHub Models client
- How to send a simple text prompt
- How to create and send a multi-message chat conversation
- How to handle the responses

## Further Reading

For more information on how to use the GitHub Models integration with LangChain Go, see the [GitHub Models documentation](https://tmc.github.io/langchaingo/docs/modules/model_io/models/llms/Integrations/githubmodels) in the LangChain Go documentation site.

The GitHub Models API documentation can be found at [GitHub Models documentation](https://docs.github.com/en/github-models).

## Using GitHub Models in Your Own Projects

To use GitHub Models in your own projects, you need to:

1. Import the GitHub Models package:
```go
import "github.com/tmc/langchaingo/llms/githubmodels"
```

2. Create a new GitHub Models LLM client:
```go
llm, err := githubmodels.New(
    githubmodels.WithToken("your_github_token"),
    githubmodels.WithModel("openai/gpt-4.1"),
)
```

3. Use the LLM for text generation or chat completion as shown in the example.
