# Google AI SDK Migration Guide

This guide helps you migrate from the legacy Google AI SDK to the official SDK in LangChain Go, addressing [Discussion #1256](https://github.com/tmc/langchaingo/discussions/1256).

## Overview

Google has deprecated the `github.com/google/generative-ai-go` package and now recommends using the official `google.golang.org/genai` SDK. LangChain Go now supports both SDKs to provide a smooth migration path.

## What's Changed

### New Default Behavior
- **v0.1.14+**: Official SDK (`google.golang.org/genai`) is the default
- **Legacy support**: Legacy SDK (`github.com/google/generative-ai-go`) is still available but deprecated

### Backward Compatibility
- **Existing code continues to work**: No immediate breaking changes
- **Same API interface**: The LangChain Go interface remains unchanged
- **Automatic SDK selection**: The system chooses the appropriate SDK based on configuration

## Migration Options

### Option 1: No Action Required (Recommended for Most Users)
If you're satisfied with the current behavior and don't need specific SDK features:

```go
// This now uses the official SDK by default
client, err := googleai.New(ctx, googleai.WithAPIKey("your-api-key"))
if err != nil {
    log.Fatal(err)
}
```

### Option 2: Explicitly Choose Official SDK
To be explicit about using the new SDK:

```go
client, err := googleai.New(ctx, 
    googleai.WithAPIKey("your-api-key"),
    googleai.WithSDKVersion(googleai.SDKOfficial), // Explicit official SDK
)
if err != nil {
    log.Fatal(err)
}

// Or use the convenience constructor
client, err := googleai.NewWithOfficialSDK(ctx, googleai.WithAPIKey("your-api-key"))
```

### Option 3: Continue Using Legacy SDK
If you need to continue using the legacy SDK temporarily:

```go
client, err := googleai.New(ctx, 
    googleai.WithAPIKey("your-api-key"),
    googleai.WithSDKVersion(googleai.SDKLegacy), // Explicit legacy SDK
)
if err != nil {
    log.Fatal(err)
}

// Or use the convenience constructor
client, err := googleai.NewWithLegacySDK(ctx, googleai.WithAPIKey("your-api-key"))
```

## Testing Your Migration

### Comparing SDK Behavior
Use the migration helper to compare behavior between SDKs:

```go
helper, err := googleai.NewMigrationHelper(ctx, "your-api-key")
if err != nil {
    log.Fatal(err)
}
defer helper.Close()

// Test with both SDKs to ensure identical behavior
legacyClient := helper.GetLegacyClient()
officialClient := helper.GetOfficialClient()

// Your testing logic here
```

### Validation Checklist
- [ ] All existing functionality works with the official SDK
- [ ] Response formats are identical
- [ ] Performance is acceptable
- [ ] Error handling behaves as expected
- [ ] Streaming functionality works correctly
- [ ] Multimodal content (images, etc.) processes correctly

## Key Differences Between SDKs

### API Compatibility
- **Interface**: LangChain Go interface remains the same
- **Responses**: Response format is preserved across SDKs
- **Features**: Feature parity maintained

### Performance
- **Official SDK**: Optimized for production use
- **Legacy SDK**: May have different performance characteristics

### Error Handling
- **Similar patterns**: Both SDKs use similar error handling
- **Error messages**: Specific error messages may differ

## Troubleshooting

### Common Issues

#### 1. Build Errors After Update
```
go: module requires Go 1.21 or later
```
**Solution**: Update to Go 1.21+ or use the legacy SDK:
```go
googleai.WithSDKVersion(googleai.SDKLegacy)
```

#### 2. Different Response Behavior
If you notice response differences:
1. Enable verbose logging
2. Compare responses using the migration helper
3. Report issues to the LangChain Go repository

#### 3. Authentication Issues
Both SDKs use the same authentication methods:
```go
// API Key
googleai.WithAPIKey("your-key")

// Service Account
googleai.WithCredentialsFile("path/to/credentials.json")

// Default credentials
googleai.WithCredentials(credentials)
```

### Performance Differences
If you experience performance issues:
1. Benchmark both SDKs with your workload
2. Consider adjusting timeout and retry settings
3. Monitor memory usage patterns

## Timeline and Deprecation

### Current Status (v0.1.14+)
- ✅ Official SDK is default
- ✅ Legacy SDK available with deprecation warnings
- ✅ Full backward compatibility

### Future Versions (v0.2.0+)
- ⚠️ Legacy SDK support may be removed
- ⚠️ Breaking change notices will be provided
- ⚠️ Migration assistance will be available

### Recommendations
1. **Immediate**: Test your application with the official SDK
2. **Short-term (1-2 months)**: Migrate to explicit official SDK usage
3. **Long-term (6+ months)**: Remove any legacy SDK dependencies

## Feature Comparison

| Feature | Legacy SDK | Official SDK | Notes |
|---------|------------|--------------|-------|
| Text Generation | ✅ | ✅ | Identical interface |
| Streaming | ✅ | ✅ | Same performance |
| Function Calling | ✅ | ✅ | Enhanced in official |
| Multimodal | ✅ | ✅ | Better support in official |
| Safety Filters | ✅ | ✅ | Same configuration |
| Vertex AI | ✅ | ✅ | Improved integration |

## Code Examples

### Basic Migration
```go
// Before (implicit legacy)
client, err := googleai.New(ctx, googleai.WithAPIKey(apiKey))

// After (explicit official - recommended)
client, err := googleai.New(ctx, 
    googleai.WithAPIKey(apiKey),
    googleai.WithSDKVersion(googleai.SDKOfficial),
)
```

### Advanced Configuration
```go
// Complete migration example
client, err := googleai.New(ctx,
    googleai.WithAPIKey(apiKey),
    googleai.WithSDKVersion(googleai.SDKOfficial),
    googleai.WithDefaultModel("gemini-pro"),
    googleai.WithDefaultTemperature(0.7),
    googleai.WithHTTPClient(customHTTPClient),
)
```

### Conditional SDK Selection
```go
// Choose SDK based on environment
var sdkVersion googleai.SDKVersion
if os.Getenv("USE_LEGACY_GOOGLE_SDK") == "true" {
    sdkVersion = googleai.SDKLegacy
} else {
    sdkVersion = googleai.SDKOfficial
}

client, err := googleai.New(ctx,
    googleai.WithAPIKey(apiKey),
    googleai.WithSDKVersion(sdkVersion),
)
```

## Getting Help

### Community Support
- [GitHub Discussions](https://github.com/tmc/langchaingo/discussions)
- [Migration Discussion #1256](https://github.com/tmc/langchaingo/discussions/1256)

### Reporting Issues
If you encounter migration issues:
1. Check this guide first
2. Search existing issues
3. Create a new issue with:
   - SDK versions tested
   - Code example
   - Expected vs actual behavior
   - Error messages

### Contributing
Help improve this migration guide:
- Share your migration experience
- Report documentation gaps
- Suggest improvements

## FAQ

**Q: Do I need to change my code immediately?**
A: No, existing code continues to work. The default has changed to the official SDK, but backward compatibility is maintained.

**Q: When will legacy SDK support be removed?**
A: Not before v0.2.0, with at least 6 months notice and migration assistance.

**Q: Are there performance differences?**
A: The official SDK is optimized for production use and may have different performance characteristics. Test with your specific workload.

**Q: Can I use both SDKs in the same application?**
A: Yes, you can configure different clients to use different SDKs, though this is not recommended for production.

**Q: What if I find a bug in the official SDK?**
A: Report it to the LangChain Go repository. You can temporarily use the legacy SDK while the issue is resolved.

## Additional Resources

- [Google AI SDK Official Documentation](https://pkg.go.dev/google.golang.org/genai)
- [LangChain Go Documentation](../../docs/)
- [Example Applications](../../examples/)
- [Migration Discussion](https://github.com/tmc/langchaingo/discussions/1256)