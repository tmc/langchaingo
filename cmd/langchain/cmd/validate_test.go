package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestValidateOpenAI(t *testing.T) {
	// Test without API key
	originalKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")

	results := validateOpenAI()

	// Should have at least one result for API key check
	if len(results) < 1 {
		t.Error("Expected at least one validation result")
	}

	// First result should be API key validation failure
	if results[0].Success {
		t.Error("API key validation should fail when key is not set")
	}

	if !strings.Contains(results[0].Error, "OPENAI_API_KEY") {
		t.Error("Error should mention OPENAI_API_KEY")
	}

	// Restore original key
	if originalKey != "" {
		os.Setenv("OPENAI_API_KEY", originalKey)
	}
}

func TestValidateOpenAIWithKey(t *testing.T) {
	// Test with fake API key
	os.Setenv("OPENAI_API_KEY", "fake-key-for-testing")
	defer os.Unsetenv("OPENAI_API_KEY")

	results := validateOpenAI()

	// Should pass API key check
	if len(results) < 1 {
		t.Error("Expected at least one validation result")
	}

	// First result should be API key validation success
	if !results[0].Success {
		t.Error("API key validation should pass when key is set")
	}

	if results[0].Test != "OpenAI API Key" {
		t.Errorf("Expected 'OpenAI API Key' test, got '%s'", results[0].Test)
	}
}

func TestValidateAnthropic(t *testing.T) {
	// Test without API key
	originalKey := os.Getenv("ANTHROPIC_API_KEY")
	os.Unsetenv("ANTHROPIC_API_KEY")

	results := validateAnthropic()

	// Should have at least one result for API key check
	if len(results) < 1 {
		t.Error("Expected at least one validation result")
	}

	// First result should be API key validation failure
	if results[0].Success {
		t.Error("API key validation should fail when key is not set")
	}

	if !strings.Contains(results[0].Error, "ANTHROPIC_API_KEY") {
		t.Error("Error should mention ANTHROPIC_API_KEY")
	}

	// Restore original key
	if originalKey != "" {
		os.Setenv("ANTHROPIC_API_KEY", originalKey)
	}
}

func TestValidateAnthropicWithKey(t *testing.T) {
	// Test with fake API key
	os.Setenv("ANTHROPIC_API_KEY", "fake-key-for-testing")
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	results := validateAnthropic()

	// Should pass API key check
	if len(results) < 1 {
		t.Error("Expected at least one validation result")
	}

	// First result should be API key validation success
	if !results[0].Success {
		t.Error("API key validation should pass when key is set")
	}

	if results[0].Test != "Anthropic API Key" {
		t.Errorf("Expected 'Anthropic API Key' test, got '%s'", results[0].Test)
	}
}

func TestTruncateString(t *testing.T) {
	testCases := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"this is a very long string", 10, "this is a ..."},
		{"exact length", 12, "exact length"},
		{"", 5, ""},
	}

	for _, tc := range testCases {
		result := truncateString(tc.input, tc.maxLen)
		if result != tc.expected {
			t.Errorf("truncateString(%q, %d) = %q, expected %q",
				tc.input, tc.maxLen, result, tc.expected)
		}
	}
}

func TestValidationResult(t *testing.T) {
	result := ValidationResult{
		Test:    "Test Name",
		Success: true,
		Message: "Success message",
		Error:   "",
	}

	if result.Test != "Test Name" {
		t.Error("ValidationResult.Test not set correctly")
	}

	if !result.Success {
		t.Error("ValidationResult.Success not set correctly")
	}

	if result.Message != "Success message" {
		t.Error("ValidationResult.Message not set correctly")
	}
}

func TestValidateCommandFlags(t *testing.T) {
	// Test that validate command has expected flags
	flags := validateCmd.Flags()

	providerFlag := flags.Lookup("provider")
	if providerFlag == nil {
		t.Error("Validate command should have --provider flag")
	}

	quickFlag := flags.Lookup("quick")
	if quickFlag == nil {
		t.Error("Validate command should have --quick flag")
	}

	// Test default values
	if validateProvider != "all" {
		t.Errorf("Default provider should be 'all', got '%s'", validateProvider)
	}

	if quickTest != false {
		t.Error("Default quick test should be false")
	}
}
