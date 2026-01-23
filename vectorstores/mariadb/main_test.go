package mariadb_test

import (
	"os"
	"testing"

	"github.com/vxcontrol/langchaingo/internal/testutil/testctr"
)

func TestMain(m *testing.M) {
	code := testctr.EnsureTestEnv()
	if code == 0 {
		code = m.Run()
	}
	os.Exit(code)
}
