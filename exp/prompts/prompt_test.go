package prompts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreatingAndUsingPromptTemplates(t *testing.T) {
	t.Parallel()

	type basicPromptTest struct {
		template          string
		inputVariables    []string
		inputValues       map[string]any
		expectedFormatted string
		expectsError      bool
	}

	basicPromptTests := []basicPromptTest{
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

	for _, test := range basicPromptTests {
		template, err := NewPromptTemplate(test.template, test.inputVariables)
		if err != nil {
			if !test.expectsError {
				assert.NoError(t, err)
			}

			continue
		}

		formatted, err := template.Format(test.inputValues)
		if err != nil {
			if !test.expectsError {
				assert.NoError(t, err)
			}

			continue
		}

		assert.Equal(t, test.expectedFormatted, formatted)
	}
}

func TestCreatingAndUsingPromptTemplatesWithOptions(t *testing.T) {
	t.Parallel()

	type promptWithOptionsTest struct {
		template          string
		inputVariables    []string
		options           []PromptTemplateOption
		optionDescription string
		inputValues       map[string]any
		expectedFormatted string
		expectsError      bool
	}

	promptWithOptionsTests := []promptWithOptionsTest{
		{
			template:          "foo {var}",
			inputVariables:    []string{"var"},
			inputValues:       map[string]any{"var": "value"},
			options:           []PromptTemplateOption{WithTemplateFormatPrompt("foo")},
			optionDescription: `WithTemplateFormat("foo")`,
			expectedFormatted: "",
			expectsError:      true,
		},

		{
			template:          "{par} foo {var}",
			inputVariables:    []string{"var"},
			inputValues:       map[string]any{"var": "value"},
			options:           []PromptTemplateOption{WithPartialVariablesPrompt(map[string]any{"par": "bar"})},
			optionDescription: `WithPartialVariables(map[string]any{"par": "bar"})`,
			expectedFormatted: "bar foo value",
			expectsError:      false,
		},

		{
			template:          "{par} foo {var}",
			inputVariables:    []string{"var"},
			inputValues:       map[string]any{"var": "value", "par": "foo"},
			options:           []PromptTemplateOption{WithPartialVariablesPrompt(map[string]any{"par": "bar"})},
			optionDescription: `WithPartialVariables(map[string]any{"par": "bar"})`,
			expectedFormatted: "foo foo value",
			expectsError:      false,
		},
	}

	for _, test := range promptWithOptionsTests {
		template, err := NewPromptTemplate(test.template, test.inputVariables, test.options...)
		if err != nil {
			if !test.expectsError {
				assert.NoError(t, err)
			}

			continue
		}

		formatted, err := template.Format(test.inputValues)
		if err != nil {
			if !test.expectsError {
				assert.NoError(t, err)
			}

			continue
		}

		assert.Equal(t, test.expectedFormatted, formatted)
	}
}
