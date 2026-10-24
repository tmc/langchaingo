package bedrockclient

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

func TestInvokeModelInputs(t *testing.T) {
	t.Run("without safety config", func(t *testing.T) {
		input, err := newInvokeModelInput("anthropic.claude-v2", []byte(`{"prompt":"hello"}`), llms.CallOptions{})
		require.NoError(t, err)
		require.NotNil(t, input)
		assert.Nil(t, input.GuardrailIdentifier)
		assert.Nil(t, input.GuardrailVersion)

		streamInput, err := newInvokeModelWithResponseStreamInput("anthropic.claude-v2", []byte(`{"prompt":"hello"}`), llms.CallOptions{})
		require.NoError(t, err)
		require.NotNil(t, streamInput)
		assert.Nil(t, streamInput.GuardrailIdentifier)
		assert.Nil(t, streamInput.GuardrailVersion)
	})

	t.Run("with safety config", func(t *testing.T) {
		var options llms.CallOptions
		llms.WithSafetyConfig(map[string]any{
			"identifier": "gr-123",
			"version":    "1",
		})(&options)

		input, err := newInvokeModelInput("anthropic.claude-v2", []byte(`{"prompt":"hello"}`), options)
		require.NoError(t, err)
		require.NotNil(t, input.GuardrailIdentifier)
		require.NotNil(t, input.GuardrailVersion)
		assert.Equal(t, "gr-123", *input.GuardrailIdentifier)
		assert.Equal(t, "1", *input.GuardrailVersion)

		streamInput, err := newInvokeModelWithResponseStreamInput("anthropic.claude-v2", []byte(`{"prompt":"hello"}`), options)
		require.NoError(t, err)
		require.NotNil(t, streamInput.GuardrailIdentifier)
		require.NotNil(t, streamInput.GuardrailVersion)
		assert.Equal(t, "gr-123", *streamInput.GuardrailIdentifier)
		assert.Equal(t, "1", *streamInput.GuardrailVersion)
	})

	t.Run("requires safety version", func(t *testing.T) {
		_, err := newInvokeModelInput("anthropic.claude-v2", nil, llms.CallOptions{
			SafetyConfig: map[string]any{"identifier": "gr-123"},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "bedrock safety version")
	})

	t.Run("requires safety identifier", func(t *testing.T) {
		_, err := newInvokeModelInput("anthropic.claude-v2", nil, llms.CallOptions{
			SafetyConfig: map[string]any{"version": "1"},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "bedrock safety identifier")
	})

	t.Run("rejects invalid safety types", func(t *testing.T) {
		_, err := newInvokeModelInput("anthropic.claude-v2", nil, llms.CallOptions{
			SafetyConfig: map[string]any{
				"identifier": 123,
				"version":    "1",
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "bedrock safety identifier")
	})

	t.Run("returns content filter error when guardrail intervenes", func(t *testing.T) {
		resp, err := invokeModel(context.Background(), &mockBedrockClient{
			invokeFunc: func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
				return &bedrockruntime.InvokeModelOutput{
					Body: []byte(`{"amazon-bedrock-guardrailAction":"INTERVENED","amazon-bedrock-trace":{"rule":"blocked"}}`),
				}, nil
			},
		}, "anthropic.claude-v2", []byte(`{"prompt":"hello"}`), llms.CallOptions{})

		require.Nil(t, resp)
		require.Error(t, err)
		assert.True(t, llms.IsContentFilterError(err))

		var llmErr *llms.Error
		require.ErrorAs(t, err, &llmErr)
		assert.Equal(t, "bedrock", llmErr.Provider)
		assert.Equal(t, "INTERVENED", llmErr.Details["guardrail_action"])
		assert.Equal(t, map[string]any{"rule": "blocked"}, llmErr.Details["trace"])
	})

	t.Run("returns content filter error for provider filtered response", func(t *testing.T) {
		resp, err := invokeModel(context.Background(), &mockBedrockClient{
			invokeFunc: func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
				return &bedrockruntime.InvokeModelOutput{
					Body: []byte(`{"results":[{"completionReason":"CONTENT_FILTERED"}]}`),
				}, nil
			},
		}, "amazon.titan-text-express-v1", []byte(`{"prompt":"hello"}`), llms.CallOptions{})

		require.Nil(t, resp)
		require.Error(t, err)
		assert.True(t, llms.IsContentFilterError(err))
	})

	t.Run("returns response when guardrail does not intervene", func(t *testing.T) {
		resp, err := invokeModel(context.Background(), &mockBedrockClient{
			invokeFunc: func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
				return &bedrockruntime.InvokeModelOutput{
					Body: []byte(`{"amazon-bedrock-guardrailAction":"NONE","outputText":"ok"}`),
				}, nil
			},
		}, "anthropic.claude-v2", []byte(`{"prompt":"hello"}`), llms.CallOptions{})

		require.NoError(t, err)
		require.NotNil(t, resp)
	})
}
