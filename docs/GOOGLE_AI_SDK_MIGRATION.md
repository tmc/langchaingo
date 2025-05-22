# Google AI SDK Migration Plan

This document outlines the migration from the legacy `github.com/google/generative-ai-go` to the official `github.com/googleapis/go-genai` SDK, addressing [Discussion #1256](https://github.com/tmc/langchaingo/discussions/1256).

## Background

Google has deprecated the `github.com/google/generative-ai-go` repository and recommends migrating to the official Google Client Library: `github.com/googleapis/go-genai`.

From the legacy repository:
> Please be advised that this repository is now considered legacy. For the latest features, performance improvements, and active development, we strongly recommend migrating to the official [Google Generative AI SDK for Go](https://github.com/googleapis/go-genai).

## Current Implementation Analysis

LangChain Go currently uses:
- Package: `github.com/google/generative-ai-go/genai`
- Location: `llms/googleai/`
- Features implemented:
  - Text generation
  - Chat completion
  - Streaming responses
  - Function calling
  - Vision/multimodal content
  - Safety settings
  - Vertex AI integration

## Migration Strategy

### Phase 1: Dual Support (Recommended)
1. **Maintain Legacy Support**: Keep existing implementation for backward compatibility
2. **Add New SDK Implementation**: Create alternative implementation using `googleapis/go-genai`
3. **Feature Parity**: Ensure all features work with both SDKs
4. **Configuration Options**: Allow users to choose which SDK to use

### Phase 2: Gradual Migration
1. **Default to New SDK**: Make the new SDK the default choice
2. **Deprecation Warnings**: Add warnings for legacy SDK usage
3. **Documentation Updates**: Update all examples and documentation

### Phase 3: Legacy Removal (Future)
1. **Breaking Change**: Remove legacy SDK support in next major version
2. **Migration Guide**: Provide comprehensive migration assistance

## Implementation Plan

### 1. New Package Structure
```
llms/googleai/
├── googleai.go              # Current implementation (legacy)
├── googleai_v2.go          # New SDK implementation  
├── options.go              # Shared options
├── internal/
│   ├── legacy/            # Legacy SDK client
│   │   └── client.go
│   └── official/          # New SDK client
│       └── client.go
└── migration_guide.md     # User migration guide
```

### 2. Configuration Options
```go
// Option to choose SDK version
type SDKVersion int

const (
    SDKLegacy SDKVersion = iota  // Use legacy google/generative-ai-go
    SDKOfficial                  // Use googleapis/go-genai (default)
)

// New option function
func WithSDKVersion(version SDKVersion) Option {
    return func(o *options) {
        o.sdkVersion = version
    }
}
```

### 3. Backward Compatible Constructor
```go
// Existing constructor - defaults to new SDK but maintains compatibility
func New(ctx context.Context, opts ...Option) (*GoogleAI, error) {
    // Default to official SDK
    options := &options{
        sdkVersion: SDKOfficial,
    }
    
    for _, opt := range opts {
        opt(options)
    }
    
    switch options.sdkVersion {
    case SDKLegacy:
        return newWithLegacySDK(ctx, options)
    case SDKOfficial:
        return newWithOfficialSDK(ctx, options)
    default:
        return newWithOfficialSDK(ctx, options)
    }
}

// Legacy constructor for explicit usage
func NewWithLegacySDK(ctx context.Context, opts ...Option) (*GoogleAI, error) {
    opts = append(opts, WithSDKVersion(SDKLegacy))
    return New(ctx, opts...)
}
```

## Key API Differences

Based on the available documentation, here are the expected differences:

### Authentication
**Legacy SDK:**
```go
client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
```

**Official SDK:**
```go
// Expected pattern (to be confirmed)
client, err := genai.NewGenerativeClient(ctx, option.WithAPIKey(apiKey))
```

### Model Initialization
**Legacy SDK:**
```go
model := client.GenerativeModel("gemini-pro")
```

**Official SDK:**
```go
// Pattern may differ - needs investigation
model := client.GenerativeModel("gemini-pro")
```

### Content Generation
The content generation patterns should be similar but may have different response structures.

## Benefits of Migration

### Performance Improvements
- Optimized for production workloads
- Better resource management
- Improved connection pooling

### Feature Completeness
- Access to latest Gemini features
- Better multimodal support
- Enhanced function calling capabilities

### Long-term Support
- Active development and maintenance
- Regular security updates
- Official Google support

## Migration Timeline

### Immediate (Phase 1)
- [ ] Research official SDK API patterns
- [ ] Implement dual SDK support
- [ ] Create comprehensive tests
- [ ] Update documentation

### Short-term (1-2 months)
- [ ] Default to official SDK
- [ ] Add deprecation warnings for legacy usage
- [ ] Update all examples

### Long-term (Next major version)
- [ ] Remove legacy SDK support
- [ ] Clean up dual implementation code

## Risk Assessment

### Low Risk Items
- New installations will use official SDK
- Existing code continues to work
- Feature parity maintained

### Medium Risk Items
- Potential API differences requiring code changes
- Different error handling patterns
- Performance characteristic changes

### High Risk Items
- Breaking changes in official SDK
- Incompatible authentication mechanisms
- Major feature differences

## Testing Strategy

### Compatibility Tests
- Run existing test suite against both SDKs
- Ensure identical behavior for all features
- Performance benchmarking

### Integration Tests
- Test with real Gemini API
- Verify streaming functionality
- Validate multimodal content handling

### Migration Tests
- Test upgrade path from legacy to official
- Verify configuration migration
- Ensure no data loss during transition

## User Communication

### Documentation Updates
- Update README with migration information
- Create migration guide for users
- Update all code examples

### Deprecation Notices
- Add warnings to legacy constructor
- Include migration timeline in warnings
- Provide clear upgrade path

### Community Support
- Monitor GitHub discussions for issues
- Provide assistance for complex migrations
- Update based on user feedback

## Implementation Status

This is a planning document. Implementation will proceed based on:
1. Community feedback on this migration plan
2. Investigation of official SDK API patterns
3. Resource availability for dual SDK maintenance

## Next Steps

1. **Research Phase**: Investigate official SDK API patterns and capabilities
2. **Prototype Phase**: Create minimal working implementation with official SDK
3. **Implementation Phase**: Build dual SDK support
4. **Testing Phase**: Comprehensive testing of both implementations
5. **Documentation Phase**: Update guides and examples
6. **Release Phase**: Ship with backward compatibility

## Community Input

This migration plan is open for community discussion. Please provide feedback on:
- Migration timeline preferences
- Feature priority concerns
- Backward compatibility requirements
- Testing and validation needs

See [Discussion #1256](https://github.com/tmc/langchaingo/discussions/1256) for ongoing conversation.