package prompts

type PromptTemplate struct {
	partialVariables map[string]any
	inputVariables   []string
	template         string
	templateFormat   string
	validateTemplate bool
}

func (p PromptTemplate) Format(inputValues map[string]any) (string, error) {
	allValues := mergePartialAndUserVariables(p.partialVariables, inputValues)
	return renderTemplate(p.template, p.templateFormat, allValues)
}

func (p PromptTemplate) FormatPromptValue(inputValues map[string]any) (PromptValue, error) {
	formattedPrompt, err := p.Format(inputValues)
	if err != nil {
		return StringPromptValue{}, err
	}
	return StringPromptValue{value: formattedPrompt}, nil
}

func (p PromptTemplate) GetInputVariables() []string {
	return p.inputVariables
}

func NewPromptTemplate(template string, inputVariables []string, options ...PromptTemplateOption) (PromptTemplate, error) {
	p := &PromptTemplate{
		partialVariables: make(map[string]any, 0),
		inputVariables:   inputVariables,
		template:         template,
		templateFormat:   "f-string",
		validateTemplate: true,
	}

	for _, option := range options {
		option(p)
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

type PromptTemplateOption func(p *PromptTemplate)

func WithPartialVariablesPrompt(partialVariables map[string]any) PromptTemplateOption {
	return func(p *PromptTemplate) {
		p.partialVariables = partialVariables
	}
}

func WithTemplateFormatPrompt(templateFormat string) PromptTemplateOption {
	return func(p *PromptTemplate) {
		p.templateFormat = templateFormat
	}
}

func WithValidateTemplatePrompt(validateTemplate bool) PromptTemplateOption {
	return func(p *PromptTemplate) {
		p.validateTemplate = validateTemplate
	}
}
