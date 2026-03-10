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

// Ref: https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters-meta.html

// metaTextGenerationInput is the input to the model.
type metaTextGenerationInput struct {
	// The prompt that you want to pass to the model. Required
	Prompt string `json:"prompt"`
	// Used to control the randomness of the generation. Optional, default = 0.5
	Temperature float64 `json:"temperature,omitempty"`
	// Used to lower value to ignore less probable options. Optional, default = 0.9
	TopP float64 `json:"top_p,omitempty"`
	// The maximum number of tokens to generate per result.
	// The model truncates the response once the generated text exceeds max_gen_len.
	// Optional, default = 512
	MaxGenLen int `json:"max_gen_len,omitempty"`
}

// metaTextGenerationOutput is the output from the model.
type metaTextGenerationOutput struct {
	// The generated text.
	Generation string `json:"generation"`
	// The number of tokens in the prompt.
	PromptTokenCount int32 `json:"prompt_token_count"`
	// The number of tokens in the generated text.
	GenerationTokenCount int32 `json:"generation_token_count"`
	// The reason why the response stopped generating text.
	// One of: ["stop", "length"]
	StopReason string `json:"stop_reason"`
}

// Meta streaming response structure
type metaStreamingResponseChunk struct {
	Generation           string `json:"generation"`
	PromptTokenCount     int32  `json:"prompt_token_count,omitempty"`
	GenerationTokenCount int32  `json:"generation_token_count,omitempty"`
	StopReason           string `json:"stop_reason,omitempty"`
	Amazon               struct {
		BedrockInvocationMetrics struct {
			InputTokenCount  int32 `json:"inputTokenCount"`
			OutputTokenCount int32 `json:"outputTokenCount"`
		} `json:"bedrock-invocationMetrics"`
	} `json:"amazon-bedrock-invocationMetrics,omitempty"`
}

// Finish reason for Meta models
const (
	MetaCompletionReasonStop   = "stop"
	MetaCompletionReasonLength = "length"
)

func createMetaCompletion(ctx context.Context,
	client *bedrockruntime.Client,
	modelID string,
	messages []Message,
	options llms.CallOptions,
) (*llms.ContentResponse, error) {
	txt := processInputMessagesGeneric(messages)

	input := &metaTextGenerationInput{
		Prompt:      txt,
		Temperature: options.GetTemperature(),
		TopP:        options.GetTopP(),
		MaxGenLen:   getMaxTokens(options.GetMaxTokens(), 512),
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
		return parseMetaStreamingResponse(ctx, client, modelInput, options)
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

	var output metaTextGenerationOutput

	err = json.Unmarshal(resp.Body, &output)
	if err != nil {
		return nil, err
	}

	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content:    output.Generation,
				StopReason: output.StopReason,
				GenerationInfo: map[string]any{
					"input_tokens":  output.PromptTokenCount,
					"output_tokens": output.GenerationTokenCount,
					// Standardized field names for cross-provider compatibility
					"PromptTokens":     output.PromptTokenCount,
					"CompletionTokens": output.GenerationTokenCount,
					"TotalTokens":      output.PromptTokenCount + output.GenerationTokenCount,
				},
			},
		},
	}, nil
}

func parseMetaStreamingResponse(ctx context.Context, client *bedrockruntime.Client, modelInput *bedrockruntime.InvokeModelWithResponseStreamInput, options llms.CallOptions) (*llms.ContentResponse, error) {
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
			var resp metaStreamingResponseChunk
			err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&resp)
			if err != nil {
				return nil, err
			}

			// Send text chunk if available
			if resp.Generation != "" {
				if err = streaming.CallWithText(ctx, options.StreamingFunc, resp.Generation); err != nil {
					return nil, err
				}
				contentchoices[0].Content += resp.Generation
			}

			// Set completion reason
			if resp.StopReason != "" {
				contentchoices[0].StopReason = resp.StopReason
			}

			// Set token counts
			if resp.PromptTokenCount > 0 {
				contentchoices[0].GenerationInfo["input_tokens"] = resp.PromptTokenCount
				contentchoices[0].GenerationInfo["PromptTokens"] = resp.PromptTokenCount
			}
			if resp.GenerationTokenCount > 0 {
				contentchoices[0].GenerationInfo["output_tokens"] = resp.GenerationTokenCount
				contentchoices[0].GenerationInfo["CompletionTokens"] = resp.GenerationTokenCount
			}
			if resp.PromptTokenCount > 0 || resp.GenerationTokenCount > 0 {
				contentchoices[0].GenerationInfo["TotalTokens"] = resp.PromptTokenCount + resp.GenerationTokenCount
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
