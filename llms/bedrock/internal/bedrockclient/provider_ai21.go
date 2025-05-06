package bedrockclient

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/tmc/langchaingo/llms"
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

func createAi21Completion(ctx context.Context, client *bedrockruntime.Client, modelID string, messages []Message, options llms.CallOptions) (*llms.ContentResponse, error) {
	txt := processInputMessagesGeneric(messages)
	inputContent := ai21TextGenerationInput{
		Prompt:        txt,
		Temperature:   options.Temperature,
		TopP:          options.TopP,
		MaxTokens:     getMaxTokens(options.MaxTokens, DefaultMaxTokenLength2048),
		StopSequences: options.StopWords,
		CountPenalty: struct {
			Scale float64 `json:"scale"`
		}{Scale: options.RepetitionPenalty},
		PresencePenalty: struct {
			Scale float64 `json:"scale"`
		}{Scale: 0},
		FrequencyPenalty: struct {
			Scale float64 `json:"scale"`
		}{Scale: 0},
		NumResults: options.CandidateCount,
	}

	body, err := json.Marshal(inputContent)
	if err != nil {
		return nil, err
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
				"input_tokens":  len(output.Prompt.Tokens),
				"output_tokens": len(completion.Data.Tokens),
			},
		}
	}

	return &llms.ContentResponse{Choices: choices}, nil
}
