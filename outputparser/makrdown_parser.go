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
	// Extract content from ```markdown ... ``` blocks
	_, afterStart, ok := strings.Cut(text, "```markdown\n")
	if !ok {
		return "", ParseError{Text: text, Reason: "no ```markdown at start of output"}
	}

	content, _, ok := strings.Cut(afterStart, "```")
	if !ok {
		return "", ParseError{Text: text, Reason: "no closing ``` at end of output"}
	}

	// Trim whitespace from content
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
