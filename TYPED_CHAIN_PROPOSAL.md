# Proposal: Typed Chain Interface with Generics

This proposal addresses GitHub discussion #1073 regarding making the Chain interface more type-safe using Go generics.

## Problem Statement

The current `Chain` interface has several usability issues:

1. **Lack of Type Safety**: Returns `map[string]any` requiring manual type assertions
2. **Error-Prone**: Easy to extract wrong keys or cast to wrong types
3. **Poor Developer Experience**: No compile-time checking of output types
4. **Documentation Burden**: Requires external documentation of output types

### Current Pattern (Problematic)

```go
// Current usage requires manual type handling
result, err := chains.Call(ctx, myChain, inputs)
if err != nil {
    return err
}

// Manual key extraction and type assertion - error prone!
text, ok := result["text"].(string)
if !ok {
    return fmt.Errorf("unexpected type for text output")
}

// What if the key changes? What if the type is wrong? No compile-time checking!
```

## Proposed Solution

### New Generic Chain Interface

```go
// Chain is a generic interface for chains that produce typed outputs
type Chain[T any] interface {
    // Call runs the logic of the chain and returns typed output + metadata
    Call(ctx context.Context, inputs map[string]any, options ...ChainCallOption) (T, map[string]any, error)
    
    // GetMemory gets the memory of the chain
    GetMemory() schema.Memory
    
    // GetInputKeys returns the input keys the chain expects
    GetInputKeys() []string
    
    // OutputType returns a description of the output type (for documentation/debugging)
    OutputType() string
}
```

### Benefits

1. **Type Safety**: Compile-time type checking
2. **Better API**: Clear separation of primary output and metadata
3. **Reduced Errors**: No manual type assertions
4. **Self-Documenting**: Output type is part of the interface
5. **IDE Support**: Better autocompletion and refactoring

### Backward Compatibility Strategy

#### Phase 1: Introduce Alongside Current Interface

```go
// Keep existing interface for backward compatibility
type LegacyChain interface {
    Call(ctx context.Context, inputs map[string]any, options ...ChainCallOption) (map[string]any, error)
    GetMemory() schema.Memory
    GetInputKeys() []string
    GetOutputKeys() []string
}

// New generic interface
type Chain[T any] interface {
    Call(ctx context.Context, inputs map[string]any, options ...ChainCallOption) (T, map[string]any, error)
    GetMemory() schema.Memory
    GetInputKeys() []string
    OutputType() string
}

// Adapter to convert between interfaces
type ChainAdapter[T any] struct {
    chain Chain[T]
}

func (a ChainAdapter[T]) Call(ctx context.Context, inputs map[string]any, options ...ChainCallOption) (map[string]any, error) {
    result, metadata, err := a.chain.Call(ctx, inputs, options...)
    if err != nil {
        return nil, err
    }
    
    // Combine typed result with metadata
    output := make(map[string]any)
    for k, v := range metadata {
        output[k] = v
    }
    output["result"] = result // or use chain-specific key
    
    return output, nil
}
```

#### Phase 2: Migrate Core Chains

```go
// Example: LLMChain becomes generic
type LLMChain[T any] struct {
    Prompt           prompts.FormatPrompter
    LLM              llms.Model
    Memory           schema.Memory
    CallbacksHandler callbacks.Handler
    OutputParser     schema.OutputParser[T] // Already generic!
    OutputKey        string
}

func (c LLMChain[T]) Call(ctx context.Context, values map[string]any, options ...ChainCallOption) (T, map[string]any, error) {
    promptValue, err := c.Prompt.FormatPrompt(values)
    if err != nil {
        var zero T
        return zero, nil, err
    }

    result, err := llms.GenerateFromSinglePrompt(ctx, c.LLM, promptValue.String(), getLLMCallOptions(options...)...)
    if err != nil {
        var zero T
        return zero, nil, err
    }

    // Parse to typed output
    finalOutput, err := c.OutputParser.ParseWithPrompt(result, promptValue)
    if err != nil {
        var zero T
        return zero, nil, err
    }

    // Return typed output + any metadata
    metadata := map[string]any{
        "prompt": promptValue.String(),
        "raw_output": result,
    }
    
    return finalOutput, metadata, nil
}

func (c LLMChain[T]) OutputType() string {
    return fmt.Sprintf("LLMChain[%T]", *new(T))
}
```

### Usage Examples

#### Simple String Chain
```go
// Create a typed string chain
stringChain := chains.NewLLMChain[string](
    llm,
    prompts.NewPromptTemplate("Question: {{.question}}", []string{"question"}),
)

// Usage with type safety
answer, metadata, err := stringChain.Call(ctx, map[string]any{
    "question": "What is the capital of France?",
})
// answer is string, no type assertion needed!
// metadata contains additional info like prompt, raw_output, etc.
```

#### Structured Output Chain
```go
type PersonInfo struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
    City string `json:"city"`
}

// Create a typed structured chain
structuredChain := chains.NewLLMChain[PersonInfo](
    llm,
    prompts.NewPromptTemplate("Extract person info: {{.text}}", []string{"text"}),
    chains.WithOutputParser(outputparser.NewStructured[PersonInfo]()),
)

// Usage with full type safety
person, metadata, err := structuredChain.Call(ctx, map[string]any{
    "text": "John is 30 years old and lives in Paris",
})
// person is PersonInfo struct, fully typed!
```

#### Chain Composition
```go
// Chains can be composed with type safety
func ComposeChains[T, U any](
    first Chain[T],
    second Chain[U],
    combiner func(T, U) U,
) Chain[U] {
    return &ComposedChain[T, U]{
        first:    first,
        second:   second,
        combiner: combiner,
    }
}
```

### Helper Functions

```go
// Typed versions of existing functions
func Call[T any](ctx context.Context, c Chain[T], inputValues map[string]any, options ...ChainCallOption) (T, map[string]any, error) {
    // Implementation similar to current Call but with type safety
}

func Run[T any](ctx context.Context, c Chain[T], input any, options ...ChainCallOption) (T, error) {
    result, _, err := Call(ctx, c, map[string]any{"input": input}, options...)
    return result, err
}

// Predict function for simple text-to-typed-output chains
func Predict[T any](ctx context.Context, c Chain[T], values map[string]any, options ...ChainCallOption) (T, error) {
    result, _, err := Call(ctx, c, values, options...)
    return result, err
}
```

### Migration Path

#### Phase 1: Foundation (Non-Breaking)
1. Add new generic `Chain[T]` interface alongside existing
2. Add adapter types for interoperability
3. Migrate helper functions to support both interfaces

#### Phase 2: Core Migration (Potentially Breaking)
1. Migrate `LLMChain` to be generic
2. Migrate other core chains (`RetrievalQA`, `SequentialChain`, etc.)
3. Update examples and documentation

#### Phase 3: Cleanup (Breaking)
1. Deprecate old interface
2. Remove adapter code
3. Finalize API

### Advanced Features

#### Type Constraints
```go
// Constrain chains to specific output types
type StringChain = Chain[string]
type DocumentChain = Chain[[]schema.Document]

// Constraint for serializable outputs
type SerializableChain[T any] interface {
    Chain[T]
    json.Marshaler
    json.Unmarshaler
}
```

#### Chain Validation
```go
// Compile-time validation of chain connections
func ValidateChainConnection[T, U any](
    source Chain[T],
    destination Chain[U],
    mapper func(T) map[string]any,
) error {
    // Validate that source output can be mapped to destination input
}
```

## Implementation Considerations

### Challenges

1. **Breaking Changes**: Core interface change affects all chains
2. **Generic Complexity**: May make simple use cases more complex
3. **Type Inference**: Go's generic type inference has limitations
4. **Migration Effort**: Large codebase to migrate

### Solutions

1. **Gradual Migration**: Introduce alongside existing interface
2. **Helper Types**: Provide common type aliases
3. **Documentation**: Extensive examples and migration guide
4. **Tooling**: Automated migration tools where possible

### Performance Impact

- **Positive**: Eliminates type assertions and map lookups
- **Neutral**: Generic instantiation cost is minimal
- **Negative**: Larger binary size due to generic instantiation

## Conclusion

This proposal provides:

1. **Type Safety**: Compile-time checking of chain outputs
2. **Better UX**: Cleaner API with less boilerplate
3. **Backward Compatibility**: Gradual migration path
4. **Future-Proof**: Foundation for more advanced typed features

The migration can be done incrementally, allowing users to adopt the new interface at their own pace while maintaining backward compatibility.

### Recommendation

Start with **Phase 1** implementation to:
1. Validate the approach with community feedback
2. Identify migration challenges early
3. Provide immediate value to users who want type safety

This addresses the core concerns raised in discussion #1073 while providing a practical migration path for the existing ecosystem.