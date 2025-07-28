package pgvector

import (
	"os"
	"testing"

	"github.com/0xDezzy/langchaingo/internal/testutil/testctr"
)

func TestMain(m *testing.M) {
	code := testctr.EnsureTestEnv()
	if code == 0 {
		code = m.Run()
	}
	os.Exit(code)
}
