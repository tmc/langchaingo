# Google AI SDK Migration Example

This example demonstrates how to migrate from the legacy Google AI SDK to the official SDK in LangChain Go, addressing [Discussion #1256](https://github.com/tmc/langchaingo/discussions/1256).

## Background

Google has deprecated `github.com/google/generative-ai-go` and recommends migrating to the official `google.golang.org/genai` SDK. This example shows how LangChain Go provides a smooth migration path.

## What This Example Demonstrates

1. **Default Behavior**: How the new default uses the official SDK
2. **Explicit SDK Selection**: Choosing between legacy and official SDKs
3. **Backward Compatibility**: Using the legacy SDK when needed
4. **Migration Testing**: Using the migration helper to compare SDK behavior
5. **Conditional Selection**: Choosing SDK based on environment variables

## Running the Example

```bash
export GOOGLE_API_KEY="your-google-api-key"
go run migration_example.go

# To test legacy SDK behavior
export USE_LEGACY_GOOGLE_SDK="true"
go run migration_example.go
```

## Key Features Showcased

### SDK Version Selection
```go
// Official SDK (default)
client, err := googleai.New(ctx, googleai.WithSDKVersion(googleai.SDKOfficial))

// Legacy SDK (for compatibility)
client, err := googleai.New(ctx, googleai.WithSDKVersion(googleai.SDKLegacy))
```

### Migration Helper
```go
helper, err := googleai.NewMigrationHelper(ctx, apiKey)
defer helper.Close()

// Compare behavior between SDKs
legacyClient := helper.GetLegacyClient()
officialClient := helper.GetOfficialClient()
```

### Conditional SDK Selection
```go
var sdkVersion googleai.SDKVersion
if os.Getenv("USE_LEGACY_GOOGLE_SDK") == "true" {
    sdkVersion = googleai.SDKLegacy
} else {
    sdkVersion = googleai.SDKOfficial
}

client, err := googleai.New(ctx, googleai.WithSDKVersion(sdkVersion))
```

## Migration Benefits

### Official SDK Advantages
- **Performance**: Optimized for production workloads
- **Features**: Access to latest Google AI capabilities
- **Support**: Official Google maintenance and updates
- **Future-proof**: Long-term compatibility guarantee

### Backward Compatibility
- **No breaking changes**: Existing code continues to work
- **Same interface**: LangChain Go API remains unchanged
- **Gradual migration**: Move at your own pace

## Real-World Migration Steps

1. **Test with Official SDK**: Run your application with the default settings
2. **Compare Behavior**: Use the migration helper to validate responses
3. **Performance Testing**: Benchmark both SDKs with your workload
4. **Gradual Rollout**: Deploy with official SDK to a subset of traffic
5. **Full Migration**: Switch all traffic to official SDK

## Troubleshooting

### Common Issues
- **Different response times**: Official SDK may have different performance characteristics
- **Error message changes**: Error formats may differ between SDKs
- **Authentication differences**: Both use the same auth methods but error messages may vary

### Solutions
- Use the migration helper to identify differences
- Implement proper error handling for both SDKs
- Test thoroughly before full migration

## Related Documentation

- [Google AI SDK Migration Plan](../../docs/GOOGLE_AI_SDK_MIGRATION.md)
- [Migration Guide](../../llms/googleai/MIGRATION_GUIDE.md)
- [Google AI Package Documentation](../../llms/googleai/)

## Contributing

This example addresses community questions about Google AI SDK migration. If you have suggestions for improvements or encounter migration issues, please contribute to the [GitHub discussion](https://github.com/tmc/langchaingo/discussions/1256).