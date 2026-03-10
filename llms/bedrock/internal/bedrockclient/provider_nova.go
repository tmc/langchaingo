package bedrockclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// Ref: https://boto3.amazonaws.com/v1/documentation/api/1.35.8/reference/services/bedrock-runtime/client/converse.html
// This was the reference for the input and output types.

// Ref: https://docs.aws.amazon.com/nova/latest/userguide/what-is-nova.html
// This was the reference for the model parameters.

// novaBinGenerationInputSource is the source of the content.
// It is used for sending binary content such as images to the model.
type novaBinGenerationInputSource struct {
	// The bytes of the content. Required if type is "image"
	Bytes []byte `json:"bytes,omitempty"`
}

// novaImageInput is the input for the image content.
type novaImageInput struct {
	// The format of the image. Required if type is "image"
	// One of: ["jpeg", "png", "webp", "gif"]
	Format string `json:"format,omitempty"`
	// The source of the content. Required if type is "image"
	Source novaBinGenerationInputSource `json:"source,omitempty"`
}

// novaTextGenerationInputContent is the content of a single input message.
// It can be either a text or an image.
type novaTextGenerationInputContent struct {
	// The text content. Required if type is "text"
	Text string `json:"text,omitempty"`
	// The image content. Required if type is "image"
	Image *novaImageInput `json:"image,omitempty"`
}

// novaTextGenerationInputMessage is a single message in the input.
type novaTextGenerationInputMessage struct {
	// The role of the message. Required
	// One of: ["user", "assistant"]
	// For system prompt, use the system field in the input
	Role string `json:"role"`
	// The content of the message. Required
	Content []novaTextGenerationInputContent `json:"content"`
}

// novaSystemPrompt is the system prompt for the input.
// It is used to provide instructions to the model.
type novaSystemPrompt struct {
	Text string `json:"text,omitempty"`
}

// novaInferenceConfigInput is the input for the text generation configuration for Amazon Nova Models.
type novaInferenceConfigInput struct {
	// The maximum number of tokens to generate per result. Optional, default = 512
	MaxTokens int `json:"maxTokens,omitempty"`
	// Use a lower value to ignore less probable options and decrease the diversity of responses. Optional, default = 1
	TopP float64 `json:"topP,omitempty"`
	// Use a lower value to decrease randomness in responses. Optional, default = 0.0
	Temperature float64 `json:"temperature,omitempty"`
	// Specify a character sequence to indicate where the model should stop.
	// Currently only supports: ["|", "User:"]
	StopSequences []string `json:"stopSequences,omitempty"`
}

// novaTextGenerationInput is the input for the text generation for Amazon Nova Models.
type novaTextGenerationInput struct {
	// The messages to send to the model. Required
	Messages []*novaTextGenerationInputMessage `json:"messages"`
	// The configuration for the text generation. Required
	InferenceConfig novaInferenceConfigInput `json:"inferenceConfig"`
	// The system prompt for the input. Optional
	System []*novaSystemPrompt `json:"system,omitempty"`
}

// novaTextGenerationOutput is the output for the text generation for Amazon Nova Models.
type novaTextGenerationOutput struct {
	Output struct {
		Message struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
			Role string `json:"role"`
		} `json:"message"`
	} `json:"output"`
	StopReason string `json:"stopReason"`
	Usage      struct {
		InputTokens               int32  `json:"inputTokens"`
		OutputTokens              int32  `json:"outputTokens"`
		TotalTokens               int32  `json:"totalTokens"`
		CacheReadInputTokenCount  *int32 `json:"cacheReadInputTokenCount"`
		CacheWriteInputTokenCount *int32 `json:"cacheWriteInputTokenCount"`
	} `json:"usage"`
}

// Nova streaming response structure
type novaStreamingResponseChunk struct {
	ContentBlockDelta struct {
		Delta struct {
			Text string `json:"text"`
		} `json:"delta"`
	} `json:"contentBlockDelta"`
	MessageStart struct {
		Role  string `json:"role"`
		Usage struct {
			InputTokens int32 `json:"inputTokens"`
		} `json:"usage"`
	} `json:"messageStart"`
	MessageDelta struct {
		StopReason string `json:"stopReason"`
		Usage      struct {
			OutputTokens int32 `json:"outputTokens"`
		} `json:"usage"`
	} `json:"messageDelta"`
	MessageStop struct{} `json:"messageStop"`
}

// Finish reason for Nova models
const (
	NovaCompletionReasonEndTurn         = "end_turn"
	NovaCompletionReasonMaxTokens       = "max_tokens"
	NovaCompletionReasonStopSequence    = "stop_sequence"
	NovaCompletionReasonContentFiltered = "content_filtered"
)

// Role attribute for the anthropic message.
const (
	NovaSystem        = "system"
	NovaRoleUser      = "user"
	NovaRoleAssistant = "assistant"
)

// Type attribute for the anthropic message.
const (
	NovaMessageTypeText  = "text"
	NovaMessageTypeImage = "image"
)

func novaInputToJSON(inputContents []*novaTextGenerationInputMessage, systemPrompt string, options llms.CallOptions) ([]byte, error) {
	input := novaTextGenerationInput{
		Messages: inputContents,
		InferenceConfig: novaInferenceConfigInput{
			MaxTokens:     options.GetMaxTokens(),
			Temperature:   options.GetTemperature(),
			TopP:          options.GetTopP(),
			StopSequences: options.StopWords,
		},
		System: []*novaSystemPrompt{{Text: systemPrompt}},
	}
	return json.Marshal(input)
}

func parseNovaResponseBody(body []byte) (*novaTextGenerationOutput, error) {
	var output novaTextGenerationOutput
	err := json.Unmarshal(body, &output)
	return &output, err
}

func createNovaCompletion(ctx context.Context,
	client *bedrockruntime.Client,
	modelID string,
	messages []Message,
	options llms.CallOptions,
) (*llms.ContentResponse, error) {
	inputContents, systemPrompt, err := processInputMessagesNova(messages)
	if err != nil {
		return nil, err
	}

	body, err := novaInputToJSON(inputContents, systemPrompt, options)
	if err != nil {
		return nil, err
	}

	if options.StreamingFunc != nil {
		modelInput := &bedrockruntime.InvokeModelWithResponseStreamInput{
			ModelId:     aws.String(modelID),
			Accept:      aws.String("*/*"),
			ContentType: aws.String("application/json"),
			Body:        body,
		}
		return parseNovaStreamingResponse(ctx, client, modelInput, options)
	}

	modelInput := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(modelID),
		Accept:      aws.String("*/*"),
		ContentType: aws.String("application/json"),
		Body:        body,
	}
	resp, err := client.InvokeModel(ctx, modelInput)
	if err != nil {
		return nil, err
	}

	output, err := parseNovaResponseBody(resp.Body)
	if err != nil {
		return nil, err
	}

	content := output.Output.Message.Content
	if len(content) == 0 {
		return nil, errors.New("no results")
	} else if stopReason := output.StopReason; stopReason != NovaCompletionReasonEndTurn &&
		stopReason != NovaCompletionReasonStopSequence &&
		stopReason != NovaCompletionReasonMaxTokens &&
		stopReason != NovaCompletionReasonContentFiltered {
		return nil, errors.New("completed due to " + stopReason + ". Maybe try increasing max tokens")
	}
	Contentchoices := make([]*llms.ContentChoice, len(content))
	for i, c := range content {
		Contentchoices[i] = &llms.ContentChoice{
			Content:    c.Text,
			StopReason: output.StopReason,
			GenerationInfo: map[string]any{
				"input_tokens":  output.Usage.InputTokens,
				"output_tokens": output.Usage.OutputTokens,
				// Standardized field names for cross-provider compatibility
				"PromptTokens":     output.Usage.InputTokens,
				"CompletionTokens": output.Usage.OutputTokens,
				"TotalTokens":      output.Usage.InputTokens + output.Usage.OutputTokens,
			},
		}
	}
	return &llms.ContentResponse{
		Choices: Contentchoices,
	}, nil
}

// process the input messages to anthropic supported input
// returns the input content and system prompt.
func processInputMessagesNova(messages []Message) ([]*novaTextGenerationInputMessage, string, error) {
	chunkedMessages := make([][]Message, 0, len(messages))
	currentChunk := make([]Message, 0, len(messages))
	var lastRole llms.ChatMessageType
	for _, message := range messages {
		if message.Role != lastRole {
			if len(currentChunk) > 0 {
				chunkedMessages = append(chunkedMessages, currentChunk)
			}
			currentChunk = make([]Message, 0, len(messages))
		}
		currentChunk = append(currentChunk, message)
		lastRole = message.Role
	}
	if len(currentChunk) > 0 {
		chunkedMessages = append(chunkedMessages, currentChunk)
	}

	inputContents := make([]*novaTextGenerationInputMessage, 0, len(messages))
	var systemPrompt string
	for _, chunk := range chunkedMessages {
		role, err := getNovaRole(chunk[0].Role)
		if err != nil {
			return nil, "", err
		}
		if role == NovaSystem {
			if systemPrompt != "" {
				return nil, "", errors.New("multiple system prompts")
			}
			for _, message := range chunk {
				c := getNovaInputContent(message)
				systemPrompt += c.Text
			}
			continue
		}
		content := make([]novaTextGenerationInputContent, 0, len(chunk))
		for _, message := range chunk {
			content = append(content, getNovaInputContent(message))
		}
		inputContents = append(inputContents, &novaTextGenerationInputMessage{
			Role:    role,
			Content: content,
		})
	}
	return inputContents, systemPrompt, nil
}

// process the role of the message to anthropic supported role.
func getNovaRole(role llms.ChatMessageType) (string, error) {
	switch role {
	case llms.ChatMessageTypeSystem:
		return NovaSystem, nil

	case llms.ChatMessageTypeAI:
		return NovaRoleAssistant, nil

	case llms.ChatMessageTypeGeneric:
		fallthrough
	case llms.ChatMessageTypeHuman:
		return NovaRoleUser, nil
	case llms.ChatMessageTypeFunction, llms.ChatMessageTypeTool:
		fallthrough
	default:
		return "", errors.New("role not supported")
	}
}

func getNovaInputContent(message Message) novaTextGenerationInputContent {
	var c novaTextGenerationInputContent
	if message.Type == NovaMessageTypeText {
		c = novaTextGenerationInputContent{
			Text: message.Content,
		}
	} else if message.Type == NovaMessageTypeImage {
		c = novaTextGenerationInputContent{}
		c.Image = &novaImageInput{
			Format: mimeTypeToFormat(message.MimeType),
			Source: novaBinGenerationInputSource{
				Bytes: []byte(message.Content),
			},
		}
	}
	return c
}

func mimeTypeToFormat(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return "jpeg"
	case "image/png":
		return "png"
	case "image/webp":
		return "webp"
	case "image/gif":
		return "gif"
	default:
		return ""
	}
}

func parseNovaStreamingResponse(ctx context.Context, client *bedrockruntime.Client, modelInput *bedrockruntime.InvokeModelWithResponseStreamInput, options llms.CallOptions) (*llms.ContentResponse, error) {
	output, err := client.InvokeModelWithResponseStream(ctx, modelInput)
	if err != nil {
		return nil, err
	}
	stream := output.GetStream()
	if stream == nil {
		return nil, errors.New("no stream")
	}
	defer stream.Close()
	defer streaming.CallWithDone(ctx, options.StreamingFunc) //nolint:errcheck

	contentchoices := []*llms.ContentChoice{{GenerationInfo: map[string]any{}}}
	for e := range stream.Events() {
		if err = stream.Err(); err != nil {
			return nil, err
		}

		if v, ok := e.(*types.ResponseStreamMemberChunk); ok {
			var resp novaStreamingResponseChunk
			err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&resp)
			if err != nil {
				return nil, err
			}

			// Check for content delta (text chunks)
			if resp.ContentBlockDelta.Delta.Text != "" {
				if err = streaming.CallWithText(ctx, options.StreamingFunc, resp.ContentBlockDelta.Delta.Text); err != nil {
					return nil, err
				}
				contentchoices[0].Content += resp.ContentBlockDelta.Delta.Text
			}

			// Check for message start (contains input tokens)
			if resp.MessageStart.Usage.InputTokens > 0 {
				contentchoices[0].GenerationInfo["input_tokens"] = resp.MessageStart.Usage.InputTokens
				contentchoices[0].GenerationInfo["PromptTokens"] = resp.MessageStart.Usage.InputTokens
			}

			// Check for message delta (contains stop reason and output tokens)
			if resp.MessageDelta.StopReason != "" {
				contentchoices[0].StopReason = resp.MessageDelta.StopReason
			}
			if resp.MessageDelta.Usage.OutputTokens > 0 {
				contentchoices[0].GenerationInfo["output_tokens"] = resp.MessageDelta.Usage.OutputTokens
				contentchoices[0].GenerationInfo["CompletionTokens"] = resp.MessageDelta.Usage.OutputTokens
			}
			if resp.MessageStart.Usage.InputTokens > 0 || resp.MessageDelta.Usage.OutputTokens > 0 {
				contentchoices[0].GenerationInfo["TotalTokens"] = resp.MessageStart.Usage.InputTokens + resp.MessageDelta.Usage.OutputTokens
			}
		}
	}
	if err = stream.Err(); err != nil {
		return nil, err
	}

	return &llms.ContentResponse{
		Choices: contentchoices,
	}, nil
}
