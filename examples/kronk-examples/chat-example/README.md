# Kronk Chat Example

An interactive chat example using the kronk abstraction layer to run a local GGUF model via llama.cpp. Unlike the [prompt-example](../prompt-example) which asks a single hard-coded question, this example maintains a conversation history so you can chat back and forth with the model.

## How to Run

The first run downloads and installs the llama.cpp libraries and model (may take several minutes). Subsequent runs load from disk.

```shell
go run chat_example.go
```

### Options

| Flag | Default | Description |
|------|---------|-------------|
| `-model` | `unsloth/Qwen3-0.6B-Q8_0` | GGUF model source (HuggingFace URL or provider/modelID) |

### Example Session

```text
Chat ready. Type your message and press Enter. Type 'quit' to exit.

YOU> What is Go?
MODEL> Go is a statically typed, compiled programming language designed at Google.

YOU> Who created it?
MODEL> Go was created by Robert Griesemer, Rob Pike, and Ken Thompson.

YOU> quit
```

Type `quit` or press `Ctrl+D` to end the conversation.
