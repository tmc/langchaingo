# MongoDB Atlas Vector Store Example with Kronk

This example demonstrates how to use a **MongoDB Atlas vector store** with **Kronk** as the local embedding model server. It follows the same flow as the OpenAI-based `mongovector-vectorstore-example` in the langchaingo repository, but replaces the OpenAI embedding client with Kronk's shared abstraction, which runs a local [EmbeddingGemma](https://ai.meta.com/models/embeddinggemma/) model.

## What this example does

1. Initializes a local EmbeddingGemma model via Kronk (auto-downloads on first run)
2. Connects to a MongoDB Atlas cluster (or local MongoDB via Docker)
3. Creates a vector search index with dot-product similarity
4. Adds city documents (Tokyo, Paris, London, etc.) with metadata (population, area)
5. Runs three similarity search queries:
   - Find cities in Japan
   - Find South American cities with a score threshold
   - Find large South American cities with metadata filters

## Quick Start (Local MongoDB via Docker)

### 1. Start MongoDB

```bash
docker compose up -d
```

This starts a MongoDB Atlas Local container on `localhost:27017`.

### 2. Run the example

```bash
go run mongovector_example.go
```

That's it! The first run will:
- Download and install llama.cpp libraries
- Download the EmbeddingGemma-300m model (~1.5 GB)
- Create the vector search index in MongoDB
- Add documents and run similarity searches

Subsequent runs are much faster since the model is cached.

Output Example

```bash
go run mongovector_example.go
```

```
Initializing kronk embedding model...
KRONK: 2026-07-16T16:25:48.058664-07:00: download-libraries: check libraries version information: arch[arm64] os[darwin] processor[metal]
KRONK: 2026-07-16T16:25:48.059575-07:00: download-libraries: check llama.cpp installation: arch[arm64] os[darwin] processor[metal] latest[b10025] current[b10025]
KRONK: 2026-07-16T16:25:48.059618-07:00: download-libraries: already installed: latest[b10025] current[b10025]
KRONK: 2026-07-16T16:25:48.08937-07:00: download-model: model file:  embeddinggemma-300m-qat-Q8_0.gguf -> already downloaded:
[{Tokyo map[area:2190 population:38] 0.9331356}]
[{Buenos Aires map[area:203 population:15.5] 0.86999905} {Rio de Janeiro map[area:1200 population:13.7] 0.8501761} {Sao Paulo map[area:1523 population:22.6] 0.84860945} {Santiago map[area:641 population:6.9] 0.8357855}]
[{Buenos Aires map[area:203 population:15.5] 0.86999905} {Sao Paulo map[area:1523 population:22.6] 0.84860945}]
```