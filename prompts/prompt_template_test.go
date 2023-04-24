package prompts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/schema"
)

func TestCheckInputVariables(t *testing.T) {
	t.Parallel()

	err := checkInputVariables([]string{"stop"})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInputVariableReserved)

	err = checkInputVariables([]string{"test"})
	require.NoError(t, err)
}

func TestCheckPartialVariables(t *testing.T) {
	t.Parallel()

	err := checkPartialVariables(map[string]interface{}{"stop": "test"})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPartialVariableReserved)

	err = checkPartialVariables(map[string]interface{}{"test": "test"})
	require.NoError(t, err)
}

var _ schema.OutputParser[any] = testEmptyOutputParser{}

type testEmptyOutputParser struct{}

func (t testEmptyOutputParser) Parse(_ string) (any, error) {
	return nil, nil
}

func (t testEmptyOutputParser) GetFormatInstructions() string {
	return ""
}

func (t testEmptyOutputParser) Type() string {
	return ""
}

func TestApplyPromptTemplateBaseOptions(t *testing.T) {
	t.Parallel()

	_, err := applyPromptTemplateBaseOptions(PromptTemplateBaseOptions{
		InputVariables: []string{"stop"},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInputVariableReserved)

	_, err = applyPromptTemplateBaseOptions(PromptTemplateBaseOptions{
		PartialVariables: map[string]any{"stop": "test"},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPartialVariableReserved)

	opts, err := applyPromptTemplateBaseOptions(PromptTemplateBaseOptions{
		InputVariables:   []string{"test"},
		OutputParser:     testEmptyOutputParser{},
		PartialVariables: map[string]any{"test": "test"},
	})
	require.NoError(t, err)
	assert.Equal(t, PromptTemplateBaseOptions{
		InputVariables:   []string{"test"},
		OutputParser:     testEmptyOutputParser{},
		PartialVariables: map[string]any{"test": "test"},
	}, opts)
}

func TestFormatPromptValue(t *testing.T) {
	t.Parallel()

	t.Run("Error", func(t *testing.T) {
		t.Parallel()

		pt, err := NewPromptTemplate(PromptTemplateOptions{
			PromptTemplateBaseOptions: PromptTemplateBaseOptions{
				InputVariables: []string{"testFormatPromptValue"},
			},
			Template:         "Hello, {{ .test }}",
			ValidateTemplate: true,
		})
		require.NoError(t, err)

		pt.opts.Template = "Hello, {{{ .testFormatPromptValue }}"
		_, err = formatPromptValue(pt, map[string]any{
			"testFormatPromptValue": "world",
		})
		require.Error(t, err)
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		pt, err := NewPromptTemplate(PromptTemplateOptions{
			PromptTemplateBaseOptions: PromptTemplateBaseOptions{
				InputVariables: []string{"testFormatPromptValue"},
			},
			Template:         "Hello, {{ .testFormatPromptValue }}",
			ValidateTemplate: true,
		})
		require.NoError(t, err)

		spv, err := formatPromptValue(pt, map[string]any{
			"testFormatPromptValue": "world",
		})
		require.NoError(t, err)
		assert.Equal(t, "Hello, world", spv.String())
	})
}

func TestMergePartialAndUserVariables(t *testing.T) {
	t.Parallel()

	type test struct {
		name             string
		partialVariables map[string]any
		userVariables    map[string]any
		mergedVariables  map[string]any
		err              error
	}

	testCases := []test{
		{
			name:             "NilPartialVariables",
			partialVariables: nil,
			userVariables: map[string]any{
				"test": "world",
			},
			mergedVariables: map[string]any{
				"test": "world",
			},
		},
		{
			name: "Success",
			partialVariables: map[string]interface{}{
				"test":  "test",
				"test2": "test2",
				"test3": func() any {
					return "test3"
				},
			},
			userVariables: map[string]any{
				"test": "world",
			},
			mergedVariables: map[string]any{
				"test":  "world",
				"test2": "test2",
				"test3": "test3",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mergedVariables, err := mergePartialAndUserVariables(tc.partialVariables, tc.userVariables)
			require.NoError(t, err)
			assert.Equal(t, tc.mergedVariables, mergedVariables)
		})
	}

	errTestCases := []test{
		{
			name: "InvalidPartialVariableType",
			partialVariables: map[string]interface{}{
				"test": 1234,
			},
			userVariables: map[string]any{
				"test": "world",
			},
			mergedVariables: make(map[string]any),
			err:             ErrInvalidPartialVariableType,
		},
	}

	for _, tc := range errTestCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mergedVariables, err := mergePartialAndUserVariables(tc.partialVariables, tc.userVariables)
			require.Error(t, err)
			assert.Empty(t, mergedVariables)
			assert.ErrorIs(t, err, tc.err)
		})
	}
}

func TestNewPromptTemplate(t *testing.T) {
	t.Parallel()

	_, err := NewPromptTemplate(PromptTemplateOptions{
		PromptTemplateBaseOptions: PromptTemplateBaseOptions{
			InputVariables: []string{"stop"},
		},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInputVariableReserved)

	_, err = NewPromptTemplate(PromptTemplateOptions{
		Template:         "Hello, {{ .test }}",
		TemplateFormat:   "unknown",
		ValidateTemplate: true,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidTemplateFormat)

	_, err = NewPromptTemplate(PromptTemplateOptions{
		Template:         "Hello, {{ .test }}",
		ValidateTemplate: false,
	})
	require.NoError(t, err)
}

func TestPromptTemplateMergePartialAndUserVariables(t *testing.T) {
	t.Parallel()

	type test struct {
		name             string
		partialVariables map[string]any
		userVariables    map[string]any
		mergedVariables  map[string]any
		err              error
	}

	testCases := []test{
		{
			name:             "NilPartialVariables",
			partialVariables: nil,
			userVariables:    map[string]any{"test": "world"},
			mergedVariables:  map[string]any{"test": "world"},
		},
		{
			name: "Success",
			partialVariables: map[string]interface{}{
				"test":  "test",
				"test2": "test2",
				"test3": func() any {
					return "test3"
				},
			},
			userVariables: map[string]any{"test": "world"},
			mergedVariables: map[string]any{
				"test":  "world",
				"test2": "test2",
				"test3": "test3",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			pt, err := NewPromptTemplate(PromptTemplateOptions{
				PromptTemplateBaseOptions: PromptTemplateBaseOptions{
					PartialVariables: tc.partialVariables,
				},
			})
			require.NoError(t, err)
			require.NotNil(t, pt)

			mergedVariables, err := pt.MergePartialAndUserVariables(tc.userVariables)
			require.NoError(t, err)
			assert.Equal(t, tc.mergedVariables, mergedVariables)
		})
	}

	errTestCases := []test{
		{
			name:             "InvalidPartialVariableType",
			partialVariables: map[string]interface{}{"test": 1234},
			userVariables:    map[string]any{"test": "world"},
			mergedVariables:  make(map[string]any),
			err:              ErrInvalidPartialVariableType,
		},
	}

	for _, tc := range errTestCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			pt, err := NewPromptTemplate(PromptTemplateOptions{
				PromptTemplateBaseOptions: PromptTemplateBaseOptions{
					PartialVariables: tc.partialVariables,
				},
			})
			require.NoError(t, err)
			require.NotNil(t, pt)

			mergedVariables, err := pt.MergePartialAndUserVariables(tc.userVariables)
			require.Error(t, err)
			assert.Empty(t, mergedVariables)
			assert.ErrorIs(t, err, tc.err)
		})
	}
}

func TestPromptTemplateGetPromptType(t *testing.T) {
	t.Parallel()

	pt, err := NewPromptTemplate(PromptTemplateOptions{})
	require.NoError(t, err)
	assert.Equal(t, TemplateTypePrompt, pt.GetPromptType())
}

func TestPromptTemplateFormat(t *testing.T) {
	t.Parallel()

	t.Run("Error", func(t *testing.T) {
		t.Parallel()

		pt, err := NewPromptTemplate(PromptTemplateOptions{
			PromptTemplateBaseOptions: PromptTemplateBaseOptions{
				InputVariables: []string{"test"},
				PartialVariables: map[string]any{
					"test": 1234,
				},
			},
			Template:         "Hello, {{ .test }}",
			ValidateTemplate: true,
		})
		if err != nil {
			t.Fatalf("Expected nil, got %v", err)
		}

		pt.opts.Template = "Hello, {{{ .test }}"
		_, err = pt.Format(map[string]any{
			"test": "world",
		})
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		pt, err := NewPromptTemplate(PromptTemplateOptions{
			PromptTemplateBaseOptions: PromptTemplateBaseOptions{
				InputVariables: []string{"test"},
			},
			Template:         "Hello, {{ .test }}",
			ValidateTemplate: true,
		})
		if err != nil {
			t.Fatalf("Expected nil, got %v", err)
		}

		str, err := pt.Format(map[string]any{
			"test": "world",
		})
		if err != nil {
			t.Fatalf("Expected nil, got %v", err)
		}
		if str != "Hello, world" {
			t.Errorf("Expected 'Hello, world', got %s", str)
		}
	})
}

func TestPromptTemplateFormatPromptValue(t *testing.T) {
	t.Parallel()

	t.Run("Error", func(t *testing.T) {
		t.Parallel()

		pt, err := NewPromptTemplate(PromptTemplateOptions{
			PromptTemplateBaseOptions: PromptTemplateBaseOptions{
				InputVariables: []string{"test"},
				PartialVariables: map[string]any{
					"test": 1234,
				},
			},
			Template:         "Hello, {{ .test }}",
			ValidateTemplate: true,
		})
		if err != nil {
			t.Fatalf("Expected nil, got %v", err)
		}

		pt.opts.Template = "Hello, {{{ .test }}"
		_, err = pt.FormatPromptValue(map[string]any{
			"test": "world",
		})
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		pt, err := NewPromptTemplate(PromptTemplateOptions{
			PromptTemplateBaseOptions: PromptTemplateBaseOptions{
				InputVariables: []string{"test"},
			},
			Template:         "Hello, {{ .test }}",
			ValidateTemplate: true,
		})
		if err != nil {
			t.Fatalf("Expected nil, got %v", err)
		}

		str, err := pt.FormatPromptValue(map[string]any{
			"test": "world",
		})
		if err != nil {
			t.Fatalf("Expected nil, got %v", err)
		}
		if str.String() != "Hello, world" {
			t.Errorf("Expected 'Hello, world', got %s", str.String())
		}
	})
}

func TestNewPromptTemplateFromExamples(t *testing.T) {
	t.Parallel()

	pt, err := NewPromptTemplateFromExamples(PromptTemplateFromExamplesOption{
		Examples:         []string{"- example 1", "- example 2"},
		Suffix:           "Thank you!",
		InputVariables:   []string{"test"},
		ExampleSeparator: "\n\n",
		Prefix:           "Hello, {{ .test }}, output as:",
	})
	require.NoError(t, err)
	assert.Equal(t, pt.opts.Template, "Hello, {{ .test }}, output as:\n\n- example 1\n\n- example 2\n\nThank you!")
}

func TestNewPromptTemplateFromFStringTemplate(t *testing.T) {
	t.Parallel()

	require.PanicsWithValue(t, "not implemented", func() {
		_, _ = NewPromptTemplateFromFStringTemplate(PromptTemplateFromFStringOption{})
	})
}

func TestNewPromptTemplateFromFile(t *testing.T) {
	t.Parallel()

	require.PanicsWithValue(t, "not implemented", func() {
		_, _ = NewPromptTemplateFromFile(PromptTemplateFromFileOption{})
	})
}
