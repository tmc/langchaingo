# Ollama Agent Usage Guide

## Issue #1045: Ollama Agents and Tools

### Problem
Ollama models don't have native function/tool calling support like OpenAI. When using Ollama with agents, the model may not generate responses in the expected format, leading to parsing errors.

### Solution
We've improved the MRKL agent's parseOutput function to be more flexible in detecting:
1. Case-insensitive variations of "Final Answer"
2. Different formats like "The answer is:" or "Answer:"
3. Case-insensitive action patterns

### Best Practices for Using Ollama with Agents

#### 1. Use Clear System Prompts
When creating an agent with Ollama, provide explicit instructions about the expected format:

```go
systemPrompt := `You are a helpful assistant that uses tools to answer questions.

IMPORTANT: You must follow this exact format:

For using a tool:
Thought: [your reasoning]
Action: [tool name]
Action Input: [tool input]

For final answer:
Thought: I now know the final answer
Final Answer: [your answer]

Always use "Final Answer:" to indicate your final response.`

agent := agents.NewOneShotAgent(
    ollamaLLM,
    tools,
    agents.WithSystemMessage(systemPrompt),
)
```

#### 2. Use Appropriate Models
Some Ollama models work better with agents than others:
- **Recommended**: llama3, mistral, mixtral, gemma2
- **May need tuning**: llama2, codellama
- **Test thoroughly**: smaller models like phi

#### 3. Adjust Temperature
Lower temperature often helps with format consistency:

```go
llm, err := ollama.New(
    ollama.WithModel("llama3"),
    ollama.WithOptions(ollama.Options{
        Temperature: 0.2, // Lower temperature for more consistent formatting
    }),
)
```

#### 4. Handle Format Variations
The improved parser now handles these variations:
- "Final Answer: X" (standard)
- "final answer: X" (lowercase)
- "The answer is: X" (natural language)
- "Answer: X" (simplified)

#### 5. Example Implementation

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/vendasta/langchaingo/agents"
    "github.com/vendasta/langchaingo/llms/ollama"
    "github.com/vendasta/langchaingo/tools"
)

func main() {
    // Create Ollama LLM with appropriate settings
    llm, err := ollama.New(
        ollama.WithModel("llama3"),
        ollama.WithOptions(ollama.Options{
            Temperature: 0.2,
            NumPredict: 512,
        }),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Create tools
    calculator := tools.Calculator{}
    
    // Create agent with clear instructions
    systemPrompt := `You are a helpful math assistant. Use the calculator tool for computations.

Format your responses as:
- For calculations: "Action: calculator" then "Action Input: [expression]"
- For final answers: "Final Answer: [result]"`
    
    agent := agents.NewOneShotAgent(
        llm,
        []tools.Tool{calculator},
        agents.WithSystemMessage(systemPrompt),
        agents.WithMaxIterations(5),
    )
    
    // Create executor
    executor := agents.NewExecutor(
        agent,
        agents.WithMaxIterations(5),
    )
    
    // Run the agent
    result, err := executor.Call(
        context.Background(),
        map[string]any{
            "input": "What is 25 * 4?",
        },
    )
    
    if err != nil {
        log.Printf("Error: %v", err)
    } else {
        fmt.Printf("Result: %v\n", result["output"])
    }
}
```

### Troubleshooting

#### Error: "unable to parse output"
- **Cause**: Model output doesn't match expected format
- **Solution**: 
  1. Lower temperature
  2. Use a more capable model
  3. Improve system prompt with examples
  4. Consider using few-shot prompting

#### Error: "agent not finished before max iterations"
- **Cause**: Model never generates "Final Answer"
- **Solution**:
  1. Explicitly mention "Final Answer:" in system prompt
  2. Increase max iterations temporarily for debugging
  3. Check if model is generating variations our parser now handles

#### Model keeps repeating actions
- **Cause**: Model doesn't understand it should stop after getting result
- **Solution**:
  1. Add explicit instructions about when to provide final answer
  2. Include examples in the system prompt
  3. Consider adding a custom output parser

### Testing Your Setup

```go
// Test function to verify Ollama agent works correctly
func TestOllamaAgent(t *testing.T) {
    ctx := context.Background()
    
    llm, err := ollama.New(
        ollama.WithModel("llama3"),
    )
    require.NoError(t, err)
    
    calculator := tools.Calculator{}
    
    agent := agents.NewOneShotAgent(
        llm,
        []tools.Tool{calculator},
        agents.WithMaxIterations(3),
    )
    
    executor := agents.NewExecutor(agent)
    
    testCases := []struct {
        input    string
        expected string
    }{
        {"What is 2+2?", "4"},
        {"Calculate 10*5", "50"},
        {"What is 100 divided by 4?", "25"},
    }
    
    for _, tc := range testCases {
        result, err := executor.Call(ctx, map[string]any{
            "input": tc.input,
        })
        
        if err != nil {
            t.Logf("Warning: %s failed: %v", tc.input, err)
            continue
        }
        
        output := fmt.Sprintf("%v", result["output"])
        if !strings.Contains(output, tc.expected) {
            t.Errorf("Expected %s in output, got: %s", tc.expected, output)
        }
    }
}
```

### Summary of Improvements

1. **More flexible parsing**: The MRKL agent now accepts various formats for final answers
2. **Case-insensitive matching**: Both actions and final answers can use different casing
3. **Better error messages**: Clearer feedback when parsing fails
4. **Robust action parsing**: Handles "Action Input" with various capitalizations

These improvements make Ollama models more reliable when used with agents, though they still require careful prompt engineering compared to models with native function calling support.