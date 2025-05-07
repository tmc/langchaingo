package callbacks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCallbackHandler(t *testing.T) {
	t.Parallel()
	// Test case 1: Context with handler
	handler := &SimpleHandler{}
	ctx := SetHandlerInContext(context.Background(), handler)

	got := GetHandlerFromContext(ctx)
	require.NotNil(t, got)
	if got != handler {
		t.Errorf("CallbackHandler() = %v, want %v", got, handler)
	}

	// Test case 2: Context without handler
	emptyCtx := context.Background()
	got = GetHandlerFromContext(emptyCtx)
	if got != nil {
		t.Errorf("CallbackHandler() = %v, want nil", got)
	}
}
