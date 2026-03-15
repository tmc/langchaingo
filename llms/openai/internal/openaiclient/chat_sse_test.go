package openaiclient

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

func TestParseStreamingChatResponse_SSEComments(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	// Test the key SSE comment patterns
	testCases := []struct {
		name            string
		body            string
		expectedContent string
	}{
		{
			name: "openrouter_comments",
			body: `data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}
: OPENROUTER PROCESSING
: OPENROUTER PROCESSING
data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{"content":" World"},"finish_reason":null}]}
data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}
data: [DONE]`,
			expectedContent: "Hello World",
		},
		{
			name: "comments_without_space",
			body: `data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{"content":"Test"},"finish_reason":null}]}
:comment-without-space
data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}
data: [DONE]`,
			expectedContent: "Test",
		},
		{
			name: "other_sse_fields",
			body: `event: message
id: 12345
data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{"content":"Data"},"finish_reason":null}]}
retry: 1000
data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}
data: [DONE]`,
			expectedContent: "Data",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(tc.body)),
			}

			req := &ChatRequest{
				StreamingFunc: func(_ context.Context, _ streaming.Chunk) error {
					return nil
				},
			}

			resp, err := parseStreamingChatResponse(ctx, r, req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp == nil {
				t.Fatal("response should not be nil")
			}
			if len(resp.Choices) == 0 {
				t.Fatal("expected at least one choice")
			}
			if got := resp.Choices[0].Message.Content; got != tc.expectedContent {
				t.Errorf("content mismatch: got %q, want %q", got, tc.expectedContent)
			}
		})
	}
}

func TestParseStreamingChatResponse_ToolCallNameCaching(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	testCases := []struct {
		name                 string
		body                 string
		expectedToolCallName string
		expectedArguments    string
		checkStreaming       bool
	}{
		{
			name: "gpt4_style_multiple_chunks_without_name",
			body: `data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}
data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{"tool_calls":[{"id":"call_123","index":0,"type":"function","function":{"name":"getCurrentWeather","arguments":""}}]},"finish_reason":null}]}
data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"type":"function","function":{"arguments":"{"}}]},"finish_reason":null}]}
data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"type":"function","function":{"arguments":"\"location\":"}}]},"finish_reason":null}]}
data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"type":"function","function":{"arguments":"\"Boston\"}"}}]},"finish_reason":null}]}
data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}]}
data: [DONE]`,
			expectedToolCallName: "getCurrentWeather",
			expectedArguments:    `{"location":"Boston"}`,
			checkStreaming:       true,
		},
		{
			name: "gemini_style_single_chunk_with_name",
			body: `data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"index":0,"id":"tool_getCurrentWeather_123","type":"function","function":{"name":"getCurrentWeather","arguments":"{\"location\":\"Boston\"}"}}]},"finish_reason":"tool_calls"}]}
data: [DONE]`,
			expectedToolCallName: "getCurrentWeather",
			expectedArguments:    `{"location":"Boston"}`,
			checkStreaming:       false,
		},
		{
			name: "parallel_tool_calls_without_names_in_subsequent_chunks",
			body: `data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}
data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{"tool_calls":[{"id":"call_1","index":0,"type":"function","function":{"name":"getCurrentWeather","arguments":""}}]},"finish_reason":null}]}
data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"type":"function","function":{"arguments":"{"}}]},"finish_reason":null}]}
data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"type":"function","function":{"arguments":"\"location\":\"Boston\"}"}}]},"finish_reason":null}]}
data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{"tool_calls":[{"id":"call_2","index":1,"type":"function","function":{"name":"getCurrentTime","arguments":""}}]},"finish_reason":null}]}
data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{"tool_calls":[{"index":1,"type":"function","function":{"arguments":"{"}}]},"finish_reason":null}]}
data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{"tool_calls":[{"index":1,"type":"function","function":{"arguments":"\"location\":\"Boston\"}"}}]},"finish_reason":null}]}
data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}]}
data: [DONE]`,
			expectedToolCallName: "getCurrentWeather", // Check first tool call
			expectedArguments:    `{"location":"Boston"}`,
			checkStreaming:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var streamedToolCalls []streaming.ToolCall

			r := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(tc.body)),
			}

			req := &ChatRequest{
				StreamingFunc: func(_ context.Context, chunk streaming.Chunk) error {
					if chunk.Type == streaming.ChunkTypeToolCall {
						streamedToolCalls = append(streamedToolCalls, chunk.ToolCall)
					}
					return nil
				},
			}

			resp, err := parseStreamingChatResponse(ctx, r, req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Greater(t, len(resp.Choices), 0)
			require.Greater(t, len(resp.Choices[0].Message.ToolCalls), 0)

			// Check final accumulated tool call
			toolCall := resp.Choices[0].Message.ToolCalls[0]
			assert.Equal(t, tc.expectedToolCallName, toolCall.Function.Name)
			assert.Equal(t, tc.expectedArguments, toolCall.Function.Arguments)

			// Check streaming callbacks if needed
			if tc.checkStreaming {
				require.Greater(t, len(streamedToolCalls), 0, "should have streamed tool calls")
				// All streamed tool calls should have the name filled in
				for i, streamedTC := range streamedToolCalls {
					assert.NotEmpty(t, streamedTC.Name, "streamed tool call %d should have name", i)
				}
			}
		})
	}
}

func TestUpdateToolCall_NameCaching(t *testing.T) {
	t.Parallel()

	t.Run("cache_name_on_first_chunk", func(t *testing.T) {
		t.Parallel()

		message := &ChatMessage{}
		nameCache := make(map[string]string)
		index := 0

		delta := &StreamedToolCall{
			ID:    "call_123",
			Type:  "function",
			Index: &index,
			Function: ToolFunction{
				Name:      "testFunction",
				Arguments: `{"arg":"value"}`,
			},
		}

		updateToolCall(message, delta, nameCache)

		assert.Equal(t, 1, len(message.ToolCalls))
		assert.Equal(t, "testFunction", message.ToolCalls[0].Function.Name)
		assert.Equal(t, "testFunction", nameCache["call_123"])
	})

	t.Run("restore_name_from_cache_on_subsequent_chunks", func(t *testing.T) {
		t.Parallel()

		message := &ChatMessage{
			ToolCalls: []ToolCall{
				{
					ID:   "call_123",
					Type: "function",
					Function: ToolFunction{
						Name:      "testFunction",
						Arguments: `{"arg":`,
					},
				},
			},
		}
		nameCache := map[string]string{
			"call_123": "testFunction",
		}
		index := 0

		// Subsequent chunk without ID and name (like GPT-4.1 style)
		delta := &StreamedToolCall{
			ID:    "", // No ID in subsequent chunk
			Type:  "function",
			Index: &index,
			Function: ToolFunction{
				Name:      "", // No name in subsequent chunk
				Arguments: `"value"}`,
			},
		}

		updateToolCall(message, delta, nameCache)

		// Name should be restored from the message's tool call
		assert.Equal(t, "testFunction", delta.Function.Name)
		assert.Equal(t, "call_123", delta.ID)
		assert.Equal(t, "testFunction", message.ToolCalls[0].Function.Name)
		assert.Equal(t, `{"arg":"value"}`, message.ToolCalls[0].Function.Arguments)
	})

	t.Run("handle_multiple_tool_calls_with_cache", func(t *testing.T) {
		t.Parallel()

		message := &ChatMessage{}
		nameCache := make(map[string]string)

		// First tool call - first chunk
		index0 := 0
		delta1 := &StreamedToolCall{
			ID:    "call_1",
			Type:  "function",
			Index: &index0,
			Function: ToolFunction{
				Name:      "function1",
				Arguments: "",
			},
		}
		updateToolCall(message, delta1, nameCache)

		// Second tool call - first chunk
		index1 := 1
		delta2 := &StreamedToolCall{
			ID:    "call_2",
			Type:  "function",
			Index: &index1,
			Function: ToolFunction{
				Name:      "function2",
				Arguments: "",
			},
		}
		updateToolCall(message, delta2, nameCache)

		assert.Equal(t, 2, len(nameCache))
		assert.Equal(t, "function1", nameCache["call_1"])
		assert.Equal(t, "function2", nameCache["call_2"])

		// First tool call - subsequent chunk without ID and name
		delta3 := &StreamedToolCall{
			ID:    "", // No ID in subsequent chunk
			Type:  "function",
			Index: &index0,
			Function: ToolFunction{
				Name:      "",
				Arguments: `{"a":"b"}`,
			},
		}
		updateToolCall(message, delta3, nameCache)

		assert.Equal(t, "function1", delta3.Function.Name)
		assert.Equal(t, "call_1", delta3.ID)
		assert.Equal(t, `{"a":"b"}`, message.ToolCalls[0].Function.Arguments)
	})

	t.Run("gpt4_style_streaming_multiple_chunks", func(t *testing.T) {
		t.Parallel()

		message := &ChatMessage{}
		nameCache := make(map[string]string)

		// First chunk: full metadata
		index0 := 0
		delta1 := &StreamedToolCall{
			ID:    "call_123",
			Type:  "function",
			Index: &index0,
			Function: ToolFunction{
				Name:      "getCurrentWeather",
				Arguments: "",
			},
		}
		updateToolCall(message, delta1, nameCache)
		assert.Equal(t, "getCurrentWeather", nameCache["call_123"])
		assert.Equal(t, "", message.ToolCalls[0].Function.Arguments)

		// Subsequent chunks: only index, type, arguments (no ID, no name)
		delta2 := &StreamedToolCall{
			ID:    "",
			Type:  "function",
			Index: &index0,
			Function: ToolFunction{
				Name:      "",
				Arguments: `{`,
			},
		}
		updateToolCall(message, delta2, nameCache)
		assert.Equal(t, "getCurrentWeather", delta2.Function.Name) // Restored from message
		assert.Equal(t, `{`, message.ToolCalls[0].Function.Arguments)

		delta3 := &StreamedToolCall{
			ID:    "",
			Index: &index0,
			Function: ToolFunction{
				Arguments: `"location":"Boston"}`,
			},
		}
		updateToolCall(message, delta3, nameCache)
		assert.Equal(t, `{"location":"Boston"}`, message.ToolCalls[0].Function.Arguments)
	})

	t.Run("gemini_style_single_complete_chunk", func(t *testing.T) {
		t.Parallel()

		message := &ChatMessage{}
		nameCache := make(map[string]string)
		index0 := 0

		// Single complete chunk with all data
		delta := &StreamedToolCall{
			ID:    "tool_getCurrentWeather_123",
			Type:  "function",
			Index: &index0,
			Function: ToolFunction{
				Name:      "getCurrentWeather",
				Arguments: `{"location":"Boston, MA"}`,
			},
		}
		updateToolCall(message, delta, nameCache)

		assert.Equal(t, 1, len(message.ToolCalls))
		assert.Equal(t, "getCurrentWeather", message.ToolCalls[0].Function.Name)
		assert.Equal(t, `{"location":"Boston, MA"}`, message.ToolCalls[0].Function.Arguments)
		assert.Equal(t, "getCurrentWeather", nameCache["tool_getCurrentWeather_123"])
	})

	t.Run("qwen_style_single_complete_chunk", func(t *testing.T) {
		t.Parallel()

		message := &ChatMessage{}
		nameCache := make(map[string]string)
		index0 := 0

		delta := &StreamedToolCall{
			ID:    "34fef3aa6",
			Type:  "function",
			Index: &index0,
			Function: ToolFunction{
				Name:      "getCurrentWeather",
				Arguments: `{"location": "Boston, MA", "unit": "celsius"}`,
			},
		}
		updateToolCall(message, delta, nameCache)

		assert.Equal(t, 1, len(message.ToolCalls))
		assert.Equal(t, "getCurrentWeather", message.ToolCalls[0].Function.Name)
		assert.Equal(t, `{"location": "Boston, MA", "unit": "celsius"}`, message.ToolCalls[0].Function.Arguments)
	})

	t.Run("claude_style_streaming_multiple_chunks", func(t *testing.T) {
		t.Parallel()

		message := &ChatMessage{}
		nameCache := make(map[string]string)
		index0 := 0

		// First chunk
		delta1 := &StreamedToolCall{
			ID:    "toolu_xxx",
			Type:  "function",
			Index: &index0,
			Function: ToolFunction{
				Name:      "getCurrentWeather",
				Arguments: "",
			},
		}
		updateToolCall(message, delta1, nameCache)

		// Subsequent chunks without ID and name
		delta2 := &StreamedToolCall{
			Index: &index0,
			Type:  "function",
			Function: ToolFunction{
				Arguments: `{`,
			},
		}
		updateToolCall(message, delta2, nameCache)

		delta3 := &StreamedToolCall{
			Index: &index0,
			Type:  "function",
			Function: ToolFunction{
				Arguments: `"location":"Boston, MA"}`,
			},
		}
		updateToolCall(message, delta3, nameCache)

		assert.Equal(t, `{"location":"Boston, MA"}`, message.ToolCalls[0].Function.Arguments)
		assert.Equal(t, "getCurrentWeather", message.ToolCalls[0].Function.Name)
	})

	t.Run("edge_case_id_without_name_and_no_cache", func(t *testing.T) {
		t.Parallel()

		message := &ChatMessage{}
		nameCache := make(map[string]string)
		index0 := 0

		// Hypothetical broken chunk: has ID but no name, and not in cache
		delta := &StreamedToolCall{
			ID:    "unknown_call",
			Type:  "function",
			Index: &index0,
			Function: ToolFunction{
				Name:      "",
				Arguments: `{"arg":"value"}`,
			},
		}
		updateToolCall(message, delta, nameCache)

		// Should still update arguments even without cache hit
		assert.Equal(t, `{"arg":"value"}`, message.ToolCalls[0].Function.Arguments)
		// Name will be empty since not in cache
		assert.Equal(t, "", delta.Function.Name)
	})

	t.Run("parallel_tools_mixed_formats", func(t *testing.T) {
		t.Parallel()

		message := &ChatMessage{}
		nameCache := make(map[string]string)

		// Tool 1: GPT-4 style (ID in first chunk, then no ID)
		index0 := 0
		delta1 := &StreamedToolCall{
			ID:    "call_1",
			Type:  "function",
			Index: &index0,
			Function: ToolFunction{
				Name:      "function1",
				Arguments: "",
			},
		}
		updateToolCall(message, delta1, nameCache)

		// Tool 2: starts before Tool 1 finishes
		index1 := 1
		delta2 := &StreamedToolCall{
			ID:    "call_2",
			Type:  "function",
			Index: &index1,
			Function: ToolFunction{
				Name:      "function2",
				Arguments: "",
			},
		}
		updateToolCall(message, delta2, nameCache)

		// Tool 1 continues (no ID)
		delta3 := &StreamedToolCall{
			Index: &index0,
			Type:  "function",
			Function: ToolFunction{
				Arguments: `{"a":"1"}`,
			},
		}
		updateToolCall(message, delta3, nameCache)

		// Tool 2 continues (no ID)
		delta4 := &StreamedToolCall{
			Index: &index1,
			Type:  "function",
			Function: ToolFunction{
				Arguments: `{"b":"2"}`,
			},
		}
		updateToolCall(message, delta4, nameCache)

		assert.Equal(t, 2, len(message.ToolCalls))
		assert.Equal(t, "function1", message.ToolCalls[0].Function.Name)
		assert.Equal(t, `{"a":"1"}`, message.ToolCalls[0].Function.Arguments)
		assert.Equal(t, "function2", message.ToolCalls[1].Function.Name)
		assert.Equal(t, `{"b":"2"}`, message.ToolCalls[1].Function.Arguments)
	})
}
