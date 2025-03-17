package outputparser

import (
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

const (
	// _markdownFormatInstruction is the instruction for the LLM to format output as markdown.
	_markdownFormatInstruction = "The output should be formatted in markdown: \n```markdown\nYour content here\n```"
)

// Markdown is a simple output parser that extracts content from markdown code blocks.
type Markdown struct{}

// NewMarkdown creates a new markdown output parser.
func NewMarkdown() Markdown {
	return Markdown{}
}

// Statically assert that Markdown implements the OutputParser interface.
var _ schema.OutputParser[any] = Markdown{}

// parse extracts content from markdown code blocks in the given text.
func (p Markdown) parse(text string) (string, error) {
	// Find the starting ```markdown
	startIdx := strings.Index(text, "```markdown\n")
	if startIdx == -1 {
		return "", ParseError{Text: text, Reason: "no ```markdown at start of output"}
	}

	// Move to the content starting position
	contentStart := startIdx + len("```markdown\n")

	// Find the last ```
	lastBacktickIdx := strings.LastIndex(text, "```")
	if lastBacktickIdx <= contentStart {
		return "", ParseError{Text: text, Reason: "no closing ``` at end of output"}
	}

	// Extract the content
	content := text[contentStart:lastBacktickIdx]

	// Trim whitespace
	content = strings.TrimSpace(content)

	return content, nil
}

// Parse extracts content from markdown code blocks.
func (p Markdown) Parse(text string) (any, error) {
	return p.parse(text)
}

// ParseWithPrompt does the same as Parse.
func (p Markdown) ParseWithPrompt(text string, _ llms.PromptValue) (any, error) {
	return p.parse(text)
}

// GetFormatInstructions returns a string explaining how the output should be formatted.
func (p Markdown) GetFormatInstructions() string {
	return _markdownFormatInstruction
}

// Type returns the type of the output parser.
func (p Markdown) Type() string {
	return "markdown_parser"
}
