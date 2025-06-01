package mysql_test

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/tmc/langchaingo/tools/sqldatabase"
	_ "github.com/tmc/langchaingo/tools/sqldatabase/mysql"
)

func Test(t *testing.T) {
	t.Parallel()

	// export LANGCHAINGO_TEST_MYSQL=user:p@ssw0rd@tcp(localhost:3306)/test
	mysqlURI := os.Getenv("LANGCHAINGO_TEST_MYSQL")
	if mysqlURI == "" {
		mysqlContainer, err := mysql.Run(
			t.Context(),
			"mysql:8.3.0",
			mysql.WithDatabase("test"),
			mysql.WithUsername("user"),
			mysql.WithPassword("p@ssw0rd"),
			mysql.WithScripts(filepath.Join("..", "testdata", "db.sql")),
		)
		// if error is no docker socket available, skip the test
		if err != nil && strings.Contains(err.Error(), "Cannot connect to the Docker daemon") {
			t.Skip("Docker not available")
		}
		require.NoError(t, err)
		defer func() {
			if err := mysqlContainer.Terminate(t.Context()); err != nil {
				t.Logf("Failed to terminate mysql container: %v", err)
			}
		}()

		mysqlURI, err = mysqlContainer.ConnectionString(t.Context())
		require.NoError(t, err)
	}

	db, err := sqldatabase.NewSQLDatabaseWithDSN("mysql", mysqlURI, nil)
	require.NoError(t, err)

	tbs := db.TableNames()
	require.NotEmpty(t, tbs)

	desc, err := db.TableInfo(t.Context(), tbs)
	require.NoError(t, err)

	t.Log(desc)

	for _, tableName := range tbs {
		_, err = db.Query(t.Context(), fmt.Sprintf("SELECT * from %s LIMIT 1", tableName))
		/* exclude no row error,
		since we only need to check if db.Query function can perform query correctly*/
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		require.NoError(t, err)
	}
}
