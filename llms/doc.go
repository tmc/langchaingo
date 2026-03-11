// Package llms provides unified support for interacting with different Language Models (LLMs) from various providers.
// Designed with an extensible architecture, the package facilitates seamless integration of LLMs
// with a focus on modularity, encapsulation, and easy configurability.
//
// The package includes the following subpackages for LLM providers:
// 1. Hugging Face:      llms/huggingface/
// 2. Mistral:           llms/mistral/
// 3. OpenAI:            llms/openai/
// 4. Google AI:         llms/googleai/
// 5. Bedrock:           llms/bedrock/
// 6. Anthropic:         llms/anthropic/
// 7. Ollama:            llms/ollama/
// 8. Cache:             llms/cache/
// 10. Streaming:        llms/streaming/
// 11. Reasoning:        llms/reasoning/
//
// Each subpackage includes provider-specific LLM implementations and helper files for communication
// with supported LLM providers. The internal directories within these subpackages contain provider-specific
// client and API implementations.
//
// The `llms.go` file contains the types and interfaces for interacting with different LLMs.
//
// The `options.go` file provides various options and functions to configure the LLMs.
package llms
