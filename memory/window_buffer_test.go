package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/schema"
)

func TestWindowBufferMemory(t *testing.T) {
	t.Parallel()

	m := NewConversationWindowBuffer(2)

	result1, err := m.LoadMemoryVariables(context.Background(), map[string]any{})
	require.NoError(t, err)
	expected1 := map[string]any{"history": ""}
	assert.Equal(t, expected1, result1)

	err = m.SaveContext(context.Background(), map[string]any{"foo": "bar1"}, map[string]any{"bar": "foo1"})
	require.NoError(t, err)

	result2, err := m.LoadMemoryVariables(context.Background(), map[string]any{})
	require.NoError(t, err)

	expected2 := map[string]any{"history": "Human: bar1\nAI: foo1"}
	assert.Equal(t, expected2, result2)

	err = m.SaveContext(context.Background(), map[string]any{"foo": "bar2"}, map[string]any{"bar": "foo2"})
	require.NoError(t, err)

	result3, err := m.LoadMemoryVariables(context.Background(), map[string]any{})
	require.NoError(t, err)

	expected3 := map[string]any{"history": "Human: bar1\nAI: foo1\nHuman: bar2\nAI: foo2"}
	assert.Equal(t, expected3, result3)

	err = m.SaveContext(context.Background(), map[string]any{"foo": "bar3"}, map[string]any{"bar": "foo3"})
	require.NoError(t, err)

	result4, err := m.LoadMemoryVariables(context.Background(), map[string]any{})
	require.NoError(t, err)

	expected4 := map[string]any{"history": "Human: bar2\nAI: foo2\nHuman: bar3\nAI: foo3"}
	assert.Equal(t, expected4, result4)
}

func TestWindowBufferMemoryReturnMessage(t *testing.T) {
	t.Parallel()
	m := NewConversationWindowBuffer(2, WithReturnMessages(true))

	err := m.SaveContext(context.Background(), map[string]any{"foo": "bar1"}, map[string]any{"bar": "foo1"})
	require.NoError(t, err)

	err = m.SaveContext(context.Background(), map[string]any{"foo": "bar2"}, map[string]any{"bar": "foo2"})
	require.NoError(t, err)

	err = m.SaveContext(context.Background(), map[string]any{"foo": "bar3"}, map[string]any{"bar": "foo3"})
	require.NoError(t, err)

	result, err := m.LoadMemoryVariables(context.Background(), map[string]any{})
	require.NoError(t, err)

	expectedChatHistory := NewChatMessageHistory(
		WithPreviousMessages([]schema.ChatMessage{
			schema.HumanChatMessage{Content: "bar2"},
			schema.AIChatMessage{Content: "foo2"},
			schema.HumanChatMessage{Content: "bar3"},
			schema.AIChatMessage{Content: "foo3"},
		}),
	)

	messages, err := expectedChatHistory.Messages(context.Background())
	require.NoError(t, err)
	expected := map[string]any{"history": messages}
	assert.Equal(t, expected, result)
}

func TestWindowBufferMemoryWithPreLoadedHistory(t *testing.T) {
	t.Parallel()

	m := NewConversationWindowBuffer(2, WithChatHistory(NewChatMessageHistory(
		WithPreviousMessages([]schema.ChatMessage{
			schema.HumanChatMessage{Content: "bar1"},
			schema.AIChatMessage{Content: "foo1"},
			schema.HumanChatMessage{Content: "bar2"},
			schema.AIChatMessage{Content: "foo2"},
			schema.HumanChatMessage{Content: "bar3"},
			schema.AIChatMessage{Content: "foo3"},
		}),
	)))

	result, err := m.LoadMemoryVariables(context.Background(), map[string]any{})
	require.NoError(t, err)
	expected := map[string]any{"history": "Human: bar2\nAI: foo2\nHuman: bar3\nAI: foo3"}
	assert.Equal(t, expected, result)
}

func TestConversationWindowBuffer_cutMessages(t *testing.T) { // nolint:funlen
	t.Parallel()
	type fields struct {
		ConversationBuffer     ConversationBuffer
		ConversationWindowSize int
	}
	type args struct {
		message []schema.ChatMessage
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantMessage []schema.ChatMessage
		isCut       bool
	}{
		{
			name: "empty messages, do not need cut",
			fields: fields{
				ConversationBuffer:     *NewConversationBuffer(),
				ConversationWindowSize: 1,
			},
			args: args{
				message: []schema.ChatMessage{},
			},
			wantMessage: []schema.ChatMessage{},
			isCut:       false,
		},
		{
			name: "message less than buffer size, do not need cut",
			fields: fields{
				ConversationBuffer:     *NewConversationBuffer(),
				ConversationWindowSize: 1,
			},
			args: args{
				message: []schema.ChatMessage{
					schema.HumanChatMessage{Content: "foo"},
					schema.AIChatMessage{Content: "bar"},
				},
			},
			wantMessage: []schema.ChatMessage{
				schema.HumanChatMessage{Content: "foo"},
				schema.AIChatMessage{Content: "bar"},
			},
			isCut: false,
		},
		{
			name: "add human message, will cut",
			fields: fields{
				ConversationBuffer:     *NewConversationBuffer(),
				ConversationWindowSize: 1,
			},
			args: args{
				message: []schema.ChatMessage{
					schema.HumanChatMessage{Content: "foo"},
					schema.AIChatMessage{Content: "bar"},
					schema.HumanChatMessage{Content: "foo1"},
				},
			},
			wantMessage: []schema.ChatMessage{
				schema.AIChatMessage{Content: "bar"},
				schema.HumanChatMessage{Content: "foo1"},
			},
			isCut: true,
		},
		{
			name: "message more than buffer size, will cut",
			fields: fields{
				ConversationBuffer:     *NewConversationBuffer(),
				ConversationWindowSize: 1,
			},
			args: args{
				message: []schema.ChatMessage{
					schema.HumanChatMessage{Content: "foo"},
					schema.AIChatMessage{Content: "bar"},
					schema.HumanChatMessage{Content: "foo1"},
					schema.AIChatMessage{Content: "bar1"},
				},
			},
			wantMessage: []schema.ChatMessage{
				schema.HumanChatMessage{Content: "foo1"},
				schema.AIChatMessage{Content: "bar1"},
			},
			isCut: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			wb := &ConversationWindowBuffer{
				ConversationBuffer:     tt.fields.ConversationBuffer,
				ConversationWindowSize: tt.fields.ConversationWindowSize,
			}
			cut, isCut := wb.cutMessages(tt.args.message)
			assert.Equalf(t, tt.wantMessage, cut, "cutMessages(%s), want:%v, get:%v", tt.name, tt.wantMessage, cut)
			assert.Equalf(t, tt.isCut, isCut, "cutMessages(%s), want:%t, get:%t", tt.name, tt.isCut, isCut)
		})
	}
}
