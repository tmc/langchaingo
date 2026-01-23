package prompts

import (
	"testing"
)

// TestVariableNamesNoLongerReserved verifies that previously "reserved" words
// can now be used as variable names after removing unnecessary restrictions.
func TestVariableNamesNoLongerReserved(t *testing.T) {
	t.Parallel()

	// These words were previously blocked but are common in real-world use
	testCases := []struct {
		name     string
		template string
		data     map[string]any
		expected string
	}{
		{
			name:     "from_variable",
			template: "Email from: {{.from}}",
			data:     map[string]any{"from": "alice@example.com"},
			expected: "Email from: alice@example.com",
		},
		{
			name:     "call_variable",
			template: "Call ID: {{.call}}",
			data:     map[string]any{"call": "SUPPORT-123"},
			expected: "Call ID: SUPPORT-123",
		},
		{
			name:     "filter_variable",
			template: "Active filter: {{.filter}}",
			data:     map[string]any{"filter": "status:active"},
			expected: "Active filter: status:active",
		},
		{
			name:     "template_variable",
			template: "Using template: {{.template}}",
			data:     map[string]any{"template": "welcome_email_v2"},
			expected: "Using template: welcome_email_v2",
		},
		{
			name:     "import_variable",
			template: "Import source: {{.import}}",
			data:     map[string]any{"import": "csv_file.csv"},
			expected: "Import source: csv_file.csv",
		},
		{
			name:     "block_variable",
			template: "Block number: {{.block}}",
			data:     map[string]any{"block": "12345"},
			expected: "Block number: 12345",
		},
		{
			name:     "multiple_previously_reserved",
			template: "From {{.from}} - Call {{.call}} - Filter {{.filter}}",
			data: map[string]any{
				"from":   "Bob",
				"call":   "789",
				"filter": "urgent",
			},
			expected: "From Bob - Call 789 - Filter urgent",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test with GoTemplate format
			result, err := RenderTemplate(tc.template, TemplateFormatGoTemplate, tc.data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

// TestEssentialValidationStillWorks verifies that essential validation
// (empty names, null bytes) still works after removing reserved word checks.
func TestEssentialValidationStillWorks(t *testing.T) {
	t.Parallel()

	t.Run("EmptyVariableName", func(t *testing.T) {
		_, err := RenderTemplate("{{.}}", TemplateFormatGoTemplate, map[string]any{
			"": "value",
		}, WithSanitization())
		if err == nil {
			t.Error("expected error for empty variable name, got nil")
		}
	})

	t.Run("NullByteInVariableName", func(t *testing.T) {
		_, err := RenderTemplate("{{.test}}", TemplateFormatGoTemplate, map[string]any{
			"test\x00name": "value",
		}, WithSanitization())
		if err == nil {
			t.Error("expected error for null byte in variable name, got nil")
		}
	})

	t.Run("InvalidIdentifierFormat", func(t *testing.T) {
		// Starting with number should still be invalid
		_, err := RenderTemplate("{{.test}}", TemplateFormatGoTemplate, map[string]any{
			"123invalid": "value",
		}, WithSanitization())
		if err == nil {
			t.Error("expected error for variable name starting with number, got nil")
		}
	})
}
