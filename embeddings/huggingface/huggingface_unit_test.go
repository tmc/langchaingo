package huggingface

import (
	"testing"

	"github.com/0xDezzy/langchaingo/llms/huggingface"
	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	t.Run("WithModel", func(t *testing.T) {
		h := &Huggingface{}
		model := "custom-model"
		WithModel(model)(h)
		assert.Equal(t, model, h.Model)
	})

	t.Run("WithTask", func(t *testing.T) {
		h := &Huggingface{}
		task := "text-classification"
		WithTask(task)(h)
		assert.Equal(t, task, h.Task)
	})

	t.Run("WithClient", func(t *testing.T) {
		h := &Huggingface{}
		hfClient := huggingface.LLM{}
		WithClient(hfClient)(h)
		assert.NotNil(t, h.client)
	})

	t.Run("WithStripNewLines", func(t *testing.T) {
		h := &Huggingface{}
		WithStripNewLines(false)(h)
		assert.Equal(t, false, h.StripNewLines)
	})

	t.Run("WithBatchSize", func(t *testing.T) {
		h := &Huggingface{}
		WithBatchSize(256)(h)
		assert.Equal(t, 256, h.BatchSize)
	})
}

func TestDefaultValues(t *testing.T) {
	assert.Equal(t, 512, _defaultBatchSize)
	assert.Equal(t, true, _defaultStripNewLines)
	assert.Equal(t, "BAAI/bge-small-en-v1.5", _defaultModel)
	assert.Equal(t, "feature-extraction", _defaultTask)
}

func TestNewHuggingface(t *testing.T) {
	// Skip tests that require API token
	t.Run("new with options", func(t *testing.T) {
		t.Skip("Skipping test that requires HuggingFace API token")
	})
}
