package bedrockclient

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/tmc/langchaingo/llms"
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
	InputTextTokenCount int `json:"inputTextTokenCount"`
	// The results of the request
	Results []struct {
		// The number of tokens in the response
		TokenCount int `json:"tokenCount"`
		// The generated text
		OutputText string `json:"outputText"`
		// The reason for the completion of the generation
		// One of: FINISH, LENGTH, CONTENT_FILTERED
		CompletionReason string `json:"completionReason"`
	} `json:"results"`
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
			MaxTokens:     getMaxTokens(options.MaxTokens, DefaultMaxTokenLength512),
			TopP:          options.TopP,
			Temperature:   options.Temperature,
			StopSequences: options.StopWords,
		},
	}

	body, err := json.Marshal(inputContent)
	if err != nil {
		return nil, err
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
			},
		}
	}

	return &llms.ContentResponse{
		Choices: contentChoices,
	}, nil
}
