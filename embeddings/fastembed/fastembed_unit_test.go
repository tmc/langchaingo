package fastembed

import (
	"context"
	"testing"

	fastembed "github.com/anush008/fastembed-go"
)

func TestNewFastEmbed(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr bool
	}{
		{
			name: "default options",
			opts: []Option{},
		},
		{
			name: "with valid model",
			opts: []Option{WithModel(fastembed.BGESmallENV15)},
		},
		{
			name: "with custom batch size",
			opts: []Option{WithBatchSize(128)},
		},
		{
			name: "with custom max length",
			opts: []Option{WithMaxLength(256)},
		},
		{
			name: "with passage embed type",
			opts: []Option{WithDocEmbedType("passage")},
		},
		{
			name:    "invalid embed type",
			opts:    []Option{WithDocEmbedType("invalid")},
			wantErr: true,
		},
		{
			name:    "invalid batch size",
			opts:    []Option{WithBatchSize(0)},
			wantErr: true,
		},
		{
			name:    "invalid max length",
			opts:    []Option{WithMaxLength(-1)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			embedder, err := NewFastEmbed(tt.opts...)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewFastEmbed() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewFastEmbed() unexpected error: %v", err)
				return
			}

			if embedder == nil {
				t.Errorf("NewFastEmbed() returned nil embedder")
				return
			}

			defer func() {
				if err := embedder.Close(); err != nil {
					t.Errorf("Close() error: %v", err)
				}
			}()
		})
	}
}

func TestFastEmbed_validateOptions(t *testing.T) {
	tests := []struct {
		name     string
		embedder *FastEmbed
		wantErr  bool
	}{
		{
			name: "valid default options",
			embedder: &FastEmbed{
				Model:        defaultModel,
				MaxLength:    defaultMaxLength,
				BatchSize:    defaultBatchSize,
				DocEmbedType: defaultDocEmbedType,
			},
			wantErr: false,
		},
		{
			name: "valid passage type",
			embedder: &FastEmbed{
				Model:        defaultModel,
				MaxLength:    defaultMaxLength,
				BatchSize:    defaultBatchSize,
				DocEmbedType: "passage",
			},
			wantErr: false,
		},
		{
			name: "invalid embed type",
			embedder: &FastEmbed{
				Model:        defaultModel,
				MaxLength:    defaultMaxLength,
				BatchSize:    defaultBatchSize,
				DocEmbedType: "invalid",
			},
			wantErr: true,
		},
		{
			name: "zero batch size",
			embedder: &FastEmbed{
				Model:        defaultModel,
				MaxLength:    defaultMaxLength,
				BatchSize:    0,
				DocEmbedType: defaultDocEmbedType,
			},
			wantErr: true,
		},
		{
			name: "negative max length",
			embedder: &FastEmbed{
				Model:        defaultModel,
				MaxLength:    -1,
				BatchSize:    defaultBatchSize,
				DocEmbedType: defaultDocEmbedType,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.embedder.validateOptions()
			if tt.wantErr && err == nil {
				t.Errorf("validateOptions() expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateOptions() unexpected error: %v", err)
			}
		})
	}
}

func TestFastEmbed_EmbedDocuments_Empty(t *testing.T) {
	embedder := &FastEmbed{
		Model:        defaultModel,
		MaxLength:    defaultMaxLength,
		BatchSize:    defaultBatchSize,
		DocEmbedType: defaultDocEmbedType,
	}

	ctx := context.Background()
	embeddings, err := embedder.EmbedDocuments(ctx, []string{})
	if err != nil {
		t.Errorf("EmbedDocuments() with empty input unexpected error: %v", err)
	}

	if len(embeddings) != 0 {
		t.Errorf("EmbedDocuments() with empty input expected empty result, got %d embeddings", len(embeddings))
	}
}
