package postgresql_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/tools/sqldatabase"
)

func Test(t *testing.T) {
	t.Parallel()

	// export LANGCHAINGO_TEST_POSTGRESQL=postgres://db_user:mysecretpassword@localhost:5438/test?sslmode=disable
	mysqlURI := os.Getenv("LANGCHAINGO_TEST_POSTGRESQL")
	if mysqlURI == "" {
		t.Skip("LANGCHAINGO_TEST_POSTGRESQL not set")
	}
	db, err := sqldatabase.NewSQLDatabaseWithDSN("pgx", mysqlURI, nil)
	require.NoError(t, err)

	tbs := db.TableNames()
	require.Greater(t, len(tbs), 0)

	desc, err := db.TableInfo(context.Background(), tbs)
	require.NoError(t, err)

	t.Log(desc)

	for _, tableName := range tbs {
		_, err = db.Query(context.Background(), fmt.Sprintf("SELECT * from %s LIMIT 1", tableName))
		/* exclude no row error,
		since we only need to check if db.Query function can perform query correctly*/
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		require.NoError(t, err)
	}

}
