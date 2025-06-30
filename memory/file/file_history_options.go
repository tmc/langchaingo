package file

// FileChatMessageHistoryOption is a function that configures a FileChatMessageHistory.
type FileChatMessageHistoryOption func(*FileChatMessageHistory)

// WithFilePath sets the file path for the chat history.
func WithFilePath(filePath string) FileChatMessageHistoryOption {
	return func(h *FileChatMessageHistory) {
		h.FilePath = filePath
	}
}

// WithCreateDirIfNotExist sets whether to create the directory if it doesn't exist.
func WithCreateDirIfNotExist(create bool) FileChatMessageHistoryOption {
	return func(h *FileChatMessageHistory) {
		h.CreateDirIfNotExist = create
	}
}

// applyChatOptions applies the given options to the FileChatMessageHistory.
func applyChatOptions(opts ...FileChatMessageHistoryOption) *FileChatMessageHistory {
	h := &FileChatMessageHistory{
		FilePath:            "chat_history.json",
		CreateDirIfNotExist: true,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}
