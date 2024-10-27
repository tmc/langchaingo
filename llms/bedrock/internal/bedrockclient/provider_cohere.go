package bedrockclient

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/starmvp/langchaingo/llms"
)

// Ref: https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters-cohere-command.html
// Also: https://docs.cohere.com/reference/generate

// cohereTextGenerationInput is the input for the text generation for Cohere Models.
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

func createCohereCompletion(ctx context.Context,
	client *bedrockruntime.Client,
	modelID string,
	messages []Message,
	options llms.CallOptions,
) (*llms.ContentResponse, error) {
	txt := processInputMessagesGeneric(messages)

	input := &cohereTextGenerationInput{
		Prompt:         txt,
		Temperature:    options.Temperature,
		P:              options.TopP,
		K:              options.TopK,
		MaxTokens:      getMaxTokens(options.MaxTokens, 20),
		StopSequences:  options.StopWords,
		NumGenerations: options.CandidateCount,
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
			GenerationInfo: map[string]interface{}{
				"generation_id": gen.ID,
				"index":         i,
			},
		}
	}

	return &llms.ContentResponse{
		Choices: choices,
	}, nil
}
