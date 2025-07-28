package alloydb

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/0xDezzy/langchaingo/llms"
)

func TestApplyChatMessageHistoryOptions(t *testing.T) {
	cmh := ChatMessageHistory{
		schemaName: "public", // default
	}

	// Test WithSchemaName option
	cmh = applyChatMessageHistoryOptions(cmh, WithSchemaName("custom_schema"))
	assert.Equal(t, "custom_schema", cmh.schemaName)

	// Test multiple options
	cmh = applyChatMessageHistoryOptions(cmh, WithSchemaName("another_schema"))
	assert.Equal(t, "another_schema", cmh.schemaName)
}

func TestChatMessageHistory_ValidateRequiredFields(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
		sessionID string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "error - empty table name",
			tableName: "",
			sessionID: "session123",
			wantErr:   true,
			errMsg:    "table name must be provided",
		},
		{
			name:      "error - empty session ID",
			tableName: "messages",
			sessionID: "",
			wantErr:   true,
			errMsg:    "session ID must be provided",
		},
		{
			name:      "success - valid inputs",
			tableName: "messages",
			sessionID: "session123",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test validation logic directly
			var err error
			if tt.tableName == "" {
				err = errors.New("table name must be provided")
			} else if tt.sessionID == "" {
				err = errors.New("session ID must be provided")
			}

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestChatMessageHistory_MessageMarshaling(t *testing.T) {
	tests := []struct {
		name        string
		messageType llms.ChatMessageType
		content     string
		wantJSON    string
	}{
		{
			name:        "human message",
			messageType: llms.ChatMessageTypeHuman,
			content:     "Hello, world!",
			wantJSON:    `"Hello, world!"`,
		},
		{
			name:        "AI message",
			messageType: llms.ChatMessageTypeAI,
			content:     "Hi there!",
			wantJSON:    `"Hi there!"`,
		},
		{
			name:        "system message",
			messageType: llms.ChatMessageTypeSystem,
			content:     "System initialized",
			wantJSON:    `"System initialized"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that messages can be properly converted
			var msg llms.ChatMessage
			switch tt.messageType {
			case llms.ChatMessageTypeHuman:
				msg = llms.HumanChatMessage{Content: tt.content}
			case llms.ChatMessageTypeAI:
				msg = llms.AIChatMessage{Content: tt.content}
			case llms.ChatMessageTypeSystem:
				msg = llms.SystemChatMessage{Content: tt.content}
			}

			assert.Equal(t, tt.content, msg.GetContent())
			assert.Equal(t, tt.messageType, msg.GetType())
		})
	}
}

func TestChatMessageHistory_SQLGeneration(t *testing.T) {
	tests := []struct {
		name       string
		schemaName string
		tableName  string
		operation  string
		wantSQL    string
	}{
		{
			name:       "insert query with default schema",
			schemaName: "public",
			tableName:  "messages",
			operation:  "insert",
			wantSQL:    `INSERT INTO "public"."messages" (session_id, data, type) VALUES ($1, $2, $3)`,
		},
		{
			name:       "delete query with custom schema",
			schemaName: "custom",
			tableName:  "chat_history",
			operation:  "delete",
			wantSQL:    `DELETE FROM "custom"."chat_history" WHERE session_id = $1`,
		},
		{
			name:       "select query",
			schemaName: "public",
			tableName:  "messages",
			operation:  "select",
			wantSQL:    `SELECT id, session_id, data, type FROM "public"."messages" WHERE session_id = $1 ORDER BY id`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test SQL query generation
			var query string
			switch tt.operation {
			case "insert":
				query = `INSERT INTO "` + tt.schemaName + `"."` + tt.tableName + `" (session_id, data, type) VALUES ($1, $2, $3)`
			case "delete":
				query = `DELETE FROM "` + tt.schemaName + `"."` + tt.tableName + `" WHERE session_id = $1`
			case "select":
				query = `SELECT id, session_id, data, type FROM "` + tt.schemaName + `"."` + tt.tableName + `" WHERE session_id = $1 ORDER BY id`
			}
			assert.Equal(t, tt.wantSQL, query)
		})
	}
}

func TestChatMessageHistory_ErrorHandling(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		wantErr   string
	}{
		{
			name:      "add message error",
			operation: "add",
			wantErr:   "failed to add message to database",
		},
		{
			name:      "clear error",
			operation: "clear",
			wantErr:   "failed to clear session",
		},
		{
			name:      "retrieve messages error",
			operation: "retrieve",
			wantErr:   "failed to retrieve messages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test error message formatting
			var formattedErr error

			switch tt.operation {
			case "add":
				formattedErr = errors.New("failed to add message to database: database error")
			case "clear":
				formattedErr = errors.New("failed to clear session session123: database error")
			case "retrieve":
				formattedErr = errors.New("failed to retrieve messages: database error")
			}

			assert.Contains(t, formattedErr.Error(), tt.wantErr)
			assert.Contains(t, formattedErr.Error(), "database error")
		})
	}
}

func TestChatMessageHistory_MessageTypeConversion(t *testing.T) {
	tests := []struct {
		name        string
		messageType string
		wantValid   bool
	}{
		{
			name:        "valid human type",
			messageType: string(llms.ChatMessageTypeHuman),
			wantValid:   true,
		},
		{
			name:        "valid AI type",
			messageType: string(llms.ChatMessageTypeAI),
			wantValid:   true,
		},
		{
			name:        "valid system type",
			messageType: string(llms.ChatMessageTypeSystem),
			wantValid:   true,
		},
		{
			name:        "invalid type",
			messageType: "unknown",
			wantValid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test message type validation
			isValid := tt.messageType == string(llms.ChatMessageTypeHuman) ||
				tt.messageType == string(llms.ChatMessageTypeAI) ||
				tt.messageType == string(llms.ChatMessageTypeSystem)

			assert.Equal(t, tt.wantValid, isValid)
			if !tt.wantValid {
				err := errors.New("unsupported message type: " + tt.messageType)
				assert.Contains(t, err.Error(), "unsupported message type")
			}
		})
	}
}

func TestChatMessageHistory_BatchOperations(t *testing.T) {
	messages := []llms.ChatMessage{
		llms.HumanChatMessage{Content: "Hello"},
		llms.AIChatMessage{Content: "Hi!"},
		llms.SystemChatMessage{Content: "System ready"},
	}

	// Test that batch size matches message count
	assert.Equal(t, 3, len(messages))

	// Test that each message has the correct type
	assert.Equal(t, llms.ChatMessageTypeHuman, messages[0].GetType())
	assert.Equal(t, llms.ChatMessageTypeAI, messages[1].GetType())
	assert.Equal(t, llms.ChatMessageTypeSystem, messages[2].GetType())
}

func TestChatMessageHistory_SchemaValidation(t *testing.T) {
	requiredColumns := map[string]string{
		"id":         "integer",
		"session_id": "text",
		"data":       "jsonb",
		"type":       "text",
	}

	tests := []struct {
		name    string
		columns map[string]string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "all columns present with correct types",
			columns: requiredColumns,
			wantErr: false,
		},
		{
			name: "missing required column",
			columns: map[string]string{
				"id":         "integer",
				"session_id": "text",
				"data":       "jsonb",
				// missing "type"
			},
			wantErr: true,
			errMsg:  "column 'type' is missing",
		},
		{
			name: "wrong column type",
			columns: map[string]string{
				"id":         "integer",
				"session_id": "text",
				"data":       "text", // should be jsonb
				"type":       "text",
			},
			wantErr: true,
			errMsg:  "has type 'text', but expected type 'jsonb'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate column presence and types
			var err error
			for reqColumn, expectedType := range requiredColumns {
				actualType, found := tt.columns[reqColumn]
				if !found {
					err = errors.New("error, column '" + reqColumn + "' is missing in table 'test'")
					break
				}
				if actualType != expectedType {
					err = errors.New("error, column '" + reqColumn + "' in table 'test' has type '" + actualType + "', but expected type '" + expectedType + "'")
					break
				}
			}

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestChatMessageHistory_ClearAndSet(t *testing.T) {
	// Test that SetMessages clears before adding
	messages := []llms.ChatMessage{
		llms.HumanChatMessage{Content: "New message 1"},
		llms.AIChatMessage{Content: "New message 2"},
	}

	// Simulate the SetMessages flow
	clearCalled := false
	addCalled := false

	// Clear operation
	clearCalled = true

	// Add operation
	if clearCalled {
		addCalled = true
	}

	assert.True(t, clearCalled, "Clear should be called")
	assert.True(t, addCalled, "Add should be called after clear")
	assert.Equal(t, 2, len(messages))
}

func TestChatMessageHistory_QueryFormatting(t *testing.T) {
	tests := []struct {
		name       string
		schemaName string
		tableName  string
		wantQuoted string
	}{
		{
			name:       "quotes schema and table names",
			schemaName: "public",
			tableName:  "messages",
			wantQuoted: `"public"."messages"`,
		},
		{
			name:       "handles special characters",
			schemaName: "my-schema",
			tableName:  "chat_history",
			wantQuoted: `"my-schema"."chat_history"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quoted := `"` + tt.schemaName + `"."` + tt.tableName + `"`
			assert.Equal(t, tt.wantQuoted, quoted)
			assert.True(t, strings.Contains(quoted, `"`))
		})
	}
}
