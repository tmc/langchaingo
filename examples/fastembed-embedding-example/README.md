# FastEmbed Embedding Example

This example demonstrates how to use the FastEmbed embedding provider with LangChain Go.

## Features

- **Local Model Execution**: Runs embedding models locally using ONNX
- **Multiple Models**: Supports BGE, sentence-transformers, and other models
- **Automatic Model Download**: Downloads and caches models on first use
- **Optimized Performance**: Uses batch processing and parallel execution
- **Document & Query Embeddings**: Different prefixing strategies for optimal search performance

## Prerequisites

You need to have the ONNX Runtime installed on your system:

### Ubuntu/Debian:
```bash
sudo apt-get install libonnxruntime-dev
```

### macOS:
```bash
brew install onnxruntime
```

### Windows:
Download the ONNX Runtime from the official repository and add it to your PATH.

## Usage

```bash
cd examples/fastembed-embedding-example
go mod tidy
go run fastembed_example.go
```

## Configuration Options

- **Model Selection**: Choose from BGE-Small, BGE-Base, All-MiniLM, etc.
- **Cache Directory**: Specify where to store downloaded models
- **Batch Size**: Configure batch processing for better performance
- **Max Length**: Set maximum token length for inputs
- **Doc Embed Type**: Choose between "default" and "passage" prefixing
- **Execution Providers**: Configure ONNX execution providers (CPU, CUDA, etc.)

## Example Output

```
Embedding documents...
Created embeddings for 4 documents
Document 1: embedding dimension = 384, first 5 values = [0.1234 -0.5678 0.9012 ...]
Document 2: embedding dimension = 384, first 5 values = [0.2345 -0.6789 0.0123 ...]
...

Embedding query: 'What is FastEmbed used for?'
Query embedding dimension: 384, first 5 values = [0.3456 -0.7890 0.1234 ...]

Calculating similarity with documents...
Document 1 similarity: 0.8521
Document 2 similarity: 0.7234
Document 3 similarity: 0.6789
Document 4 similarity: 0.9012
```

## Model Storage

Models are automatically downloaded and cached in the specified cache directory (default: `./local_cache`). Once downloaded, the models run entirely offline.

## Supported Models

- `fastembed.BGESmallENV15` (default) - Fast, 384-dim English model
- `fastembed.BGEBaseENV15` - Better quality, 768-dim English model  
- `fastembed.BGESmallZH` - Chinese language model
- `fastembed.AllMiniLML6V2` - Sentence-transformers model

See the [FastEmbed documentation](https://github.com/qdrant/fastembed/) for the full list of supported models.