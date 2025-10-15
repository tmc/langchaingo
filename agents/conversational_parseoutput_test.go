package agents

import (
	"regexp"
	"strings"
	"testing"
)

// TestConversationalParseOutputSingleLineInput verifies parseOutput handles simple single-line inputs.
func TestConversationalParseOutputSingleLineInput(t *testing.T) {
	t.Parallel()

	agent := &ConversationalAgent{OutputKey: "output"}
	actions, finish, err := agent.parseOutput("Action: search\nAction Input: hello world")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if finish != nil {
		t.Fatal("expected action, not finish")
	}
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].Tool != "search" {
		t.Errorf("expected tool 'search', got %q", actions[0].Tool)
	}
	if actions[0].ToolInput != "hello world" {
		t.Errorf("expected input 'hello world', got %q", actions[0].ToolInput)
	}
}

// TestConversationalParseOutputMultilineJSONInput verifies parseOutput correctly handles multiline JSON inputs.
func TestConversationalParseOutputMultilineJSONInput(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          string
		expectedTool   string
		expectedInput  string
	}{
		{
			name:           "multiline JSON with newlines in payload",
			input:          "Action: api_call\nAction Input: {\"endpoint\": \"/api\",\n\"method\": \"GET\"}",
			expectedTool:   "api_call",
			expectedInput:  "{\"endpoint\": \"/api\",\n\"method\": \"GET\"}",
		},
		{
			name:           "complex multiline JSON (issue #1414 Bicep case)",
			input:          "Action: azd_error_troubleshooting\nAction Input: {\"errorMessage\": \"initializing provisioning manager: failed to compile bicep template: failed running bicep build: exit code: 1, stdout: , stderr: /home/vscode/test/infra/main.bicep(26,32) : Error BCP033\"}",
			expectedTool:   "azd_error_troubleshooting",
			expectedInput:  "{\"errorMessage\": \"initializing provisioning manager: failed to compile bicep template: failed running bicep build: exit code: 1, stdout: , stderr: /home/vscode/test/infra/main.bicep(26,32) : Error BCP033\"}",
		},
		{
			name:           "action input spanning many lines",
			input:          "Action: process\nAction Input: {\n  \"line1\": \"value1\",\n  \"line2\": \"value2\",\n  \"line3\": \"value3\"\n}",
			expectedTool:   "process",
			expectedInput:  "{\n  \"line1\": \"value1\",\n  \"line2\": \"value2\",\n  \"line3\": \"value3\"\n}",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			agent := &ConversationalAgent{OutputKey: "output"}
			actions, finish, err := agent.parseOutput(tc.input)

			if err != nil {
				t.Fatalf("parseOutput failed for multiline input: %v", err)
			}
			if finish != nil {
				t.Fatal("expected action, not finish")
			}
			if len(actions) != 1 {
				t.Fatalf("expected 1 action, got %d", len(actions))
			}
			if actions[0].Tool != tc.expectedTool {
				t.Errorf("tool mismatch: expected %q, got %q", tc.expectedTool, actions[0].Tool)
			}
			// TrimSpace is called on the input in parseOutput, so we need to compare trimmed versions
			if strings.TrimSpace(actions[0].ToolInput) != strings.TrimSpace(tc.expectedInput) {
				t.Errorf("input mismatch: expected %q, got %q", tc.expectedInput, actions[0].ToolInput)
			}
		})
	}
}

// TestConversationalParseOutputFinalAnswer verifies parseOutput correctly identifies AI final answers.
func TestConversationalParseOutputFinalAnswer(t *testing.T) {
	t.Parallel()

	agent := &ConversationalAgent{OutputKey: "output"}
	actions, finish, err := agent.parseOutput("AI: The answer is 42")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if actions != nil {
		t.Fatal("expected finish, not action")
	}
	if finish == nil {
		t.Fatal("expected finish to be non-nil")
	}
	// Note: parseOutput includes the leading space after the split on "AI:"
	expected := " The answer is 42"
	if result := finish.ReturnValues["output"]; result != expected {
		t.Errorf("final answer mismatch: expected %q, got %q", expected, result)
	}
}

// TestConversationalParseOutputInvalid verifies parseOutput returns error for invalid input.
func TestConversationalParseOutputInvalid(t *testing.T) {
	t.Parallel()

	agent := &ConversationalAgent{OutputKey: "output"}
	actions, finish, err := agent.parseOutput("This is not a valid action format")

	if err == nil {
		t.Fatal("expected error for invalid input, got nil")
	}
	if actions != nil {
		t.Fatal("expected nil actions for invalid input")
	}
	if finish != nil {
		t.Fatal("expected nil finish for invalid input")
	}
}

// TestRegexPatternMultiline verifies the regex pattern used by parseOutput correctly handles multiline input.
// This test specifically validates the (?s) DOTALL flag behavior.
func TestRegexPatternMultiline(t *testing.T) {
	t.Parallel()

	// The new regex pattern with (?s) DOTALL flag
	newPattern := regexp.MustCompile(`Action: (.*?)[\n]*(?s)Action Input: (.*)`)

	testCases := []struct {
		name           string
		input          string
		expectedTool   string
		expectedInput  string
		shouldMatch    bool
	}{
		{
			name:          "simple single line",
			input:         "Action: search\nAction Input: hello",
			expectedTool:  "search",
			expectedInput: "hello",
			shouldMatch:   true,
		},
		{
			name:          "multiline JSON",
			input:         "Action: api\nAction Input: {\"a\": \"b\",\n\"c\": \"d\"}",
			expectedTool:  "api",
			expectedInput: "{\"a\": \"b\",\n\"c\": \"d\"}",
			shouldMatch:   true,
		},
		{
			name:        "no match",
			input:       "No action here",
			shouldMatch: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matches := newPattern.FindStringSubmatch(tc.input)

			if tc.shouldMatch {
				if len(matches) != 3 {
					t.Fatalf("expected full match + 2 capture groups, got %d matches", len(matches))
				}
				if matches[1] != tc.expectedTool {
					t.Errorf("tool mismatch: expected %q, got %q", tc.expectedTool, matches[1])
				}
				if matches[2] != tc.expectedInput {
					t.Errorf("input mismatch: expected %q, got %q", tc.expectedInput, matches[2])
				}
			} else {
				if len(matches) != 0 {
					t.Fatalf("expected no matches, got %d", len(matches))
				}
			}
		})
	}
}
