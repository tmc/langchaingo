package perplexity

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*testing.T) []Option
		wantErr bool
	}{
		{
			name: "no api key",
			setup: func(t *testing.T) []Option {
				t.Setenv("PERPLEXITY_API_KEY", "")
				return nil
			},
			wantErr: true,
		},
		{
			name: "with env api key",
			setup: func(t *testing.T) []Option {
				t.Setenv("PERPLEXITY_API_KEY", "test-key")
				return nil
			},
			wantErr: false,
		},
		{
			name: "with option api key",
			setup: func(t *testing.T) []Option {
				t.Setenv("PERPLEXITY_API_KEY", "")
				return []Option{WithAPIKey("test-key")}
			},
			wantErr: false,
		},
		{
			name: "with custom model",
			setup: func(t *testing.T) []Option {
				t.Setenv("PERPLEXITY_API_KEY", "test-key")
				return []Option{WithModel(ModelLlamaSonarLarge)}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			opts := tt.setup(t)
			
			tool, err := New(opts...)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, tool)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tool)
			}
		})
	}
}

func TestTool_Integration(t *testing.T) {
	if os.Getenv("PERPLEXITY_API_KEY") == "" {
		t.Skip("PERPLEXITY_API_KEY not set")
	}

	tool, err := New()
	require.NoError(t, err)
	require.NotNil(t, tool)

	// Test Name and Description
	assert.Equal(t, "PerplexityAI", tool.Name())
	assert.NotEmpty(t, tool.Description())

	// Test Call functionality
	ctx := context.Background()
	response, err := tool.Call(ctx, "what is the largest country in the world by total area?")
	require.NoError(t, err)
	assert.Contains(t, response, "Russia")
}
