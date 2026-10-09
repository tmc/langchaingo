package databricksclientsllama31

import (
	"github.com/tmc/langchaingo/llms"
)

// MapRole maps ChatMessageType to LlamaRole.
func MapRole(chatRole llms.ChatMessageType) Role {
	switch chatRole {
	case llms.ChatMessageTypeAI:
		return RoleAssistant
	case llms.ChatMessageTypeHuman:
		return RoleUser
	case llms.ChatMessageTypeSystem:
		return RoleSystem
	case llms.ChatMessageTypeFunction, llms.ChatMessageTypeTool:
		return RoleIPython // Mapping tools and functions to ipython
	case llms.ChatMessageTypeGeneric:
		return RoleUser // Defaulting generic to user
	default:
		return Role(chatRole)
	}
}
