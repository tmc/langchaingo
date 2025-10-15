package main

import (
	"fmt"
	"log"

	"github.com/vendasta/langchaingo/prompts"
)

func main() {
	fmt.Println("LangChain Go - Prompt Template Example")
	fmt.Println("=====================================\n")

	// Example 1: Basic template
	fmt.Println("1. Basic Template:")
	template := prompts.NewPromptTemplate(
		"Write a {{.length}} {{.style}} story about {{.topic}}.",
		[]string{"length", "style", "topic"},
	)

	result, err := template.Format(map[string]any{
		"length": "short",
		"style":  "funny",
		"topic":  "a robot learning to cook",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   %s\n\n", result)

	// Example 2: Using different template formats
	fmt.Println("2. Different Template Formats:")

	// F-string format (Python-style)
	fstringTemplate := prompts.PromptTemplate{
		Template:       "Hello {name}! Your score is {score}%.",
		InputVariables: []string{"name", "score"},
		TemplateFormat: prompts.TemplateFormatFString,
	}

	result, err = fstringTemplate.Format(map[string]any{
		"name":  "Alice",
		"score": 95,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   F-string: %s\n", result)

	// Jinja2 format
	jinja2Template := prompts.PromptTemplate{
		Template:       "Hello {{ name }}! Your score is {{ score }}%.",
		InputVariables: []string{"name", "score"},
		TemplateFormat: prompts.TemplateFormatJinja2,
	}

	result, err = jinja2Template.Format(map[string]any{
		"name":  "Bob",
		"score": 88,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Jinja2: %s\n\n", result)

	// Example 3: Partial variables
	fmt.Println("3. Partial Variables:")
	partialTemplate := prompts.PromptTemplate{
		Template:       "{{.greeting}}, {{.name}}! {{.message}}",
		InputVariables: []string{"name", "message"},
		TemplateFormat: prompts.TemplateFormatGoTemplate,
		PartialVariables: map[string]any{
			"greeting": "Welcome",
		},
	}

	result, err = partialTemplate.Format(map[string]any{
		"name":    "Charlie",
		"message": "Hope you're having a great day!",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   %s\n\n", result)

	// Example 4: Chat prompt template
	fmt.Println("4. Chat Prompt Template:")
	chatTemplate := prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
		prompts.NewSystemMessagePromptTemplate(
			"You are a helpful assistant that translates {{.input_language}} to {{.output_language}}.",
			[]string{"input_language", "output_language"},
		),
		prompts.NewHumanMessagePromptTemplate(
			"{{.text}}",
			[]string{"text"},
		),
	})

	messages, err := chatTemplate.FormatMessages(map[string]any{
		"input_language":  "English",
		"output_language": "French",
		"text":            "Hello, how are you?",
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, msg := range messages {
		fmt.Printf("   [%s]: %s\n", msg.GetType(), msg.GetContent())
	}

	// Example 5: Using sanitization for untrusted data
	fmt.Println("\n5. Optional Sanitization:")
	unsafeData := map[string]any{
		"user_input": "<script>alert('xss')</script>",
	}

	// Without sanitization (default)
	result, err = prompts.RenderTemplate(
		"User said: {{.user_input}}",
		prompts.TemplateFormatGoTemplate,
		unsafeData,
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Without sanitization: %s\n", result)

	// With sanitization
	result, err = prompts.RenderTemplate(
		"User said: {{.user_input}}",
		prompts.TemplateFormatGoTemplate,
		unsafeData,
		prompts.WithSanitization(),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   With sanitization: %s\n", result)
}
