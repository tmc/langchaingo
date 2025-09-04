package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		showUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "check":
		checkCmd()
	case "list-packages":
		listPackagesCmd()
	case "help", "--help", "-h":
		showUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		showUsage()
		os.Exit(1)
	}
}

func checkCmd() {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	dirFlag := fs.String("dir", ".", "directory to process")

	fs.Parse(os.Args[2:])

	fmt.Printf("Checking httprr file compression status in %s...\n", *dirFlag)

	var uncompressedFiles []string
	err := filepath.Walk(*dirFlag, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip devtools directory itself
		if strings.Contains(path, "/internal/devtools/") {
			return nil
		}

		if !info.IsDir() && strings.HasSuffix(path, ".httprr") && !strings.HasSuffix(path, ".httprr.gz") {
			uncompressedFiles = append(uncompressedFiles, path)
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking files: %v\n", err)
		os.Exit(1)
	}

	if len(uncompressedFiles) == 0 {
		fmt.Println("✓ All httprr files are properly compressed")
		os.Exit(0)
	} else {
		fmt.Printf("⚠ Found %d uncompressed httprr files:\n", len(uncompressedFiles))
		for _, file := range uncompressedFiles {
			relPath, _ := filepath.Rel(*dirFlag, file)
			fmt.Printf("  - %s\n", relPath)
		}
		os.Exit(1)
	}
}

func listPackagesCmd() {
	fs := flag.NewFlagSet("list-packages", flag.ExitOnError)
	dirFlag := fs.String("dir", ".", "directory to process")
	formatFlag := fs.String("format", "paths", "output format: 'paths' or 'command'")

	fs.Parse(os.Args[2:])

	packages, err := findPackagesWithHttprr(*dirFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(packages) == 0 {
		if *formatFlag == "command" {
			fmt.Print("# No packages using httprr found")
		} else {
			fmt.Print("# No packages using httprr found")
		}
		return
	}

	switch *formatFlag {
	case "command":
		fmt.Printf("go test -httprecord=. %s", strings.Join(packages, " "))
	case "paths":
		fmt.Print(strings.Join(packages, " "))
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown format '%s'. Use 'paths' or 'command'\n", *formatFlag)
		os.Exit(1)
	}
}

// findPackagesWithHttprr scans for Go packages that import httprr
func findPackagesWithHttprr(rootDir string) ([]string, error) {
	var packages []string
	seen := make(map[string]bool)

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-Go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip vendor directories
		if strings.Contains(path, "/vendor/") {
			return nil
		}

		// Check if file imports httprr
		hasHttprr, err := fileImportsHttprr(path)
		if err != nil {
			return err
		}

		if hasHttprr {
			// Get the package path relative to the module root
			pkgDir := filepath.Dir(path)

			// Convert to Go module path format
			relPath, err := filepath.Rel(rootDir, pkgDir)
			if err != nil {
				return err
			}

			// Convert to Go package path format
			pkgPath := "./" + filepath.ToSlash(relPath)
			if pkgPath == "./" {
				pkgPath = "."
			}

			if !seen[pkgPath] {
				packages = append(packages, pkgPath)
				seen[pkgPath] = true
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort packages for consistent output
	sort.Strings(packages)
	return packages, nil
}

// fileImportsHttprr checks if a Go file imports the httprr package
func fileImportsHttprr(filePath string) (bool, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ImportsOnly)
	if err != nil {
		// If we can't parse the file, assume it doesn't import httprr
		return false, nil
	}

	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		if strings.Contains(importPath, "httprr") {
			return true, nil
		}
	}

	return false, nil
}

func showUsage() {
	fmt.Print(`Usage: rrtool <command> [options]

Commands:
  clean          Remove duplicate files when both compressed/uncompressed exist
  list-packages  List Go packages that use httprr
  help           Show this help message

Options:
  -dir string    Directory to process (default ".")
  -r             Process directories recursively (pack/unpack)
  -dry-run       Show what would be done without doing it (clean only)
  -format string Output format for list-packages: 'paths' or 'command' (default "paths")
`)
}
