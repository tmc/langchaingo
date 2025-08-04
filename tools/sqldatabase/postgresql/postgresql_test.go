package postgresql_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/yincongcyincong/langchaingo/tools/sqldatabase"
)

func Test(t *testing.T) {
	t.Parallel()
	
	// export LANGCHAINGO_TEST_POSTGRESQL=postgres://db_user:mysecretpassword@localhost:5438/test?sslmode=disable
	pgURI := os.Getenv("LANGCHAINGO_TEST_POSTGRESQL")
	if pgURI == "" {
		pgContainer, err := postgres.RunContainer(
			context.Background(),
			testcontainers.WithImage("postgres:16.2"),
			postgres.WithDatabase("test"),
			postgres.WithUsername("db_user"),
			postgres.WithPassword("p@mysecretpassword"),
			postgres.WithInitScripts(filepath.Join("..", "testdata", "db.sql")),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(5*time.Second)),
		)
		if err != nil && strings.Contains(err.Error(), "Cannot connect to the Docker daemon") {
			t.Skip("Docker not available")
		}
		require.NoError(t, err)
		defer func() {
			require.NoError(t, pgContainer.Terminate(context.Background()))
		}()
		
		pgURI, err = pgContainer.ConnectionString(context.Background(), "sslmode=disable")
		require.NoError(t, err)
	}
	
	db, err := sqldatabase.NewSQLDatabaseWithDSN("pgx", pgURI, nil)
	require.NoError(t, err)
	
	tbs := db.TableNames()
	require.NotEmpty(t, tbs)
	
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
