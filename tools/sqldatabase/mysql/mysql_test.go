package mysql_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/tools/sqldatabase"
	_ "github.com/tmc/langchaingo/tools/sqldatabase/mysql"
)

func Test(t *testing.T) {
	// export LANGCHAINGO_TEST_MYSQL=user:p@ssw0rd@tcp(localhost:3306)/test
	mysqlURI := os.Getenv("LANGCHAINGO_TEST_MYSQL")
	if mysqlURI == "" {
		t.Skip("LANGCHAINGO_TEST_MYSQL not set")
	}
	db, err := sqldatabase.NewSQLDatabaseWithDSN("mysql", mysqlURI, nil)
	require.NoError(t, err)

	tbs := db.TableNames()
	require.Greater(t, len(tbs), 0)

	desc, err := db.TableInfo(context.Background(), tbs)
	require.NoError(t, err)

	t.Log(desc)
}
