# Prompt Template Example

This example demonstrates the core features of LangChain Go's prompt template system.

## Features Demonstrated

1. **Basic Templates** - Simple variable substitution using Go templates
2. **Template Formats** - Using different template syntaxes (Go, Jinja2, F-string)
3. **Partial Variables** - Pre-filling some template variables
4. **Chat Templates** - Creating structured chat prompts
5. **Optional Sanitization** - HTML escaping for untrusted data

## Running the Example

```bash
go run main.go
```

## Key Concepts

### Template Formats

LangChain Go supports three template formats:
- **Go Templates** (default): `{{ .variable }}`
- **Jinja2**: `{{ variable }}`
- **F-strings**: `{variable}`

### Security

By default, templates render without sanitization for maximum compatibility. When working with untrusted user input, enable sanitization:

```go
result, err := prompts.RenderTemplate(
    template,
    prompts.TemplateFormatGoTemplate,
    data,
    prompts.WithSanitization(), // Enables HTML escaping
)
```

### Partial Variables

Partial variables let you pre-fill template values:

```go
template := prompts.PromptTemplate{
    Template: "{{.greeting}}, {{.name}}!",
    InputVariables: []string{"name"},
    PartialVariables: map[string]any{
        "greeting": "Hello",
    },
}
```