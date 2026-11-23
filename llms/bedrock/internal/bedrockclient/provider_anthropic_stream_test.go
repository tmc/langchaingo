package bedrockclient

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/tmc/langchaingo/llms"
)

// anthropicToolStreamEvents is a real InvokeModelWithResponseStream reply from
// us.anthropic.claude-sonnet-4-6 to "Get the weather in Paris and in Tokyo",
// with a get_weather tool: a text block, then two tool_use blocks.
var anthropicToolStreamEvents = []string{
	`{"type":"message_start","message":{"model":"claude-sonnet-4-6","id":"msg_bdrk_pjonxmfm47k7dqhfckpgts53bisfbmt2twqfenw4375v4t36dd5a","type":"message","role":"assistant","content":[],"container":null,"stop_reason":null,"stop_sequence":null,"stop_details":null,"usage":{"input_tokens":576,"cache_creation_input_tokens":0,"cache_read_input_tokens":0,"cache_creation":{"ephemeral_5m_input_tokens":0,"ephemeral_1h_input_tokens":0},"output_tokens":1,"service_tier":"standard"}}}`,
	`{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
	`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Sure"}}`,
	`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"! Let me fetch"}}`,
	`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" the weather for both cities"}}`,
	`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" simultaneously"}}`,
	`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"!"}}`,
	`{"type":"content_block_stop","index":0}`,
	`{"type":"content_block_start","index":1,"content_block":{"type":"tool_use","id":"toolu_bdrk_01Pvfgh1kFecPX5cG8Qrkjgh","name":"get_weather","input":{}}}`,
	`{"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":""}}`,
	`{"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":"{\"city\": \"Pa"}}`,
	`{"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":"ris\"}"}}`,
	`{"type":"content_block_stop","index":1}`,
	`{"type":"content_block_start","index":2,"content_block":{"type":"tool_use","id":"toolu_bdrk_01R7w7Sw5sQDm5wrotaCkB7D","name":"get_weather","input":{}}}`,
	`{"type":"content_block_delta","index":2,"delta":{"type":"input_json_delta","partial_json":""}}`,
	`{"type":"content_block_delta","index":2,"delta":{"type":"input_json_delta","partial_json":"{\"city\": \""}}`,
	`{"type":"content_block_delta","index":2,"delta":{"type":"input_json_delta","partial_json":"Tokyo\"}"}}`,
	`{"type":"content_block_stop","index":2}`,
	`{"type":"message_delta","delta":{"stop_reason":"tool_use","stop_sequence":null,"stop_details":null,"container":null},"usage":{"input_tokens":576,"cache_creation_input_tokens":0,"cache_read_input_tokens":0,"output_tokens":102}}`,
	`{"type":"message_stop","amazon-bedrock-invocationMetrics":{"inputTokenCount":576,"outputTokenCount":102,"invocationLatency":2309,"firstByteLatency":700}}`,
}

type eventStreamTransport struct{ body []byte }

func (t eventStreamTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/vnd.amazon.eventstream"}},
		Body:       io.NopCloser(bytes.NewReader(t.body)),
		Request:    req,
	}, nil
}

// newStreamReplayClient returns a client whose InvokeModelWithResponseStream
// replies with events, encoded as Bedrock's event-stream chunks.
func newStreamReplayClient(t *testing.T, events []string) *bedrockruntime.Client {
	t.Helper()
	var body bytes.Buffer
	encoder := eventstream.NewEncoder()
	for _, event := range events {
		payload, err := json.Marshal(map[string]string{"bytes": base64.StdEncoding.EncodeToString([]byte(event))})
		if err != nil {
			t.Fatal(err)
		}
		msg := eventstream.Message{
			Headers: eventstream.Headers{
				{Name: ":event-type", Value: eventstream.StringValue("chunk")},
				{Name: ":message-type", Value: eventstream.StringValue("event")},
				{Name: ":content-type", Value: eventstream.StringValue("application/json")},
			},
			Payload: payload,
		}
		if err := encoder.Encode(&body, msg); err != nil {
			t.Fatal(err)
		}
	}
	return bedrockruntime.New(bedrockruntime.Options{
		Region:      "us-east-1",
		Credentials: credentials.NewStaticCredentialsProvider("test", "test", ""),
		HTTPClient:  &http.Client{Transport: eventStreamTransport{body: body.Bytes()}},
	})
}

func TestAnthropicStreamingReturnsToolCalls(t *testing.T) {
	client := newStreamReplayClient(t, anthropicToolStreamEvents)

	var streamed string
	options := llms.CallOptions{
		StreamingFunc: func(_ context.Context, chunk []byte) error {
			streamed += string(chunk)
			return nil
		},
	}
	input := &bedrockruntime.InvokeModelWithResponseStreamInput{
		ModelId: aws.String("us.anthropic.claude-sonnet-4-6"),
		Body:    []byte(`{}`),
	}
	resp, err := parseStreamingCompletionResponse(context.Background(), client, input, options)
	if err != nil {
		t.Fatal(err)
	}

	choice := resp.Choices[0]
	if want := "Sure! Let me fetch the weather for both cities simultaneously!"; choice.Content != want || streamed != want {
		t.Errorf("content = %q, streamed = %q, want %q", choice.Content, streamed, want)
	}
	if choice.StopReason != "tool_use" {
		t.Errorf("stop reason = %q, want tool_use", choice.StopReason)
	}

	type call struct{ id, name, args string }
	var got []call
	for _, tc := range choice.ToolCalls {
		got = append(got, call{tc.ID, tc.FunctionCall.Name, tc.FunctionCall.Arguments})
	}
	want := []call{
		{"toolu_bdrk_01Pvfgh1kFecPX5cG8Qrkjgh", "get_weather", `{"city":"Paris"}`},
		{"toolu_bdrk_01R7w7Sw5sQDm5wrotaCkB7D", "get_weather", `{"city":"Tokyo"}`},
	}
	if len(got) != len(want) {
		t.Fatalf("tool calls = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("tool call %d = %v, want %v", i, got[i], want[i])
		}
	}
	if choice.FuncCall == nil || choice.FuncCall.Name != "get_weather" {
		t.Errorf("FuncCall = %v, want the first tool call", choice.FuncCall)
	}
	if choice.GenerationInfo["input_tokens"] != 576 || choice.GenerationInfo["output_tokens"] != 102 {
		t.Errorf("usage = %v, want 576 input / 102 output tokens", choice.GenerationInfo)
	}
}
