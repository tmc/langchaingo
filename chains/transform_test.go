package chains

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransform(t *testing.T) {
	t.Parallel()

	c := NewTransform(
		func(_ context.Context, m map[string]any, _ ...ChainCallOption) (map[string]any, error) {
			input, ok := m["input"].(string)
			if !ok {
				return nil, fmt.Errorf("%w: %w", ErrInvalidInputValues, ErrInputValuesWrongType)
			}

			return map[string]any{
				"output": input + "foo",
			}, nil
		},
		[]string{"input"},
		[]string{"output"},
	)

	output, err := Run(context.Background(), c, "baz")
	require.NoError(t, err)
	require.Equal(t, "bazfoo", output)
}
