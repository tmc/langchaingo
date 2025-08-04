package alloydb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	
	"github.com/jackc/pgx/v5"
	"github.com/yincongcyincong/langchaingo/llms"
	"github.com/yincongcyincong/langchaingo/schema"
	"github.com/yincongcyincong/langchaingo/util/alloydbutil"
)

type ChatMessageHistory struct {
	engine     alloydbutil.PostgresEngine
	sessionID  string
	tableName  string
	schemaName string
}

var _ schema.ChatMessageHistory = &ChatMessageHistory{}

// NewChatMessageHistory creates a new NewChatMessageHistory with options.
func NewChatMessageHistory(ctx context.Context,
	engine alloydbutil.PostgresEngine,
	tableName,
	sessionID string,
	opts ...ChatMessageHistoryStoresOption,
) (ChatMessageHistory, error) {
	var err error
	// Ensure required fields are set
	if engine.Pool == nil {
		return ChatMessageHistory{}, errors.New("alloyDB engine must be provided")
	}
	if tableName == "" {
		return ChatMessageHistory{}, errors.New("table name must be provided")
	}
	if sessionID == "" {
		return ChatMessageHistory{}, errors.New("session ID must be provided")
	}
	cmh := ChatMessageHistory{
		engine:    engine,
		tableName: tableName,
		sessionID: sessionID,
	}
	cmh = applyChatMessageHistoryOptions(cmh, opts...)
	
	err = cmh.validateTable(ctx)
	if err != nil {
		return ChatMessageHistory{}, fmt.Errorf("error validating table '%s' in schema '%s': %w", tableName, cmh.schemaName, err)
	}
	return cmh, nil
}

// validateTable validates if a table with a specific schema exist and it
// contains the required columns.
func (c *ChatMessageHistory) validateTable(ctx context.Context) error {
	tableExistsQuery := `SELECT EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = $1 AND table_name = $2);`
	
	var exists bool
	err := c.engine.Pool.QueryRow(ctx, tableExistsQuery, c.schemaName, c.tableName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error validating the existence of table '%s' in schema '%s': %w", c.tableName, c.schemaName, err)
	}
	if !exists {
		return fmt.Errorf("table '%s' does not exist in schema '%s'", c.tableName, c.schemaName)
	}
	
	// Required columns with their types
	requiredColumns := map[string]string{
		"id":         "integer",
		"session_id": "text",
		"data":       "jsonb",
		"type":       "text",
	}
	
	columns := make(map[string]string)
	
	// Get the columns from the table
	columnsQuery := `
    	 	SELECT column_name, data_type
    	 	FROM information_schema.columns
   	 		WHERE table_schema = $1 AND table_name = $2;`
	
	rows, err := c.engine.Pool.Query(ctx, columnsQuery, c.schemaName, c.tableName)
	if err != nil {
		return fmt.Errorf("error fetching columns from table '%s' in schema '%s': %w", c.tableName, c.schemaName, err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var columnName, dataType string
		if err := rows.Scan(&columnName, &dataType); err != nil {
			return fmt.Errorf("error scanning column names from table '%s' in schema '%s': %w", c.tableName, c.schemaName, err)
		}
		columns[columnName] = dataType
	}
	
	// Validate column names and types
	for reqColumn, expectedType := range requiredColumns {
		actualType, found := columns[reqColumn]
		if !found {
			return fmt.Errorf("error, column '%s' is missing in table '%s'. Expected columns: %v", reqColumn, c.tableName, requiredColumns)
		}
		if actualType != expectedType {
			return fmt.Errorf("error, column '%s' in table '%s' has type '%s', but expected type '%s'",
				reqColumn, c.tableName, actualType, expectedType)
		}
	}
	return nil
}

// addMessage adds a new message into the ChatMessageHistory for a given
// session.
func (c *ChatMessageHistory) addMessage(ctx context.Context, content string, messageType llms.ChatMessageType) error {
	data, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to serialize content to JSON: %w", err)
	}
	query := fmt.Sprintf(`INSERT INTO %q.%q (session_id, data, type) VALUES ($1, $2, $3)`,
		c.schemaName, c.tableName)
	
	_, err = c.engine.Pool.Exec(ctx, query, c.sessionID, data, messageType)
	if err != nil {
		return fmt.Errorf("failed to add message to database: %w", err)
	}
	return nil
}

// AddMessage adds a message to the ChatMessageHistory.
func (c *ChatMessageHistory) AddMessage(ctx context.Context, message llms.ChatMessage) error {
	return c.addMessage(ctx, message.GetContent(), message.GetType())
}

// AddAIMessage adds an AI-generated message to the ChatMessageHistory.
func (c *ChatMessageHistory) AddAIMessage(ctx context.Context, content string) error {
	return c.addMessage(ctx, content, llms.ChatMessageTypeAI)
}

// AddUserMessage adds a user-generated message to the ChatMessageHistory.
func (c *ChatMessageHistory) AddUserMessage(ctx context.Context, content string) error {
	return c.addMessage(ctx, content, llms.ChatMessageTypeHuman)
}

// Clear removes all messages associated with a session from the
// ChatMessageHistory.
func (c *ChatMessageHistory) Clear(ctx context.Context) error {
	query := fmt.Sprintf(`DELETE FROM %q.%q WHERE session_id = $1`,
		c.schemaName, c.tableName)
	
	_, err := c.engine.Pool.Exec(ctx, query, c.sessionID)
	if err != nil {
		return fmt.Errorf("failed to clear session %s: %w", c.sessionID, err)
	}
	return err
}

// AddMessages adds multiple messages to the ChatMessageHistory for a given
// session.
func (c *ChatMessageHistory) AddMessages(ctx context.Context, messages []llms.ChatMessage) error {
	b := &pgx.Batch{}
	query := fmt.Sprintf(`INSERT INTO %q.%q (session_id, data, type) VALUES ($1, $2, $3)`,
		c.schemaName, c.tableName)
	
	for _, message := range messages {
		data, err := json.Marshal(message.GetContent())
		if err != nil {
			return fmt.Errorf("failed to serialize content to JSON: %w", err)
		}
		b.Queue(query, c.sessionID, data, message.GetType())
	}
	return c.engine.Pool.SendBatch(ctx, b).Close()
}

// Messages retrieves all messages associated with a session from the
// ChatMessageHistory.
func (c *ChatMessageHistory) Messages(ctx context.Context) ([]llms.ChatMessage, error) {
	query := fmt.Sprintf(
		`SELECT id, session_id, data, type FROM %q.%q WHERE session_id = $1 ORDER BY id`,
		c.schemaName, c.tableName,
	)
	
	rows, err := c.engine.Pool.Query(ctx, query, c.sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve messages: %w", err)
	}
	defer rows.Close()
	
	var messages []llms.ChatMessage
	for rows.Next() {
		var id int
		var sessionID, data, messageType string
		
		if err := rows.Scan(&id, &sessionID, &data, &messageType); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		
		// Variable to hold the deserialized content
		var content string
		
		// Unmarshal the JSON data into the content variable
		err := json.Unmarshal([]byte(data), &content)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal data: %w", err)
		}
		switch messageType {
		case string(llms.ChatMessageTypeAI):
			messages = append(messages, llms.AIChatMessage{Content: content})
		case string(llms.ChatMessageTypeHuman):
			messages = append(messages, llms.HumanChatMessage{Content: content})
		case string(llms.ChatMessageTypeSystem):
			messages = append(messages, llms.SystemChatMessage{Content: content})
		default:
			return nil, fmt.Errorf("unsupported message type: %s", messageType)
		}
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}
	
	return messages, nil
}

// SetMessages clears the current messages from the ChatMessageHistory for a
// given session and then adds new messages to it.
func (c *ChatMessageHistory) SetMessages(ctx context.Context, messages []llms.ChatMessage) error {
	err := c.Clear(ctx)
	if err != nil {
		return err
	}
	
	b := &pgx.Batch{}
	query := fmt.Sprintf(`INSERT INTO %q.%q (session_id, data, type) VALUES ($1, $2, $3)`,
		c.schemaName, c.tableName)
	
	for _, message := range messages {
		data, err := json.Marshal(message.GetContent())
		if err != nil {
			return fmt.Errorf("failed to serialize content to JSON: %w", err)
		}
		b.Queue(query, c.sessionID, data, message.GetType())
	}
	return c.engine.Pool.SendBatch(ctx, b).Close()
}
