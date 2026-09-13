package anthropicclient

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseStreamingMessageResponse_withEmptyInput(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	response := createSSEResponse(SSEDataWithEmptyInput)
	defer response.Body.Close()
	payload := &messagePayload{}

	result, err := parseStreamingMessageResponse(ctx, response, payload)

	// Verify results
	require.NoError(t, err, "Parsing should complete without errors")
	require.NotNil(t, result, "Result should not be nil")

	// Additional assertions could verify specific content parsed from the SSE stream
	require.Equal(t, "msg_01KpsxABJ1CZwpfVuT6XFz7T", result.ID, "Message ID should match expected value")
	require.Equal(t, "claude-3-7-sonnet-latest", result.Model, "Model should match expected value")
	require.Equal(t, "assistant", result.Role, "Role should be 'assistant'")
	require.Len(t, result.Content, 2, "Content should contain two blocks")

	firstContent, ok := result.Content[0].(*TextContent)
	require.True(t, ok, "First content block should be of type TextContent")
	require.Equal(t, "I can help you find your current IP address. Let me retrieve that information for you.", firstContent.Text, "First content block text should match expected value")

	secondContent, ok := result.Content[1].(*ToolUseContent)
	require.True(t, ok, "Second content block should be of type ToolUseContent")
	require.Equal(t, "get_current_ip_address", secondContent.Name, "Tool use name should match expected value")
	require.Empty(t, secondContent.Input, "Tool use input should be empty")
}

func TestCreateMessageReasoningOnlyUsesStreamDecoder(t *testing.T) {
	events := []string{
		`{"type":"message_start","message":{"id":"m","type":"message","role":"assistant","content":[],"model":"claude-test","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":1,"output_tokens":0}}}`,
		`{"type":"content_block_start","index":0,"content_block":{"type":"thinking","thinking":""}}`,
		`{"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"consider"}}`,
		`{"type":"content_block_stop","index":0}`,
		`{"type":"content_block_start","index":1,"content_block":{"type":"text","text":""}}`,
		`{"type":"content_block_delta","index":1,"delta":{"type":"text_delta","text":"answer"}}`,
		`{"type":"content_block_stop","index":1}`,
		`{"type":"message_stop"}`,
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		for _, e := range events {
			_, _ = io.WriteString(w, "data: "+e+"\n\n")
		}
	}))
	defer server.Close()
	client, err := New("test-key", "claude-test", server.URL)
	require.NoError(t, err)

	var reasoning []string
	resp, err := client.createMessage(context.Background(), &messagePayload{
		Model: "claude-test",
		StreamingReasoningFunc: func(_ context.Context, reasoningChunk, _ []byte) error {
			reasoning = append(reasoning, string(reasoningChunk))
			return nil
		},
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"consider"}, reasoning)
	require.Len(t, resp.Content, 2)
	thinking, ok := resp.Content[0].(*ThinkingContent)
	require.True(t, ok, "content[0] is %T", resp.Content[0])
	assert.Equal(t, "consider", thinking.Thinking)
	text, ok := resp.Content[1].(*TextContent)
	require.True(t, ok, "content[1] is %T", resp.Content[1])
	assert.Equal(t, "answer", text.Text)
}

func Test_parseStreamingMessageResponse_withInputJSONDeltas(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	response := createSSEResponse(SSEDataWithInputJSONDeltas)
	defer response.Body.Close()
	payload := &messagePayload{}

	result, err := parseStreamingMessageResponse(ctx, response, payload)

	// Verify results
	require.NoError(t, err, "Parsing should complete without errors")
	require.NotNil(t, result, "Result should not be nil")

	// Additional assertions could verify specific content parsed from the SSE stream
	require.Equal(t, "msg_01QdDq6hdDLd5v9fndWvs43Z", result.ID, "Message ID should match expected value")
	require.Equal(t, "claude-3-7-sonnet-latest", result.Model, "Model should match expected value")
	require.Equal(t, "assistant", result.Role, "Role should be 'assistant'")
	require.Len(t, result.Content, 2, "Content should contain two blocks")

	firstContent, ok := result.Content[0].(*TextContent)
	require.True(t, ok, "First content block should be of type TextContent")
	require.Equal(t, "I can help you get the current time. Let me check that for you.", firstContent.Text, "First content block text should match expected value")

	secondContent, ok := result.Content[1].(*ToolUseContent)
	require.True(t, ok, "Second content block should be of type ToolUseContent")
	require.Equal(t, "get_current_time", secondContent.Name, "Tool use name should match expected value")
	require.Equal(t, map[string]interface{}{
		"format": "2006-01-02 15:04:05",
	}, secondContent.Input, "Tool use input should match expected value")
}

// createAnthropicSSEResponse creates an HTTP response containing a simulated
// Anthropic API server-sent events (SSE) stream.
func createSSEResponse(data string) *http.Response {
	recorder := httptest.NewRecorder()
	recorder.Header().Set("Content-Type", "application/json")
	recorder.WriteHeader(http.StatusOK)
	if _, err := recorder.WriteString(data); err != nil {
		panic(err)
	}

	return recorder.Result()
}

const SSEDataWithEmptyInput = `event: message_start
data: {"type":"message_start","message":{"id":"msg_01KpsxABJ1CZwpfVuT6XFz7T","type":"message","role":"assistant","model":"claude-3-7-sonnet-latest","content":[],"stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":417,"cache_creation_input_tokens":0,"cache_read_input_tokens":0,"output_tokens":2}}        }

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: ping
data: {"type": "ping"}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"I can"}   }

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" help you find your current IP address. Let me retrieve"}   }

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" that information for you."}}

event: content_block_stop
data: {"type":"content_block_stop","index":0        }

event: content_block_start
data: {"type":"content_block_start","index":1,"content_block":{"type":"tool_use","id":"toolu_01Lz8gVHwSEMLBTTDbTqGcia","name":"get_current_ip_address","input":{}}           }

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":""}           }

event: content_block_stop
data: {"type":"content_block_stop","index":1          }

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"tool_use","stop_sequence":null},"usage":{"output_tokens":59}   }

event: message_stop
data: {"type":"message_stop"            }`

const SSEDataWithInputJSONDeltas = `event: message_start
data: {"type":"message_start","message":{"id":"msg_01QdDq6hdDLd5v9fndWvs43Z","type":"message","role":"assistant","model":"claude-3-7-sonnet-latest","content":[],"stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":463,"cache_creation_input_tokens":0,"cache_read_input_tokens":0,"output_tokens":2}}    }

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}        }

event: ping
data: {"type": "ping"}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"I can"}      }

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" help you get the current time. Let"}        }

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" me check that for you."}        }

event: content_block_stop
data: {"type":"content_block_stop","index":0   }

event: content_block_start
data: {"type":"content_block_start","index":1,"content_block":{"type":"tool_use","id":"toolu_01HSrVQU8QDxAsVwuAdbja45","name":"get_current_time","input":{}}             }

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":""}    }

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":"{\"for"}      }

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":"mat\": \"20"}          }

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":"06-01-0"}  }

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":"2 15:04:"}          }

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":"05\"}"}           }

event: content_block_stop
data: {"type":"content_block_stop","index":1          }

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"tool_use","stop_sequence":null},"usage":{"output_tokens":83}            }

event: message_stop
data: {"type":"message_stop"           }`

func Test_messagePayload_TemperatureSerialization(t *testing.T) {
	base := messagePayload{
		Model: "claude-fable-5",
		Messages: []ChatMessage{
			{Role: "user", Content: "hi"},
		},
		MaxTokens: 100,
	}

	t.Run("nil temperature is omitted", func(t *testing.T) {
		data, err := json.Marshal(base)
		require.NoError(t, err)
		assert.NotContains(t, string(data), "temperature")
	})

	t.Run("explicit zero temperature is sent", func(t *testing.T) {
		payload := base
		temperature := 0.0
		payload.Temperature = &temperature
		data, err := json.Marshal(payload)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"temperature":0`)
	})

	t.Run("adaptive thinking omits budget_tokens", func(t *testing.T) {
		payload := base
		payload.Thinking = &ThinkingConfig{Type: "adaptive"}
		data, err := json.Marshal(payload)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"thinking":{"type":"adaptive"}`)
		assert.NotContains(t, string(data), "budget_tokens")
	})
}

func Test_messagePayload_SystemSerialization(t *testing.T) {
	base := messagePayload{
		Model: "claude-fable-5",
		Messages: []ChatMessage{
			{Role: "user", Content: "hi"},
		},
		MaxTokens: 100,
	}
	c := &Client{}

	t.Run("empty system is omitted", func(t *testing.T) {
		payload := base
		payload.System = ""
		c.setMessageDefaults(&payload)
		data, err := json.Marshal(payload)
		require.NoError(t, err)
		assert.NotContains(t, string(data), "system")
	})

	t.Run("plain string keeps string encoding", func(t *testing.T) {
		payload := base
		payload.System = "You are helpful"
		c.setMessageDefaults(&payload)
		data, err := json.Marshal(payload)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"system":"You are helpful"`)
	})

	t.Run("blocks with cache control and ttl", func(t *testing.T) {
		payload := base
		payload.System = []TextContent{
			{Type: "text", Text: "Big prefix", CacheControl: &CacheControl{Type: "ephemeral", TTL: "1h"}},
		}
		c.setMessageDefaults(&payload)
		data, err := json.Marshal(payload)
		require.NoError(t, err)
		assert.Contains(t, string(data),
			`"system":[{"type":"text","text":"Big prefix","cache_control":{"type":"ephemeral","ttl":"1h"}}]`)
	})
}
