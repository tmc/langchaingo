package documentloaders

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/0xDezzy/langchaingo/textsplitter"
)

func TestNewEPUB(t *testing.T) {
	loader := NewEPUB("test.epub")
	if loader.filePath != "test.epub" {
		t.Errorf("Expected filePath to be 'test.epub', got %s", loader.filePath)
	}
	if loader.mode != "single" {
		t.Errorf("Expected default mode to be 'single', got %s", loader.mode)
	}
}

func TestNewEPUBWithMode(t *testing.T) {
	loader := NewEPUB("test.epub", WithMode("elements"))
	if loader.mode != "elements" {
		t.Errorf("Expected mode to be 'elements', got %s", loader.mode)
	}
}

func TestNewEPUBFromBytes(t *testing.T) {
	data := []byte("test data")
	loader := NewEPUBFromBytes(data)
	if loader.data == nil {
		t.Error("Expected data to be set")
	}
	if loader.mode != "single" {
		t.Errorf("Expected default mode to be 'single', got %s", loader.mode)
	}
}

func TestNewEPUBFromReader(t *testing.T) {
	reader := strings.NewReader("test content")
	loader := NewEPUBFromReader(reader)
	if loader.reader == nil {
		t.Error("Expected reader to be set")
	}
	if loader.mode != "single" {
		t.Errorf("Expected default mode to be 'single', got %s", loader.mode)
	}
}

func TestEPUBLoad_NoValidInput(t *testing.T) {
	loader := EPUB{}
	_, err := loader.Load(context.Background())
	if err == nil {
		t.Error("Expected error for loader with no valid input")
	}
	if !strings.Contains(err.Error(), "no valid input provided") {
		t.Errorf("Expected error message about no valid input, got: %s", err.Error())
	}
}

func TestEPUBLoad_InvalidFile(t *testing.T) {
	loader := NewEPUB("nonexistent.epub")
	_, err := loader.Load(context.Background())
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestEPUBLoadAndSplit_ErrorHandling(t *testing.T) {
	// Test that LoadAndSplit calls Load and then splits the documents
	loader := NewEPUB("nonexistent.epub")
	splitter := textsplitter.NewRecursiveCharacter()

	_, err := loader.LoadAndSplit(context.Background(), splitter)
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
	// The error should come from Load method, not from splitting
	if !strings.Contains(err.Error(), "failed to open EPUB") {
		t.Errorf("Expected error from Load method, got: %s", err.Error())
	}
}

// TestEPUBModes tests that the mode option affects the behavior
func TestEPUBModes(t *testing.T) {
	tests := []struct {
		name string
		mode string
	}{
		{"single mode", "single"},
		{"elements mode", "elements"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewEPUB("test.epub", WithMode(tt.mode))
			if loader.mode != tt.mode {
				t.Errorf("Expected mode %s, got %s", tt.mode, loader.mode)
			}
		})
	}
}

// TestEPUBLoad_RealFiles tests loading various EPUB files
func TestEPUBLoad_RealFiles(t *testing.T) {
	testCases := []struct {
		name        string
		filename    string
		expectError bool
	}{
		{
			name:        "standard EPUB",
			filename:    "testdata/test.epub",
			expectError: false,
		},
		{
			name:        "wrong mimetype EPUB",
			filename:    "testdata/wrong_mimetype.epub",
			expectError: true, // Invalid EPUB file, should error
		},
		{
			name:        "wrong TOC attribute EPUB",
			filename:    "testdata/wrong_toc_attribute.epub",
			expectError: false, // Should still work, just with wrong TOC attribute
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			loader := NewEPUB(tc.filename)
			docs, err := loader.Load(context.Background())

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", tc.name)
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error loading %s, got: %s", tc.name, err.Error())
			}

			if len(docs) != 1 {
				t.Errorf("Expected 1 document in single mode for %s, got %d", tc.name, len(docs))
			}

			doc := docs[0]
			if doc.PageContent == "" {
				t.Errorf("Expected non-empty page content for %s", tc.name)
			}

			// Check metadata
			if doc.Metadata["mode"] != "single" {
				t.Errorf("Expected mode 'single' for %s, got %v", tc.name, doc.Metadata["mode"])
			}

			if doc.Metadata["source"] != tc.filename {
				t.Errorf("Expected source '%s', got %v", tc.filename, doc.Metadata["source"])
			}

			t.Logf("%s: Title=%v, Author=%v, Chapters=%v, Content length=%d",
				tc.name, doc.Metadata["title"], doc.Metadata["author"],
				doc.Metadata["chapters"], len(doc.PageContent))
		})
	}
}

// TestEPUBLoad_ElementsMode tests loading multiple EPUB files in elements mode
func TestEPUBLoad_ElementsMode(t *testing.T) {
	testCases := []struct {
		filename    string
		expectError bool
	}{
		{"testdata/test.epub", false},
		{"testdata/wrong_mimetype.epub", true}, // Invalid EPUB
		{"testdata/wrong_toc_attribute.epub", false},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			loader := NewEPUB(tc.filename, WithMode("elements"))
			docs, err := loader.Load(context.Background())

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", tc.filename)
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error loading %s, got: %s", tc.filename, err.Error())
			}

			if len(docs) == 0 {
				t.Errorf("Expected at least one document in elements mode for %s", tc.filename)
			}

			t.Logf("%s: Found %d chapters", tc.filename, len(docs))

			// Check that each document has chapter metadata
			for i, doc := range docs {
				if doc.Metadata["chapter"] == nil {
					t.Errorf("Document %d missing chapter metadata in %s", i, tc.filename)
				}

				if doc.Metadata["total_chapters"] == nil {
					t.Errorf("Document %d missing total_chapters metadata in %s", i, tc.filename)
				}

				if doc.Metadata["mode"] != "elements" {
					t.Errorf("Document %d expected mode 'elements', got %v in %s", i, doc.Metadata["mode"], tc.filename)
				}

				if i < 3 { // Log first 3 chapters for debugging
					t.Logf("  Chapter %d: Title=%v, Content length=%d",
						i+1, doc.Metadata["chapter_title"], len(doc.PageContent))
				}
			}
		})
	}
}

// TestEPUBLoadAndSplit tests LoadAndSplit with multiple EPUB files
func TestEPUBLoadAndSplit(t *testing.T) {
	testCases := []struct {
		filename    string
		expectError bool
	}{
		{"testdata/test.epub", false},
		{"testdata/wrong_mimetype.epub", true}, // Invalid EPUB
		{"testdata/wrong_toc_attribute.epub", false},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			loader := NewEPUB(tc.filename)
			splitter := textsplitter.NewRecursiveCharacter(textsplitter.WithChunkSize(200))

			docs, err := loader.LoadAndSplit(context.Background(), splitter)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", tc.filename)
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error with LoadAndSplit for %s, got: %s", tc.filename, err.Error())
			}

			if len(docs) == 0 {
				t.Errorf("Expected at least one document after splitting for %s", tc.filename)
			}

			// Check that documents were split (should have more docs than original)
			originalDocs, _ := loader.Load(context.Background())
			if len(docs) <= len(originalDocs) {
				t.Logf("Note: %s may be too short to split. Original: %d, Split: %d",
					tc.filename, len(originalDocs), len(docs))
			}

			t.Logf("%s: Original docs=%d, Split docs=%d", tc.filename, len(originalDocs), len(docs))

			// Verify each split document has metadata
			for i, doc := range docs {
				if doc.Metadata["source"] == nil {
					t.Errorf("Split document %d missing source metadata for %s", i, tc.filename)
				}
				if i < 3 { // Log first 3 splits
					t.Logf("  Split %d: Content length=%d", i+1, len(doc.PageContent))
				}
			}
		})
	}
}

// TestEPUBFromBytes tests loading EPUB from byte data
func TestEPUBFromBytes(t *testing.T) {
	testCases := []struct {
		filename    string
		expectError bool
	}{
		{"testdata/test.epub", false},
		{"testdata/wrong_mimetype.epub", true}, // Invalid EPUB
		{"testdata/wrong_toc_attribute.epub", false},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			// Read file into bytes
			data, err := os.ReadFile(tc.filename)
			if err != nil {
				t.Fatalf("Failed to read test file %s: %v", tc.filename, err)
			}

			// Test loading from bytes
			loader := NewEPUBFromBytes(data)
			docs, err := loader.Load(context.Background())

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", tc.filename)
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error loading %s from bytes, got: %s", tc.filename, err.Error())
			}

			if len(docs) != 1 {
				t.Errorf("Expected 1 document for %s, got %d", tc.filename, len(docs))
			}

			if docs[0].PageContent == "" {
				t.Errorf("Expected non-empty content for %s", tc.filename)
			}

			// Source should not be set when loading from bytes
			if docs[0].Metadata["source"] != nil {
				t.Errorf("Expected no source in metadata when loading from bytes, got: %v",
					docs[0].Metadata["source"])
			}

			t.Logf("%s from bytes: Title=%v, Content length=%d",
				tc.filename, docs[0].Metadata["title"], len(docs[0].PageContent))
		})
	}
}

// TestEPUBCompareLoadingMethods compares loading the same file via different methods
func TestEPUBCompareLoadingMethods(t *testing.T) {
	filename := "testdata/test.epub"

	// Load from file path
	fileLoader := NewEPUB(filename)
	fileDocs, err := fileLoader.Load(context.Background())
	if err != nil {
		t.Fatalf("Failed to load from file path: %v", err)
	}

	// Load from bytes
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	bytesLoader := NewEPUBFromBytes(data)
	bytesDocs, err := bytesLoader.Load(context.Background())
	if err != nil {
		t.Fatalf("Failed to load from bytes: %v", err)
	}

	// Load from reader
	reader, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer reader.Close()
	readerLoader := NewEPUBFromReader(reader)
	readerDocs, err := readerLoader.Load(context.Background())
	if err != nil {
		t.Fatalf("Failed to load from reader: %v", err)
	}

	// Compare content (should be identical)
	if fileDocs[0].PageContent != bytesDocs[0].PageContent {
		t.Error("File and bytes loading produced different content")
	}

	if fileDocs[0].PageContent != readerDocs[0].PageContent {
		t.Error("File and reader loading produced different content")
	}

	// Compare metadata (source should differ)
	if fileDocs[0].Metadata["title"] != bytesDocs[0].Metadata["title"] {
		t.Error("File and bytes loading produced different titles")
	}

	t.Logf("All three loading methods produced consistent content (%d chars)",
		len(fileDocs[0].PageContent))
}

// TestEPUBImplementsLoader verifies that EPUB implements the Loader interface
func TestEPUBImplementsLoader(t *testing.T) {
	var _ Loader = EPUB{}
}
