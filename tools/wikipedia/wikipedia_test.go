package wikipedia

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

const _userAgent = "langchaingo test (https://github.com/tmc/langchaingo)"

func TestWikipedia(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	tool := New(_userAgent)
	_, err := tool.Call(ctx, "america")
	require.NoError(t, err)
}
