# Streaming with iter.Seq2 Design

This document outlines the design for adding iter.Seq2 support to LangChainGo streaming APIs.

## Goals

1. **Modern Go Idioms**: Use Go 1.23+ iter.Seq2 for streaming
2. **Backwards Compatibility**: Keep existing callback-based APIs working
3. **Composability**: Enable chaining, filtering, transforming streams
4. **Type Safety**: Strongly typed events vs raw bytes
5. **Resource Management**: Proper cleanup and cancellation

## Design Overview

### Core Types

```go
package llms

import "iter"

// StreamEvent represents a single event in the stream
type StreamEvent struct {
    Type      EventType
    Content   string
    Reasoning string
    Metadata  map[string]any
    TokenInfo *TokenInfo
    Error     error
}

type EventType int

const (
    EventContent EventType = iota
    EventReasoning
    EventToolCall
    EventToolResult
    EventMetadata
    EventDone
    EventError
)

// TokenInfo tracks token usage during streaming
type TokenInfo struct {
    InputTokens     int
    OutputTokens    int
    ReasoningTokens int
    CachedTokens    int
    TotalTokens     int
}
```

### Streaming Interface

```go
// Add to Model interface (optional)
type StreamingModel interface {
    Model
    GenerateContentStream(ctx context.Context, messages []MessageContent, options ...CallOption) iter.Seq2[StreamEvent, error]
}
```

### Usage Examples

#### Basic Streaming
```go
stream := llm.GenerateContentStream(ctx, messages)
for event, err := range stream {
    if err != nil {
        return err
    }
    
    switch event.Type {
    case llms.EventContent:
        fmt.Print(event.Content)
    case llms.EventReasoning:
        log.Printf("Thinking: %s", event.Reasoning)
    }
}
```

#### Filtered Streaming
```go
contentStream := llms.FilterEvents(
    llm.GenerateContentStream(ctx, messages),
    llms.EventContent,
)

for event, err := range contentStream {
    fmt.Print(event.Content)
}
```

## Implementation Phases

### Phase 1: Core Types and Build Tags
- Add StreamEvent types
- Use build tags for Go 1.23+ support
- No breaking changes

### Phase 2: Provider Implementation
- Implement GenerateContentStream for OpenAI, Anthropic
- Use internal adapters to bridge callbacks → iterators
- Add unit tests

### Phase 3: Helper Functions
- FilterEvents, BufferEvents, MapEvents
- Composition utilities
- Documentation and examples

### Phase 4: Deprecation Path
- Mark callback APIs as deprecated (with timeline)
- Provide migration guide
- Keep both APIs working during transition

## Technical Details

### Build Tag Strategy
```go
//go:build go1.23

package llms

import "iter"

func (o *LLM) GenerateContentStream(...) iter.Seq2[StreamEvent, error] {
    // Implementation
}
```

### Adapter Pattern
```go
func callbackToIterator(
    ctx context.Context,
    callGen func(context.Context, func([]byte) error) error,
) iter.Seq2[StreamEvent, error] {
    return func(yield func(StreamEvent, error) bool) {
        err := callGen(ctx, func(chunk []byte) error {
            event := StreamEvent{
                Type: EventContent,
                Content: string(chunk),
            }
            if !yield(event, nil) {
                return fmt.Errorf("cancelled")
            }
            return nil
        })
        if err != nil {
            yield(StreamEvent{}, err)
        }
    }
}
```

### Composability Helpers
```go
// Filter events by type
func FilterEvents(stream iter.Seq2[StreamEvent, error], types ...EventType) iter.Seq2[StreamEvent, error]

// Transform events
func MapEvents(stream iter.Seq2[StreamEvent, error], transform func(StreamEvent) StreamEvent) iter.Seq2[StreamEvent, error]

// Buffer events in batches
func BufferEvents(stream iter.Seq2[StreamEvent, error], size int) iter.Seq2[[]StreamEvent, error]

// Take only first N events
func TakeEvents(stream iter.Seq2[StreamEvent, error], n int) iter.Seq2[StreamEvent, error]

// Combine multiple streams
func MergeEvents(streams ...iter.Seq2[StreamEvent, error]) iter.Seq2[StreamEvent, error]
```

## Benefits

1. **Native Go Idioms**: Works with range loops, natural cancellation
2. **Composable**: Unix pipe-like composition of stream operations
3. **Type Safe**: Structured events instead of raw bytes
4. **Resource Safe**: Automatic cleanup when iteration stops
5. **Future Proof**: Aligns with Go's evolution

## Migration Timeline

- **Q1 2025**: Implement core types and build tag support
- **Q2 2025**: Complete provider implementations  
- **Q3 2025**: Promote as primary streaming API
- **Q4 2025**: Begin deprecation warnings for callbacks
- **2026**: Remove callbacks in v2.0

## Compatibility

- Go 1.23+ for iter.Seq2 support
- Go 1.18+ continues working with callbacks
- No breaking changes during transition
- Automatic feature detection based on Go version

## Open Questions

1. Should we support streaming cancellation via context?
2. How to handle partial events (e.g., streaming tool arguments)?
3. Should EventError be separate or embedded in StreamEvent?
4. Do we need rate limiting/backpressure helpers?

## Example Implementation

See the OpenAI provider implementation in `llms/openai/streaming.go` (to be created).