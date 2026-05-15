package bedrockclient

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/tmc/langchaingo/llms"
)

const (
	bedrockAcceptAll       = "*/*"
	bedrockJSONContentType = "application/json"
)

type guardrailResponse struct {
	AmazonBedrockGuardrailAction string          `json:"amazon-bedrock-guardrailAction"`
	AmazonBedrockTrace           json.RawMessage `json:"amazon-bedrock-trace"`
}

type streamingCompletionResponseChunk struct {
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
		OutputTokens             int `json:"output_tokens"`
		CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
		CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
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
			InputTokens              int `json:"input_tokens"`
			OutputTokens             int `json:"output_tokens"`
			CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
			CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
		} `json:"usage"`
	} `json:"message"`
	guardrailResponse
}

type invokeModelAPI interface {
	InvokeModel(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error)
}

type invokeModelWithResponseStreamAPI interface {
	InvokeModelWithResponseStream(ctx context.Context, params *bedrockruntime.InvokeModelWithResponseStreamInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelWithResponseStreamOutput, error)
}

type guardrailConfigurableAdapter interface {
	body() []byte
	setGuardrailIdentifier(*string)
	setGuardrailVersion(*string)
	setBody([]byte)
}

type guardrailConfigurableInput struct {
	*bedrockruntime.InvokeModelInput
}

var _ guardrailConfigurableAdapter = (*guardrailConfigurableInput)(nil)

func (i *guardrailConfigurableInput) setGuardrailIdentifier(identifier *string) {
	i.GuardrailIdentifier = identifier
}
func (i *guardrailConfigurableInput) setGuardrailVersion(version *string) {
	i.GuardrailVersion = version
}
func (i *guardrailConfigurableInput) body() []byte        { return i.Body }
func (i *guardrailConfigurableInput) setBody(body []byte) { i.Body = body }

type guardrailConfigurableStreamInput struct {
	*bedrockruntime.InvokeModelWithResponseStreamInput
}

var _ guardrailConfigurableAdapter = (*guardrailConfigurableStreamInput)(nil)

func (i *guardrailConfigurableStreamInput) setGuardrailIdentifier(identifier *string) {
	i.GuardrailIdentifier = identifier
}
func (i *guardrailConfigurableStreamInput) setGuardrailVersion(version *string) {
	i.GuardrailVersion = version
}
func (i *guardrailConfigurableStreamInput) body() []byte        { return i.Body }
func (i *guardrailConfigurableStreamInput) setBody(body []byte) { i.Body = body }

func newInvokeModelInput(modelID string, body []byte, options llms.CallOptions) (*bedrockruntime.InvokeModelInput, error) {
	input := &guardrailConfigurableInput{
		&bedrockruntime.InvokeModelInput{
			ModelId:     aws.String(modelID),
			Accept:      aws.String(bedrockAcceptAll),
			ContentType: aws.String(bedrockJSONContentType),
			Body:        body,
		},
	}
	if err := handleGuardrailParams(input, options); err != nil {
		return nil, err
	}
	return input.InvokeModelInput, nil
}

func newInvokeModelWithResponseStreamInput(modelID string, body []byte, options llms.CallOptions) (*bedrockruntime.InvokeModelWithResponseStreamInput, error) {
	input := &guardrailConfigurableStreamInput{
		&bedrockruntime.InvokeModelWithResponseStreamInput{
			ModelId:     aws.String(modelID),
			Accept:      aws.String(bedrockAcceptAll),
			ContentType: aws.String(bedrockJSONContentType),
			Body:        body,
		},
	}
	if err := handleGuardrailParams(input, options); err != nil {
		return nil, err
	}
	return input.InvokeModelWithResponseStreamInput, nil
}

func invokeModel(ctx context.Context, client invokeModelAPI, modelID string, body []byte, options llms.CallOptions) (*bedrockruntime.InvokeModelOutput, error) {
	input, err := newInvokeModelInput(modelID, body, options)
	if err != nil {
		return nil, err
	}

	output, err := client.InvokeModel(ctx, input)
	if err != nil {
		return nil, err
	}
	if err := checkGuardrailError(output.Body); err != nil {
		return nil, err
	}

	return output, nil
}

func invokeModelWithResponseStream(ctx context.Context, client invokeModelWithResponseStreamAPI, modelID string, body []byte, options llms.CallOptions) (*bedrockruntime.InvokeModelWithResponseStreamOutput, error) {
	input, err := newInvokeModelWithResponseStreamInput(modelID, body, options)
	if err != nil {
		return nil, err
	}
	return client.InvokeModelWithResponseStream(ctx, input)
}
