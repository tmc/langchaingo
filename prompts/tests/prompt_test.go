package prompts_test

import (
	"testing"

	"github.com/tmc/langchaingo/prompts"
)

type basicPromptTest struct {
	template          string
	inputVariables    []string
	inputValues       map[string]any
	expectedFormatted string
	expectsError      bool
}

var basicPromptTests = []basicPromptTest{
	{
		template:          "foo {var}",
		inputVariables:    []string{"var"},
		inputValues:       map[string]any{"var": "value"},
		expectedFormatted: "foo value",
		expectsError:      false,
	},

	{
		template:          "foo {var} bar foo {var2} bar {var}",
		inputVariables:    []string{"var", "var2"},
		inputValues:       map[string]any{"var": "value", "var2": "value2"},
		expectedFormatted: "foo value bar foo value2 bar value",
		expectsError:      false,
	},

	{
		template:          "foo {{var}}",
		inputVariables:    []string{},
		inputValues:       map[string]any{},
		expectedFormatted: "foo {var}",
		expectsError:      false,
	},

	{
		template:          "{{pre{var}post}}",
		inputVariables:    []string{"var"},
		inputValues:       map[string]any{"var": "bar"},
		expectedFormatted: "{prebarpost}",
		expectsError:      false,
	},

	{
		template:          "foo {var}}",
		inputVariables:    []string{"var"},
		inputValues:       map[string]any{"var": "value"},
		expectedFormatted: "",
		expectsError:      true,
	},

	{
		template:          "{ foo {var}",
		inputVariables:    []string{"var"},
		inputValues:       map[string]any{"var": "value"},
		expectedFormatted: "",
		expectsError:      true,
	},

	{
		template:          "foo {var} is {var2}",
		inputVariables:    []string{"var"},
		inputValues:       map[string]any{"var": "value"},
		expectedFormatted: "",
		expectsError:      true,
	},
}

func TestCreatingAndUsingPromptTemplates(t *testing.T) {
	for _, test := range basicPromptTests {
		template, err := prompts.NewPromptTemplate(test.template, test.inputVariables)
		if err != nil {
			if !test.expectsError {
				t.Errorf("Unexpected error %s", err.Error())
			}

			continue
		}

		formatted, err := template.Format(test.inputValues)
		if err != nil {
			if !test.expectsError {
				t.Errorf("Unexpected error %s", err.Error())
			}

			continue
		}

		if formatted != test.expectedFormatted {
			t.Errorf("Prompt template (%s) with input variables %v and input values %v, not equal to expected %s when formatted. Got: %s", test.template, test.inputVariables, test.inputValues, test.expectedFormatted, formatted)
		}
	}
}

type promptWithOptionsTest struct {
	template          string
	inputVariables    []string
	options           []prompts.PromptTemplateOption
	optionDescription string
	inputValues       map[string]any
	expectedFormatted string
	expectsError      bool
}

var promptWithOptionsTests = []promptWithOptionsTest{
	{
		template:          "foo {var}",
		inputVariables:    []string{"var"},
		inputValues:       map[string]any{"var": "value"},
		options:           []prompts.PromptTemplateOption{prompts.WithTemplateFormatPrompt("foo")},
		optionDescription: `WithTemplateFormat("foo")`,
		expectedFormatted: "",
		expectsError:      true,
	},

	{
		template:          "{par} foo {var}",
		inputVariables:    []string{"var"},
		inputValues:       map[string]any{"var": "value"},
		options:           []prompts.PromptTemplateOption{prompts.WithPartialVariablesPrompt(map[string]any{"par": "bar"})},
		optionDescription: `WithPartialVariables(map[string]any{"par": "bar"})`,
		expectedFormatted: "bar foo value",
		expectsError:      false,
	},

	{
		template:          "{par} foo {var}",
		inputVariables:    []string{"var"},
		inputValues:       map[string]any{"var": "value", "par": "foo"},
		options:           []prompts.PromptTemplateOption{prompts.WithPartialVariablesPrompt(map[string]any{"par": "bar"})},
		optionDescription: `WithPartialVariables(map[string]any{"par": "bar"})`,
		expectedFormatted: "foo foo value",
		expectsError:      false,
	},
}

func TestCreatingAndUsingPromptTemplatesWithOptions(t *testing.T) {
	for _, test := range promptWithOptionsTests {
		template, err := prompts.NewPromptTemplate(test.template, test.inputVariables, test.options...)
		if err != nil {
			if !test.expectsError {
				t.Errorf("Unexpected error %s", err.Error())
			}

			continue
		}

		formatted, err := template.Format(test.inputValues)
		if err != nil {
			if !test.expectsError {
				t.Errorf("Unexpected error %s", err.Error())
			}

			continue
		}

		if formatted != test.expectedFormatted {
			t.Errorf("Prompt template (%s) with input variables %v, input values %v and options %s, not equal to expected (%s) when formatted. Got: %s", test.template, test.inputVariables, test.inputValues, test.optionDescription, test.expectedFormatted, formatted)
		}
	}
}
