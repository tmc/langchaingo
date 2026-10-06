# Kronk examples

These examples run local GGUF models through the
[Kronk](https://github.com/ardanlabs/kronk) SDK and the shared LangChainGo
adapter in [`kronk`](./kronk).

Run commands from this directory. The first run downloads the required native
libraries and model; later runs use the local Kronk cache.

```bash
# Single prompt with streaming output
go run ./prompt-example

# Direct streaming GenerateContent call
go run ./stream-example

# Interactive multi-turn chat (type quit to exit)
go run ./chat-example

# Native LangChainGo tool call and tool-result round trip
go run ./function-example -v

# Grammar-constrained JSON generation
go run ./structured-output-example
```

The MongoDB vector-search example additionally requires MongoDB Atlas or
MongoDB Atlas Local:

```bash
cd mongovector-example
docker compose up -d
go run mongovector_example.go
docker compose down
```

Use an Atlas cluster instead by setting `MONGODB_URI` before running the
MongoDB example.

To compile and test the complete module:

```bash
go test ./...
go vet ./...
go build ./...
```
