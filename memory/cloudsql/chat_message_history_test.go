package cloudsql_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory/cloudsql"
	"github.com/tmc/langchaingo/util/cloudsqlutil"
)

type chatMsg struct{}

func (chatMsg) GetType() llms.ChatMessageType {
	return llms.ChatMessageTypeHuman
}

func (chatMsg) GetContent() string {
	return "test content"
}

func getEnvVariables(t *testing.T) (string, string, string, string, string, string) {
	t.Helper()

	username := os.Getenv("CLOUDSQL_USERNAME")
	if username == "" {
		t.Skip("CLOUDSQL_USERNAME environment variable not set")
	}
	password := os.Getenv("CLOUDSQL_PASSWORD")
	if password == "" {
		t.Skip("CLOUDSQL_PASSWORD environment variable not set")
	}
	database := os.Getenv("CLOUDSQL_DATABASE")
	if database == "" {
		t.Skip("CLOUDSQL_DATABASE environment variable not set")
	}
	projectID := os.Getenv("CLOUDSQL_PROJECT_ID")
	if projectID == "" {
		t.Skip("CLOUDSQL_PROJECT_ID environment variable not set")
	}
	region := os.Getenv("CLOUDSQL_REGION")
	if region == "" {
		t.Skip("CLOUDSQL_REGION environment variable not set")
	}
	instance := os.Getenv("CLOUDSQL_INSTANCE")
	if instance == "" {
		t.Skip("CLOUDSQL_INSTANCE environment variable not set")
	}

	return username, password, database, projectID, region, instance
}

func setEngine(ctx context.Context, t *testing.T) (cloudsqlutil.PostgresEngine, error) {
	t.Helper()
	username, password, database, projectID, region, instance := getEnvVariables(t)

	pgEngine, err := cloudsqlutil.NewPostgresEngine(ctx,
		cloudsqlutil.WithUser(username),
		cloudsqlutil.WithPassword(password),
		cloudsqlutil.WithDatabase(database),
		cloudsqlutil.WithCloudSQLInstance(projectID, region, instance),
	)

	return pgEngine, err
}

func TestValidateTable(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)
	engine, err := setEngine(ctx, t)
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
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
			desc:      "Creation of Chat Message History with missing table",
			tableName: "",
			sessionID: "session",
			err:       "table name must be provided",
		},
		{
			desc:      "Creation of Chat Message History with missing session ID",
			tableName: "items",
			sessionID: "",
			err:       "session ID must be provided",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			chatMsgHistory, err := cloudsql.NewChatMessageHistory(ctx, engine, tc.tableName, tc.sessionID)
			if tc.err != "" && (err == nil || !strings.Contains(err.Error(), tc.err)) {
				t.Fatalf("unexpected error: got %q, want %q", err, tc.err)
			} else {
				errStr := err.Error()
				if errStr != tc.err {
					t.Fatalf("unexpected error: got %q, want %q", errStr, tc.err)
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
