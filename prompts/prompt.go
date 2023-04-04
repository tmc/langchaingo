package prompts

import "github.com/tmc/langchaingo/memory"

type PromptTemplate struct {
	partialVariables map[string]string
	inputVariables   []string
	template         string
	templateFormat   string
	validateTemplate bool
}

func (p PromptTemplate) Format(inputValues map[string]any) (string, error) {
	allValues := p.mergePartialAndUserVariables(inputValues)
	return renderTemplate(p.template, p.templateFormat, allValues)
}

func NewTemplate(template string, inputVariables []string, options ...Option) (PromptTemplate, error) {
	p := &PromptTemplate{
		partialVariables: make(map[string]string, 0),
		inputVariables:   inputVariables,
		template:         template,
		templateFormat:   "f-string",
		validateTemplate: true,
	}

	for _, opt := range options {
		opt(p)
	}

	if p.validateTemplate {
		totalInputVariables := make([]string, len(p.inputVariables)+len(p.partialVariables))
		totalInputVariables = append(totalInputVariables, p.inputVariables...)

		for variable := range p.partialVariables {
			totalInputVariables = append(totalInputVariables, variable)
		}

		if err := checkValidTemplate(p.template, p.templateFormat, totalInputVariables); err != nil {
			return *p, err
		}
	}

	return *p, nil
}

type Option func(p *PromptTemplate)

func WithPartialVariables(partialVariables map[string]string) Option {
	return func(p *PromptTemplate) {
		p.partialVariables = partialVariables
	}
}

func WithTemplateFormat(templateFormat string) Option {
	return func(p *PromptTemplate) {
		p.templateFormat = templateFormat
	}
}

func WithValidateTemplate(validateTemplate bool) Option {
	return func(p *PromptTemplate) {
		p.validateTemplate = validateTemplate
	}
}

func (p PromptTemplate) mergePartialAndUserVariables(userVariables map[string]any) map[string]any {
	allValues := make(map[string]any)
	for variable, value := range p.partialVariables {
		allValues[variable] = value
	}

	for variable, value := range userVariables {
		allValues[variable] = value
	}

	return allValues
}

type PromptValue interface {
	String() string
	ToChatMessages() []memory.ChatMessage
}

type StringPromptValue struct {
	value string
}

func (v StringPromptValue) String() string { return v.value }
func (v StringPromptValue) ToChatMessages() []memory.ChatMessage {
	return []memory.ChatMessage{memory.HumanChatMessage{Text: v.value}}
}

func (p PromptTemplate) FormatPromptValue(inputValues map[string]any) (PromptValue, error) {
	formattedPrompt, err := p.Format(inputValues)
	if err != nil {
		return StringPromptValue{}, err
	}
	return StringPromptValue{value: formattedPrompt}, nil
}
