# Google AI SDK Migration Plan

## Overview
This document outlines the plan to migrate from the legacy `github.com/google/generative-ai-go` package to the official `google.golang.org/genai` package as recommended in the [GitHub advisory](https://github.com/google/generative-ai-go).

## Current State
- Currently using: `github.com/google/generative-ai-go v0.15.1`
- Target: `google.golang.org/genai v1.6.0`

## API Differences Identified

### 1. Client API Changes
- Old: `client.GenerativeModel(modelName)`
- New: Different client initialization and model access patterns

### 2. Content/Part Structure
- Old: `genai.Text()` returns a Part
- New: `genai.Text()` returns `[]*Content`, different type hierarchy

### 3. Response Structure
- Old: `usage.CandidatesTokenCount`
- New: Different usage metadata structure

### 4. Type Changes
- Old: `genai.GenerativeModel` type
- New: Different model type structure

## Migration Strategy

### Phase 1: API Compatibility Layer
1. Create adapter/wrapper types to bridge the API differences
2. Implement compatibility functions that map old API calls to new ones
3. Maintain backward compatibility for existing users

### Phase 2: Gradual Migration
1. Update internal implementations to use new SDK
2. Add new SDK-specific features while maintaining old API surface
3. Add deprecation warnings for old patterns

### Phase 3: Full Migration
1. Update all method signatures to use new types
2. Remove compatibility layer
3. Update documentation and examples

## Implementation Plan

### Files to Update
- `googleai.go` - Main implementation
- `embeddings.go` - Embedding functionality  
- `new.go` - Client creation
- Tests and examples

### Testing Strategy
- Ensure existing tests pass with compatibility layer
- Add integration tests with new SDK
- Test migration path thoroughly

## Breaking Changes
This will be a major version change due to API differences. We should:
1. Increment major version
2. Provide clear migration guide
3. Support both versions during transition period

## Timeline
- Phase 1: 2-3 weeks (compatibility layer)
- Phase 2: 1-2 weeks (gradual migration)  
- Phase 3: 1 week (cleanup and documentation)

## Related Issues
- Discussion: https://github.com/tmc/langchaingo/discussions/1256
- Google Advisory: Repository migration notice

## Notes
The new SDK has significant architectural differences that require careful planning to avoid breaking existing users. A compatibility layer approach will provide the smoothest migration path.