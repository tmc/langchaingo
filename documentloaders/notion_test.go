package documentloaders

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/schema"
)

func TestNotionDirectoryLoader_Load(t *testing.T) {
	t.Parallel()

	// Create a temporary test directory
	tempDir, err := os.MkdirTemp("", "notion_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create sample Markdown files in the temporary directory
	testFiles := []struct {
		name     string
		content  string
		expected schema.Document
	}{
		{
			name:    "test1.md",
			content: "# Test Document 1\nThis is test document 1.",
			expected: schema.Document{
				PageContent: "# Test Document 1\nThis is test document 1.",
				Metadata:    map[string]interface{}{"source": filepath.Join(tempDir, "test1.md")},
			},
		},
		{
			name:    "test2.md",
			content: "# Test Document 2\nThis is test document 2.",
			expected: schema.Document{
				PageContent: "# Test Document 2\nThis is test document 2.",
				Metadata:    map[string]interface{}{"source": filepath.Join(tempDir, "test2.md")},
			},
		},
	}

	for _, file := range testFiles {
		filePath := filepath.Join(tempDir, file.name)
		err := os.WriteFile(filePath, []byte(file.content), 0o600)
		require.NoError(t, err)
	}

	// Create a NotionDirectoryLoader instance
	loader := NewNotionDirectory(tempDir)

	// Load documents from the test directory
	docs, err := loader.Load()
	require.NoError(t, err)

	// Verify the loaded documents match the expected ones
	require.Len(t, docs, len(testFiles))
	for i, expected := range testFiles {
		assert.Equal(t, expected.expected, docs[i])
	}
}
