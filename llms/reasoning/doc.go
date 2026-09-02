// Package reasoning provides primitives for working with reasoning content.
//
// This package implements utilities for extracting and processing reasoning
// (step-by-step thinking) content from language model responses. It operates in two modes:
//
//  1. Streaming mode: Uses a stateful ChunkContentSplitter to process content chunks
//     as they arrive, separating text from reasoning blocks marked with <think> or
//     <thinking> tags. The splitter maintains state between calls to handle cases
//     where reasoning blocks span multiple chunks.
//
//  2. Complete content mode: Uses SplitContent function to separate reasoning content
//     from a complete response, handling cases where the LLM backend hasn't explicitly
//     divided reasoning from regular content.
//
// The package recognizes reasoning content enclosed in <think>...</think> or
// <thinking>...</thinking> tags, which various LLM providers use to indicate
// step-by-step thinking processes.
//
// Beyond content splitting, the package is the low-level source of truth for model
// capabilities, shared by every provider adapter so the wire shape is resolved the
// same way everywhere. It sits below llms and must not import it. It provides:
//
//   - Reasoning-model detection: IsReasoningModel / DefaultIsReasoningModel.
//   - Per-family capability resolvers keyed by the distinctive part of the model
//     name (so they work with and without an OpenRouter provider prefix):
//     ClaudeReasoningKindFor and the Claude_* helpers (adaptive vs budget thinking,
//     sampling rules, effort-with-budget, structured-output support),
//     OpenAIReasoningCapsFor / ClampEffort, and the Gemini* helpers.
//   - ResolveOff: the single, provider-aware decision of how to disable thinking
//     (an explicit disable wire, a zero budget, or ErrReasoningOffUnsupported for
//     models that cannot be disabled, e.g. always-on Claude or Bedrock defaults).
//
// Unrecognized models are treated as optimistic pass-through: the provider API,
// not a local table, is the final arbiter, so a newer model never regresses.
//
// Example usage for streaming mode:
//
//	splitter := reasoning.NewChunkContentSplitter()
//
//	for chunk := range responseChunks {
//	    text, reasoning := splitter.Split(chunk)
//
//	    if reasoning != "" {
//	        // Process reasoning content (e.g., display as step-by-step thinking)
//	        fmt.Println("Reasoning:", reasoning)
//	    }
//
//	    if text != "" {
//	        // Process regular text content
//	        fmt.Println("Content:", text)
//	    }
//	}
//
// Example usage for complete content mode:
//
//	response := "Here's what I found: <thinking>First, I need to analyze the data.
//	The pattern shows increasing values.</thinking> The trend is clearly upward."
//
//	reasoning, content := reasoning.SplitContent(response)
//
//	fmt.Println("Reasoning:", reasoning)  // "First, I need to analyze the data. The pattern shows increasing values."
//	fmt.Println("Content:", content)      // "Here's what I found: The trend is clearly upward."
//
// See also:
//   - IsReasoningModel: Checks if a model supports reasoning
//   - DefaultIsReasoningModel: Provides the default reasoning model detection logic
package reasoning

//go:generate go run ./internal/cmd/modelsdev -out testdata/models_dev.json
//go:generate go run ./internal/cmd/mistralmodels -out testdata/mistral_models.json
