package documentloaders

import (
	"context"
	"os"
	"strings"
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

	expected := "Lorem ipsum dolor sit amet, dico fastidii omnesque mea in. Eam ut iusto fastidii, id qui audire abhorreant"
	segments := strings.Split(docs[0].PageContent, ".")
	assert.Contains(t, expected, segments[0])
	assert.Equal(t, "1", docs[0].Metadata["Pages"])
}
