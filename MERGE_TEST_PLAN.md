# Merge Test Plan - tmc/langchaingo upstream merge

This test plan covers areas affected by merge conflict resolutions and ensures no regressions were introduced.

## Priority 1: Critical Functionality Tests

### 1.1 GPT-4.1 Model Token Counting
**What changed:** Added back GPT-4.1 models with 1M+ token context windows

**Tests:**
- [ ] Test `GetModelContextSize("gpt-4.1")` returns `1047576`
- [ ] Test `GetModelContextSize("gpt-4.1-mini")` returns `1047576`
- [ ] Test `GetModelContextSize("gpt-4.1-nano")` returns `1047576`
- [ ] Test `CountTokens("gpt-4.1", "sample text")` works correctly
- [ ] Test `CalculateMaxTokens("gpt-4.1", "sample text")` returns correct value

**Run:**
```bash
go test ./llms -run TestCountTokens -v
go test ./llms -run TestCalculateMaxTokens -v
```

**Manual verification:**
```bash
# If you have services using GPT-4.1 models, verify they can calculate tokens
```

---

### 1.2 OpenAI User and ParallelToolCalls Options
**What changed:** Restored User and ParallelToolCalls fields to ChatRequest

**Tests:**
- [x] Unit tests pass (already verified)
- [ ] Integration test: Create LLM with `WithUser()` and verify it's sent to API
- [ ] Integration test: Create LLM with `WithParallelToolCalls(true)` and verify behavior
- [ ] Test with function calling to ensure parallel execution works

**Run:**
```bash
go test ./llms/openai -run TestWithUserAndParallelToolCalls -v
```

**Integration test (requires API key):**
```bash
export OPENAI_API_KEY="your-key"
go test ./llms/openai -run TestFunctionCall -v
```

---

### 1.3 OpenAI Metadata Filtering
**What changed:** Internal metadata (openai:*, thinking_config) now filtered before API calls

**Tests:**
- [ ] Test that `openai:use_legacy_max_tokens` metadata is NOT sent to API
- [ ] Test that `thinking_config` metadata is NOT sent to API
- [ ] Test that custom metadata IS sent to API
- [ ] Test that empty metadata after filtering becomes nil

**Test code:**
```go
// Create a test that mocks the API and verifies metadata filtering
opts := []llms.CallOption{
    llms.WithMetadata(map[string]any{
        "openai:internal": "should-be-filtered",
        "custom_field": "should-be-sent",
    }),
}
// Verify only custom_field makes it to the API request
```

---

## Priority 2: Token Tracking & Reporting

### 2.1 Enhanced Token Usage Reporting
**What changed:** Added comprehensive token tracking fields

**Tests:**
- [ ] Test response includes `ThinkingContent` field
- [ ] Test response includes `ThinkingTokens` field
- [ ] Test response includes `PromptCachedTokens` field
- [ ] Test response includes `CompletionAudioTokens` field
- [ ] Test response includes prediction token fields
- [ ] Verify backward compatibility - old fields still present

**Run:**
```bash
# Integration test with o1 model for reasoning tokens
export OPENAI_API_KEY="your-key"
go test ./llms/openai -run TestOpenAI -v
```

**Manual check:**
```go
// In your services, after calling GenerateContent, inspect GenerationInfo:
response, _ := llm.GenerateContent(ctx, messages)
info := response.Choices[0].GenerationInfo
fmt.Printf("Thinking tokens: %v\n", info["ThinkingTokens"])
fmt.Printf("Cached tokens: %v\n", info["PromptCachedTokens"])
```

---

### 2.2 O1/O3 Model System Message Handling
**What changed:** Added ModelCapability system that handles models without system message support

**Tests:**
- [ ] Test o1-preview automatically converts system messages to user messages
- [ ] Test o1-mini automatically converts system messages to user messages
- [ ] Test o3 models handle system messages correctly
- [ ] Test regular GPT-4 models still use system messages normally
- [ ] Verify system content is prepended to first user message for o1/o3

**Run:**
```bash
# If you use o1 models, test with system messages
export OPENAI_API_KEY="your-key"
# Create test with system message and verify no API errors
```

---

## Priority 3: Google Vertex AI

### 3.1 Vertex Tool Conversion
**What changed:** Fixed duplicate genaiTools declaration

**Tests:**
- [ ] Test `convertTools()` with empty tool list returns nil
- [ ] Test `convertTools()` with single tool works
- [ ] Test `convertTools()` with multiple tools works
- [ ] Test tool calling with Vertex AI models

**Run:**
```bash
go test ./llms/googleai/vertex -v
```

**Integration (requires GCP credentials):**
```bash
export GOOGLE_APPLICATION_CREDENTIALS="path/to/creds.json"
go test ./llms/googleai/vertex -run TestVertex -v
```

---

## Priority 4: Updated Models & Context Sizes

### 4.1 GPT Model Context Sizes
**What changed:** Updated context sizes for various models

**Verify:**
- [ ] GPT-3.5-turbo: 4096 → 16385 ✓
- [ ] GPT-4-turbo: 128000 ✓
- [ ] GPT-4o: 128000 ✓
- [ ] GPT-4o-mini: 128000 ✓

**Impact check:**
```bash
# If your services have hardcoded context size assumptions, verify they work with new sizes
# Check for any max_tokens calculations that might be affected
grep -r "4096" your-service/
grep -r "MaxTokens" your-service/
```

---

## Priority 5: Regression Testing

### 5.1 Existing Functionality
**Run existing test suites:**

```bash
# Core LLM tests
go test ./llms/... -v

# OpenAI specific tests
go test ./llms/openai/... -v

# Google AI tests  
go test ./llms/googleai/... -v

# Chain tests (use LLMs)
go test ./chains/... -v

# Agent tests (use LLMs)
go test ./agents/... -v

# Memory tests
go test ./memory/... -v
```

---

## Priority 6: Service Integration Tests

### 6.1 Your Production Services
For each service using langchaingo:

**Pre-deployment checklist:**
- [ ] Add `replace` directive to go.mod
- [ ] Update to use new fork version
- [ ] Run service's test suite
- [ ] Verify no import errors
- [ ] Check logs for any new warnings/errors

**Example replace directive:**
```go
// go.mod
replace github.com/tmc/langchaingo => github.com/vendasta/langchaingo v0.x.x
```

### 6.2 Common Integration Points
- [ ] LLM initialization with custom options
- [ ] Function/tool calling workflows
- [ ] Streaming responses
- [ ] Token counting for billing/metrics
- [ ] System message usage patterns
- [ ] Multi-turn conversations with memory

---

## Quick Smoke Test Script

```bash
#!/bin/bash
# Run this for quick validation

echo "=== Building all packages ==="
go build ./... || exit 1

echo "=== Running unit tests ==="
go test ./llms/count_tokens_test.go -v || exit 1
go test ./llms/openai -run TestWithUserAndParallelToolCalls -v || exit 1

echo "=== Checking for common issues ==="
# Check if any files still reference vendasta in imports
echo "Checking for vendasta imports..."
VENDASTA_COUNT=$(grep -r "github.com/vendasta/langchaingo" --include="*.go" . | wc -l)
if [ $VENDASTA_COUNT -gt 0 ]; then
    echo "WARNING: Found vendasta imports in go files!"
    grep -r "github.com/vendasta/langchaingo" --include="*.go" .
    exit 1
fi

echo "=== Basic smoke test passed! ==="
echo "Next: Run integration tests with API keys"
```

---

## Known Changes to Document

### Breaking Changes: NONE ✓
All changes are backward compatible.

### New Features Available:
1. **GPT-4.1 model support** - Use models with 1M+ token contexts
2. **Enhanced token tracking** - Get detailed breakdowns including:
   - Thinking/reasoning tokens (o1/o3 models)
   - Cached prompt tokens (when caching enabled)
   - Audio tokens (for audio-enabled models)
   - Prediction tokens (for speculative decoding)
3. **Automatic o1/o3 handling** - System messages automatically converted
4. **Better metadata filtering** - Internal flags don't leak to API

### Deprecations: NONE ✓

---

## Test Priority by Service Type

**If your service uses:**

| Feature | Priority Tests |
|---------|---------------|
| GPT-4.1 models | 1.1, 4.1 |
| Function calling | 1.2, 1.3 |
| o1/o3 reasoning models | 2.1, 2.2 |
| Vertex AI | 3.1 |
| Token counting/billing | 1.1, 2.1, 4.1 |
| System messages | 2.2 |
| Custom metadata | 1.3 |

---

## Rollback Plan

If issues are found:

1. **Immediate:** Don't deploy to production
2. **Investigation:** Check specific test that failed
3. **Quick fix:** If minor, fix and retest
4. **Rollback option:** Revert to previous fork version
   ```bash
   git revert <merge-commit>
   # Or checkout previous stable tag
   ```

---

## Sign-off Checklist

Before merging to main/deploying:

- [ ] All Priority 1 tests pass
- [ ] No regression in existing tests
- [ ] Integration tests with real API pass
- [ ] At least one service successfully tested
- [ ] Team reviewed merge changes
- [ ] Documentation updated (if needed)
- [ ] Rollback plan communicated

---

## Notes

- Tests marked with `[x]` have already been verified
- Integration tests require API keys (set env vars)
- Some tests may incur small API costs
- The User/ParallelToolCalls fix is critical if you use these features

