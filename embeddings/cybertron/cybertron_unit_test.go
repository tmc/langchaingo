package cybertron

import (
	"testing"

	"github.com/nlpodyssey/cybertron/pkg/models/bert"
	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	t.Run("WithModel", func(t *testing.T) {
		c := &Cybertron{}
		modelName := "test-model"
		WithModel(modelName)(c)
		assert.Equal(t, modelName, c.Model)
	})

	t.Run("WithModelsDir", func(t *testing.T) {
		c := &Cybertron{}
		modelsDir := "/test/models"
		WithModelsDir(modelsDir)(c)
		assert.Equal(t, modelsDir, c.ModelsDir)
	})

	t.Run("WithPoolingStrategy", func(t *testing.T) {
		c := &Cybertron{}
		strategy := bert.ClsTokenPooling
		WithPoolingStrategy(strategy)(c)
		assert.Equal(t, strategy, c.PoolingStrategy)
	})
}

func TestDefaultValues(t *testing.T) {
	assert.Equal(t, "models", _defaultModelsDir)
	assert.Equal(t, "sentence-transformers/all-MiniLM-L6-v2", _defaultModel)
	assert.Equal(t, bert.MeanPooling, _defaultPoolingStrategy)
}

func TestApplyOptions(t *testing.T) {
	// Test that applyOptions is skipped if it requires downloading models
	t.Skip("Skipping test that requires downloading models")
}
