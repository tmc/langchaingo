package mysql_test

import (
	"context"
	"os"
	"testing"

	"github.com/tmc/langchaingo/tools/sqldatabase"
	_ "github.com/tmc/langchaingo/tools/sqldatabase/mysql"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

var (
	ctx = context.Background()
)

func Test(t *testing.T) {
	// Same as export LANGCHAINGO_TEST_MYSQL=user:p@ssw0rd@tcp(localhost:3306)/test
	godotenv.Overload()

	mysqlURI := os.Getenv("LANGCHAINGO_TEST_MYSQL")
	if mysqlURI == "" {
		t.Skip("LANGCHAINGO_TEST_MYSQL not set")
	}
	db, err := sqldatabase.NewSQLDatabaseWithDSN("mysql", mysqlURI, nil)
	require.NoError(t, err)

	tbs := db.TableNames()
	require.Greater(t, len(tbs), 0)

	desc, err := db.TableInfo(ctx, tbs)
	require.NoError(t, err)

	t.Log(desc)
}
