# Google AI SDK Migration Guide

## Overview

Google has announced that the `github.com/google/generative-ai-go` library is now considered legacy and recommends migrating to the official [Google Generative AI SDK for Go](https://github.com/googleapis/go-genai) for latest features, performance improvements, and active development.

## Current Status

The current LangChain Go implementation uses:
- `github.com/google/generative-ai-go/genai` (legacy)

## Migration Plan

### Phase 1: Compatibility Layer
- Add support for the new SDK alongside the existing implementation
- Provide option to choose between legacy and new SDK
- Maintain backward compatibility

### Phase 2: New SDK Implementation  
- Create new client implementation using `github.com/googleapis/go-genai`
- Migrate all existing functionality to the new SDK
- Add new features available in the official SDK

### Phase 3: Deprecation
- Mark legacy implementation as deprecated
- Provide migration guide for users
- Eventually remove legacy implementation

## Implementation Details

### New Package Structure
```
llms/googleai/
├── googleai.go          # Main interface (unchanged)
├── legacy/              # Legacy implementation
│   ├── client.go
│   └── options.go
├── official/            # New official SDK implementation
│   ├── client.go
│   └── options.go
└── MIGRATION.md        # This file
```

### Configuration Options
```go
// Use legacy SDK (default for backward compatibility)
llm, err := googleai.New(ctx, googleai.WithLegacySDK())

// Use official SDK
llm, err := googleai.New(ctx, googleai.WithOfficialSDK())
```

## Breaking Changes Expected

When migrating to the official SDK, users may encounter:

1. **Different authentication methods**
2. **Updated model names/identifiers**
3. **Changed API response structures**
4. **New configuration options**

## Timeline

- **Phase 1**: Add compatibility layer (current PR)
- **Phase 2**: Implement new SDK support (Q2 2025)
- **Phase 3**: Deprecate legacy SDK (Q4 2025)

## Related Issues

- GitHub Discussion: #1256 "How integrate googleapis/go-genai with langchaingo"
- Legacy SDK repository: https://github.com/google/generative-ai-go
- Official SDK repository: https://github.com/googleapis/go-genai

## Contributing

Community contributions are welcome for this migration. Please see the main CONTRIBUTING.md for guidelines.