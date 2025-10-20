# OpenAI o1/o3 Reasoning Example

This example demonstrates how to use OpenAI's o1 and o3 reasoning models with LangChainGo's unified reasoning API.

## Overview

OpenAI's o1 and o3 models are specifically designed for complex reasoning tasks. They use "reasoning effort" to control the depth of thinking, which affects both quality and token usage.

## Features Demonstrated

1. **Reasoning Strength Control**: Use `llms.WithReasoningStrength()` to control reasoning effort
   - `0.0 - 0.33`: Low effort (faster, fewer tokens)
   - `0.34 - 0.66`: Medium effort (balanced)
   - `0.67 - 1.0`: High effort (slower, more tokens, higher quality)

2. **Reasoning Token Tracking**: Extract reasoning token usage with `llms.ExtractReasoningUsage()`

3. **Model Capability Detection**: Use `llm.SupportsReasoning()` to check if a model supports reasoning

## Usage

```bash
# Use default settings (o1 model, high reasoning strength)
go run main.go

# Use o1-mini with medium reasoning
go run main.go -model=o1-mini -strength=0.5

# Use o3 with low reasoning (fastest)
go run main.go -model=o3 -strength=0.2

# Use o3-mini with high reasoning
go run main.go -model=o3-mini -strength=1.0
```

## Examples Included

### 1. Complex Math Problem
Optimization problem requiring step-by-step reasoning and calculus.

### 2. Logic Puzzle
Classic logic puzzle requiring understanding of context and inference.

### 3. Reasoning Strength Comparison
Demonstrates how different reasoning strengths affect token usage and response quality.

## Reasoning Strength Mapping

LangChainGo automatically maps the 0.0-1.0 strength value to OpenAI's reasoning_effort parameter:

| Strength Range | OpenAI reasoning_effort | Description |
|---------------|------------------------|-------------|
| 0.0 - 0.33    | "low"                  | Minimal reasoning, fastest |
| 0.34 - 0.66   | "medium"               | Balanced reasoning |
| 0.67 - 1.0    | "high"                 | Deep reasoning, highest quality |

## Cost Considerations

- **Reasoning tokens** are counted separately from output tokens
- Higher reasoning effort → more reasoning tokens → higher cost
- Use lower reasoning strength for simple problems to save costs
- Use higher reasoning strength for complex problems requiring careful analysis

## API Key

Set your OpenAI API key:
```bash
export OPENAI_API_KEY=your-api-key-here
```

## Output Format

The example shows:
1. The model's answer
2. Reasoning token usage
3. Raw generation info for debugging

## Cross-Provider Compatibility

This example uses LangChainGo's unified reasoning API. Similar code works with:
- **Anthropic Claude 4**: Use `llms.WithThinkingMode()` instead
- **Google Gemini 2.0+**: Use `llms.WithThinkingMode()` instead

See the reasoning.go documentation for details on cross-provider usage.

## Related Examples

- `openai-o1-example`: Basic o1 usage without reasoning controls
- `anthropic-interleaved-thinking`: Anthropic's thinking feature
- `structured-output-example`: Combining reasoning with structured output

## Learn More

- [OpenAI o1 Documentation](https://platform.openai.com/docs/guides/reasoning)
- [LangChainGo Reasoning API](../../llms/reasoning.go)
- [SDK Recommendations](../../SDK_RECOMMENDATIONS.md)
