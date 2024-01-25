package llms

import "github.com/tmc/langchaingo/schema"

// TextParts is a helper function to create a MessageContent with a role and a
// list of text parts.
func TextParts(role schema.ChatMessageType, parts ...string) MessageContent {
	result := MessageContent{
		Role:  role,
		Parts: []ContentPart{},
	}
	for _, part := range parts {
		result.Parts = append(result.Parts, TextContent{
			Text: part,
		})
	}
	return result
}
