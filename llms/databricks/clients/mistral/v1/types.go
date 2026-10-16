package databricksclientsmistralv1

// Role the role of the chat message.
type Role string

// Role the role of the chat message.
const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
	RoleTool      Role = "tool"
)

// FinishReason the reason that a chat message was finished.
type FinishReason string

// FinishReason the reason that a chat message was finished.
const (
	FinishReasonStop   FinishReason = "stop"
	FinishReasonLength FinishReason = "length"
	FinishReasonError  FinishReason = "error"
)

// ResponseFormat the format that the response must adhere to.
type ResponseFormat string

// ResponseFormat the format that the response must adhere to.
const (
	ResponseFormatText       ResponseFormat = "text"
	ResponseFormatJSONObject ResponseFormat = "json_object"
)

// ToolType type of tool defined for the llm.
type ToolType string

// ToolType type of tool defined for the llm.
const (
	ToolTypeFunction ToolType = "function"
)

// ToolChoice the choice of tool to use.
type ToolChoice string

// ToolChoice the choice of tool to use.
const (
	ToolChoiceAny  ToolChoice = "any"
	ToolChoiceAuto ToolChoice = "auto"
	ToolChoiceNone ToolChoice = "none"
)

// Tool definition of a tool that the llm can call.
type Tool struct {
	Type     ToolType `json:"type"`
	Function Function `json:"function"`
}

// Function definition of a function that the llm can call including its parameters.
type Function struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

// FunctionCall represents a request to call an external tool by the llm.
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolCall represents the call to a tool by the llm.
type ToolCall struct {
	ID       string       `json:"id"`
	Type     ToolType     `json:"type"`
	Function FunctionCall `json:"function"`
}

// ChatMessage represents a single message in a chat.
type ChatMessage struct {
	Role      Role       `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}
