package documentloaders

import (
	"context"
	"os"
	"testing"

	"github.com/ledongthuc/pdf"
	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/textsplitter"
)

func TestPDFLoader(t *testing.T) {
	t.Parallel()

	page1Content := " A Simple PDF File  This is a small demonstration .pdf file -  " +
		"just for use in the Virtual Mechanics tutorials. More text. And more  text. And more " +
		"text. And more text. And more text.  And more text. And more text. And more text. " +
		"And more text. And more  text. And more text. Boring, zzzzz. And more text. And more " +
		"text. And  more text. And more text. And more text. And more text. And more text.  " +
		"And more text. And more text.  And more text. And more text. And more text. And more " +
		"text. And more  text. And more text. And more text. Even more. Continued on page 2 ..."

	page2Content := " Simple PDF File 2  ...continued from page 1. Yet more text. And more " +
		"text. And more text.  And more text. And more text. And more text. And more text. And more " +
		" text. Oh, how boring typing this stuff. But not as boring as watching  paint dry. And more " +
		"text. And more text. And more text. And more text.  Boring.  More, a little more text. " +
		"The end, and just as well. "

	expectedResults := []struct {
		content  string
		metadata map[string]any
	}{
		{content: page1Content, metadata: map[string]any{"page": 1, "total_pages": 2}},
		{content: page2Content, metadata: map[string]any{"page": 2, "total_pages": 2}},
	}

	t.Run("PDFLoad", func(t *testing.T) {
		t.Parallel()
		f, err := os.Open("./testdata/sample.pdf")
		assert.NoError(t, err)
		defer f.Close()
		finfo, err := f.Stat()
		assert.NoError(t, err)
		p := NewPDF(f, finfo.Size())
		docs, err := p.Load(context.Background())
		assert.NoError(t, err)

		assert.Len(t, docs, 2)

		for r := range expectedResults {
			assert.Equal(t, expectedResults[r].content, docs[r].PageContent)
			assert.Equal(t, expectedResults[r].metadata, docs[r].Metadata)
		}
	})

	t.Run("PDFLoadPassword", func(t *testing.T) {
		t.Parallel()
		f, err := os.Open("./testdata/sample_password.pdf")
		assert.NoError(t, err)
		defer f.Close()
		finfo, err := f.Stat()
		assert.NoError(t, err)
		p := NewPDF(f, finfo.Size(), WithPassword("password"))
		docs, err := p.Load(context.Background())
		assert.NoError(t, err)

		assert.Len(t, docs, 2)

		for r := range expectedResults {
			assert.Equal(t, expectedResults[r].content, docs[r].PageContent)
			assert.Equal(t, expectedResults[r].metadata, docs[r].Metadata)
		}
	})

	t.Run("PDFLoadPasswordWrong", func(t *testing.T) {
		t.Parallel()
		f, err := os.Open("./testdata/sample_password.pdf")
		assert.NoError(t, err)
		defer f.Close()
		finfo, err := f.Stat()
		assert.NoError(t, err)
		p := NewPDF(f, finfo.Size(), WithPassword("password1"))
		docs, err := p.Load(context.Background())
		assert.Errorf(t, err, pdf.ErrInvalidPassword.Error())

		assert.Len(t, docs, 0)
	})
}

func TestPDFTextSplit(t *testing.T) {
	t.Parallel()
	page1_1Content := "A Simple PDF File  This is a small demonstration .pdf file -  " +
		"just for use in the Virtual Mechanics tutorials. More text. And more  text. And more " +
		"text. And more text. And more text.  And more text. And more text. And more text. And " +
		"more text. And more  text. And more text. Boring, zzzzz. And more"
	page1_2Content := "text. Boring, zzzzz. And more text. And more text. And  more text. And " +
		"more text. And more text. And more text. And more text.  And more text. And more text.  And " +
		"more text. And more text. And more text. And more text. And more  text. And more text. And " +
		"more text. Even more. Continued on page 2 ..."

	page2_1Content := "Simple PDF File 2  ...continued from page 1. Yet more text. And more text. " +
		"And more text.  And more text. And more text. And more text. And more text. And more  text. " +
		"Oh, how boring typing this stuff. But not as boring as watching  paint dry. And more text. " +
		"And more text. And more text. And more"
	page2_2Content := "text. And more text. And more text.  Boring.  More, a little more text. The end, and just as well."

	expectedResults := []struct {
		content  string
		metadata map[string]any
	}{
		{content: page1_1Content, metadata: map[string]any{"page": 1, "total_pages": 2}},
		{content: page1_2Content, metadata: map[string]any{"page": 1, "total_pages": 2}},
		{content: page2_1Content, metadata: map[string]any{"page": 2, "total_pages": 2}},
		{content: page2_2Content, metadata: map[string]any{"page": 2, "total_pages": 2}},
	}

	t.Run("PDFTextSplit", func(t *testing.T) {
		t.Parallel()
		f, err := os.Open("./testdata/sample.pdf")
		assert.NoError(t, err)
		defer f.Close()
		finfo, err := f.Stat()
		assert.NoError(t, err)
		p := NewPDF(f, finfo.Size())
		split := textsplitter.NewRecursiveCharacter()
		split.ChunkSize = 300
		split.ChunkOverlap = 30
		docs, err := p.LoadAndSplit(context.Background(), split)
		assert.NoError(t, err)

		assert.Len(t, docs, 4)

		for r := range expectedResults {
			assert.Equal(t, expectedResults[r].content, docs[r].PageContent)
			assert.Equal(t, expectedResults[r].metadata, docs[r].Metadata)
		}
	})
}
