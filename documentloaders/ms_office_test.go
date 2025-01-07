package documentloaders

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMSOfficeLoader(test *testing.T) {
	test.Parallel()

	docExpectedContent := "This is a .doc test file."
	docxExpectedContent := "This is a .docx test file."
	xlsxExpectedContent := "This is an .xlsx test file"
	pptxExpectedContent := "This is a .pptx test file"

	test.Run("Load .doc", func(t *testing.T) {
		t.Parallel()

		file, err := os.Open("./testdata/test.doc")
		require.NoError(t, err)
		defer file.Close()

		fileInfo, err := file.Stat()
		require.NoError(t, err)

		loader := NewOffice(file, fileInfo.Name(), fileInfo.Size())
		docs, err := loader.Load(context.Background())
		require.NoError(t, err)

		assert.Len(t, docs, 1)
		assert.True(t, strings.Contains(docs[0].PageContent, docExpectedContent))
	})

	test.Run("Load .docx", func(t *testing.T) {
		t.Parallel()

		file, err := os.Open("./testdata/test.docx")
		require.NoError(t, err)
		defer file.Close()

		fileInfo, err := file.Stat()
		require.NoError(t, err)

		loader := NewOffice(file, fileInfo.Name(), fileInfo.Size())
		docs, err := loader.Load(context.Background())
		require.NoError(t, err)

		assert.Len(t, docs, 1)
		assert.True(t, strings.Contains(docs[0].PageContent, docxExpectedContent))
	})

	test.Run("Load .xlsx", func(t *testing.T) {
		t.Parallel()

		file, err := os.Open("./testdata/test.xlsx")
		require.NoError(t, err)
		defer file.Close()

		fileInfo, err := file.Stat()
		require.NoError(t, err)

		loader := NewOffice(file, fileInfo.Name(), fileInfo.Size())
		docs, err := loader.Load(context.Background())
		require.NoError(t, err)

		assert.Len(t, docs, 1)
		assert.True(t, strings.Contains(docs[0].PageContent, xlsxExpectedContent))
	})

	test.Run("Load .pptx", func(t *testing.T) {
		t.Parallel()

		file, err := os.Open("./testdata/test.pptx")
		require.NoError(t, err)
		defer file.Close()

		fileInfo, err := file.Stat()
		require.NoError(t, err)

		loader := NewOffice(file, fileInfo.Name(), fileInfo.Size())
		docs, err := loader.Load(context.Background())
		require.NoError(t, err)

		assert.Len(t, docs, 1)
		assert.True(t, strings.Contains(docs[0].PageContent, pptxExpectedContent))
	})
}
