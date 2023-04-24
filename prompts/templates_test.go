package prompts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInterpolateGoTemplate(t *testing.T) {
	t.Parallel()

	type tests struct {
		name           string
		template       string
		templateValues map[string]any
		expected       string
		errValue       string
	}

	testCases := []tests{
		{
			name:     "Single",
			template: "Hello {{ .key }}",
			templateValues: map[string]any{
				"key": "world",
			},
			expected: "Hello world",
		},
		{
			name:     "Multiple",
			template: "Hello {{ .key1 }} and {{ .key2 }}",
			templateValues: map[string]any{
				"key1": "world",
				"key2": "universe",
			},
			expected: "Hello world and universe",
		},
		{
			name:     "Nested",
			template: "Hello {{ .key1.key2 }}",
			templateValues: map[string]any{
				"key1": map[string]any{
					"key2": "world",
				},
			},
			expected: "Hello world",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual, err := interpolateGoTemplate(tc.template, tc.templateValues)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}

	errTestCases := []tests{
		{
			name:     "ParseErrored",
			template: "Hello {{{ .key1 }}",
			expected: "",
			errValue: "template: template:1: unexpected \"{\" in command",
		},
		{
			name:     "ExecuteErrored",
			template: "Hello {{ .key1 .key2 }}",
			expected: "",
			errValue: "template: template:1:9: executing \"template\" at <.key1>: key1 is not a method but has arguments",
		},
	}

	for _, tc := range errTestCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := interpolateGoTemplate(tc.template, map[string]any{})
			require.Error(t, err)
			assert.EqualError(t, err, tc.errValue)
		})
	}
}

func TestCheckValidTemplate(t *testing.T) {
	t.Parallel()

	t.Run("NoTemplateAvailable", func(t *testing.T) {
		t.Parallel()

		err := CheckValidTemplate("Hello, {test}", "unknown", []string{"test"})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidTemplateFormat)
		assert.EqualError(t, err, "invalid template format, got: unknown, should be one of [go-template]")
	})

	t.Run("TemplateErrored", func(t *testing.T) {
		t.Parallel()

		err := CheckValidTemplate("Hello, {{{ test }}", TemplateFormatGoTemplate, []string{"test"})
		require.Error(t, err)
		assert.EqualError(t, err, "template: template:1: unexpected \"{\" in command")
	})

	t.Run("TemplateValid", func(t *testing.T) {
		t.Parallel()

		err := CheckValidTemplate("Hello, {{ .test }}", TemplateFormatGoTemplate, []string{"test"})
		require.NoError(t, err)
	})
}

func TestRenderTemplate(t *testing.T) {
	t.Parallel()

	t.Run("TemplateAvailable", func(t *testing.T) {
		t.Parallel()

		actual, err := RenderTemplate(
			"Hello {{ .key }}",
			TemplateFormatGoTemplate,
			map[string]any{
				"key": "world",
			},
		)
		require.NoError(t, err)
		assert.Equal(t, "Hello world", actual)
	})

	t.Run("TemplateNotAvailable", func(t *testing.T) {
		t.Parallel()

		_, err := RenderTemplate(
			"Hello {key}",
			"unknown",
			map[string]any{
				"key": "world",
			},
		)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidTemplateFormat)
	})
}
