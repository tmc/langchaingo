package tests

import (
	"context"
	"runtime"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

type streamRecv struct {
	Content string
}

func (s *streamRecv) streamFunc(_ context.Context, chunk []byte) error {
	s.Content += string(chunk)
	return nil
}

func newStreamRecv() *streamRecv {
	return &streamRecv{}
}

func printStack(t *testing.T) { // nolint: unused
	t.Helper()
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			t.Logf("Stack trace:\n%s", buf[:n])
			return
		}
		buf = make([]byte, 2*len(buf))
	}
}

func logTools(t *testing.T, tools []llms.ToolCall) {
	t.Helper()
	for _, tool := range tools {
		t.Logf(
			"Tool Name: %s , Tool ID: %s, Tool Args: %s",
			tool.FunctionCall.Name,
			tool.ID,
			tool.FunctionCall.Arguments,
		)
	}
}
