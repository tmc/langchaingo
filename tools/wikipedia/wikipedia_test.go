package wikipedia

import (
	"context"
	"testing"
	
	"github.com/stretchr/testify/require"
)

const _userAgent = "langchaingo test (https://github.com/yincongcyincong/langchaingo)"

func TestWikipedia(t *testing.T) {
	t.Parallel()
	
	tool := New(_userAgent)
	_, err := tool.Call(context.Background(), "america")
	require.NoError(t, err)
}
