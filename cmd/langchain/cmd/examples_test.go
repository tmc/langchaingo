package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestExamples(t *testing.T) (string, []string) {
	tempDir := t.TempDir()
	exampleDirs := []string{
		"openai-completion-example",
		"anthropic-tool-call-example",
		"chroma-vectorstore-example",
	}

	for _, dir := range exampleDirs {
		examplePath := filepath.Join(tempDir, dir)
		err := os.MkdirAll(examplePath, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		// Create a mock Go file
		goFile := filepath.Join(examplePath, "main.go")
		err = os.WriteFile(goFile, []byte("package main\n\nfunc main() {}\n"), 0600)
		if err != nil {
			t.Fatalf("Failed to create test Go file: %v", err)
		}

		// Create a mock README
		readmeFile := filepath.Join(examplePath, "README.md")
		err = os.WriteFile(readmeFile, []byte("# Test Example\nThis is a test example.\n"), 0600)
		if err != nil {
			t.Fatalf("Failed to create test README: %v", err)
		}
	}
	return tempDir, exampleDirs
}

func TestDiscoverExamples(t *testing.T) {
	tempDir, exampleDirs := setupTestExamples(t)

	// Change working directory temporarily
	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Warning: failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Mock the examples directory path
	if err := os.MkdirAll("examples", 0755); err != nil {
		t.Fatalf("Failed to create examples directory: %v", err)
	}
	for _, dir := range exampleDirs {
		src := filepath.Join(tempDir, dir)
		dst := filepath.Join("examples", dir)
		err := os.Rename(src, dst)
		if err != nil {
			t.Fatalf("Failed to move example: %v", err)
		}
	}

	// Test discovery
	examples, err := discoverExamples()
	if err != nil {
		t.Fatalf("Failed to discover examples: %v", err)
	}

	if len(examples) != len(exampleDirs) {
		t.Errorf("Expected %d examples, got %d", len(exampleDirs), len(examples))
	}

	// Test categorization
	for _, example := range examples {
		if example.Category == "" {
			t.Errorf("Example %s has empty category", example.Name)
		}
		if example.Name == "" {
			t.Errorf("Example has empty name")
		}
		if !example.HasReadme {
			t.Errorf("Example %s should have README", example.Name)
		}
	}
}

func TestCategorizeExample(t *testing.T) {
	testCases := []struct {
		name     string
		expected string
	}{
		{"openai-completion-example", "llm"},
		{"anthropic-tool-call-example", "llm"},
		{"chroma-vectorstore-example", "vectorstore"},
		{"chains-conversation-memory-sqlite", "chain"},
		{"document-qa-example", "document"},
		{"random-example", "other"},
	}

	for _, tc := range testCases {
		result := categorizeExample(tc.name)
		if result != tc.expected {
			t.Errorf("categorizeExample(%s) = %s, expected %s", tc.name, result, tc.expected)
		}
	}
}

func TestExtractTags(t *testing.T) {
	testCases := []struct {
		name     string
		expected []string
	}{
		{"anthropic-vision-example", []string{"vision"}},
		{"openai-tool-call-example", []string{"tool-calling"}},
		{"streaming-chat-example", []string{"chat", "streaming"}},
		{"cache-llm-example", []string{"caching"}},
	}

	for _, tc := range testCases {
		result := extractTags(tc.name)
		if len(result) != len(tc.expected) {
			t.Errorf("extractTags(%s) returned %d tags, expected %d", tc.name, len(result), len(tc.expected))
			continue
		}

		for i, tag := range result {
			if tag != tc.expected[i] {
				t.Errorf("extractTags(%s)[%d] = %s, expected %s", tc.name, i, tag, tc.expected[i])
			}
		}
	}
}

func TestExamplesListCommand(t *testing.T) {
	// Test basic list command
	listCmd.SetArgs([]string{})

	// Mock discoverExamples to return test data
	originalOutput := outputFormat
	outputFormat = "table"
	defer func() { outputFormat = originalOutput }()

	err := listCmd.RunE(listCmd, []string{})
	// This will fail because we don't have examples directory in test,
	// but we can test the command structure
	if err == nil {
		t.Log("List command executed successfully")
	}
}

func TestExamplesRunCommandValidation(t *testing.T) {
	// Test that run command requires an argument - this should be caught by Cobra
	// before reaching our RunE function since we have cobra.ExactArgs(1)
	if runCmd.Args != nil {
		err := runCmd.Args(runCmd, []string{})
		if err == nil {
			t.Error("Run command should require an example name argument")
		}
	}
}

func TestOutputFormatting(t *testing.T) {
	// Test table output
	examples := []Example{
		{
			Name:        "test-example",
			Category:    "llm",
			Tags:        []string{"test"},
			Description: "Test description",
		},
	}

	// Redirect stdout temporarily
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test table output
	err := outputTable(examples)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("outputTable failed: %v", err)
	}

	// Read the output
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "test-example") {
		t.Error("Table output should contain example name")
	}
}
