// Package main provides a tool to normalize version information in httprr recordings.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	dryRun  = flag.Bool("dry-run", false, "Show what would be changed without modifying files")
	verbose = flag.Bool("v", false, "Verbose output")
)

func main() {
	flag.Parse()

	// Find all .httprr files
	recordings, err := filepath.Glob("**/*.httprr")
	if err != nil {
		log.Fatal(err)
	}
	
	// Also check common test directories
	testDirs := []string{
		"chains/testdata",
		"llms/*/testdata",
		"vectorstores/*/testdata",
		"embeddings/*/testdata",
	}
	
	for _, pattern := range testDirs {
		matches, err := filepath.Glob(pattern + "/*.httprr")
		if err == nil {
			recordings = append(recordings, matches...)
		}
	}
	
	// Deduplicate
	seen := make(map[string]bool)
	var unique []string
	for _, r := range recordings {
		if !seen[r] {
			seen[r] = true
			unique = append(unique, r)
		}
	}
	recordings = unique
	
	if len(recordings) == 0 {
		fmt.Println("No .httprr files found")
		return
	}
	
	fmt.Printf("Found %d .httprr files\n", len(recordings))
	
	for _, file := range recordings {
		if err := normalizeRecording(file); err != nil {
			log.Printf("Error processing %s: %v", file, err)
		}
	}
}

func normalizeRecording(filename string) error {
	// Read the file
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	
	// Check if it's a valid httprr file
	if !strings.HasPrefix(string(content), "httprr trace v1\n") {
		return fmt.Errorf("not a valid httprr trace file")
	}
	
	originalContent := string(content)
	normalized := normalizeHTTPRRFile(originalContent)
	
	if normalized == originalContent {
		if *verbose {
			fmt.Printf("No changes needed for %s\n", filename)
		}
		return nil
	}
	
	if *dryRun {
		fmt.Printf("Would update %s:\n", filename)
		// Show a sample of changes
		showChanges(originalContent, normalized)
		return nil
	}
	
	// Write the normalized content back
	if err := os.WriteFile(filename, []byte(normalized), 0644); err != nil {
		return err
	}
	
	fmt.Printf("Updated %s\n", filename)
	return nil
}

func normalizeHTTPRRFile(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) < 2 || lines[0] != "httprr trace v1" {
		return content // Not a valid httprr file
	}
	
	var result strings.Builder
	result.WriteString("httprr trace v1\n")
	
	i := 1
	for i < len(lines) {
		// Skip empty lines
		if i >= len(lines) || strings.TrimSpace(lines[i]) == "" {
			i++
			continue
		}
		
		// Parse the byte count line (e.g., "773 1369")
		parts := strings.Fields(lines[i])
		if len(parts) != 2 {
			// Malformed, return original
			return content
		}
		
		reqBytes, err1 := strconv.Atoi(parts[0])
		respBytes, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			return content // Malformed
		}
		
		i++ // Move past the count line
		
		// Extract the request and response based on byte counts
		startPos := strings.Index(content[strings.Index(content, lines[i]):], lines[i])
		if startPos < 0 {
			return content
		}
		
		// Find the actual position in the full content
		fullPos := 0
		for j := 0; j < i && j < len(lines); j++ {
			fullPos += len(lines[j]) + 1 // +1 for newline
		}
		
		remaining := content[fullPos:]
		if len(remaining) < reqBytes+respBytes {
			return content // Not enough data
		}
		
		request := remaining[:reqBytes]
		response := remaining[reqBytes:reqBytes+respBytes]
		
		// Normalize the request and response
		normalizedReq := normalizeContent(request)
		normalizedResp := normalizeContent(response)
		
		// Write the updated byte counts and data
		result.WriteString(fmt.Sprintf("%d %d\n", len(normalizedReq), len(normalizedResp)))
		result.WriteString(normalizedReq)
		result.WriteString(normalizedResp)
		
		// Move to the next record
		// We need to find where we are in the lines array after consuming the bytes
		consumed := reqBytes + respBytes
		pos := fullPos + consumed
		
		// Find which line we're on now
		currentPos := 0
		for j := 0; j < len(lines); j++ {
			if currentPos >= pos {
				i = j
				break
			}
			currentPos += len(lines[j]) + 1
		}
		
		if i >= len(lines)-1 {
			break
		}
	}
	
	return result.String()
}

func normalizeContent(content string) string {
	// Normalize x-goog-api-client header
	googClientPattern := regexp.MustCompile(`(x-goog-api-client: )gl-go/[\d.]+ gccl/v?[\d.]+ genai-go/[\d.]+ gapic/[\d.]+ gax/[\d.]+`)
	content = googClientPattern.ReplaceAllString(content, "${1}gl-go/X.X.X gccl/X.X.X genai-go/X.X.X gapic/X.X.X gax/X.X.X")
	
	// Normalize other version patterns in x-goog-api-client
	googClientVersionPattern := regexp.MustCompile(`(x-goog-api-client: [^\r\n]*)/v?\d+\.\d+(\.\d+)?`)
	content = googClientVersionPattern.ReplaceAllString(content, "${1}/X.X.X")
	
	// Normalize x-amz-user-agent header  
	amzPattern := regexp.MustCompile(`(x-amz-user-agent: [^\r\n]*)\bv?\d+\.\d+(\.\d+)?(-[a-zA-Z0-9.]+)?`)
	content = amzPattern.ReplaceAllString(content, "${1}X.X.X")
	
	// Normalize Go version in various headers
	goVersionPattern := regexp.MustCompile(`\bgo\d+\.\d+(\.\d+)?\b`)
	content = goVersionPattern.ReplaceAllString(content, "goX.X.X")
	
	return content
}

func showChanges(original, normalized string) {
	// Show first few differences
	origLines := strings.Split(original, "\n")
	normLines := strings.Split(normalized, "\n")
	
	shown := 0
	maxShow := 5
	
	for i := 0; i < len(origLines) && i < len(normLines); i++ {
		if origLines[i] != normLines[i] && shown < maxShow {
			fmt.Printf("  Line %d:\n", i+1)
			fmt.Printf("    - %s\n", origLines[i])
			fmt.Printf("    + %s\n", normLines[i])
			shown++
		}
	}
	
	if shown == maxShow {
		fmt.Println("  ... (more changes)")
	}
}