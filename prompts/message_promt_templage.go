package prompts

import (
	"github.com/tmc/langchaingo/load"
	"github.com/tmc/langchaingo/schema"
)

// SystemMessagePromptTemplate is a message formatter that returns a system message.
type SystemMessagePromptTemplate struct {
	Prompt PromptTemplate `json:"systemPrompt,omitempty"`
}

var _ MessageFormatter = SystemMessagePromptTemplate{}

// FormatMessages formats the message with the values given.
func (p SystemMessagePromptTemplate) FormatMessages(values map[string]any) ([]schema.ChatMessage, error) {
	text, err := p.Prompt.Format(values)
	return []schema.ChatMessage{schema.SystemChatMessage{Text: text}}, err
}

// GetInputVariables returns the input variables the prompt expects.
func (p SystemMessagePromptTemplate) GetInputVariables() []string {
	return p.Prompt.InputVariables
}

func (p SystemMessagePromptTemplate) Save(path string) error {
	err := load.ToFile(p, path)
	if err != nil {
		return err
	}
	return nil
}

// NewSystemMessagePromptTemplate creates a new system message prompt template.
func NewSystemMessagePromptTemplate(template string, inputVariables []string) SystemMessagePromptTemplate {
	return SystemMessagePromptTemplate{
		Prompt: NewPromptTemplate(template, inputVariables),
	}
}

func NewSystemMessagePromptFromFile(path string) (SystemMessagePromptTemplate, error) {
	var promptTemplate SystemMessagePromptTemplate
	err := load.FromFile(&promptTemplate, path)
	if err != nil {
		return SystemMessagePromptTemplate{}, err
	}
	return promptTemplate, nil
}

// AIMessagePromptTemplate is a message formatter that returns a AI message.
type AIMessagePromptTemplate struct {
	Prompt PromptTemplate `json:"aiPrompt,omitempty"`
}

var _ MessageFormatter = AIMessagePromptTemplate{}

// FormatMessages formats the message with the values given.
func (p AIMessagePromptTemplate) FormatMessages(values map[string]any) ([]schema.ChatMessage, error) {
	text, err := p.Prompt.Format(values)
	return []schema.ChatMessage{schema.AIChatMessage{Text: text}}, err
}

// GetInputVariables returns the input variables the prompt expects.
func (p AIMessagePromptTemplate) GetInputVariables() []string {
	return p.Prompt.InputVariables
}

func (p AIMessagePromptTemplate) Save(path string) error {
	err := load.ToFile(p, path)
	if err != nil {
		return err
	}
	return nil
}

// NewAIMessagePromptTemplate creates a new AI message prompt template.
func NewAIMessagePromptTemplate(template string, inputVariables []string) AIMessagePromptTemplate {
	return AIMessagePromptTemplate{
		Prompt: NewPromptTemplate(template, inputVariables),
	}
}

// HumanMessagePromptTemplate is a message formatter that returns a human message.
type HumanMessagePromptTemplate struct {
	Prompt PromptTemplate `json:"humanPrompt,omitempty"`
}

var _ MessageFormatter = HumanMessagePromptTemplate{}

// FormatMessages formats the message with the values given.
func (p HumanMessagePromptTemplate) FormatMessages(values map[string]any) ([]schema.ChatMessage, error) {
	text, err := p.Prompt.Format(values)
	return []schema.ChatMessage{schema.HumanChatMessage{Text: text}}, err
}

// GetInputVariables returns the input variables the prompt expects.
func (p HumanMessagePromptTemplate) GetInputVariables() []string {
	return p.Prompt.InputVariables
}

func (p HumanMessagePromptTemplate) Save(path string) error {
	err := load.ToFile(p, path)
	if err != nil {
		return err
	}
	return nil
}

// NewHumanMessagePromptTemplate creates a new human message prompt template.
func NewHumanMessagePromptTemplate(template string, inputVariables []string) HumanMessagePromptTemplate {
	return HumanMessagePromptTemplate{
		Prompt: NewPromptTemplate(template, inputVariables),
	}
}

func NewHumanMessagePromptFromFile(path string) (HumanMessagePromptTemplate, error) {
	var promptTemplate HumanMessagePromptTemplate
	err := load.FromFile(&promptTemplate, path)
	if err != nil {
		return HumanMessagePromptTemplate{}, err
	}
	return promptTemplate, nil
}

// GenericMessagePromptTemplate is a message formatter that returns message with the specified speaker.
type GenericMessagePromptTemplate struct {
	Prompt PromptTemplate `json:"prompt,omitempty"`
	Role   string         `json:"role,omitempty"`
}

var _ MessageFormatter = GenericMessagePromptTemplate{}

// FormatMessages formats the message with the values given.
func (p GenericMessagePromptTemplate) FormatMessages(values map[string]any) ([]schema.ChatMessage, error) {
	text, err := p.Prompt.Format(values)
	return []schema.ChatMessage{schema.GenericChatMessage{Text: text, Role: p.Role}}, err
}

// GetInputVariables returns the input variables the prompt expects.
func (p GenericMessagePromptTemplate) GetInputVariables() []string {
	return p.Prompt.InputVariables
}

func (p GenericMessagePromptTemplate) Save(path string) error {
	err := load.ToFile(p, path)
	if err != nil {
		return err
	}
	return nil
}

// NewGenericMessagePromptTemplate creates a new generic message prompt template.
func NewGenericMessagePromptTemplate(role, template string, inputVariables []string) GenericMessagePromptTemplate {
	return GenericMessagePromptTemplate{
		Prompt: NewPromptTemplate(template, inputVariables),
		Role:   role,
	}
}
