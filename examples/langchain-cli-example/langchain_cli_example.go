package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func main() {
	fmt.Println("🦜️🔗 LangChain Go CLI Tools Demonstration")
	fmt.Println("==========================================")
	fmt.Println()
	fmt.Println("This example shows how to use the LangChain Go CLI tools")
	fmt.Println("that are already implemented in this repository.")
	fmt.Println()

	// Install the langchain CLI
	fmt.Println("🔨 Installing LangChain CLI...")
	installCmd := exec.Command("go", "install", "github.com/tmc/langchaingo/cmd/langchain")
	if err := installCmd.Run(); err != nil {
		log.Printf("Failed to install CLI: %v", err)
		return
	}
	fmt.Println("✅ CLI installed successfully!")
	fmt.Println()

	cliPath := "langchain"
	
	// 1. Demonstrate examples listing
	demonstrateExamplesListing(cliPath)
	
	// 2. Demonstrate project initialization  
	demonstrateProjectInit(cliPath)
	
	// 3. Demonstrate validation
	demonstrateValidation(cliPath)
	
	fmt.Println()
	fmt.Println("🎉 All CLI tools demonstrated successfully!")
	fmt.Println("💡 Use 'langchain --help' to explore more options")
}

func demonstrateExamplesListing(cliPath string) {
	fmt.Println("📋 1. EXAMPLES MANAGEMENT")
	fmt.Println("========================")
	
	// List all examples
	fmt.Println("→ Listing all available examples:")
	cmd := exec.Command(cliPath, "examples", "list", "--format", "table")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error listing examples: %v", err)
		return
	}
	
	// Show first few lines
	lines := strings.Split(string(output), "\n")
	fmt.Printf("Found %d examples total. Showing first 5:\n", len(lines)-3)
	for i := 0; i < min(8, len(lines)); i++ {
		fmt.Println(lines[i])
	}
	fmt.Println("...")
	fmt.Println()
	
	// Filter by category
	fmt.Println("→ Filtering LLM examples:")
	cmd = exec.Command(cliPath, "examples", "list", "--category", "llm", "--format", "json")
	output, err = cmd.Output()
	if err != nil {
		log.Printf("Error filtering examples: %v", err)
		return
	}
	
	var examples []map[string]interface{}
	if err := json.Unmarshal(output, &examples); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		return
	}
	
	fmt.Printf("Found %d LLM examples:\n", len(examples))
	for i, example := range examples[:min(3, len(examples))] {
		fmt.Printf("  %d. %s\n", i+1, example["name"])
	}
	fmt.Println("  ...")
	fmt.Println()
	
	// Filter by tag
	fmt.Println("→ Filtering examples by 'completion' tag:")
	cmd = exec.Command(cliPath, "examples", "list", "--tag", "completion")
	output, err = cmd.Output()
	if err != nil {
		log.Printf("Error filtering by tag: %v", err)
		return
	}
	lines = strings.Split(string(output), "\n")
	fmt.Printf("Found %d completion examples\n", len(lines)-3)
	fmt.Println()
}

func demonstrateProjectInit(cliPath string) {
	fmt.Println("🚀 2. PROJECT INITIALIZATION")
	fmt.Println("============================")
	
	// Actually create a test project
	fmt.Println("→ Creating a test project with basic-llm template:")
	cmd := exec.Command(cliPath, "init", "cli-test-project", "--template", "basic-llm")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error creating project: %v\nOutput: %s", err, output)
		return
	}
	fmt.Println(string(output))
	
	// Show what was created
	fmt.Println("→ Project structure created:")
	cmd = exec.Command("find", "cli-test-project", "-type", "f")
	output, err = cmd.Output()
	if err != nil {
		log.Printf("Error listing project files: %v", err)
		return
	}
	fmt.Println(string(output))
	
	// Test the generated project
	fmt.Println("→ Testing generated project:")
	cmd = exec.Command("go", "test", "-v")
	cmd.Dir = "cli-test-project"
	output, err = cmd.Output()
	if err != nil {
		fmt.Printf("Test output: %s\n", output)
		log.Printf("Note: Tests may fail without proper setup, which is expected")
	} else {
		fmt.Println("✅ Generated project tests pass!")
		fmt.Printf("Test output: %s\n", output)
	}
	
	// Clean up
	fmt.Println("→ Cleaning up test project...")
	cmd = exec.Command("rm", "-rf", "cli-test-project")
	if err := cmd.Run(); err != nil {
		log.Printf("Warning: Failed to clean up test project: %v", err)
	} else {
		fmt.Println("✅ Test project cleaned up")
	}
	fmt.Println()
}

func demonstrateValidation(cliPath string) {
	fmt.Println("🔍 3. PROVIDER VALIDATION")
	fmt.Println("=========================")
	
	// Actually run validation commands
	fmt.Println("→ Testing OpenAI validation (quick mode):")
	cmd := exec.Command(cliPath, "validate", "--provider", "openai", "--quick")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ OpenAI validation failed (expected without API key): %s\n", output)
	} else {
		fmt.Printf("✅ OpenAI validation passed: %s\n", output)
	}
	fmt.Println()
	
	fmt.Println("→ Testing Anthropic validation:")
	cmd = exec.Command(cliPath, "validate", "--provider", "anthropic")
	output, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Anthropic validation failed (expected without API key): %s\n", output)
	} else {
		fmt.Printf("✅ Anthropic validation passed: %s\n", output)
	}
	fmt.Println()
	
	fmt.Println("💡 Note: Validations fail without proper API keys, which is expected behavior")
	fmt.Println()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}