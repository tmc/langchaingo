// Package llms provides unified support for interacting with different Language Models (LLMs) from various providers.
// Designed with an extensible architecture, the package facilitates seamless integration of LLMs
// with a focus on modularity, encapsulation, and easy configurability.
//
// The package includes the following subpackages for LLM providers:
// 1. Hugging Face:      llms/huggingface/
// 2. Local LLM:         llms/local/
// 3. OpenAI:            llms/openai/
// 4. Google AI:         llms/googleai/
// 5. Cohere:            llms/cohere/
//
// Each subpackage includes provider-specific LLM implementations and helper files for communication
// with supported LLM providers. The internal directories within these subpackages contain provider-specific
// client and API implementations.
//
// The `llms.go` file contains the types and interfaces for interacting with different LLMs.
//
// The `options.go` file provides various options and functions to configure the LLMs.
package llms
