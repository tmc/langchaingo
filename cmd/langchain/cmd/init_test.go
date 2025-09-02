package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateProjectDirectory(t *testing.T) {
	tempDir := t.TempDir()
	projectName := "test-project"
	projectPath := filepath.Join(tempDir, projectName)

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Warning: failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	err := createProjectDirectory(projectName)
	if err != nil {
		t.Fatalf("createProjectDirectory failed: %v", err)
	}

	// Check directory was created
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		t.Error("Project directory was not created")
	}
}

func TestCreateProjectDirectoryExists(t *testing.T) {
	tempDir := t.TempDir()
	projectName := "test-project"
	projectPath := filepath.Join(tempDir, projectName)

	// Create directory first
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Warning: failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Should fail without force flag
	forceCreate = false
	err := createProjectDirectory(projectName)
	if err == nil {
		t.Error("createProjectDirectory should fail when directory exists")
	}

	// Should succeed with force flag
	forceCreate = true
	err = createProjectDirectory(projectName)
	if err != nil {
		t.Errorf("createProjectDirectory should succeed with force flag: %v", err)
	}
}

func TestGenerateFromTemplate(t *testing.T) {
	tempDir := t.TempDir()
	projectName := "test-project"

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Warning: failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create project directory
	err := os.MkdirAll(projectName, 0755)
	if err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Test with basic-llm template
	err = generateFromTemplate(projectName, "basic-llm")
	if err != nil {
		t.Fatalf("generateFromTemplate failed: %v", err)
	}

	// Check that files were created
	expectedFiles := []string{
		"main.go",
		"go.mod",
		"README.md",
		".env.example",
	}

	for _, file := range expectedFiles {
		filePath := filepath.Join(projectName, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not created", file)
		}
	}

	// Check content of main.go
	mainGoPath := filepath.Join(projectName, "main.go")
	content, err := os.ReadFile(mainGoPath)
	if err != nil {
		t.Fatalf("Failed to read main.go: %v", err)
	}

	if !strings.Contains(string(content), "package main") {
		t.Error("main.go should contain 'package main'")
	}

	if !strings.Contains(string(content), "langchaingo") {
		t.Error("main.go should import langchaingo")
	}
}

func TestGenerateFromInvalidTemplate(t *testing.T) {
	tempDir := t.TempDir()
	projectName := "test-project"

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Warning: failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create project directory
	if err := os.MkdirAll(projectName, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Test with invalid template
	err := generateFromTemplate(projectName, "invalid-template")
	if err == nil {
		t.Error("generateFromTemplate should fail with invalid template")
	}

	if !strings.Contains(err.Error(), "template 'invalid-template' not found") {
		t.Error("Error message should mention template not found")
	}
}

func TestGetTemplateNames(t *testing.T) {
	names := getTemplateNames()

	expectedTemplates := []string{"basic-llm", "chat-bot", "rag-system"}

	if len(names) != len(expectedTemplates) {
		t.Errorf("Expected %d templates, got %d", len(expectedTemplates), len(names))
	}

	for _, expected := range expectedTemplates {
		found := false
		for _, name := range names {
			if name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Template %s not found in template names", expected)
		}
	}
}

func TestTemplateContent(t *testing.T) {
	// Test that templates contain expected content
	for templateName, template := range templates {
		// Check that main.go exists
		if _, exists := template.Files["main.go"]; !exists {
			t.Errorf("Template %s should have main.go", templateName)
		}

		// Check that go.mod exists
		if _, exists := template.Files["go.mod"]; !exists {
			t.Errorf("Template %s should have go.mod", templateName)
		}

		// Check that README.md exists
		if _, exists := template.Files["README.md"]; !exists {
			t.Errorf("Template %s should have README.md", templateName)
		}

		// Check main.go content
		mainContent := template.Files["main.go"]
		if !strings.Contains(mainContent, "package main") {
			t.Errorf("Template %s main.go should contain 'package main'", templateName)
		}

		if !strings.Contains(mainContent, "langchaingo") {
			t.Errorf("Template %s main.go should import langchaingo", templateName)
		}
	}
}

func TestInitCommandValidation(t *testing.T) {
	// Test that init command requires project name - this should be caught by Cobra
	// before reaching our RunE function since we have cobra.ExactArgs(1)
	if initCmd.Args != nil {
		err := initCmd.Args(initCmd, []string{})
		if err == nil {
			t.Error("Init command should require project name argument")
		}
	}
}
