package bedrock_test

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/vxcontrol/langchaingo/internal/httprr"
	"github.com/vxcontrol/langchaingo/llms/bedrock"
	"github.com/vxcontrol/langchaingo/testing/llmtest"
)

func TestLLM(t *testing.T) {
	cassette := filepath.Join("testdata", "TestLLM.httprr")
	recording, err := httprr.Recording(cassette)
	if err != nil {
		t.Fatal(err)
	}
	if _, statErr := os.Stat(cassette); statErr != nil && !recording {
		t.Skip("no httprr recording for TestLLM; re-run with -httprecord=. and real AWS credentials")
	}

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	if !rr.Recording() {
		t.Parallel()
	}

	client, err := setUpTestWithTransport(rr)
	if err != nil {
		t.Fatal(err)
	}

	llm, err := bedrock.New(bedrock.WithClient(client), bedrock.WithModel(bedrock.ModelAnthropicClaudeHaiku45))
	if err != nil {
		t.Fatalf("Failed to create Bedrock LLM: %v", err)
	}

	llmtest.TestLLM(t, llm)
}
