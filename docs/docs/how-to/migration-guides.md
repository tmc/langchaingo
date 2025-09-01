# Migration Guides

This document provides migration guides for major changes in LangChainGo.

## OpenAI Default Model Change (v0.1.14+)

### What Changed

Starting with LangChainGo v0.1.14+, the default OpenAI model has changed from **`gpt-3.5-turbo`** to **`gpt-4o-mini`**.

### Why This Change

- `gpt-4o-mini` offers better performance and capabilities than `gpt-3.5-turbo`
- It's more cost-effective and faster than full GPT-4 models
- Provides better reasoning and instruction-following capabilities

### Impact Assessment

**If you're not explicitly setting a model:**
```go
// This will now use gpt-4o-mini instead of gpt-3.5-turbo
llm, err := openai.New()
```

**Your code will continue to work**, but responses might be:
- Slightly different in style or format
- More accurate and detailed
- Potentially use slightly more tokens

### Migration Options

#### Option 1: Accept the new default (Recommended)
Most applications will benefit from the improved capabilities. No code changes needed.

#### Option 2: Explicitly use the old model
If you need exact compatibility with previous behavior:

```go
// Before (implicit)
llm, err := openai.New()

// After (explicit)
llm, err := openai.New(openai.WithModel("gpt-3.5-turbo"))
```

#### Option 3: Upgrade to a better model
Take this opportunity to upgrade to an even better model:

```go
// Upgrade to full GPT-4o for best quality
llm, err := openai.New(openai.WithModel("gpt-4o"))

// Or use GPT-4o-mini explicitly
llm, err := openai.New(openai.WithModel("gpt-4o-mini"))
```

### Testing Your Application

1. **Test with default model**: Run your application without specifying a model
2. **Compare outputs**: Check if the new outputs meet your requirements
3. **Monitor token usage**: `gpt-4o-mini` usage should be similar to `gpt-3.5-turbo`
4. **Update tests**: If you have tests that depend on specific response formats

### Example Migration

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/tmc/langchaingo/llms/openai"
)

func main() {
    ctx := context.Background()
    
    // ✅ Modern approach - uses gpt-4o-mini by default
    llm, err := openai.New()
    if err != nil {
        log.Fatal(err)
    }
    
    // ✅ If you need the old behavior
    // llm, err := openai.New(openai.WithModel("gpt-3.5-turbo"))
    
    // ✅ If you want the best available model
    // llm, err := openai.New(openai.WithModel("gpt-4o"))
    
    response, err := llm.Call(ctx, "What is the capital of France?")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(response)
}
```

---

## Milvus SDK Migration

### What Changed

The official Milvus Go SDK repository has been deprecated and archived. The new SDK is now part of the main Milvus repository.

### Migration Steps

#### 1. Update Dependencies

In your `go.mod`:

```diff
- github.com/milvus-io/milvus-sdk-go/v2 v2.4.0
+ github.com/milvus-io/milvus/client/v2 v2.4.2
```

#### 2. Update Import Statements

```diff
- import "github.com/milvus-io/milvus-sdk-go/v2/client"
- import "github.com/milvus-io/milvus-sdk-go/v2/entity"
+ import "github.com/milvus-io/milvus/client/v2/client"  
+ import "github.com/milvus-io/milvus/client/v2/entity"
```

#### 3. Run Migration Commands

```bash
# Update dependencies
go get github.com/milvus-io/milvus/client/v2@latest
go mod tidy

# Test the changes
go test ./vectorstores/milvus/...
```

### Files That Need Updates

If you're using Milvus vector store directly:
- Any files importing the old SDK
- Integration tests
- Example applications

The API remains largely compatible, so most code should work without changes.

---

## Breaking Changes Checklist

When migrating between versions, check for these common breaking changes:

### v0.1.14+
- [ ] Default OpenAI model changed to `gpt-4o-mini`
- [ ] Review any hardcoded model assumptions
- [ ] Test applications with new default behavior

### v0.1.13+
- [ ] Check Milvus vector store imports
- [ ] Update deprecated Milvus SDK references

### General Migration Tips

1. **Pin your versions**: Use specific versions in production
   ```go
   // go.mod
   github.com/tmc/langchaingo v0.1.13
   ```

2. **Test thoroughly**: Always test migrations in a staging environment

3. **Monitor changes**: Subscribe to the [GitHub releases](https://github.com/tmc/langchaingo/releases)

4. **Update incrementally**: Don't skip major versions if possible

5. **Read release notes**: Check the changelog for each version

---

## Getting Help

If you encounter issues during migration:

1. Check the [GitHub Issues](https://github.com/tmc/langchaingo/issues)
2. Search for existing solutions
3. Create a new issue with:
   - Your current version
   - Target version
   - Specific error messages
   - Minimal reproduction code

---

## Version Compatibility Matrix

| LangChainGo Version | OpenAI Default Model | Milvus SDK | Go Version |
|-------------------|---------------------|------------|------------|
| v0.1.14+ | gpt-4o-mini | New SDK | 1.21+ |
| v0.1.13 | gpt-3.5-turbo | New SDK | 1.21+ |
| v0.1.12 | gpt-3.5-turbo | Old SDK | 1.21+ |

---

*This migration guide is updated regularly. Last updated: 2025-08-29*