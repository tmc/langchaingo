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
