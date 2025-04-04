package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/averikitsch/langchaingo/memory/alloydb"
	"github.com/averikitsch/langchaingo/util/alloydbutil"
	"github.com/tmc/langchaingo/llms"
)

// getEnvVariables loads the necessary environment variables for the AlloyDB connection
// and the chat message history creation.
func getEnvVariables() (string, string, string, string, string, string, string, string, string) {
	// Requires environment variable ALLOYDB_USERNAME to be set.
	username := os.Getenv("ALLOYDB_USERNAME")
	if username == "" {
		log.Fatal("environment variable ALLOYDB_USERNAME is empty")
	}
	// Requires environment variable ALLOYDB_PASSWORD to be set.
	password := os.Getenv("ALLOYDB_PASSWORD")
	if password == "" {
		log.Fatal("environment variable ALLOYDB_PASSWORD is empty")
	}
	// Requires environment variable ALLOYDB_DATABASE to be set.
	database := os.Getenv("ALLOYDB_DATABASE")
	if database == "" {
		log.Fatal("environment variable ALLOYDB_DATABASE is empty")
	}
	// Requires environment variable PROJECT_ID to be set.
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		log.Fatal("environment variable PROJECT_ID is empty")
	}
	// Requires environment variable ALLOYDB_REGION to be set.
	region := os.Getenv("ALLOYDB_REGION")
	if region == "" {
		log.Fatal("environment variable ALLOYDB_REGION is empty")
	}
	// Requires environment variable ALLOYDB_INSTANCE to be set.
	instance := os.Getenv("ALLOYDB_INSTANCE")
	if instance == "" {
		log.Fatal("environment variable ALLOYDB_INSTANCE is empty")
	}
	// Requires environment variable ALLOYDB_CLUSTER to be set.
	cluster := os.Getenv("ALLOYDB_CLUSTER")
	if cluster == "" {
		log.Fatal("environment variable ALLOYDB_CLUSTER is empty")
	}
	// Requires environment variable ALLOYDB_TABLE to be set.
	tableName := os.Getenv("ALLOYDB_TABLE")
	if tableName == "" {
		log.Fatal("environment variable ALLOYDB_TABLE is empty")
	}
	// Requires environment variable ALLOYDB_SESSION_ID to be set.
	sessionID := os.Getenv("ALLOYDB_SESSION_ID")
	if sessionID == "" {
		log.Fatal("environment variable ALLOYDB_SESSION_ID is empty")
	}

	return username, password, database, projectID, region, instance, cluster, tableName, sessionID
}

func printMessages(ctx context.Context, cmh alloydb.ChatMessageHistory) {
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
	username, password, database, projectID, region, instance, cluster, tableName, sessionID := getEnvVariables()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pgEngine, err := alloydbutil.NewPostgresEngine(ctx,
		alloydbutil.WithUser(username),
		alloydbutil.WithPassword(password),
		alloydbutil.WithDatabase(database),
		alloydbutil.WithAlloyDBInstance(projectID, region, cluster, instance),
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
	cmh, err := alloydb.NewChatMessageHistory(ctx, pgEngine, tableName, sessionID)
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
