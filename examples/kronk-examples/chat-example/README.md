# Kronk Chat Example

A minimal chat example using the kronk abstraction layer to run a local GGUF model via llama.cpp.

## How to Run

The first run downloads and installs the llama.cpp libraries and model (may take several minutes). Subsequent runs load from disk.

```shell
go run chat_example.go
```

### Options

| Flag | Default | Description |
|------|---------|-------------|
| `-model` | `unsloth/Qwen3-0.6B-Q8_0` | GGUF model source (HuggingFace URL or provider/modelID) |
| `-prompt` | _"In one sentence, what makes Go a good language for concurrent programming?"_ | Question to ask the model |

### Example

```shell
go run chat_example.go -model unsloth/Qwen3-0.6B-Q8_0 -prompt "What is 2+2?"
```
