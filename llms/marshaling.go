package llms

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// UnmarshalYAML custom unmarshaling logic for MessageContent.
func (mc *MessageContent) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var singleText struct {
		Role ChatMessageType `yaml:"role"`
		Text string          `yaml:"text"`
	}

	if err := unmarshal(&singleText); err == nil && singleText.Text != "" {
		mc.Role = singleText.Role
		mc.Parts = []ContentPart{TextContent{Text: singleText.Text}}
		return nil
	}

	var raw struct {
		Role  ChatMessageType          `yaml:"role"`
		Parts []map[string]interface{} `yaml:"parts"`
	}

	if err := unmarshal(&raw); err != nil {
		return err
	}

	mc.Role = raw.Role

	for _, part := range raw.Parts {
		switch part["type"] {
		case "text", nil:
			var content TextContent
			if err := mapToStruct(part, &content); err != nil {
				return err
			}
			mc.Parts = append(mc.Parts, content)
		case "image_url":
			var content ImageURLContent
			content.URL, _ = part["url"].(string)
			mc.Parts = append(mc.Parts, content)
		case "binary":
			var content BinaryContent
			data, err := base64.StdEncoding.DecodeString(part["data"].(string))
			if err != nil {
				return err
			}
			content.MIMEType, _ = part["mime_type"].(string)
			content.Data = data
			mc.Parts = append(mc.Parts, content)
		case "tool_call":
			var content ToolCall
			tc, ok := part["tool_call"].(map[string]interface{})
			if !ok {
				return fmt.Errorf("tool_call part is not a map")
			}
			if err := mapToStruct(tc, &content); err != nil {
				return err
			}
			mc.Parts = append(mc.Parts, content)
		case "tool_response":
			var content ToolCallResponse
			tr, ok := part["tool_response"].(map[string]interface{})
			if !ok {
				return fmt.Errorf("tool_response part is not a map")
			}
			if err := mapToStruct(tr, &content); err != nil {
				return err
			}
			mc.Parts = append(mc.Parts, content)
		default:
			return fmt.Errorf("unknown content type: %s", part["type"])
		}
	}

	return nil
}

func (mc *MessageContent) UnmarshalJSON(data []byte) error {
	var m struct {
		Role  ChatMessageType `json:"role"`
		Text  string          `json:"text"`
		Parts []struct {
			Type     string `json:"type"`
			Text     string `json:"text,omitempty"`
			ImageURL struct {
				URL string `json:"url"`
			} `json:"image_url,omitempty"`
			Binary struct {
				Data     string `json:"data"`
				MIMEType string `json:"mime_type"`
			} `json:"binary,omitempty"`
			ID       string `json:"id"`
			ToolCall struct {
				ID           string        `json:"id"`
				Type         string        `json:"type"`
				FunctionCall *FunctionCall `json:"function"`
			} `json:"tool_call"`
			ToolResponse struct {
				ToolCallID string `json:"tool_call_id"`
				Name       string `json:"name"`
				Content    string `json:"content"`
			} `json:"tool_response"`
		} `json:"parts"`
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	mc.Role = m.Role

	for _, part := range m.Parts {
		switch part.Type {
		case "text":
			mc.Parts = append(mc.Parts, TextContent{Text: part.Text})
		case "image_url":
			mc.Parts = append(mc.Parts, ImageURLContent{URL: part.ImageURL.URL})
		case "binary":
			decoded, err := base64.StdEncoding.DecodeString(part.Binary.Data)
			if err != nil {
				return fmt.Errorf("failed to decode binary data: %w", err)
			}
			mc.Parts = append(mc.Parts, BinaryContent{MIMEType: part.Binary.MIMEType, Data: decoded})
		case "tool_call":
			mc.Parts = append(mc.Parts, ToolCall{
				ID:           part.ToolCall.ID,
				Type:         part.ToolCall.Type,
				FunctionCall: part.ToolCall.FunctionCall,
			})
		case "tool_response":
			mc.Parts = append(mc.Parts, ToolCallResponse{
				ToolCallID: part.ToolResponse.ToolCallID,
				Name:       part.ToolResponse.Name,
				Content:    string(part.ToolResponse.Content),
			})
		default:
			return fmt.Errorf("unknown content type: %s", part.Type)
		}
	}
	// Special case: handle single text part directly:
	if len(mc.Parts) == 0 && m.Text != "" {
		mc.Parts = []ContentPart{TextContent{Text: m.Text}}
	}
	return nil
}

// MarshalYAML custom marshaling logic for MessageContent.
func (mc MessageContent) MarshalYAML() (interface{}, error) {
	// Special case: handle single text part directly
	if len(mc.Parts) == 1 {
		if content, ok := mc.Parts[0].(TextContent); ok {
			return map[string]interface{}{
				"role": mc.Role,
				"text": content.Text,
			}, nil
		}
	}

	var parts []map[string]interface{}
	for _, part := range mc.Parts {
		switch content := part.(type) {
		case TextContent:
			parts = append(parts, map[string]interface{}{
				"type": "text",
				"text": content.Text,
			})
		case ImageURLContent:
			parts = append(parts, map[string]interface{}{
				"type": "image_url",
				"url":  content.URL,
			})
		case BinaryContent:
			parts = append(parts, map[string]interface{}{
				"type":      "binary",
				"mime_type": content.MIMEType,
				"data":      base64.StdEncoding.EncodeToString(content.Data),
			})
		case ToolCall:
			parts = append(parts, map[string]interface{}{
				"type":      "tool_call",
				"tool_call": content,
			})
		case ToolCallResponse:
			parts = append(parts, map[string]interface{}{
				"type":          "tool_response",
				"tool_response": content,
			})
		default:
			return nil, fmt.Errorf("unknown content type: %T", content)
		}
	}

	raw := make(map[string]interface{})
	raw["role"] = mc.Role
	raw["parts"] = parts
	return raw, nil
}

func (mc MessageContent) MarshalJSON() ([]byte, error) {
	hasSingleTextPart := false
	if len(mc.Parts) == 1 {
		_, hasSingleTextPart = mc.Parts[0].(TextContent)
	}
	if hasSingleTextPart {
		tp, _ := mc.Parts[0].(TextContent)
		return json.Marshal(struct {
			Role ChatMessageType `json:"role"`
			Text string          `json:"text"`
		}{Role: mc.Role, Text: tp.Text})
	}

	return json.Marshal(struct {
		Role  ChatMessageType `json:"role"`
		Parts []ContentPart   `json:"parts"`
	}{
		Role:  mc.Role,
		Parts: mc.Parts,
	})
}

// Helper function to map raw data to struct.
func mapToStruct(data map[string]interface{}, target interface{}) error {
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(bytes, target)
}
