package anthropicclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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
