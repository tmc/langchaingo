package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/memory/sqlite3"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func prepare(ctx context.Context, db *sql.DB) error {
	// check if db has any records
	var count int
	res := db.QueryRowContext(ctx, "SELECT count(id) FROM langchaingo_messages")
	if err := res.Err(); err != nil {
		return err
	}

	if err := res.Scan(&count); err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	_, err := db.ExecContext(
		ctx,
		"INSERT INTO langchaingo_messages(session, content, type) VALUES (?, ?, ?)",
		"example",
		"Hi there, my name is Murilo!",
		llms.ChatMessageTypeHuman,
	)
	return err
}

func run() error {
	llm, err := openai.New()
	if err != nil {
		return err
	}

	db, err := sql.Open("sqlite3", "history_example.db")
	if err != nil {
		return err
	}

	chatHistory := sqlite3.NewSqliteChatMessageHistory(
		sqlite3.WithSession("example"),
		sqlite3.WithDB(db),
	)
	conversationBuffer := memory.NewConversationBuffer(memory.WithChatHistory(chatHistory))
	llmChain := chains.NewConversation(llm, conversationBuffer)
	ctx := context.Background()

	// prepare the db with some sample data
	if err := prepare(ctx, db); err != nil {
		return err
	}

	out, err := chains.Run(ctx, llmChain, "What's my name? How many times did I ask this?")
	if err != nil {
		return err
	}

	fmt.Println(out)
	return nil
}
