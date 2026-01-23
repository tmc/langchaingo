# Extended Capabilities Example

This example demonstrates Claude 3.7+'s combined extended capabilities:
- **Extended Thinking**: Deep reasoning with configurable thinking budget
- **Extended Output**: Generate up to 128K tokens (vs standard 8K limit)

## Features Demonstrated

1. **Combined Capabilities**: Shows how to use both extended thinking AND extended output together
2. **Complex Task**: Generates a comprehensive distributed systems guide requiring both deep reasoning and extensive output
3. **Token Metrics**: Displays detailed token usage including thinking tokens and output tokens

## Running the Example

```bash
# Set your API key
export ANTHROPIC_API_KEY=your-api-key

# Run the example
go run .
```

## Key Implementation Points

```go
// Enable both capabilities together
opts := []llms.CallOption{
    // Extended thinking for complex reasoning
    llms.WithThinkingMode(llms.ThinkingModeHigh),
    
    // Extended output for up to 128K tokens
    anthropic.WithExtendedOutput(),
    
    // Set high token limit
    llms.WithMaxTokens(50000),
}
```

## What to Expect

- The model will use extended thinking to reason about the complex distributed systems topic
- It will generate a comprehensive guide that may exceed standard token limits
- Token metrics will show both thinking tokens used and total output generated
- If the response is large (>10K chars), you'll have the option to save it to a file

## Requirements

- Claude 3.7 Sonnet model (`claude-3-7-sonnet-20250219`)
- Valid Anthropic API key with access to Claude 3.7
- Both extended thinking and extended output features enabled on your account