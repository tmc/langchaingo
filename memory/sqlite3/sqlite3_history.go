// Package sqlite3 adds support for
// chat message history using sqlite3.
package sqlite3

import (
	"bytes"
	"context"
	"database/sql"
	"strings"

	_ "github.com/mattn/go-sqlite3" // sqlite3 driver.
	"github.com/vendasta/langchaingo/llms"
	"github.com/vendasta/langchaingo/schema"
)

// SqliteChatMessageHistory is a struct that stores chat messages.
type SqliteChatMessageHistory struct {
	// DB is the database connection.
	DB *sql.DB
	// Ctx is a context that can be used for the schema exec.
	//nolint:containedctx // This is used only when execing schema.
	Ctx context.Context
	// DBAddress is the address or file path for connecting the db.
	DBAddress string
	// TableName is the name of the messages table.
	TableName string
	// Limit is the max number of records per select.
	Limit int
	// Session defines a session name or id for a conversation.
	Session string
	// Schema defines a initial schema to be run.
	Schema []byte
	// Overwrite is a safety flag used for SetMessages and Clear functions.
	Overwrite bool
}

// Statically assert that SqliteChatMessageHistory implement the chat message history interface.
var _ schema.ChatMessageHistory = &SqliteChatMessageHistory{}

// NewSqliteChatMessageHistory creates a new SqliteChatMessageHistory using chat message options.
func NewSqliteChatMessageHistory(options ...SqliteChatMessageHistoryOption) *SqliteChatMessageHistory {
	return applyChatOptions(options...)
}

// Messages returns all messages stored.
func (h *SqliteChatMessageHistory) Messages(ctx context.Context) ([]llms.ChatMessage, error) {
	querytpl := []string{
		"SELECT content,type,created FROM ",
		" WHERE session = ? ORDER BY created ASC LIMIT ?;",
	}
	query := strings.Join(querytpl, h.TableName)
	res, err := h.DB.QueryContext(ctx, query, h.Session, h.Limit)
	if err != nil {
		return nil, err
	}

	defer res.Close()

	var msgs []llms.ChatMessage
	for res.Next() {
		var content, msgtype string
		var created interface{}

		if err = res.Scan(&content, &msgtype, &created); err != nil {
			return nil, err
		}

		switch msgtype {
		case string(llms.ChatMessageTypeAI):
			msgs = append(msgs, llms.AIChatMessage{Content: content})
		case string(llms.ChatMessageTypeHuman):
			msgs = append(msgs, llms.HumanChatMessage{Content: content})
		case string(llms.ChatMessageTypeSystem):
			msgs = append(msgs, llms.SystemChatMessage{Content: content})
		default:
		}
	}

	if err := res.Err(); err != nil {
		return nil, err
	}

	return msgs, nil
}

func (h *SqliteChatMessageHistory) addMessage(ctx context.Context, text string, role llms.ChatMessageType) error {
	querytpl := []string{
		"INSERT INTO ",
		" (session, content, type) VALUES (?, ?, ?);",
	}
	query := strings.Join(querytpl, h.TableName)
	_, err := h.DB.ExecContext(ctx, query, h.Session, text, role)
	return err
}

// AddMessage adds a message to the chat message history.
func (h *SqliteChatMessageHistory) AddMessage(ctx context.Context, message llms.ChatMessage) error {
	return h.addMessage(ctx, message.GetContent(), message.GetType())
}

// AddAIMessage adds an AIMessage to the chat message history.
func (h *SqliteChatMessageHistory) AddAIMessage(ctx context.Context, text string) error {
	return h.addMessage(ctx, text, llms.ChatMessageTypeAI)
}

// AddUserMessage adds a user to the chat message history.
func (h *SqliteChatMessageHistory) AddUserMessage(ctx context.Context, text string) error {
	return h.addMessage(ctx, text, llms.ChatMessageTypeHuman)
}

// Clear resets messages.
func (h *SqliteChatMessageHistory) Clear(ctx context.Context) error {
	if !h.Overwrite {
		return nil
	}

	querytpl := []string{
		"DELETE FROM ",
		" WHERE session = ?;",
	}
	query := strings.Join(querytpl, h.TableName)
	_, err := h.DB.ExecContext(ctx, query, h.Session)
	return err
}

// SetMessages resets chat history and bulk insert new messages into it.
func (h *SqliteChatMessageHistory) SetMessages(ctx context.Context, messages []llms.ChatMessage) error {
	if !h.Overwrite {
		return nil
	}

	/*
	 BEGIN TRANSACTION;
	 DELETE FROM table WHERE session = ?;
	 INSERT INTO table (session, content, type)
	 VALUES (?, ?, ?), ...;
	 COMMIT;`
	*/
	buf := bytes.NewBufferString("BEGIN TRANSACTION;")
	buf.WriteString(" DELETE FROM ")
	buf.WriteString(h.TableName)
	buf.WriteString(" WHERE session = ?;")
	buf.WriteString(" INSERT INTO ")
	buf.WriteString(h.TableName)
	buf.WriteString(" (session, content, type) VALUES ")

	inputs := make([]string, len(messages))
	values := []interface{}{h.Session}

	for i, msg := range messages {
		inputs[i] = "(?, ?, ?)"
		values = append(values, h.Session, msg.GetContent(), string(msg.GetType()))
	}

	buf.WriteString(strings.Join(inputs, ", "))
	buf.WriteString("; COMMIT;")

	_, err := h.DB.ExecContext(ctx, buf.String(), values...)
	return err
}
