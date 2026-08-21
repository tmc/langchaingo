package bedrock_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/bedrock"
)

func truncationLLM(t *testing.T, stopReason string) *bedrock.LLM {
	t.Helper()
	return truncationLLMWithBody(t,
		fmt.Sprintf(`{"output":{"message":{"role":"assistant","content":[{"text":"half an ans"}]}},"stopReason":%q,"usage":{"inputTokens":1,"outputTokens":1,"totalTokens":2}}`, stopReason),
		bedrock.WithModel("anthropic.claude-3-sonnet-20240229-v1:0"), bedrock.WithConverseAPI())
}

func truncationLLMWithBody(t *testing.T, body string, opts ...bedrock.Option) *bedrock.LLM {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)

	client := bedrockruntime.NewFromConfig(aws.Config{
		Region:      "us-east-1",
		Credentials: credentials.NewStaticCredentialsProvider("unit", "test", ""),
	}, func(o *bedrockruntime.Options) {
		o.BaseEndpoint = aws.String(srv.URL)
	})

	llm, err := bedrock.New(append([]bedrock.Option{bedrock.WithClient(client)}, opts...)...)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return llm
}

func TestTruncationIsDerivedWithoutRewritingStopReason(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name          string
		stopReason    string
		wantTruncated bool
	}{
		{"out of budget", "max_tokens", true},
		{"finished on its own", "end_turn", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			llm := truncationLLM(t, tc.stopReason)

			resp, err := llm.GenerateContent(context.Background(),
				[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")})
			if err != nil {
				t.Fatalf("GenerateContent: %v", err)
			}
			if got := resp.Choices[0].Truncated; got != tc.wantTruncated {
				t.Fatalf("Truncated = %v, want %v", got, tc.wantTruncated)
			}
			if got := resp.Choices[0].StopReason; got != tc.stopReason {
				t.Fatalf("StopReason = %q, want the vendor's own %q", got, tc.stopReason)
			}
		})
	}
}

func TestFailOnTruncationKeepsThePartialAnswer(t *testing.T) {
	t.Parallel()

	llm := truncationLLM(t, "max_tokens")
	msgs := []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}

	resp, err := llm.GenerateContent(context.Background(), msgs, llms.WithFailOnTruncation())
	if err == nil {
		t.Fatal("want an error once the caller opted in")
	}
	if !llms.IsTruncatedError(err) {
		t.Fatalf("want a truncation error, got %v", err)
	}
	if resp == nil || resp.Choices[0].Content != "half an ans" {
		t.Fatalf("want the partial answer alongside the error, got %+v", resp)
	}

	if _, err := llm.GenerateContent(context.Background(), msgs); err != nil {
		t.Fatalf("without the option the same response must stay a success: %v", err)
	}
}

func TestTruncationOnTheLegacyDoor(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name          string
		model         string
		body          string
		wantStop      string
		wantTruncated bool
	}{
		{
			name:  "anthropic out of budget",
			model: "anthropic.claude-3-sonnet-20240229-v1:0",
			body: `{"type":"message","role":"assistant","content":[{"type":"text","text":"half an ans"}],` +
				`"stop_reason":"max_tokens","usage":{"input_tokens":1,"output_tokens":1}}`,
			wantStop: "max_tokens", wantTruncated: true,
		},
		{
			name:  "anthropic finished on its own",
			model: "anthropic.claude-3-sonnet-20240229-v1:0",
			body: `{"type":"message","role":"assistant","content":[{"type":"text","text":"a whole answer"}],` +
				`"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`,
			wantStop: "end_turn", wantTruncated: false,
		},
		{
			name:     "amazon spells it LENGTH",
			model:    "amazon.titan-text-express-v1",
			body:     `{"inputTextTokenCount":1,"results":[{"tokenCount":1,"outputText":"half an ans","completionReason":"LENGTH"}]}`,
			wantStop: "LENGTH", wantTruncated: true,
		},
		{
			name:     "amazon spells finishing FINISH",
			model:    "amazon.titan-text-express-v1",
			body:     `{"inputTextTokenCount":1,"results":[{"tokenCount":1,"outputText":"a whole answer","completionReason":"FINISH"}]}`,
			wantStop: "FINISH", wantTruncated: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			llm := truncationLLMWithBody(t, tc.body, bedrock.WithModel(tc.model))

			resp, err := llm.GenerateContent(context.Background(),
				[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")})
			if err != nil {
				t.Fatalf("GenerateContent: %v", err)
			}
			if got := resp.Choices[0].Truncated; got != tc.wantTruncated {
				t.Fatalf("Truncated = %v, want %v", got, tc.wantTruncated)
			}
			if got := resp.Choices[0].StopReason; got != tc.wantStop {
				t.Fatalf("StopReason = %q, want the vendor's own %q", got, tc.wantStop)
			}
		})
	}
}
