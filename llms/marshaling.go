package llms

import (
	"encoding/base64"
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
			if err := mapToStruct(part, &content); err != nil {
				return err
			}
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
		default:
			return fmt.Errorf("unknown content type: %s", part["type"])
		}
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
		default:
			return nil, fmt.Errorf("unknown content type")
		}
	}

	raw := make(map[string]interface{})
	raw["role"] = mc.Role
	raw["parts"] = parts

	return raw, nil
}

// Helper function to map raw data to struct.
func mapToStruct(data map[string]interface{}, target interface{}) error {
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(bytes, target)
}
