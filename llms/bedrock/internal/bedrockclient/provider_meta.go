package bedrockclient

import (
	"context"
	"encoding/json"
	
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/yincongcyincong/langchaingo/llms"
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
	PromptTokenCount int `json:"prompt_token_count"`
	// The number of tokens in the generated text.
	GenerationTokenCount int `json:"generation_token_count"`
	// The reason why the response stopped generating text.
	// One of: ["stop", "length"]
	StopReason string `json:"stop_reason"`
}

// Finish reason for the completion of the generation.
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
		Temperature: options.Temperature,
		TopP:        options.TopP,
		MaxGenLen:   getMaxTokens(options.MaxTokens, 512),
	}
	
	body, err := json.Marshal(input)
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
				GenerationInfo: map[string]interface{}{
					"input_tokens":  output.PromptTokenCount,
					"output_tokens": output.GenerationTokenCount,
				},
			},
		},
	}, nil
}
