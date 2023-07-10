package documentloaders

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPDFLoader(t *testing.T) {
	t.Parallel()
	file, err := os.Open("./testdata/lorem-ipsum.pdf")
	assert.NoError(t, err)

	loader := NewPDF(file)

	docs, err := loader.Load(context.Background())
	require.NoError(t, err)

	expectedText := "Te quo illum phaedrum salutatus, has in quis alii vide."
	assert.Contains(t, docs[0].PageContent, expectedText)
	assert.Equal(t, docs[0].Metadata["Pages"], "1")
}
