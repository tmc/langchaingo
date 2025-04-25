package alloydb

const (
	defaultSchemaName = "public"
)

// ChatMessageHistoryStoresOption is a function for creating chat message
// history with other than the default values.
type ChatMessageHistoryStoresOption func(c *ChatMessageHistory)

// WithSchemaName sets the schemaName field for the ChatMessageHistory.
func WithSchemaName(schemaName string) ChatMessageHistoryStoresOption {
	return func(c *ChatMessageHistory) {
		c.schemaName = schemaName
	}
}

// applyChatMessageHistoryOptions applies the given options to the
// ChatMessageHistory.
func applyChatMessageHistoryOptions(cmh ChatMessageHistory, opts ...ChatMessageHistoryStoresOption) ChatMessageHistory {
	cmh.schemaName = defaultSchemaName

	// Check for optional values.
	for _, opt := range opts {
		opt(&cmh)
	}
	return cmh
}
