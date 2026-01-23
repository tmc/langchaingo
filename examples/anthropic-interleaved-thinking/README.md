# Interleaved Thinking Example

This example demonstrates Claude 3.7+'s interleaved thinking capability, which allows the model to use thinking tokens between tool calls for better multi-step reasoning.

## What is Interleaved Thinking?

Interleaved thinking enables Claude to:
- **Think between tool calls**: Use reasoning tokens to plan which tool to use next
- **Interpret results**: Process tool outputs before deciding on next steps
- **Synthesize information**: Combine results from multiple tools coherently

## Features Demonstrated

1. **Multi-step problem solving** with quarterly sales analysis
2. **Tool orchestration** with calculate, search, and analyze functions
3. **Thinking tokens** used for planning and interpretation
4. **Token metrics** showing thinking vs visible output

## Running the Example

```bash
# Set your API key
export ANTHROPIC_API_KEY=your-api-key

# Run the example
go run .
```

## Key Implementation

```go
// Enable interleaved thinking for tool use
opts := []llms.CallOption{
    // Thinking mode for reasoning
    llms.WithThinkingMode(llms.ThinkingModeMedium),
    
    // Enable interleaved thinking beta feature
    anthropic.WithInterleavedThinking(),
    
    // Provide tools for the model to use
    llms.WithTools(tools),
}
```

## Example Flow

The demo presents a multi-step analysis task:
1. Calculate year-over-year growth rates
2. Analyze data trends
3. Search for explanatory factors
4. Make predictions based on findings

Between each tool call, Claude uses thinking tokens to:
- Decide which tool to use next
- Interpret the results from the previous tool
- Plan the next step in the analysis

## Token Metrics

The example displays detailed token usage:
- **Thinking Tokens**: Used for internal reasoning between tools
- **Visible Output**: The actual response content
- **Thinking Ratio**: Percentage of tokens used for thinking

## Requirements

- Claude 3.7 Sonnet model (`claude-3-7-sonnet-20250219`)
- Valid Anthropic API key
- Interleaved thinking feature access