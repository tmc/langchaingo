package outputparser_test

import (
	"testing"

	"github.com/tmc/langchaingo/outputparser"
)

func TestMarkdown(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "Valid markdown block with no preceding text",
			input:    "```markdown\nSimple content\nwith multiple lines\n```",
			expected: "Simple content\nwith multiple lines",
			hasError: false,
		},
		{
			name:     "Missing opening markdown tag",
			input:    "Here is some content:\n\nThis is not properly formatted.\n```",
			expected: "",
			hasError: true,
		},
		{
			name:     "Missing closing tag",
			input:    "Here is some content:\n\n```markdown\nThis doesn't have a closing tag.",
			expected: "",
			hasError: true,
		},
		{
			name:     "Empty markdown block",
			input:    "```markdown\n```",
			expected: "",
			hasError: false,
		},
		{
			name:     "Markdown blocks and code blocks",
			input:    "First block:\n```markdown\nContent 1\n\nCode block:\n```csharp\nContent 2\n``` 13456```",
			expected: "Content 1\n\nCode block:\n```csharp\nContent 2\n``` 13456",
			hasError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := outputparser.NewMarkdown()

			result, err := parser.Parse(tc.input)

			// Check error status
			if tc.hasError && err == nil {
				t.Errorf("Expected an error but got none")
			}

			if !tc.hasError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// If we're not expecting an error, check the result
			if !tc.hasError {
				content, ok := result.(string)
				if !ok {
					t.Errorf("Expected result to be string, got %T", result)
				}

				if content != tc.expected {
					t.Errorf("Expected:\n%s\n\nGot:\n%s", tc.expected, content)
				}
			}
		})
	}
}
