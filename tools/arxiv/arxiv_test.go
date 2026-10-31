package arxiv

import (
	"context"
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()
	tool, err := New(10, DefaultUserAgent)
	if err != nil {
		t.Fatal(err)
	}
	call, err := tool.Call(context.Background(), "electron")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(call)
}
