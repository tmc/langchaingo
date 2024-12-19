package llms

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockModel struct {
	Model
	stopSequences []string
	stopWords     []string
	content       string
}

func (m *mockModel) GenerateContent(_ context.Context, messages []MessageContent, options ...CallOption) (*ContentResponse, error) {
	opts := &CallOptions{}
	for _, opt := range options {
		opt(opts)
	}
	m.stopSequences = opts.StopSequences
	m.stopWords = opts.StopWords
	return &ContentResponse{
		Choices: []*ContentChoice{
			{
				Content: m.content,
			},
		},
	}, nil
}

func TestStopSequences(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		stopWords     []string
		stopSequences []string
		content       string
		want          []string
	}{
		{
			name:          "StopSequences takes precedence",
			stopWords:     []string{"stop1", "stop2"},
			stopSequences: []string{"seq1", "seq2"},
			content:       "test content",
			want:          []string{"seq1", "seq2"},
		},
		{
			name:          "Fallback to StopWords",
			stopWords:     []string{"stop1", "stop2"},
			stopSequences: nil,
			content:       "test content",
			want:          []string{"stop1", "stop2"},
		},
		{
			name:          "Empty StopWords and StopSequences",
			stopWords:     nil,
			stopSequences: nil,
			content:       "test content",
			want:          nil,
		},
		{
			name:          "Empty StopWords with StopSequences",
			stopWords:     nil,
			stopSequences: []string{"seq1"},
			content:       "test content",
			want:          []string{"seq1"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			model := &mockModel{content: tt.content}
			messages := []MessageContent{
				{
					Role: ChatMessageTypeHuman,
					Parts: []ContentPart{
						TextContent{Text: "test prompt"},
					},
				},
			}

			opts := []CallOption{
				WithStopWords(tt.stopWords),
				WithStopSequences(tt.stopSequences),
			}

			_, err := model.GenerateContent(context.Background(), messages, opts...)
			assert.NoError(t, err)

			if tt.stopSequences != nil {
				assert.Equal(t, tt.want, model.stopSequences)
			} else {
				assert.Equal(t, tt.want, model.stopWords)
			}
		})
	}
}

func TestStopSequencesStreaming(t *testing.T) {
	t.Parallel()
	model := &mockModel{content: "test content"}
	messages := []MessageContent{
		{
			Role: ChatMessageTypeHuman,
			Parts: []ContentPart{
				TextContent{Text: "test prompt"},
			},
		},
	}

	streamingContent := ""
	streamingFunc := func(_ context.Context, chunk []byte) error {
		streamingContent += string(chunk)
		return nil
	}

	stopSequences := []string{"seq1", "seq2"}
	opts := []CallOption{
		WithStopSequences(stopSequences),
		WithStreamingFunc(streamingFunc),
	}

	_, err := model.GenerateContent(context.Background(), messages, opts...)
	assert.NoError(t, err)
	assert.Equal(t, stopSequences, model.stopSequences)
}
