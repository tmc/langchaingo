package cloudsql_test

import (
	"os"
	"testing"

	"github.com/tmc/langchaingo/internal/testutil/testctr"
)

func TestMain(m *testing.M) {
	testctr.EnsureTestEnv()
	os.Exit(m.Run())
}
