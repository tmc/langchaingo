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

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/vendasta/langchaingo/tools/sqldatabase"
)

func Test(t *testing.T) {
	testctr.SkipIfDockerNotAvailable(t)

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Parallel()
	ctx := context.Background()

	// export LANGCHAINGO_TEST_POSTGRESQL=postgres://db_user:mysecretpassword@localhost:5438/test?sslmode=disable
	pgURI := os.Getenv("LANGCHAINGO_TEST_POSTGRESQL")
	if pgURI == "" {
		pgContainer, err := postgres.Run(
			ctx,
			"postgres:17", // TODO: lets add a text matrix for this (or do so in this test)
			postgres.WithDatabase("test"),
			postgres.WithUsername("db_user"),
			postgres.WithPassword("p@mysecretpassword"),
			postgres.WithInitScripts(filepath.Join("..", "testdata", "db.sql")),
			postgres.BasicWaitStrategies(),
			testcontainers.WithLogger(log.TestLogger(t)),
		)
		if err != nil && strings.Contains(err.Error(), "Cannot connect to the Docker daemon") {
			t.Skip("Docker not available")
		}
		require.NoError(t, err)
		defer func() {
			require.NoError(t, pgContainer.Terminate(ctx))
		}()

		pgURI, err = pgContainer.ConnectionString(ctx, "sslmode=disable")
		require.NoError(t, err)
	}

	db, err := sqldatabase.NewSQLDatabaseWithDSN("pgx", pgURI, nil)
	require.NoError(t, err)

	tbs := db.TableNames()
	require.NotEmpty(t, tbs)

	desc, err := db.TableInfo(ctx, tbs)
	require.NoError(t, err)

	t.Log(desc)

	for _, tableName := range tbs {
		_, err = db.Query(ctx, fmt.Sprintf("SELECT * from %s LIMIT 1", tableName))
		/* exclude no row error,
		since we only need to check if db.Query function can perform query correctly*/
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		require.NoError(t, err)
	}
}
