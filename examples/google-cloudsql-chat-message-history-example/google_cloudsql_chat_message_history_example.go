package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/vendasta/langchaingo/llms"
	"github.com/vendasta/langchaingo/memory/cloudsql"
	"github.com/vendasta/langchaingo/util/cloudsqlutil"
)

// getEnvVariables loads the necessary environment variables for the CloudSQL connection
// and the chat message history creation.
func getEnvVariables() (string, string, string, string, string, string, string, string) {
	// Requires environment variable POSTGRES_USERNAME to be set.
	username := os.Getenv("POSTGRES_USERNAME")
	if username == "" {
		log.Fatal("environment variable POSTGRES_USERNAME is empty")
	}
	// Requires environment variable POSTGRES_PASSWORD to be set.
	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		log.Fatal("environment variable POSTGRES_PASSWORD is empty")
	}
	// Requires environment variable POSTGRES_DATABASE to be set.
	database := os.Getenv("POSTGRES_DATABASE")
	if database == "" {
		log.Fatal("environment variable POSTGRES_DATABASE is empty")
	}
	// Requires environment variable PROJECT_ID to be set.
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		log.Fatal("environment variable PROJECT_ID is empty")
	}
	// Requires environment variable POSTGRES_REGION to be set.
	region := os.Getenv("POSTGRES_REGION")
	if region == "" {
		log.Fatal("environment variable POSTGRES_REGION is empty")
	}
	// Requires environment variable POSTGRES_INSTANCE to be set.
	instance := os.Getenv("POSTGRES_INSTANCE")
	if instance == "" {
		log.Fatal("environment variable POSTGRES_INSTANCE is empty")
	}
	// Requires environment variable POSTGRES_TABLE to be set.
	tableName := os.Getenv("POSTGRES_TABLE")
	if tableName == "" {
		log.Fatal("environment variable POSTGRES_TABLE is empty")
	}
	// Requires environment variable POSTGRES_SESSION_ID to be set.
	sessionID := os.Getenv("POSTGRES_SESSION_ID")
	if sessionID == "" {
		log.Fatal("environment variable POSTGRES_SESSION_ID is empty")
	}

	return username, password, database, projectID, region, instance, tableName, sessionID
}

func printMessages(ctx context.Context, cmh cloudsql.ChatMessageHistory) {
	msgs, err := cmh.Messages(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, msg := range msgs {
		fmt.Println("Message:", msg)
	}
}

func main() {
	// Requires that the Environment variables to be set as indicated in the getEnvVariables function.
	username, password, database, projectID, region, instance, tableName, sessionID := getEnvVariables()
	ctx := context.Background()

	pgEngine, err := cloudsqlutil.NewPostgresEngine(ctx,
		cloudsqlutil.WithUser(username),
		cloudsqlutil.WithPassword(password),
		cloudsqlutil.WithDatabase(database),
		cloudsqlutil.WithCloudSQLInstance(projectID, region, instance),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Creates a new table in the Postgres database, which will be used for storing Chat History.
	err = pgEngine.InitChatHistoryTable(ctx, tableName)
	if err != nil {
		log.Fatal(err)
	}

	// Creates a new Chat Message History
	cmh, err := cloudsql.NewChatMessageHistory(ctx, pgEngine, tableName, sessionID)
	if err != nil {
		log.Fatal(err)
	}

	// Creates individual messages and adds them to the chat message history.
	aiMessage := llms.AIChatMessage{Content: "test AI message"}
	humanMessage := llms.HumanChatMessage{Content: "test HUMAN message"}
	// Adds a user message to the chat message history.
	err = cmh.AddUserMessage(ctx, aiMessage.GetContent())
	if err != nil {
		log.Fatal(err)
	}
	// Adds a user message to the chat message history.
	err = cmh.AddUserMessage(ctx, humanMessage.GetContent())
	if err != nil {
		log.Fatal(err)
	}

	printMessages(ctx, cmh)

	// Create multiple messages and store them in the chat message history at the same time.
	multipleMessages := []llms.ChatMessage{
		llms.AIChatMessage{Content: "first AI test message from AddMessages"},
		llms.AIChatMessage{Content: "second AI test message from AddMessages"},
		llms.HumanChatMessage{Content: "first HUMAN test message from AddMessages"},
	}

	// Adds multiple messages to the chat message history.
	err = cmh.AddMessages(ctx, multipleMessages)
	if err != nil {
		log.Fatal(err)
	}

	printMessages(ctx, cmh)

	// Create messages that will overwrite the existing ones
	overWrittingMessages := []llms.ChatMessage{
		llms.AIChatMessage{Content: "overwritten AI test message"},
		llms.HumanChatMessage{Content: "overwritten HUMAN test message"},
	}
	// Overwrites the existing messages with new ones.
	err = cmh.SetMessages(ctx, overWrittingMessages)
	if err != nil {
		log.Fatal(err)
	}

	printMessages(ctx, cmh)

	// Clear all the messages from the current session.
	err = cmh.Clear(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
