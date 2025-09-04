package prompts

import (
	"embed"
	"fmt"
	"log"
)

// Example_basicTemplateRendering demonstrates basic template usage with automatic security.
func Example_basicTemplateRendering() {
	// Basic template rendering - all formats supported
	data := map[string]any{
		"name":    "Alice",
		"role":    "Developer",
		"company": "Acme Corp",
	}

	// F-String format (fastest, simple variable substitution)
	result, err := RenderTemplate(
		"Hello {name}! You're a {role} at {company}.",
		TemplateFormatFString,
		data,
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("F-String:", result)

	// Go Template format (recommended, supports conditionals and loops)
	result, err = RenderTemplate(
		"Hello {{.name}}! You're a {{.role}} at {{.company}}.",
		TemplateFormatGoTemplate,
		data,
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Go Template:", result)

	// Output:
	// F-String: Hello Alice! You're a Developer at Acme Corp.
	// Go Template: Hello Alice! You're a Developer at Acme Corp.
}

// Example_optionalSecurity demonstrates how to enable security for untrusted input.
func Example_optionalSecurity() {
	// User input that contains potential security threats
	userInput := map[string]any{
		"username": "alice<script>alert('xss')</script>",
		"bio":      "I love coding! <b>Bold text</b>",
		"website":  "https://example.com",
	}

	template := `
Profile for {{.username}}
Bio: {{.bio}}
Website: {{.website}}`

	// Enable sanitization for untrusted user input
	result, err := RenderTemplate(template, TemplateFormatGoTemplate, userInput, WithSanitization())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)

	// Output:
	// Profile for alice&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;
	// Bio: I love coding! &lt;b&gt;Bold text&lt;/b&gt;
	// Website: https://example.com
}

// Example_templateWithLogic demonstrates templates with conditionals and loops.
func Example_templateWithLogic() {
	data := map[string]any{
		"user": map[string]any{
			"name":  "Alice",
			"score": 92,
		},
		"items": []string{"apple", "banana", "cherry"},
	}

	template := `Welcome {{.user.name}}!

{{if gt .user.score 90 -}}
🌟 Excellent performance! ({{.user.score}}%)
{{- else -}}
📈 Keep up the good work! ({{.user.score}}%)
{{- end}}

Your items:
{{range .items}}• {{.}}
{{end}}`

	result, err := RenderTemplate(template, TemplateFormatGoTemplate, data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)

	// Output:
	// Welcome Alice!
	//
	// 🌟 Excellent performance! (92%)
	//
	// Your items:
	// • apple
	// • banana
	// • cherry
}

//go:embed testdata/*.j2
var templateFiles embed.FS

// Example_templateWithIncludes demonstrates safe template composition with filesystem access.
func Example_templateWithIncludes() {
	// Using embed.FS ensures templates are bundled at compile time
	// and provides a secure filesystem boundary

	data := map[string]any{
		"title":   "Welcome Guide",
		"user":    "Alice",
		"company": "Acme Corp",
	}

	// This template can safely include other templates within the embedded filesystem
	result, err := RenderTemplateFS(
		templateFiles,
		"testdata/main.j2", // Template that includes other templates
		TemplateFormatJinja2,
		data,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)

	// Output:
	// Welcome Guide
	// Hello Alice! Welcome to Acme Corp.
	// This is a safe template composition example.
}

// Example_promptTemplate demonstrates using PromptTemplate for LLM integration.
func Example_promptTemplate() {
	// Create a prompt template for an AI assistant
	template := NewPromptTemplate(
		`You are a helpful assistant for {{.company}}.
User: {{.username}} ({{.role}})
Query: {{.query}}

Please provide a helpful response appropriate for their role.`,
		[]string{"company", "username", "role", "query"},
	)

	// User input (automatically secured)
	data := map[string]any{
		"company":  "Acme Corp",
		"username": "alice",
		"role":     "developer",
		"query":    "How do I optimize database queries?",
	}

	prompt, err := template.FormatPrompt(data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(prompt.String())

	// Output:
	// You are a helpful assistant for Acme Corp.
	// User: alice (developer)
	// Query: How do I optimize database queries?
	//
	// Please provide a helpful response appropriate for their role.
}

// Example_errorHandling demonstrates proper error handling with templates.
func Example_errorHandling() {
	// This will demonstrate different types of errors

	// 1. Invalid template syntax
	_, err := RenderTemplate("Hello {{.name", TemplateFormatGoTemplate, map[string]any{"name": "Alice"})
	if err != nil {
		fmt.Println("Template syntax error:", err)
	}

	// 2. Missing required variable
	_, err = RenderTemplate("Hello {{.name}}!", TemplateFormatGoTemplate, map[string]any{})
	if err != nil {
		fmt.Println("Missing variable error:", err)
	}

	// 3. Invalid identifier format (starts with number)
	_, err = RenderTemplate("{{.data}}", TemplateFormatGoTemplate, map[string]any{"123invalid": "value"}, WithSanitization())
	if err != nil {
		fmt.Println("Invalid variable name:", err)
	}

	// Output:
	// Template syntax error: template parse failure: template: template:1: unclosed action
	// Missing variable error: template execution failure: template: template:1:8: executing "template" at <.name>: map has no entry for key "name"
	// Invalid variable name: template execution failure: template validation failure: invalid variable name: 123invalid
}

// Example_migration demonstrates migrating to the new security model.
func Example_migration() {
	// OLD WAY: Templates rendered without sanitization by default
	// result, _ := RenderTemplate(template, format, userInput)

	// NEW WAY: Enable sanitization when needed for untrusted input
	userInput := map[string]any{
		"name": "<script>alert('xss')</script>",
	}

	template := "Hello {{.name}}!"

	// Enable sanitization for untrusted data
	result, err := RenderTemplate(template, TemplateFormatGoTemplate, userInput, WithSanitization())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)

	// Output:
	// Hello &lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;!
}
