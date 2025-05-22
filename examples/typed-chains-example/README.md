# Typed Chains Example

This example demonstrates the proposed typed Chain interface that would provide compile-time type safety for chain outputs.

## Current Problems

The existing Chain interface returns `map[string]any`, requiring:
- Manual type assertions
- Knowledge of output keys
- Runtime error checking
- Poor IDE support

```go
// Current usage (error-prone)
result, err := chains.Call(ctx, myChain, inputs)
text, ok := result["text"].(string)  // Manual casting, what if key is wrong?
if !ok {
    // Handle type error at runtime
}
```

## Proposed Solution

Typed chains provide compile-time type safety:

```go
// Proposed usage (type-safe)  
result, metadata, err := typedChain.Call(ctx, inputs)
// result is already the correct type, no casting needed!
```

## Examples in This Demo

1. **Simple String Chain** - Basic typed string output
2. **Structured Output Chain** - Complex struct output with full type safety
3. **Legacy Compatibility** - Adapter pattern for backward compatibility
4. **Type Safety Benefits** - Demonstrates compile-time checking

## Running the Example

```bash
go run main.go
```

## Key Benefits Demonstrated

- ✅ **Compile-time type checking** - Wrong types caught at build time
- ✅ **No manual casting** - Results are already the correct type
- ✅ **Better IDE support** - Autocompletion and refactoring work properly
- ✅ **Cleaner APIs** - Separation of primary result and metadata
- ✅ **Backward compatibility** - Can work alongside existing chains

## Implementation Status

This is a **proof-of-concept** implementation showing:
- How the typed interface would work
- Backward compatibility strategy
- Migration path from current chains
- Type safety benefits

The actual implementation would require:
1. Community discussion and approval
2. Gradual migration of existing chains
3. Documentation and migration guides
4. Testing with real-world usage patterns