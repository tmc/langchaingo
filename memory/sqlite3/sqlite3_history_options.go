package sqlite3

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3" // sqlite3 driver.
)

// DefaultLimit sets a default limit for select queries.
const DefaultLimit = 1000

// DefaultTableName sets a default table name.
const DefaultTableName = "langchaingo_messages"

// DefaultSchema sets a default schema to be run after connecting.
const DefaultSchema = `CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY,
		name TEXT,
		session TEXT NOT NULL,
		content TEXT NOT NULL,
		type TEXT NOT NULL,
		created TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_langchaingo_id ON %s (id);
CREATE INDEX IF NOT EXISTS idx_langchaingo_session ON %s (session);`

// SqliteChatMessageHistoryOption is a function for creating new
// chat message history with other than the default values.
type SqliteChatMessageHistoryOption func(m *SqliteChatMessageHistory)

// WithDB is an option for NewSqliteChatMessageHistory for adding
// a database connection.
func WithDB(db *sql.DB) SqliteChatMessageHistoryOption {
	return func(m *SqliteChatMessageHistory) {
		m.DB = db
	}
}

// WithContext is an option for NewSqliteChatMessageHistory
// to use a context internally when running Schema.
func WithContext(ctx context.Context) SqliteChatMessageHistoryOption {
	return func(m *SqliteChatMessageHistory) {
		m.Ctx = ctx
	}
}

// WithLimit is an option for NewSqliteChatMessageHistory for
// defining a limit number for select queries.
func WithLimit(limit int) SqliteChatMessageHistoryOption {
	return func(m *SqliteChatMessageHistory) {
		m.Limit = limit
	}
}

// WithSchema is an option for NewSqliteChatMessageHistory for
// running a schema when connected. Useful for migrations for example.
func WithSchema(schema []byte) SqliteChatMessageHistoryOption {
	return func(m *SqliteChatMessageHistory) {
		m.Schema = schema
	}
}

// WithOverwrite is an option for NewSqliteChatMessageHistory for
// allowing dangerous operations like SetMessages or Clear.
func WithOverwrite() SqliteChatMessageHistoryOption {
	return func(m *SqliteChatMessageHistory) {
		m.Overwrite = true
	}
}

// WithDBAddress is an option for NewSqliteChatMessageHistory for
// specifying an address or file path for when connecting the db.
func WithDBAddress(addr string) SqliteChatMessageHistoryOption {
	return func(m *SqliteChatMessageHistory) {
		m.DBAddress = addr
	}
}

// WithTableName is an option for NewSqliteChatMessageHistory for
// running a schema when connected. Useful for migrations for example.
func WithTableName(name string) SqliteChatMessageHistoryOption {
	return func(m *SqliteChatMessageHistory) {
		m.TableName = name
	}
}

// WithSession is an option for NewSqliteChatMessageHistory for
// setting a session name or id for the history.
func WithSession(session string) SqliteChatMessageHistoryOption {
	return func(m *SqliteChatMessageHistory) {
		m.Session = session
	}
}

func applyChatOptions(options ...SqliteChatMessageHistoryOption) *SqliteChatMessageHistory {
	h := &SqliteChatMessageHistory{}

	for _, option := range options {
		option(h)
	}

	if h.TableName == "" {
		h.TableName = DefaultTableName
	}

	if h.Limit < 1 {
		h.Limit = DefaultLimit
	}

	if h.Schema == nil {
		h.Schema = []byte(fmt.Sprintf(DefaultSchema, h.TableName, h.TableName, h.TableName))
	}

	if h.Ctx == nil {
		h.Ctx = context.Background()
	}

	if h.DBAddress == "" {
		h.DBAddress = ":memory:"
	}

	if h.Session == "" {
		h.Session = "default"
	}

	if h.DB == nil {
		db, err := sql.Open("sqlite3", h.DBAddress)
		if err != nil {
			panic(err)
		}
		h.DB = db
	}

	if _, err := h.DB.ExecContext(h.Ctx, string(h.Schema)); err != nil {
		panic(err)
	}

	return h
}
