# llmman Completion Example

This example shows how to use [llmman](https://github.com/llmmanorg/llmman) with LangChain Go.

llmman is a local model runner that serves the Ollama API (alongside OpenAI- and Anthropic-compatible ones) on port 17434. Because the API is the same, the existing `llms/ollama` package works unchanged; the only difference from the Ollama example is `ollama.WithServerURL("http://localhost:17434")`.

## What Does This Example Do?

1. **Points the Ollama client at llmman**: `ollama.New` is configured with the llmman server URL and the `gemma4` model.
2. **Generates a Completion**: Asks "Who was the first man to walk on the moon?"
3. **Streams the Output**: Prints the response as it is generated.

## How to Run

1. Install llmman:
   ```shell
   curl -fsSL https://raw.githubusercontent.com/llmmanorg/llmman/main/install.sh | sh
   ```
2. Pull a model and start the server:
   ```shell
   llmman pull gemma4
   llmman serve
   ```
3. Run the example: `go run llmman_completion_example.go`

llmman listens on `127.0.0.1:17434` by default; set `LLMMAN_HOST` (`[host][:port]`) to change it, and pass the matching URL to `ollama.WithServerURL`.

Models can also be pulled from OCI registries or directly from Hugging Face, e.g. `llmman pull hf.co/unsloth/Qwen3.5-0.8B-GGUF`.
