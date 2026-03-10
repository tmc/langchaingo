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

// Ref: https://docs.ai21.com/reference/j2-complete-ref
type ai21TextGenerationInput struct {
	// The text which the model is requested to continue.
	Prompt string `json:"prompt"`
	// Modifies the distribution from which tokens are sampled. Optional, default = 0.7
	Temperature float64 `json:"temperature,omitempty"`
	// Sample tokens from the corresponding top percentile of probability mass. Optional, default = 1
	TopP float64 `json:"topP,omitempty"`
	// The maximum number of tokens to generate per result. Optional, default = 16
	MaxTokens int `json:"maxTokens,omitempty"`
	// Stops decoding if any of the strings is generated. Optional.
	StopSequences []string `json:"stopSequences,omitempty"`

	// The scale factor for the count penalty
	CountPenalty struct {
		Scale float64 `json:"scale"`
	} `json:"countPenalty"`
	// The scale factor for the presence penalty
	PresencePenalty struct {
		Scale float64 `json:"scale"`
	} `json:"presencePenalty"`
	// The scale factor for the frequency penalty
	FrequencyPenalty struct {
		Scale float64 `json:"scale"`
	} `json:"frequencyPenalty"`

	// The number of results to generate. Optional, default = 1
	NumResults int `json:"numResults,omitempty"`
}

// Ref: https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters-jamba.html
type ai21JambaInput struct {
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
	MaxTokens   int      `json:"max_tokens,omitempty"`
	Temperature float64  `json:"temperature,omitempty"`
	TopP        float64  `json:"top_p,omitempty"`
	Stop        []string `json:"stop,omitempty"`
	N           int      `json:"n,omitempty"`
}

// AI21 Jamba response structure (OpenAI-like format)
type ai21JambaOutput struct {
	ID      string `json:"id"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role      string `json:"role"`
			Content   string `json:"content"`
			ToolCalls any    `json:"tool_calls"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int32 `json:"prompt_tokens"`
		CompletionTokens int32 `json:"completion_tokens"`
		TotalTokens      int32 `json:"total_tokens"`
	} `json:"usage"`
	Meta  any    `json:"meta"`
	Model string `json:"model"`
}

// AI21 streaming response structure (for Jamba models)
type ai21StreamingResponseChunk struct {
	Text         string `json:"text"`
	FinishReason string `json:"finish_reason,omitempty"`
	Index        int    `json:"index,omitempty"`
	Usage        struct {
		PromptTokens     int32 `json:"prompt_tokens"`
		CompletionTokens int32 `json:"completion_tokens"`
	} `json:"usage,omitempty"`
}

// Legacy AI21 response structure for J2 models
type ai21TextGenerationOutput struct {
	// The ID of the request
	ID any `json:"id"` // Docs say it's a string, got number
	// The prompt that was used for the request

	// The input fields of the request (minified)
	Prompt struct {
		// The input tokens
		Tokens []struct{} `json:"tokens"` // for counting only
	} `json:"prompt"`

	// The completions of the request (minified)
	Completions []struct {
		// The generated data
		Data struct {
			// The generated text
			Text string `json:"text"`
			// The generated tokens
			Tokens []struct{} `json:"tokens"` // for counting only
		} `json:"data"`

		// The reason the generation was stopped
		FinishReason struct {
			// The reason the generation was stopped
			// One of: "length", "stop", "endoftext"
			Reason string `json:"reason"`
		} `json:"finishReason"`
	} `json:"completions"`
}

// Finish reason for the completion of the generation for AI21 Models.
const (
	Ai21CompletionReasonLength    = "length"
	Ai21CompletionReasonStop      = "stop"
	Ai21CompletionReasonEndOfText = "endoftext"
)

// AI21 Jamba role constants
const (
	Ai21RoleUser      = "user"
	Ai21RoleAssistant = "assistant"
	Ai21RoleSystem    = "system"
)

func getAi21Role(role llms.ChatMessageType) (string, error) {
	switch role {
	case llms.ChatMessageTypeSystem:
		return Ai21RoleSystem, nil
	case llms.ChatMessageTypeAI:
		return Ai21RoleAssistant, nil
	case llms.ChatMessageTypeGeneric:
		fallthrough
	case llms.ChatMessageTypeHuman:
		return Ai21RoleUser, nil
	case llms.ChatMessageTypeFunction, llms.ChatMessageTypeTool:
		fallthrough
	default:
		return "", errors.New("role not supported")
	}
}

func createAi21Completion(ctx context.Context, client *bedrockruntime.Client, modelID string, messages []Message, options llms.CallOptions) (*llms.ContentResponse, error) {
	// Check if this is a Jamba model (use messages API)
	if strings.Contains(modelID, "jamba") {
		return createAi21JambaCompletion(ctx, client, modelID, messages, options)
	}

	// Legacy J2 models (use prompt API)
	txt := processInputMessagesGeneric(messages)
	inputContent := ai21TextGenerationInput{
		Prompt:        txt,
		Temperature:   options.GetTemperature(),
		TopP:          options.GetTopP(),
		MaxTokens:     getMaxTokens(options.GetMaxTokens(), 2048),
		StopSequences: options.StopWords,
		CountPenalty: struct {
			Scale float64 `json:"scale"`
		}{Scale: options.GetRepetitionPenalty()},
		PresencePenalty: struct {
			Scale float64 `json:"scale"`
		}{Scale: options.GetPresencePenalty()},
		FrequencyPenalty: struct {
			Scale float64 `json:"scale"`
		}{Scale: options.GetFrequencyPenalty()},
		NumResults: options.GetCandidateCount(),
	}

	body, err := json.Marshal(inputContent)
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
		return parseAi21StreamingResponse(ctx, client, modelInput, options)
	}

	modelInput := bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(modelID),
		Body:        body,
		Accept:      aws.String("*/*"),
		ContentType: aws.String("application/json"),
	}

	resp, err := client.InvokeModel(ctx, &modelInput)
	if err != nil {
		return nil, err
	}

	var output ai21TextGenerationOutput
	err = json.Unmarshal(resp.Body, &output)
	if err != nil {
		return nil, err
	}

	choices := make([]*llms.ContentChoice, len(output.Completions))
	for i, completion := range output.Completions {
		choices[i] = &llms.ContentChoice{
			Content:    completion.Data.Text,
			StopReason: completion.FinishReason.Reason,
			GenerationInfo: map[string]any{
				"id":            output.ID,
				"input_tokens":  int32(len(output.Prompt.Tokens)),
				"output_tokens": int32(len(completion.Data.Tokens)),
				// Standardized field names for cross-provider compatibility
				"PromptTokens":     int32(len(output.Prompt.Tokens)),
				"CompletionTokens": int32(len(completion.Data.Tokens)),
				"TotalTokens":      int32(len(output.Prompt.Tokens)) + int32(len(completion.Data.Tokens)),
			},
		}
	}

	return &llms.ContentResponse{Choices: choices}, nil
}

func createAi21JambaCompletion(ctx context.Context, client *bedrockruntime.Client, modelID string, messages []Message, options llms.CallOptions) (*llms.ContentResponse, error) {
	jambaMessages := make([]struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}, len(messages))

	for i, msg := range messages {
		role, err := getAi21Role(msg.Role)
		if err != nil {
			return nil, err
		}
		jambaMessages[i].Role = role
		jambaMessages[i].Content = msg.Content
	}

	inputContent := ai21JambaInput{
		Messages:    jambaMessages,
		MaxTokens:   getMaxTokens(options.GetMaxTokens(), 4096),
		Temperature: options.GetTemperature(),
		TopP:        options.GetTopP(),
		Stop:        options.StopWords,
		N:           options.GetCandidateCount(),
	}

	body, err := json.Marshal(inputContent)
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
		return parseAi21StreamingResponse(ctx, client, modelInput, options)
	}

	modelInput := bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(modelID),
		Body:        body,
		Accept:      aws.String("*/*"),
		ContentType: aws.String("application/json"),
	}

	resp, err := client.InvokeModel(ctx, &modelInput)
	if err != nil {
		return nil, err
	}

	var output ai21JambaOutput
	err = json.Unmarshal(resp.Body, &output)
	if err != nil {
		return nil, err
	}

	choices := make([]*llms.ContentChoice, len(output.Choices))
	for i, choice := range output.Choices {
		choices[i] = &llms.ContentChoice{
			Content:    choice.Message.Content,
			StopReason: choice.FinishReason,
			GenerationInfo: map[string]any{
				"id":            output.ID,
				"input_tokens":  output.Usage.PromptTokens,
				"output_tokens": output.Usage.CompletionTokens,
			},
		}
	}

	return &llms.ContentResponse{Choices: choices}, nil
}

func parseAi21StreamingResponse(ctx context.Context, client *bedrockruntime.Client, modelInput *bedrockruntime.InvokeModelWithResponseStreamInput, options llms.CallOptions) (*llms.ContentResponse, error) {
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
			var resp ai21StreamingResponseChunk
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

			// Set token counts if available
			if resp.Usage.PromptTokens > 0 {
				contentchoices[0].GenerationInfo["input_tokens"] = resp.Usage.PromptTokens
				contentchoices[0].GenerationInfo["PromptTokens"] = resp.Usage.PromptTokens
			}
			if resp.Usage.CompletionTokens > 0 {
				contentchoices[0].GenerationInfo["output_tokens"] = resp.Usage.CompletionTokens
				contentchoices[0].GenerationInfo["CompletionTokens"] = resp.Usage.CompletionTokens
			}
			if resp.Usage.PromptTokens > 0 || resp.Usage.CompletionTokens > 0 {
				contentchoices[0].GenerationInfo["TotalTokens"] = resp.Usage.PromptTokens + resp.Usage.CompletionTokens
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
