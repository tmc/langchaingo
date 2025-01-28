package tools

import (
	"context"
	"testing"
)

type SomeTool struct {
}

func (st *SomeTool) Name() string {
	return "An awesome tool"
}
func (st *SomeTool) Description() string {
	return "This tool is awesome"
}
func (st *SomeTool) Call(ctx context.Context, input string) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	return "test", nil
}
func TestTool(t *testing.T) {
	t.Run("Tool Exists in Kit", func(t *testing.T) {
		kit := Kit{
			&SomeTool{},
		}
		_, err := kit.UseTool(context.Background(), "An awesome tool", "test")
		if err != nil {
			t.Errorf("Error using tool: %v", err)
		}
	})
	t.Run("Tool Does Not Exist in Kit", func(t *testing.T) {
		kit := Kit{
			&SomeTool{},
		}
		_, err := kit.UseTool(context.Background(), "A tool that does not exist", "test")
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}
