# Anthropic Structured Output Example

This example demonstrates how to use structured output with the Anthropic provider in LangChainGo.

## Overview

Unlike OpenAI and Google, Anthropic does NOT have native structured output support. This implementation **simulates** structured output using tool calling:

1. The schema is converted to a tool definition
2. The model is forced to call that tool using `ToolChoice`
3. The JSON from the tool call is extracted and returned as structured output

From the user's perspective, the API is **identical** to OpenAI and Google:

```go
resp, err := llm.GenerateContent(ctx, messages,
    llms.WithStructuredOutput(schema),
)
```

## How It Works

### Backend Implementation

When you call `WithStructuredOutput()` on Anthropic:

1. **Schema → Tool**: The schema is converted to an Anthropic tool definition
2. **Force Tool Use**: `ToolChoice` is set to force calling this specific tool
3. **Extract JSON**: The tool's input arguments are returned as the structured output

### Comparison with Native Support

**OpenAI (Native):**
- Uses `response_format` with `json_schema` type
- Strict mode available for exact matching
- Guaranteed JSON output

**Google (Native):**
- Uses `ResponseSchema` parameter
- Built into the generation API
- Native JSON mode

**Anthropic (Simulated):**
- Uses tool calling under the hood
- Converts schema → tool definition
- Extracts JSON from tool arguments
- **Transparent to the user** - same API

## Running the Example

Set your Anthropic API key:

```bash
export ANTHROPIC_API_KEY=your_api_key_here
```

Run the example:

```bash
go run main.go
```

## Examples in This File

### Example 1: Extract User Information

Extracts structured user data from natural language text:

```go
userSchema := &llms.StructuredOutputDefinition{
    Name: "user_info",
    Schema: &llms.StructuredOutputSchema{
        Type: llms.SchemaTypeObject,
        Properties: map[string]*llms.StructuredOutputSchema{
            "name":  {Type: llms.SchemaTypeString},
            "age":   {Type: llms.SchemaTypeInteger},
            "email": {Type: llms.SchemaTypeString},
        },
        Required: []string{"name", "age"},
    },
}

resp, _ := llm.GenerateContent(ctx, messages,
    llms.WithStructuredOutput(userSchema),
)

var user UserInfo
json.Unmarshal([]byte(resp.Choices[0].Content), &user)
```

### Example 2: Generate Recipe

Generates a structured recipe with arrays and nested data:

```go
recipeSchema := &llms.StructuredOutputDefinition{
    Name: "recipe",
    Schema: &llms.StructuredOutputSchema{
        Type: llms.SchemaTypeObject,
        Properties: map[string]*llms.StructuredOutputSchema{
            "title":       {Type: llms.SchemaTypeString},
            "ingredients": {
                Type:  llms.SchemaTypeArray,
                Items: &llms.StructuredOutputSchema{Type: llms.SchemaTypeString},
            },
            "steps": {
                Type:  llms.SchemaTypeArray,
                Items: &llms.StructuredOutputSchema{Type: llms.SchemaTypeString},
            },
        },
        Required: []string{"title", "ingredients", "steps"},
    },
}
```

## Limitations

Since this is **simulated** via tool calling:

1. **Performance**: Slightly more latency due to tool call overhead
2. **Tokens**: Uses additional tokens for tool definition
3. **Not 100% Guaranteed**: While highly reliable, it's not as strict as OpenAI's native mode
4. **Tool Call Mechanics**: Under the hood, it's using Anthropic's tool calling API

## When to Use

Use structured output when you need:

- Consistent JSON responses
- Type-safe parsing
- Validation against a schema
- Multi-provider code (same API for OpenAI, Google, Anthropic)

## Best Practices

1. **Define Required Fields**: Always specify which fields are required
2. **Add Descriptions**: Help the model understand what each field represents
3. **Set MaxTokens**: Ensure enough tokens for the full response
4. **Validate Output**: Even with structured output, validate the parsed result

## See Also

- OpenAI structured output example: `../openai-structured-output-example/`
- Google structured output example: `../googleai-structured-output-example/`
- Anthropic tool calling: `../anthropic-tool-call-example/`
