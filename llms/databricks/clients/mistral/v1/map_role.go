package databricksclientsmistralv1

import "github.com/tmc/langchaingo/llms"

// mapRole maps llms.ChatMessageType to Role.
// Map function.
func MapRole(chatRole llms.ChatMessageType) Role {
	switch chatRole {
	case llms.ChatMessageTypeAI:
		return RoleAssistant
	case llms.ChatMessageTypeHuman, llms.ChatMessageTypeGeneric:
		return RoleUser
	case llms.ChatMessageTypeSystem:
		return RoleSystem
	case llms.ChatMessageTypeTool, llms.ChatMessageTypeFunction:
		return RoleTool
	default:
		return Role(chatRole)
	}
}
