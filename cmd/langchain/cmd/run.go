package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func executeExample(examplePath, runFile string) error {
	cmd := exec.Command("go", "run", runFile)
	cmd.Dir = examplePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Set up environment
	cmd.Env = os.Environ()

	fmt.Printf("Executing: go run %s\n", runFile)
	fmt.Println(strings.Repeat("-", 50))

	err := cmd.Run()
	if err != nil {
		fmt.Println(strings.Repeat("-", 50))
		return fmt.Errorf("example execution failed: %w", err)
	}

	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("Example completed successfully!")
	return nil
}

func checkExampleDependencies(examplePath string) error {
	// Check if go.mod exists
	goModPath := filepath.Join(examplePath, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return fmt.Errorf("go.mod not found in example directory")
	}

	// Check if dependencies are downloaded
	cmd := exec.Command("go", "mod", "download")
	cmd.Dir = examplePath
	return cmd.Run()
}
