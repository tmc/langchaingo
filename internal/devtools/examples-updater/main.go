// Package main provides a tool for updating example dependencies.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	version = flag.String("version", "", "Version to update to (e.g., v0.1.14-pre.1)")
	dryRun  = flag.Bool("dry-run", false, "Show what would be changed without making changes")
	verbose = flag.Bool("v", false, "Verbose output")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Update langchaingo version in all example go.mod files.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Update all examples to v0.1.14-pre.1\n")
		fmt.Fprintf(os.Stderr, "  go run ./internal/devtools/examples-updater -version v0.1.14-pre.1\n\n")
		fmt.Fprintf(os.Stderr, "  # Dry run to see what would change\n")
		fmt.Fprintf(os.Stderr, "  go run ./internal/devtools/examples-updater -version v0.1.14-pre.1 -dry-run\n")
	}
	flag.Parse()

	if *version == "" {
		log.Fatal("Version is required. Use -version flag to specify the target version.")
	}

	if err := updateExamples(*version, *dryRun); err != nil {
		log.Fatal(err)
	}
}

func updateExamples(newVersion string, dryRun bool) error {
	// Find all go.mod files in examples directory
	modFiles, err := findGoModFiles("examples")
	if err != nil {
		return fmt.Errorf("failed to find go.mod files: %w", err)
	}

	if len(modFiles) == 0 {
		return fmt.Errorf("no go.mod files found in examples directory")
	}

	fmt.Printf("Found %d go.mod files to update\n", len(modFiles))
	if dryRun {
		fmt.Println("DRY RUN - no changes will be made")
	}
	fmt.Println()

	// Track files that need replace directive removal
	var replaceFiles []string

	for _, modFile := range modFiles {
		if *verbose {
			fmt.Printf("Processing %s\n", modFile)
		}

		// Read the file
		content, err := os.ReadFile(modFile)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", modFile, err)
		}

		original := string(content)
		updated := original

		// Update version
		versionRe := regexp.MustCompile(`github\.com/tmc/langchaingo v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?`)
		if versionRe.MatchString(updated) {
			oldVersion := versionRe.FindString(updated)
			updated = versionRe.ReplaceAllString(updated, "github.com/tmc/langchaingo "+newVersion)
			if *verbose && oldVersion != "github.com/tmc/langchaingo "+newVersion {
				fmt.Printf("  Version: %s → %s\n", strings.TrimPrefix(oldVersion, "github.com/tmc/langchaingo "), newVersion)
			}
		}

		// Check for replace directives that can be removed
		if strings.Contains(updated, "replace github.com/tmc/langchaingo =>") {
			replaceFiles = append(replaceFiles, modFile)
			// Remove replace directive and its comment
			updated = removeReplaceDirective(updated)
			if *verbose {
				fmt.Printf("  Removing replace directive\n")
			}
		}

		// Only write if changed
		if original != updated {
			if !dryRun {
				if err := os.WriteFile(modFile, []byte(updated), 0644); err != nil {
					return fmt.Errorf("failed to write %s: %w", modFile, err)
				}
			}
			fmt.Printf("✓ Updated %s\n", modFile)
		} else if *verbose {
			fmt.Printf("  No changes needed\n")
		}
	}

	// Summary
	fmt.Println()
	if len(replaceFiles) > 0 {
		fmt.Printf("Replace directives removed from %d files:\n", len(replaceFiles))
		for _, f := range replaceFiles {
			fmt.Printf("  - %s\n", f)
		}
	}

	if dryRun {
		fmt.Println("\nDRY RUN complete - no files were modified")
	} else {
		fmt.Printf("\n✅ Successfully updated %d go.mod files to %s\n", len(modFiles), newVersion)
	}

	return nil
}

func findGoModFiles(root string) ([]string, error) {
	var modFiles []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == "go.mod" {
			modFiles = append(modFiles, path)
		}
		return nil
	})
	return modFiles, err
}

func removeReplaceDirective(content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var result strings.Builder
	var skipNext bool

	for scanner.Scan() {
		line := scanner.Text()

		// Skip comment lines before replace directive
		if strings.Contains(line, "Temporary replace directive") ||
			strings.Contains(line, "TODO: Remove after") {
			skipNext = true
			continue
		}

		// Skip the replace directive itself
		if skipNext && strings.HasPrefix(strings.TrimSpace(line), "replace github.com/tmc/langchaingo") {
			skipNext = false
			continue
		}

		// Skip empty lines after replace directive was removed
		if skipNext && strings.TrimSpace(line) == "" {
			skipNext = false
			continue
		}

		result.WriteString(line)
		result.WriteString("\n")
	}

	// Clean up any trailing newlines
	output := result.String()
	output = strings.TrimRight(output, "\n")
	output += "\n"

	return output
}
