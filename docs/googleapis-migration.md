# Google GenAI SDK Migration Analysis

## Overview

Google has deprecated the `github.com/google/generative-ai-go` package in favor of the new `github.com/googleapis/go-genai` package. This document analyzes the impact on langchaingo and provides a migration plan.

## Current State

LangChainGo currently uses:
- Package: `github.com/google/generative-ai-go` v0.15.1
- Status: **Legacy/Deprecated**
- Used in: `llms/googleai/` package

### Affected Files
- `llms/googleai/googleai.go`
- `llms/googleai/new.go` 
- `llms/googleai/embeddings.go`
- `llms/googleai/shared_test/shared_test.go`

## Migration Target

Google's new package:
- Package: `github.com/googleapis/go-genai`
- Status: **Official/Recommended**
- Benefits: Latest features, performance improvements, active development

## Impact Analysis

### 1. Import Changes Required
```go
// Current (legacy)
import "github.com/google/generative-ai-go/genai"

// Target (new)
import "github.com/googleapis/go-genai/genai"
```

### 2. API Compatibility
- **Unknown**: Need to analyze API differences between packages
- **Likely**: Some breaking changes in API surface
- **Impact**: May require updates to langchaingo's Google AI wrapper

### 3. Feature Parity
- **Risk**: New package may have different feature set
- **Opportunity**: Access to latest Google AI capabilities
- **Testing**: Extensive testing required to ensure compatibility

## Migration Strategy

### Phase 1: Investigation & Planning
1. ✅ **Document current usage** - Catalog all uses of legacy package
2. ⏳ **Analyze new package** - Study API differences and capabilities  
3. ⏳ **Create compatibility matrix** - Map old APIs to new APIs
4. ⏳ **Identify breaking changes** - Document what will break

### Phase 2: Implementation
1. ⏳ **Create migration branch** - Separate branch for migration work
2. ⏳ **Update dependencies** - Replace legacy package with new one
3. ⏳ **Update imports and APIs** - Modify code to use new package
4. ⏳ **Update tests** - Ensure all tests pass with new package

### Phase 3: Validation & Release
1. ⏳ **Integration testing** - Test with real Google AI services
2. ⏳ **Performance comparison** - Benchmark old vs new
3. ⏳ **Documentation updates** - Update examples and docs
4. ⏳ **Release planning** - Coordinate major version if needed

## Risks & Considerations

### Technical Risks
- **Breaking Changes**: New package may have incompatible APIs
- **Feature Gaps**: Some current features may not be available
- **Performance**: Migration might impact performance (positive or negative)

### Project Risks  
- **Compatibility**: Existing user code may break
- **Timeline**: Migration could take significant development time
- **Testing**: Comprehensive testing required across all Google AI features

### User Impact
- **Dependency Changes**: Users may need to update their dependencies
- **Code Changes**: User code using Google AI features may need updates
- **Version Planning**: May require major version bump

## Recommendation

### Immediate Actions (High Priority)
1. **Create GitHub Issue**: Track this migration work transparently
2. **API Analysis**: Deep dive into new package API surface
3. **Proof of Concept**: Small migration test to understand complexity
4. **Community Input**: Get feedback from users about migration timeline

### Timeline Considerations
- **Legacy Support**: Google may maintain legacy package for some time
- **User Migration**: Give users adequate time to migrate their code
- **Testing Period**: Allow extended testing period before release

## Implementation Plan

### Option 1: Direct Migration (Breaking Change)
```go
// Pros: Clean, simple, uses latest APIs
// Cons: Breaking change for users, requires major version bump

// Before
import "github.com/google/generative-ai-go/genai"

// After  
import "github.com/googleapis/go-genai/genai"
```

### Option 2: Compatibility Layer (Gradual Migration)
```go
// Pros: Backward compatible, gradual migration
// Cons: More complex, maintenance overhead

// Support both packages during transition period
// Detect which package user prefers and use accordingly
```

### Option 3: New Package (Side-by-side)
```go
// Pros: No breaking changes, users can choose
// Cons: Code duplication, maintenance overhead

// llms/googleai/     - legacy package support
// llms/googleaiv2/   - new package support
```

## Next Steps

1. **Research**: Study the new `github.com/googleapis/go-genai` package
2. **Experiment**: Create proof-of-concept migration
3. **Plan**: Develop detailed migration timeline
4. **Communicate**: Update GitHub discussion #1256 with findings

## References

- [GitHub Discussion #1256](https://github.com/tmc/langchaingo/discussions/1256)
- [Legacy Package](https://github.com/google/generative-ai-go) (Deprecated)
- [New Package](https://github.com/googleapis/go-genai) (Recommended)
- [Google AI Documentation](https://ai.google.dev/docs)

---

**Status**: Investigation Phase  
**Last Updated**: 2025-01-22  
**Next Review**: After new package analysis