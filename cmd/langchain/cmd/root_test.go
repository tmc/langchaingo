package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommand(t *testing.T) {
	// Test that root command executes without error
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	// Capture output
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Root command execution failed: %v", err)
	}
}

func TestRootCommandHelp(t *testing.T) {
	// Test help output
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Help command failed: %v", err)
	}

	output := out.String()

	// Check for key elements in help output
	expectedStrings := []string{
		"LangChain Go CLI",
		"examples",
		"init",
		"validate",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Help output missing expected string: %s", expected)
		}
	}
}

func TestCommandStructure(t *testing.T) {
	// Test that all expected commands are registered
	expectedCommands := []string{"examples", "init", "validate"}

	for _, cmdName := range expectedCommands {
		cmd, _, err := rootCmd.Find([]string{cmdName})
		if err != nil {
			t.Errorf("Command %s not found: %v", cmdName, err)
			continue
		}

		if cmd.Name() != cmdName {
			t.Errorf("Expected command name %s, got %s", cmdName, cmd.Name())
		}
	}
}

func TestSubcommandStructure(t *testing.T) {
	// Test examples subcommands
	examplesCmd, _, err := rootCmd.Find([]string{"examples"})
	if err != nil {
		t.Fatalf("Examples command not found: %v", err)
	}

	expectedSubcommands := []string{"list", "run"}
	for _, subCmd := range expectedSubcommands {
		found := false
		for _, child := range examplesCmd.Commands() {
			if child.Name() == subCmd {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Examples subcommand %s not found", subCmd)
		}
	}
}
