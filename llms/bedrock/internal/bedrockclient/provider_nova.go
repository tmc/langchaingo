package bedrockclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/tmc/langchaingo/llms"
)

// novaBinGenerationInputSource is the source of the content.
type novaBinGenerationInputSource struct {
	Bytes []byte `json:"bytes,omitempty"`
}

type novaImageInput struct {
	Format string                        `json:"format,omitempty"`
	Source *novaBinGenerationInputSource `json:"source,omitempty"`
}

// anthropicTextGenerationInputContent is a single message in the input.
type novaTextGenerationInputContent struct {
	// The text content. Required if type is "text"
	Text string `json:"text,omitempty"`
	// The image content. Required if type is "image"
	Image *novaImageInput `json:"image,omitempty"`
}

type novaTextGenerationInputMessage struct {
	// The role of the message. Required
	// One of: ["user", "assistant"]
	// For system prompt, use the system field in the input
	Role string `json:"role"`
	// The content of the message. Required
	Content []novaTextGenerationInputContent `json:"content"`
}

type novaSystemPrompt struct {
	Text string `json:"text,omitempty"`
}

// novaTextGenerationConfigInput is the input for the text generation configuration for Amazon Models.
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
type novaTextGenerationInput struct {
	Messages        []*novaTextGenerationInputMessage `json:"messages"`
	InferenceConfig novaInferenceConfigInput          `json:"inferenceConfig"`
	System          []*novaSystemPrompt               `json:"system,omitempty"`
}

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
		InputTokens               int  `json:"inputTokens"`
		OutputTokens              int  `json:"outputTokens"`
		TotalTokens               int  `json:"totalTokens"`
		CacheReadInputTokenCount  *int `json:"cacheReadInputTokenCount"`
		CacheWriteInputTokenCount *int `json:"cacheWriteInputTokenCount"`
	} `json:"usage"`
}

// Finish reason for the completion of the generation.
const (
	NovaCompletionReasonEndTurn      = "end_turn"
	NovaCompletionReasonMaxTokens    = "max_tokens"
	NovaCompletionReasonStopSequence = "stop_sequence"
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
			MaxTokens:     options.MaxTokens,
			Temperature:   options.Temperature,
			TopP:          options.TopP,
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
		return parseStreamingCompletionResponse(ctx, client, modelInput, options)
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
	} else if stopReason := output.StopReason; stopReason != NovaCompletionReasonEndTurn && stopReason != NovaCompletionReasonStopSequence {
		return nil, errors.New("completed due to " + stopReason + ". Maybe try increasing max tokens")
	}
	Contentchoices := make([]*llms.ContentChoice, len(content))
	for i, c := range content {
		Contentchoices[i] = &llms.ContentChoice{
			Content:    c.Text,
			StopReason: output.StopReason,
			GenerationInfo: map[string]interface{}{
				"input_tokens":  output.Usage.InputTokens,
				"output_tokens": output.Usage.OutputTokens,
			},
		}
	}
	return &llms.ContentResponse{
		Choices: Contentchoices,
	}, nil
}

type novaStreamingCompletionResponseChunk struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
	Delta struct {
		Type         string `json:"type"`
		Text         string `json:"text"`
		StopReason   string `json:"stop_reason"`
		StopSequence any    `json:"stop_sequence"`
	} `json:"delta"`
	AmazonBedrockInvocationMetrics struct {
		InputTokenCount   int `json:"inputTokenCount"`
		OutputTokenCount  int `json:"outputTokenCount"`
		InvocationLatency int `json:"invocationLatency"`
		FirstByteLatency  int `json:"firstByteLatency"`
	} `json:"amazon-bedrock-invocationMetrics"`
	Usage struct {
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Message struct {
		ID           string `json:"id"`
		Type         string `json:"type"`
		Role         string `json:"role"`
		Content      []any  `json:"content"`
		Model        string `json:"model"`
		StopReason   any    `json:"stop_reason"`
		StopSequence any    `json:"stop_sequence"`
		Usage        struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	} `json:"message"`
}

func parseNovaStreamingCompletionResponse(ctx context.Context, client *bedrockruntime.Client, modelInput *bedrockruntime.InvokeModelWithResponseStreamInput, options llms.CallOptions) (*llms.ContentResponse, error) {
	output, err := client.InvokeModelWithResponseStream(ctx, modelInput)
	if err != nil {
		return nil, err
	}
	stream := output.GetStream()
	if stream == nil {
		return nil, errors.New("no stream")
	}
	defer stream.Close()

	contentchoices := []*llms.ContentChoice{{GenerationInfo: map[string]interface{}{}}}
	for e := range stream.Events() {
		if err = stream.Err(); err != nil {
			return nil, err
		}

		if v, ok := e.(*types.ResponseStreamMemberChunk); ok {
			var resp novaStreamingCompletionResponseChunk
			err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&resp)
			if err != nil {
				return nil, err
			}

			switch resp.Type {
			case "message_start":
				contentchoices[0].GenerationInfo["input_tokens"] = resp.Message.Usage.InputTokens
			case "content_block_delta":
				if err = options.StreamingFunc(ctx, []byte(resp.Delta.Text)); err != nil {
					return nil, err
				}
				contentchoices[0].Content += resp.Delta.Text
			case "message_delta":
				contentchoices[0].StopReason = resp.Delta.StopReason
				contentchoices[0].GenerationInfo["output_tokens"] = resp.Usage.OutputTokens
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
			Source: &novaBinGenerationInputSource{
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
	default:
		return ""
	}
}
