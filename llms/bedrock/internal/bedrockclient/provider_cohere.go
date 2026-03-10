package bedrockclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// Ref: https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters-cohere-command.html
// Legacy cohereTextGenerationInput is the input for the text generation for legacy Cohere Models.
type cohereTextGenerationInput struct {
	// The prompt that you want to pass to the model. Required
	Prompt string `json:"prompt"`
	// Use a lower value to decrease randomness in the response. Optional, default = 0.9
	Temperature float64 `json:"temperature,omitempty"`
	// Use a lower value to ignore less probable options. Optional, default = 0.75
	P float64 `json:"p,omitempty"`
	// Specify the number of token choices the model uses to generate the next token.
	// If both p and k are enabled, p acts after k
	// Optional, default = 0
	K int `json:"k,omitempty"`
	// Specify the maximum number of tokens to use in the generated response.
	// Optional, default = 20
	MaxTokens int `json:"max_tokens,omitempty"`
	// Configure up to four sequences that the model recognizes. After a stop sequence, the model stops generating further tokens.
	// The returned text doesn't contain the stop sequence.
	StopSequences  []string `json:"stop_sequences,omitempty"`
	NumGenerations int      `json:"num_generations,omitempty"`
}

// Ref: https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters-cohere-command-r-plus.html
// cohereCommandRInput is the input for Command R models
type cohereCommandRInput struct {
	Message     string `json:"message"`
	ChatHistory []struct {
		Role    string `json:"role"`
		Message string `json:"message"`
	} `json:"chat_history,omitempty"`
	MaxTokens     int      `json:"max_tokens,omitempty"`
	Temperature   float64  `json:"temperature,omitempty"`
	P             float64  `json:"p,omitempty"`
	K             int      `json:"k,omitempty"`
	StopSequences []string `json:"stop_sequences,omitempty"`
}

// Cohere role constants
const (
	CohereRoleUser    = "USER"
	CohereRoleChatbot = "CHATBOT"
	CohereRoleSystem  = "SYSTEM"
)

func getCohereRole(role llms.ChatMessageType) (string, error) {
	switch role {
	case llms.ChatMessageTypeSystem:
		return CohereRoleSystem, nil
	case llms.ChatMessageTypeAI:
		return CohereRoleChatbot, nil
	case llms.ChatMessageTypeGeneric:
		fallthrough
	case llms.ChatMessageTypeHuman:
		return CohereRoleUser, nil
	case llms.ChatMessageTypeFunction, llms.ChatMessageTypeTool:
		fallthrough
	default:
		return "", errors.New("role not supported")
	}
}

// Finish reason for the completion of the generation for Cohere Models.
const (
	CohereCompletionReasonComplete   = "COMPLETE"
	CohereCompletionReasonMaxTokens  = "MAX_TOKENS"
	CohereCompletionReasonError      = "ERROR"
	CohereCompletionReasonErrorToxic = "ERROR_TOXIC"
)

// cohereTextGenerationOutput is the output for the text generation for Cohere Models.
type cohereTextGenerationOutput struct {
	// The ID of the response.
	ID string `json:"id"`
	// The generations of the response.
	Generations []*cohereTextGenerationOutputGeneration `json:"generations"`
	// The text of the response (for Command R models)
	Text string `json:"text"`
	// The reason the generation finished (for Command R models)
	FinishReason string `json:"finish_reason"`
}

// cohereTextGenerationOutputGeneration is the generation output for the text generation for Cohere Models.
type cohereTextGenerationOutputGeneration struct {
	// The ID of the generation.
	ID string `json:"id"`
	// The index of the generation.
	Index int `json:"index"`
	// The reason the generation finished.
	FinishReason string `json:"finish_reason"`
	// The text of the generation.
	Text string `json:"text"`
}

// Cohere streaming response structure
type cohereStreamingResponseChunk struct {
	Text         string `json:"text"`
	FinishReason string `json:"finish_reason,omitempty"`
	IsFinished   bool   `json:"is_finished,omitempty"`
	Index        int    `json:"index,omitempty"`
	EventType    string `json:"event_type,omitempty"`
	GenerationId string `json:"generation_id,omitempty"`
}

func createCohereCompletion(ctx context.Context,
	client *bedrockruntime.Client,
	modelID string,
	messages []Message,
	options llms.CallOptions,
) (*llms.ContentResponse, error) {
	// Check if this is a Command R model
	if strings.Contains(modelID, "command-r") {
		return createCohereCommandRCompletion(ctx, client, modelID, messages, options)
	}

	// Legacy models (command-text-v14, command-light-text-v14)
	txt := processInputMessagesGeneric(messages)

	input := &cohereTextGenerationInput{
		Prompt:         txt,
		Temperature:    options.GetTemperature(),
		P:              options.GetTopP(),
		K:              options.GetTopK(),
		MaxTokens:      getMaxTokens(options.GetMaxTokens(), 20),
		StopSequences:  options.StopWords,
		NumGenerations: options.GetCandidateCount(),
	}

	body, err := json.Marshal(input)
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
		return parseCohereStreamingResponse(ctx, client, modelInput, options)
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

	var output cohereTextGenerationOutput

	err = json.Unmarshal(resp.Body, &output)
	if err != nil {
		return nil, err
	}

	choices := make([]*llms.ContentChoice, len(output.Generations))

	for i, gen := range output.Generations {
		choices[i] = &llms.ContentChoice{
			Content:    gen.Text,
			StopReason: gen.FinishReason,
			GenerationInfo: map[string]any{
				"generation_id": gen.ID,
				"index":         i,
			},
		}
	}

	return &llms.ContentResponse{
		Choices: choices,
	}, nil
}

func createCohereCommandRCompletion(ctx context.Context,
	client *bedrockruntime.Client,
	modelID string,
	messages []Message,
	options llms.CallOptions,
) (*llms.ContentResponse, error) {
	// Convert messages to Cohere format
	var currentMessage string
	var chatHistory []struct {
		Role    string `json:"role"`
		Message string `json:"message"`
	}

	for i, msg := range messages {
		role, err := getCohereRole(msg.Role)
		if err != nil {
			return nil, err
		}

		if i == len(messages)-1 {
			// Last message becomes the current message
			currentMessage = msg.Content
		} else {
			// Previous messages become chat history, but skip system messages
			if role != CohereRoleSystem {
				chatHistory = append(chatHistory, struct {
					Role    string `json:"role"`
					Message string `json:"message"`
				}{
					Role:    role,
					Message: msg.Content,
				})
			}
		}
	}

	input := &cohereCommandRInput{
		Message:       currentMessage,
		ChatHistory:   chatHistory,
		MaxTokens:     getMaxTokens(options.GetMaxTokens(), 512),
		Temperature:   options.GetTemperature(),
		P:             options.GetTopP(),
		K:             options.GetTopK(),
		StopSequences: options.StopWords,
	}

	body, err := json.Marshal(input)
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
		return parseCohereStreamingResponse(ctx, client, modelInput, options)
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

	var output cohereTextGenerationOutput
	err = json.Unmarshal(resp.Body, &output)
	if err != nil {
		return nil, err
	}

	// Command R models return text directly
	choices := []*llms.ContentChoice{
		{
			Content:    output.Text,
			StopReason: output.FinishReason,
			GenerationInfo: map[string]any{
				"id": output.ID,
			},
		},
	}

	return &llms.ContentResponse{
		Choices: choices,
	}, nil
}

func parseCohereStreamingResponse(ctx context.Context, client *bedrockruntime.Client, modelInput *bedrockruntime.InvokeModelWithResponseStreamInput, options llms.CallOptions) (*llms.ContentResponse, error) {
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
			var resp cohereStreamingResponseChunk
			err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&resp)
			if err != nil {
				return nil, err
			}

			// Send text chunk if available
			if resp.Text != "" {
				if err = streaming.CallWithText(ctx, options.StreamingFunc, resp.Text); err != nil {
					return nil, err
				}
				contentchoices[0].Content += resp.Text
			}

			// Set completion reason
			if resp.FinishReason != "" {
				contentchoices[0].StopReason = resp.FinishReason
			}

			// Set generation info
			if resp.GenerationId != "" {
				contentchoices[0].GenerationInfo["generation_id"] = resp.GenerationId
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
