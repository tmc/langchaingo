package milvus

import (
	"os"
	"testing"

	"github.com/vendasta/langchaingo/internal/testutil/testctr"
)

func TestMain(m *testing.M) {
	testctr.EnsureTestEnv()
	os.Exit(m.Run())
}
