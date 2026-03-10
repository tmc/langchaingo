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

// Ref: https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters-titan-text.html

// amazonTextGenerationConfigInput is the input for the text generation configuration for Amazon Models.
type amazonTextGenerationConfigInput struct {
	// The maximum number of tokens to generate per result. Optional, default = 512
	MaxTokens int `json:"maxTokenCount,omitempty"`
	// Use a lower value to ignore less probable options and decrease the diversity of responses. Optional, default = 1
	TopP float64 `json:"topP,omitempty"`
	// Use a lower value to decrease randomness in responses. Optional, default = 0.0
	Temperature float64 `json:"temperature,omitempty"`
	// Specify a character sequence to indicate where the model should stop.
	// Currently only supports: ["|", "User:"]
	StopSequences []string `json:"stopSequences,omitempty"`
}

// amazonTextGenerationInput is the input for the text generation for Amazon Models.
type amazonTextGenerationInput struct {
	// The text which the model is requested to continue.
	InputText string `json:"inputText"`
	// The configuration for the text generation
	TextGenerationConfig amazonTextGenerationConfigInput `json:"textGenerationConfig"`
}

// amazonTextGenerationOutput is the output for the text generation for Amazon Models.
type amazonTextGenerationOutput struct {
	// The number of tokens in the prompt
	InputTextTokenCount int32 `json:"inputTextTokenCount"`
	// The results of the request
	Results []struct {
		// The number of tokens in the response
		TokenCount int32 `json:"tokenCount"`
		// The generated text
		OutputText string `json:"outputText"`
		// The reason for the completion of the generation
		// One of: FINISH, LENGTH, CONTENT_FILTERED
		CompletionReason string `json:"completionReason"`
	} `json:"results"`
}

// Amazon Titan streaming response structure
type amazonStreamingResponseChunk struct {
	OutputText           string `json:"outputText"`
	Index                int    `json:"index"`
	TotalOutputTextCount int    `json:"totalOutputTextCount"`
	CompletionReason     string `json:"completionReason"`
	InputTextTokenCount  int32  `json:"inputTextTokenCount"`
	OutputTextTokenCount int32  `json:"outputTextTokenCount"`
}

// Finish reason for the completion of the generation for Amazon Models.
const (
	AmazonCompletionReasonFinish          = "FINISH"
	AmazonCompletionReasonMaxTokens       = "LENGTH"
	AmazonCompletionReasonContentFiltered = "CONTENT_FILTERED"
)

func createAmazonCompletion(ctx context.Context,
	client *bedrockruntime.Client,
	modelID string,
	messages []Message,
	options llms.CallOptions,
) (*llms.ContentResponse, error) {
	txt := processInputMessagesGeneric(messages)

	inputContent := amazonTextGenerationInput{
		InputText: txt,
		TextGenerationConfig: amazonTextGenerationConfigInput{
			MaxTokens:     getMaxTokens(options.GetMaxTokens(), 512),
			TopP:          options.GetTopP(),
			Temperature:   options.GetTemperature(),
			StopSequences: options.StopWords,
		},
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
		return parseAmazonStreamingResponse(ctx, client, modelInput, options)
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

	var output amazonTextGenerationOutput
	err = json.Unmarshal(resp.Body, &output)
	if err != nil {
		return nil, err
	}

	if len(output.Results) == 0 {
		return nil, errors.New("no results")
	}

	contentChoices := make([]*llms.ContentChoice, len(output.Results))

	for i, result := range output.Results {
		contentChoices[i] = &llms.ContentChoice{
			Content:    result.OutputText,
			StopReason: result.CompletionReason,
			GenerationInfo: map[string]any{
				"input_tokens":  output.InputTextTokenCount,
				"output_tokens": result.TokenCount,
				// Standardized field names for cross-provider compatibility
				"PromptTokens":     output.InputTextTokenCount,
				"CompletionTokens": result.TokenCount,
				"TotalTokens":      output.InputTextTokenCount + result.TokenCount,
			},
		}
	}

	return &llms.ContentResponse{
		Choices: contentChoices,
	}, nil
}

func parseAmazonStreamingResponse(ctx context.Context, client *bedrockruntime.Client, modelInput *bedrockruntime.InvokeModelWithResponseStreamInput, options llms.CallOptions) (*llms.ContentResponse, error) {
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
			var resp amazonStreamingResponseChunk
			err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&resp)
			if err != nil {
				return nil, err
			}

			// Send text chunk if available
			if resp.OutputText != "" {
				if err = streaming.CallWithText(ctx, options.StreamingFunc, resp.OutputText); err != nil {
					return nil, err
				}
				contentchoices[0].Content += resp.OutputText
			}

			// Set completion reason
			if resp.CompletionReason != "" {
				contentchoices[0].StopReason = resp.CompletionReason
			}

			// Set token counts
			if resp.InputTextTokenCount > 0 {
				contentchoices[0].GenerationInfo["input_tokens"] = resp.InputTextTokenCount
				contentchoices[0].GenerationInfo["PromptTokens"] = resp.InputTextTokenCount
			}
			if resp.OutputTextTokenCount > 0 {
				contentchoices[0].GenerationInfo["output_tokens"] = resp.OutputTextTokenCount
				contentchoices[0].GenerationInfo["CompletionTokens"] = resp.OutputTextTokenCount
			}
			if resp.InputTextTokenCount > 0 || resp.OutputTextTokenCount > 0 {
				contentchoices[0].GenerationInfo["TotalTokens"] = resp.InputTextTokenCount + resp.OutputTextTokenCount
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
