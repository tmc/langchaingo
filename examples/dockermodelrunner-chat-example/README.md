# Docker Model Runner Chat Example

Welcome to this cheerful example of using the LangChain Go library with Docker Model Runner! 🐳

## What This Example Does

This fun little program demonstrates how to use the LangChain Go library to generate text completions using Docker Model Runner's capabilities to run local language models. Here's what it does:

1. 🚀 Sets up an OpenAI-compatible language model client, using the Docker Model Runner that is embedded in the Docker Desktop app.
2. 🧠 Generates a completion for the prompt "Tell me about the Anime series called Attack on Titan".
3. 🎨 Uses a temperature of 0.8 for some creative variety in the output.
4. 🛑 Sets a max tokens of 1024.
5. 📝 Prints the generated completion to the console with a streaming function.

## How It Works

The program uses the `langchaingo` library to interact with the OpenAI-compatible Docker Model Runner's APIs through a socat proxy. It creates a new OpenAI client, sets up a context and the base URL for the Docker Model Runner, obtained from the socat proxy, and then generates a completion based on the given prompt.

The Docker Model Runner is embedded in the Docker Desktop app, and at the moment it's only available for Mac and Windows. The OpenAI-compatible API is available thanks to the Testcontainers Go module for the Docker Model Runner, which exposes a port on the host machine and maps it to the Docker Model Runner's 80 HTTP port. Finally, the LLM client is configured to use the Testcontainers Go module using the `/engines/v1` endpoint of the Docker Model Runner.

> [!NOTE]
> You can find more information about the Docker Model Runner APIs [here](https://docs.docker.com/desktop/features/model-runner/#what-api-endpoints-are-available).

The `GenerateContent` function is used with some interesting options:
- `WithTemperature(0.8)`: This adds a bit of randomness to the output.
- `WithMaxTokens(1024)`: This limits the response to 1024 tokens.

## Running the Example

To run this example, make sure you have Docker Desktop running and in the `v4.40+` version, which is the version that includes the Docker Model Runner. Then pull the `ai/smollm2:360M-Q4_K_M` model, using the following command:

```bash
docker model pull ai/smollm2:360M-Q4_K_M
```

Finally, simply execute the program and watch as it generates a creative response about the Attack on Titan anime!

```bash
go run -v .
```

Have fun exploring the possibilities with LangChain and the Docker Model Runner! 🐳
