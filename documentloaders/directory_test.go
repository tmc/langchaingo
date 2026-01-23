package documentloaders

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRecursiveDirLoader_Options(t *testing.T) {
	t.Run("default option", func(t *testing.T) {
		l := NewRecursiveDirLoader()
		assert.Equal(t, ".", l.root)
		assert.Equal(t, 1, l.maxDepth)
		assert.Empty(t, l.allowExt)
	})

	t.Run("custom option", func(t *testing.T) {
		l := NewRecursiveDirLoader(
			WithRoot("./testdata"),
			WithMaxDepth(2),
			WithAllowExts(".txt", ".csv"),
			WithCSVOpts([]string{"name", "age", "city", "country"}),
			WithPDFOpts("password1"),
		)
		assert.Contains(t, "./testdata", l.root)
		assert.Equal(t, 2, l.maxDepth)
		assert.Len(t, l.allowExt, 2)
		assert.Contains(t, l.allowExt, ".txt")
		assert.Contains(t, l.allowExt, ".csv")
		assert.Equal(t, []string{"name", "age", "city", "country"}, l.Columns)
		assert.Equal(t, "password1", l.PDFPassword)
	})
}

func TestLoad_AllSupportedTypes(t *testing.T) {
	ctx := context.Background()
	root := "./testdata"

	t.Run("txt", func(t *testing.T) {
		l := NewRecursiveDirLoader(
			WithRoot(root),
			WithAllowExts(".txt"),
		)
		docs, err := l.Load(ctx)
		require.NoError(t, err)
		require.Len(t, docs, 1)
		assert.Contains(t, docs[0].PageContent, "Foo Bar Baz")
	})

	t.Run("csv", func(t *testing.T) {
		l := NewRecursiveDirLoader(
			WithRoot(root),
			WithCSVOpts([]string{"name", "age", "city", "country"}),
			WithAllowExts(".csv"),
		)
		docs, err := l.Load(ctx)
		require.NoError(t, err)
		require.Len(t, docs, 20)
		assert.Contains(t, docs[0].PageContent, "name: John Doe\nage: 25\ncity: New York\ncountry: United States")
	})

	t.Run("pdf valid password", func(t *testing.T) {
		l := NewRecursiveDirLoader(
			WithRoot(root),
			WithMaxDepth(2),
			WithAllowExts(".pdf"),
			WithPDFOpts("password"),
		)
		docs, err := l.Load(ctx)
		require.NoError(t, err)
		assert.Len(t, docs, 4)
		assert.Contains(t, docs[0].PageContent, "Simple PDF")
	})

	t.Run("pdf invalid password", func(t *testing.T) {
		l := NewRecursiveDirLoader(
			WithRoot(root),
			WithMaxDepth(2),
			WithAllowExts(".pdf"),
			WithPDFOpts("password1"),
		)
		docs, err := l.Load(ctx)
		require.NoError(t, err)
		assert.Len(t, docs, 2)
	})

	t.Run("depth limit", func(t *testing.T) {
		l := NewRecursiveDirLoader(
			WithRoot(root),
			WithMaxDepth(2),
			WithAllowExts(".md"),
		)
		docs, err := l.Load(ctx)
		require.NoError(t, err)
		assert.Len(t, docs, 1)
		assert.Contains(t, docs[0].PageContent, "hello md")
	})

	t.Run("all extensions", func(t *testing.T) {
		l := NewRecursiveDirLoader(
			WithRoot(root),
			WithMaxDepth(3),
			WithAllowExts(".txt", ".csv", ".html", ".md", ".pdf"),
		)
		docs, err := l.Load(ctx)
		require.NoError(t, err)
		require.Len(t, docs, 25)
	})

	t.Run("empty allowExt => all files", func(t *testing.T) {
		l := NewRecursiveDirLoader(
			WithRoot(root),
			WithMaxDepth(3),
		)
		docs, err := l.Load(ctx)
		require.NoError(t, err)
		assert.Len(t, docs, 25)
	})
}
