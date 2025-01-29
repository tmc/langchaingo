package tools

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type someTool struct{}

func (st *someTool) Name() string {
	return "An awesome tool"
}

func (st *someTool) Description() string {
	return "This tool is awesome"
}

func (st *someTool) Call(ctx context.Context, _ string) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	return "test", nil
}

func TestToolWithTestify(t *testing.T) {
	t.Parallel()
	kit := Kit{
		&someTool{},
	}

	// Test when the tool exists
	t.Run("Tool Exists in Kit", func(t *testing.T) {
		result, err := kit.UseTool(context.Background(), "An awesome tool", "test")
		assert.NoError(t, err)
		assert.Equal(t, "test", result)
	})

	// Test when the tool does not exist
	t.Run("Tool Does Not Exist in Kit", func(t *testing.T) {
		_, err := kit.UseTool(context.Background(), "A tool that does not exist", "test")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidTool, err.Error())
	})
}
