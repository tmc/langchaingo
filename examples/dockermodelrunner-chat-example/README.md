# Docker Model Runner Chat Example

Welcome to this cheerful example of using the LangChain Go library with Docker Model Runner! 🐳

## What This Example Does

This fun little program demonstrates how to use the LangChain Go library to generate text completions using Docker Model Runner's caoabilities to run local language models. Here's what it does:

1. 🚀 Sets up an OpenAI-compatible language model client, using the Docker Model Runner that is embedded in the Docker Desktop app.
2. 🧠 Generates a completion for the prompt "What would be a good company name a company that makes colorful clothes for whales?".
3. 🎨 Uses a temperature of 0.8 for some creative variety in the output.
4. 🛑 Sets a max tokens of 1024.
5. 📝 Prints the generated completion to the console with a streaming function.

## How It Works

The program uses the `langchaingo` library to interact with the OpenAI-compatible Docker Model Runner's APIs. It creates a new OpenAI client, sets up a context and the base URL for the Docker Model Runner, and then generates a completion based on the given prompt.

The Docker Model Runner is embedded in the Docker Desktop app, and at the moment it's only available for Mac. The OpenAI-compatible API is available at `http://localhost:12434/engines/v1`. The 12434 TCP port is disabled by default, so you need to enable it in the Docker Desktop settings.

> [!NOTE]
> You can find more information about the Docker Model Runner APIs [here](https://docs.docker.com/desktop/features/model-runner/#what-api-endpoints-are-available).

> [!NOTE]
> If you don't want to expose the 12434 TCP port to the outside world, you can use the `socat` command to forward the port to localhost. You can find an example of how to do this in the [dockermodelrunner-socat-chat-example](../dockermodelrunner-socat-chat-example) directory.

The `GenerateFromSinglePrompt` function is used with some interesting options:
- `WithTemperature(0.8)`: This adds a bit of randomness to the output.
- `WithMaxTokens(1024)`: This limits the response to 1024 tokens.

## Running the Example

To run this example, make sure you have Docker Desktop running and in the `v4.40+` version, which is the version that includes the Docker Model Runner. Then pull the `ai/llama3.2:3b` model, using the following command:

```bash
docker model pull ai/llama3.2:3b
```

Finally, simply execute the program and watch as it generates a creative response about company names!

```bash
# Set the OPENAI_API_KEY environment variable to any value
#so that the OpenAI client doesn't complain.
OPENAI_API_KEY=foo go run -v .
```

Have fun exploring the possibilities with LangChain and the Docker Model Runner! 🐳
