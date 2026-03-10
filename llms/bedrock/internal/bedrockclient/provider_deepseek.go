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

type deepSeekTextGenerationInput struct {
	Prompt      string   `json:"prompt"`
	Temperature float64  `json:"temperature,omitempty"`
	TopP        float64  `json:"top_p,omitempty"`
	MaxTokens   int      `json:"max_tokens,omitempty"`
	Stop        []string `json:"stop,omitempty"`
}

type deepSeekTextGenerationOutput struct {
	Choices []struct {
		Text       string `json:"text"`
		StopReason string `json:"stop_reason"`
	} `json:"choices"`
}

type deepSeekStreamingResponseChunk struct {
	Choices []struct {
		Text       string `json:"text"`
		StopReason string `json:"stop_reason,omitempty"`
	} `json:"choices"`
}

func createDeepSeekCompletion(ctx context.Context,
	client *bedrockruntime.Client,
	modelID string,
	messages []Message,
	options llms.CallOptions,
) (*llms.ContentResponse, error) {
	// Format prompt with DeepSeek-R1 required formatting
	prompt := formatDeepSeekPrompt(messages)

	input := deepSeekTextGenerationInput{
		Prompt:      prompt,
		Temperature: options.GetTemperature(),
		TopP:        options.GetTopP(),
		MaxTokens:   getMaxTokens(options.GetMaxTokens(), 512),
		Stop:        options.StopWords,
	}

	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	if options.StreamingFunc != nil {
		modelInput := &bedrockruntime.InvokeModelWithResponseStreamInput{
			ModelId:     aws.String(modelID),
			Accept:      aws.String("application/json"),
			ContentType: aws.String("application/json"),
			Body:        body,
		}
		return parseDeepSeekStreamingResponse(ctx, client, modelInput, options)
	}

	modelInput := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(modelID),
		Accept:      aws.String("application/json"),
		ContentType: aws.String("application/json"),
		Body:        body,
	}

	resp, err := client.InvokeModel(ctx, modelInput)
	if err != nil {
		return nil, err
	}

	var output deepSeekTextGenerationOutput
	err = json.Unmarshal(resp.Body, &output)
	if err != nil {
		return nil, err
	}

	choices := make([]*llms.ContentChoice, 0, len(output.Choices))
	for _, choice := range output.Choices {
		choices = append(choices, &llms.ContentChoice{
			Content:    choice.Text,
			StopReason: choice.StopReason,
		})
	}

	return &llms.ContentResponse{
		Choices: choices,
	}, nil
}

func formatDeepSeekPrompt(messages []Message) string {
	var sb strings.Builder

	// Start with system message if available
	for _, message := range messages {
		if message.Role == llms.ChatMessageTypeSystem {
			sb.WriteString("<|begin_of_text|><|start_header_id|>system<|end_header_id|>\n")
			sb.WriteString(message.Content)
			sb.WriteString("<|eot_id|>\n")
			break
		}
	}

	// Add user messages
	for _, message := range messages {
		switch message.Role {
		case llms.ChatMessageTypeHuman:
			sb.WriteString("<|start_header_id|>user<|end_header_id|>\n")
			sb.WriteString(message.Content)
			sb.WriteString("<|eot_id|>\n")
		case llms.ChatMessageTypeAI:
			sb.WriteString("<|start_header_id|>assistant<|end_header_id|>\n")
			sb.WriteString(message.Content)
			sb.WriteString("<|eot_id|>\n")
		}
	}

	// End with assistant prompt
	sb.WriteString("<|start_header_id|>assistant<|end_header_id|>\n")

	return sb.String()
}

func parseDeepSeekStreamingResponse(ctx context.Context, client *bedrockruntime.Client, modelInput *bedrockruntime.InvokeModelWithResponseStreamInput, options llms.CallOptions) (*llms.ContentResponse, error) {
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
			var resp deepSeekStreamingResponseChunk
			err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&resp)
			if err != nil {
				return nil, err
			}

			if len(resp.Choices) > 0 && resp.Choices[0].Text != "" {
				if err = streaming.CallWithText(ctx, options.StreamingFunc, resp.Choices[0].Text); err != nil {
					return nil, err
				}
				contentchoices[0].Content += resp.Choices[0].Text
			}

			if len(resp.Choices) > 0 && resp.Choices[0].StopReason != "" {
				contentchoices[0].StopReason = resp.Choices[0].StopReason
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
