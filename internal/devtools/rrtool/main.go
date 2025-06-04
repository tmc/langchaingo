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

	"github.com/tmc/langchaingo/internal/httprr"
)

func main() {
	if len(os.Args) < 2 {
		showUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "pack":
		packCmd()
	case "unpack":
		unpackCmd()
	case "check":
		checkCmd()
	case "clean":
		cleanCmd()
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

func packCmd() {
	fs := flag.NewFlagSet("pack", flag.ExitOnError)
	dirFlag := fs.String("dir", ".", "directory to process")
	recursiveFlag := fs.Bool("r", false, "process directories recursively")

	fs.Parse(os.Args[2:])

	if err := pack(*dirFlag, *recursiveFlag); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func unpackCmd() {
	fs := flag.NewFlagSet("unpack", flag.ExitOnError)
	dirFlag := fs.String("dir", ".", "directory to process")
	recursiveFlag := fs.Bool("r", false, "process directories recursively")

	fs.Parse(os.Args[2:])

	if err := unpack(*dirFlag, *recursiveFlag); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
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

func cleanCmd() {
	fs := flag.NewFlagSet("clean", flag.ExitOnError)
	dirFlag := fs.String("dir", ".", "directory to process")
	dryRunFlag := fs.Bool("dry-run", false, "show what would be removed without removing")

	fs.Parse(os.Args[2:])

	fmt.Printf("Cleaning up duplicate httprr files in %s...\n", *dirFlag)

	duplicates := make(map[string][]string)

	// Find all httprr files
	err := filepath.Walk(*dirFlag, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (strings.HasSuffix(path, ".httprr") || strings.HasSuffix(path, ".httprr.gz")) {
			// Get base name without extensions
			base := path
			if strings.HasSuffix(base, ".httprr.gz") {
				base = strings.TrimSuffix(base, ".httprr.gz")
			} else {
				base = strings.TrimSuffix(base, ".httprr")
			}

			duplicates[base] = append(duplicates[base], path)
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding files: %v\n", err)
		os.Exit(1)
	}

	// Process duplicates
	cleaned := 0
	for _, files := range duplicates {
		if len(files) > 1 {
			// We have both compressed and uncompressed versions
			var compressed, uncompressed string
			for _, f := range files {
				if strings.HasSuffix(f, ".httprr.gz") {
					compressed = f
				} else {
					uncompressed = f
				}
			}

			if compressed != "" && uncompressed != "" {
				// Determine which to remove based on modification time
				compressedInfo, _ := os.Stat(compressed)
				uncompressedInfo, _ := os.Stat(uncompressed)

				var toRemove string
				if compressedInfo.ModTime().After(uncompressedInfo.ModTime()) {
					toRemove = uncompressed
					fmt.Printf("Removing older uncompressed file: %s\n", uncompressed)
				} else {
					toRemove = compressed
					fmt.Printf("Removing older compressed file: %s\n", compressed)
				}

				if !*dryRunFlag {
					if err := os.Remove(toRemove); err != nil {
						fmt.Fprintf(os.Stderr, "Failed to remove %s: %v\n", toRemove, err)
					} else {
						cleaned++
					}
				} else {
					cleaned++
				}
			}
		}
	}

	if cleaned == 0 {
		fmt.Println("No duplicate files found")
	} else {
		if *dryRunFlag {
			fmt.Printf("Would clean up %d duplicate file pairs\n", cleaned)
		} else {
			fmt.Printf("Cleaned up %d duplicate file pairs\n", cleaned)
		}
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

func pack(dir string, recursive bool) error {
	fmt.Printf("Packing httprr files to compressed format in %s...\n", dir)

	var totalFiles, convertedFiles int

	if recursive {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".httprr") && !strings.HasSuffix(path, ".httprr.gz") {
				totalFiles++
				fmt.Printf("Compressing: %s\n", path)
				if err := httprr.CompressFile(path); err != nil {
					fmt.Printf("Failed to compress %s: %v\n", path, err)
				} else {
					convertedFiles++
					fmt.Printf("Successfully compressed httprr files in %s\n", filepath.Dir(path))
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	} else {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				path := filepath.Join(dir, entry.Name())
				if strings.HasSuffix(path, ".httprr") && !strings.HasSuffix(path, ".httprr.gz") {
					totalFiles++
					fmt.Printf("Compressing: %s\n", path)
					if err := httprr.CompressFile(path); err != nil {
						fmt.Printf("Failed to compress %s: %v\n", path, err)
					} else {
						convertedFiles++
						fmt.Printf("Successfully compressed httprr files in %s\n", dir)
					}
				}
			}
		}
	}

	if totalFiles == 0 {
		fmt.Println("No uncompressed .httprr files found")
	} else {
		fmt.Printf("Compressed %d/%d httprr files\n", convertedFiles, totalFiles)
	}

	return nil
}

func unpack(dir string, recursive bool) error {
	fmt.Printf("Unpacking httprr files to uncompressed format in %s...\n", dir)

	var totalFiles, convertedFiles int

	if recursive {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".httprr.gz") {
				totalFiles++
				fmt.Printf("Decompressing: %s\n", path)
				if err := httprr.DecompressFile(path); err != nil {
					fmt.Printf("Failed to decompress %s: %v\n", path, err)
				} else {
					convertedFiles++
					fmt.Printf("Successfully decompressed httprr files in %s\n", filepath.Dir(path))
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	} else {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				path := filepath.Join(dir, entry.Name())
				if strings.HasSuffix(path, ".httprr.gz") {
					totalFiles++
					fmt.Printf("Decompressing: %s\n", path)
					if err := httprr.DecompressFile(path); err != nil {
						fmt.Printf("Failed to decompress %s: %v\n", path, err)
					} else {
						convertedFiles++
						fmt.Printf("Successfully decompressed httprr files in %s\n", dir)
					}
				}
			}
		}
	}

	if totalFiles == 0 {
		fmt.Println("No compressed .httprr.gz files found")
	} else {
		fmt.Printf("Decompressed %d/%d httprr files\n", convertedFiles, totalFiles)
	}

	return nil
}

func showUsage() {
	fmt.Print(`Usage: rrtool <command> [options]

Commands:
  pack           Compress .httprr files to .httprr.gz format
  unpack         Decompress .httprr.gz files to .httprr format
  check          Check compression status (exit 1 if uncompressed files found)
  clean          Remove duplicate files when both compressed/uncompressed exist
  list-packages  List Go packages that use httprr
  help           Show this help message

Options:
  -dir string    Directory to process (default ".")
  -r             Process directories recursively (pack/unpack)
  -dry-run       Show what would be done without doing it (clean only)
  -format string Output format for list-packages: 'paths' or 'command' (default "paths")

Examples:
  # Compress all httprr files in current directory
  rrtool pack

  # Compress all httprr files recursively
  rrtool pack -r

  # Decompress files in specific directory
  rrtool unpack -dir ./testdata

  # Decompress all files recursively
  rrtool unpack -r

  # Check if all files are compressed (for CI/linting)
  rrtool check

  # Clean up duplicate files
  rrtool clean

  # Preview what clean would remove
  rrtool clean -dry-run

  # List all packages that use httprr
  rrtool list-packages

  # Generate go test command for all packages using httprr
  rrtool list-packages -format=command

  # Use the command output directly
  $(rrtool list-packages -format=command)

`)
}
