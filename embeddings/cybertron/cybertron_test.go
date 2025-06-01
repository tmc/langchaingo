package cybertron

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCybertronEmbeddings(t *testing.T) {
	t.Parallel()

	_, err := os.Stat(_defaultModelsDir)
	if os.IsNotExist(err) && os.Getenv("CYBERTRON_DO_DOWNLOAD") == "" {
		// Cybertron downloads the embedding model and caches it in ModelsDir. Doing this as
		// part of the tests would be costly in terms of time and bandwidth and likely make
		// the test flaky.
		t.Skipf("ModelsDir %q doesn't exist", _defaultModelsDir)
	}

	emb, err := NewCybertron(
		WithModelsDir(_defaultModelsDir),
		WithModel(_defaultModel),
		WithPoolingStrategy(_defaultPoolingStrategy),
	)
	require.NoError(t, err)

	res, err := emb.CreateEmbedding(t.Context(), []string{
		"Hello world", "The world is ending", "good bye",
	})
	require.NoError(t, err)
	require.Len(t, res, 3)
}
