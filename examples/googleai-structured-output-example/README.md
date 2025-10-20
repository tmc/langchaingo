# Google AI Structured Output Example

This example demonstrates how to use Google AI (Gemini) with structured output support in LangChainGo.

## Overview

Structured output allows you to define a JSON schema and have the model return responses that conform to that schema. This is useful for:

- Data extraction from unstructured text
- Generating structured data (recipes, user profiles, etc.)
- Ensuring consistent output format
- Type-safe parsing of LLM responses

## Features Demonstrated

1. **Simple Object Extraction**: Extract user information from text
2. **Nested Arrays**: Generate recipes with ingredients and steps
3. **Nested Objects**: Extract company and CEO information
4. **Enum Validation**: Sentiment analysis with constrained values

## Prerequisites

- Google AI API key (set as `GOOGLE_API_KEY` environment variable)
- Go 1.23 or higher

## Running the Example

```bash
export GOOGLE_API_KEY=your-api-key-here
go run main.go
```

## How It Works

### 1. Define a Schema

```go
schema := &llms.StructuredOutputDefinition{
    Name:        "user_info",
    Description: "Extract user information",
    Schema: &llms.StructuredOutputSchema{
        Type: llms.SchemaTypeObject,
        Properties: map[string]*llms.StructuredOutputSchema{
            "name": {
                Type:        llms.SchemaTypeString,
                Description: "User's full name",
            },
            "age": {
                Type:        llms.SchemaTypeInteger,
                Description: "User's age in years",
            },
        },
        Required: []string{"name", "age"},
    },
}
```

### 2. Use WithStructuredOutput Option

```go
resp, err := llm.GenerateContent(ctx, messages,
    llms.WithStructuredOutput(schema),
    llms.WithMaxTokens(200),
)
```

### 3. Parse JSON Response

```go
var userInfo struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}
json.Unmarshal([]byte(resp.Choices[0].Content), &userInfo)
```

## Schema Features

### Supported Types

- `SchemaTypeObject`: Objects with properties
- `SchemaTypeArray`: Arrays of items
- `SchemaTypeString`: String values
- `SchemaTypeInteger`: Integer numbers
- `SchemaTypeNumber`: Floating-point numbers
- `SchemaTypeBoolean`: Boolean values

### Advanced Features

- **Nested Objects**: Objects can contain other objects
- **Nested Arrays**: Arrays can contain objects or primitives
- **Required Fields**: Mark fields as required
- **Enum Values**: Constrain string values to specific options
- **Descriptions**: Add descriptions for better extraction

## Cross-Provider Compatibility

This same API works with:

- **Google AI (Gemini)**: Native support via ResponseSchema
- **OpenAI**: Native support via ResponseFormat
- **Anthropic**: Simulated via tool calling

You can write provider-agnostic code and switch providers without changing your schema definitions.

## Example Output

```
Example 1: User Information Extraction
---------------------------------------
Input: Extract information: Sarah Johnson is a 28-year-old software engineer...

Extracted Information:
  Name: Sarah Johnson
  Age: 28
  Email: sarah.j@techcorp.com
  Occupation: software engineer

Example 2: Recipe Extraction
-----------------------------
Input: Give me a simple recipe for chocolate chip cookies

Recipe: Classic Chocolate Chip Cookies
Servings: 24
Prep Time: 15 minutes

Ingredients:
  1. 2 1/4 cups all-purpose flour
  2. 1 tsp baking soda
  3. 1 cup butter, softened
  ...

Steps:
  1. Preheat oven to 375°F
  2. Mix flour and baking soda in a bowl
  3. Beat butter and sugars until creamy
  ...
```

## Notes

- Google's structured output requires `gemini-2.0-flash` or newer models
- The model automatically sets `ResponseMIMEType` to `application/json`
- Validation errors are caught early with `llms.ValidateSchema()`
- Complex schemas with deep nesting are supported

## Learn More

- [LangChainGo Documentation](https://github.com/tmc/langchaingo)
- [Google AI Documentation](https://ai.google.dev/docs)
- [Structured Output Guide](../../llms/structured.go)
