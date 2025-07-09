//nolint:contextcheck
package cloudsql_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/tmc/langchaingo/memory/cloudsql"
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

func setEngineWithImage(t *testing.T) cloudsqlutil.PostgresEngine {
	t.Helper()
	pgvectorURL := preCheckEnvSetting(t)
	ctx := context.Background()
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

	return pgEngine
}

func TestValidateTableWithContainer(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	engine := setEngineWithImage(t)

	t.Cleanup(func() {
		cancel()
		engine.Close()
	})
	tcs := []struct {
		desc      string
		tableName string
		sessionID string
		err       string
	}{
		{
			desc:      "Successful creation of Chat Message History",
			tableName: "items",
			sessionID: "session",
			err:       "",
		},

		{
			desc:      "Creation of Chat Message History with missing session ID",
			tableName: "testchattable",
			sessionID: "",
			err:       "session ID must be provided",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			err := engine.InitChatHistoryTable(ctx, tc.tableName)
			if err != nil {
				t.Fatal("Failed to create chat msg table", err)
			}
			chatMsgHistory, err := cloudsql.NewChatMessageHistory(ctx, engine, tc.tableName, tc.sessionID)
			if tc.err != "" && (err == nil || !strings.Contains(err.Error(), tc.err)) {
				t.Fatalf("unexpected error: got %q, want %q", err, tc.err)
			} else {
				if err != nil {
					errStr := err.Error()
					if errStr != tc.err {
						t.Fatalf("unexpected error: got %q, want %q", errStr, tc.err)
					}
				}
			}
			// if the chat message history was created successfully, continue with the other methods tests
			if err == nil {
				err = chatMsgHistory.AddMessage(ctx, chatMsg{})
				if err != nil {
					t.Fatal(err)
				}
				err = chatMsgHistory.AddAIMessage(ctx, "AI message")
				if err != nil {
					t.Fatal(err)
				}
				err = chatMsgHistory.AddUserMessage(ctx, "user message")
				if err != nil {
					t.Fatal(err)
				}
				err = chatMsgHistory.Clear(ctx)
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
