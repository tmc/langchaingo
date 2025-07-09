//nolint:contextcheck
package cloudsql

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/tmc/langchaingo/util/cloudsqlutil"
)

func preCheckEnvSetting(t *testing.T) string {
	t.Helper()

	pgvectorURL := os.Getenv("PGVECTOR_CONNECTION_STRING")
	ctx := context.Background()
	if pgvectorURL == "" {
		pgVectorContainer, err := tcpostgres.RunContainer(
			ctx,
			testcontainers.WithImage("docker.io/pgvector/pgvector:pg16"),
			tcpostgres.WithDatabase("db_test"),
			tcpostgres.WithUsername("user"),
			tcpostgres.WithPassword("passw0rd!"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(30*time.Second)),
		)
		if err != nil && strings.Contains(err.Error(), "Cannot connect to the Docker daemon") {
			t.Skip("Docker not available")
		}
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, pgVectorContainer.Terminate(ctx))
		})

		str, err := pgVectorContainer.ConnectionString(ctx, "sslmode=disable")
		require.NoError(t, err)

		pgvectorURL = str
	}

	return pgvectorURL
}

func setEngineWithImage(ctx context.Context, t *testing.T) (cloudsqlutil.PostgresEngine, error) {
	t.Helper()
	pgvectorURL := preCheckEnvSetting(t)
	myPool, err := pgxpool.New(ctx, pgvectorURL)
	if err != nil {
		t.Fatal("Could not set Engine: ", err)
	}
	// Call NewPostgresEngine to initialize the database connection
	pgEngine, err := cloudsqlutil.NewPostgresEngine(ctx,
		cloudsqlutil.WithPool(myPool),
	)
	if err != nil {
		t.Fatal("Could not set Engine: ", err)
	}

	return pgEngine, err
}

func TestValidateTableWithContainer(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	engine, err := setEngineWithImage(ctx, t)
	if err != nil {
		t.Fatal(fmt.Printf("setEngineWithImage Error: %s", err))
	}
	cmhTestCases(ctx, t, engine, cancel)
}
