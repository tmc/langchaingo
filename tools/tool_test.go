package tools
// The 'input' parameter in the Call method is unused, so we will rename it to '_'
// to indicate that it is intentionally ignored.
func (st *SomeTool) Call(ctx context.Context, _ string) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	return "test", nil
}

// Additionally, we will modify the TestTool function to run tests in parallel
func TestTool(t *testing.T) {
	t.Parallel() // Call to method parallel

	t.Run("Tool Exists in Kit", func(t *testing.T) {
		t.Parallel() // Call to method parallel in the test run
		kit := Kit{
			&SomeTool{},
		}
		_, err := kit.UseTool(context.Background(), "An awesome tool", "test")
		if err != nil {
			t.Errorf("Error using tool: %v", err)
		}
	})

	t.Run("Tool Does Not Exist in Kit", func(t *testing.T) {
		t.Parallel() // Call to method parallel in the test run
		kit := Kit{
			&SomeTool{},
		}
		_, err := kit.UseTool(context.Background(), "A tool that does not exist", "test")
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
