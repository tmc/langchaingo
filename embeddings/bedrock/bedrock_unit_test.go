package bedrock

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/stretchr/testify/assert"
)

func TestNewBedrock(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr bool
	}{
		{
			name: "with model option",
			opts: []Option{WithModel(ModelTitanEmbedG1)},
		},
		{
			name: "with batch size option",
			opts: []Option{WithBatchSize(256)},
		},
		{
			name: "with strip new lines option",
			opts: []Option{WithStripNewLines(false)},
		},
		{
			name: "with multiple options",
			opts: []Option{
				WithModel(ModelCohereEn),
				WithBatchSize(128),
				WithStripNewLines(true),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip if AWS credentials are not available
			t.Skip("Skipping test that requires AWS credentials")
		})
	}
}

func TestBedrockOptions(t *testing.T) {
	t.Run("WithModel", func(t *testing.T) {
		b := &Bedrock{}
		WithModel(ModelTitanEmbedG1)(b)
		assert.Equal(t, ModelTitanEmbedG1, b.ModelID)
	})

	t.Run("WithBatchSize", func(t *testing.T) {
		b := &Bedrock{}
		WithBatchSize(256)(b)
		assert.Equal(t, 256, b.BatchSize)
	})

	t.Run("WithStripNewLines", func(t *testing.T) {
		b := &Bedrock{}
		WithStripNewLines(false)(b)
		assert.Equal(t, false, b.StripNewLines)
	})

	t.Run("WithClient", func(t *testing.T) {
		b := &Bedrock{}
		client := &bedrockruntime.Client{}
		WithClient(client)(b)
		assert.Equal(t, client, b.client)
	})
}

func TestGetProvider(t *testing.T) {
	tests := []struct {
		name     string
		modelID  string
		expected string
	}{
		{
			name:     "amazon titan model",
			modelID:  ModelTitanEmbedG1,
			expected: "amazon",
		},
		{
			name:     "amazon titan text v1",
			modelID:  ModelTitanEmbedG1,
			expected: "amazon",
		},
		{
			name:     "cohere english model",
			modelID:  ModelCohereEn,
			expected: "cohere",
		},
		{
			name:     "cohere multilingual model",
			modelID:  ModelCohereMulti,
			expected: "cohere",
		},
		{
			name:     "unknown model",
			modelID:  "unknown-model",
			expected: "unknown-model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getProvider(tt.modelID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultValues(t *testing.T) {
	assert.Equal(t, 512, _defaultBatchSize)
	assert.Equal(t, true, _defaultStripNewLines)
	assert.Equal(t, ModelTitanEmbedG1, _defaultModel)
}
