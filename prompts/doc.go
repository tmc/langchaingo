// Package prompts provides utilities for creating and managing prompts for Large Language Models (LLMs).
//
// # Basic Usage
//
// The simplest way to use this package is with prompt templates:
//
//	// Create a basic prompt template
//	template := prompts.NewPromptTemplate(
//		"Write a {{.style}} story about {{.topic}}.",
//		[]string{"style", "topic"},
//	)
//
//	// Format the template with values
//	result, err := template.Format(map[string]any{
//		"style": "funny",
//		"topic": "a robot learning to cook",
//	})
//	// Result: "Write a funny story about a robot learning to cook."
//
// # Template Formats
//
// Three template formats are supported:
//
//   - Go Templates (default): `{{ .variable }}` - Native Go text/template with sprig functions
//   - Jinja2: `{{ variable }}` - Python-style templates with filters and logic
//   - F-Strings: `{variable}` - Simple Python-style variable substitution
//
// Example using different formats:
//
//	// Go template (default)
//	goTemplate := prompts.NewPromptTemplate(
//		"Hello {{ .name }}!",
//		[]string{"name"},
//	)
//
//	// Jinja2 format
//	jinja2Template := prompts.PromptTemplate{
//		Template:       "Hello {{ name }}!",
//		InputVariables: []string{"name"},
//		TemplateFormat: prompts.TemplateFormatJinja2,
//	}
//
//	// F-string format
//	fstringTemplate := prompts.PromptTemplate{
//		Template:       "Hello {name}!",
//		InputVariables: []string{"name"},
//		TemplateFormat: prompts.TemplateFormatFString,
//	}
//
// # Chat Prompts
//
// For conversational AI, use ChatPromptTemplate:
//
//	chatTemplate := prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
//		prompts.NewSystemMessagePromptTemplate(
//			"You are a helpful assistant.",
//			nil,
//		),
//		prompts.NewHumanMessagePromptTemplate(
//			"{{.question}}",
//			[]string{"question"},
//		),
//	})
//
//	messages, err := chatTemplate.FormatMessages(map[string]any{
//		"question": "What is the capital of France?",
//	})
//
// # Partial Variables
//
// Pre-fill some template variables while leaving others for runtime:
//
//	template := prompts.PromptTemplate{
//		Template:       "{{.greeting}}, {{.name}}!",
//		InputVariables: []string{"name"},
//		PartialVariables: map[string]any{
//			"greeting": "Welcome",
//		},
//	}
//
//	result, err := template.Format(map[string]any{
//		"name": "Alice",
//	})
//	// Result: "Welcome, Alice!"
//
// # Security
//
// By default, templates render without sanitization for backward compatibility.
// When working with untrusted user input, enable HTML escaping:
//
//	// Enable sanitization for untrusted data
//	result, err := prompts.RenderTemplate(
//		"User said: {{.input}}",
//		prompts.TemplateFormatGoTemplate,
//		map[string]any{"input": userInput},
//		prompts.WithSanitization(), // Escapes HTML special characters
//	)
//
// Templates always block filesystem access for security. Use RenderTemplateFS
// for controlled template inheritance:
//
//	//go:embed templates/*
//	var templateFS embed.FS
//
//	result, err := prompts.RenderTemplateFS(
//		templateFS,
//		"email.j2",
//		prompts.TemplateFormatJinja2,
//		data,
//	)
//
// # Advanced Features
//
// # Few-Shot Learning
//
// Create prompts with examples for better model performance:
//
//	fewShot := &prompts.FewShotPrompt{
//		ExamplePrompt: prompts.NewPromptTemplate(
//			"Q: {{.question}}\nA: {{.answer}}",
//			[]string{"question", "answer"},
//		),
//		Examples: []map[string]string{
//			{"question": "What is 2+2?", "answer": "4"},
//			{"question": "What is 3+3?", "answer": "6"},
//		},
//		Suffix: "\nQ: {{.question}}\nA:",
//		InputVariables: []string{"question"},
//	}
//
// # Performance Considerations
//
// - Go templates are fastest and recommended for production
// - Template compilation is cached for repeated use
// - Use RenderTemplateFS with embed.FS for optimal production deployments
package prompts
