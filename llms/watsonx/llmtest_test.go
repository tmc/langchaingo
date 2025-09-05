package watsonx

import (
	"os"
	"testing"

	wx "github.com/IBM/watsonx-go/pkg/models"
	"github.com/tmc/langchaingo/testing/llmtest"
)

func TestLLM(t *testing.T) {
	if os.Getenv("WATSONX_API_KEY") == "" || os.Getenv("WATSONX_PROJECT_ID") == "" {
		t.Skip("WATSONX_API_KEY or WATSONX_PROJECT_ID not set")
	}

	llm, err := New(
		"ibm/granite-13b-instruct-v2",
		wx.WithWatsonxAPIKey(wx.WatsonxAPIKey(os.Getenv("WATSONX_API_KEY"))),
		wx.WithWatsonxProjectID(wx.WatsonxProjectID(os.Getenv("WATSONX_PROJECT_ID"))),
	)
	if err != nil {
		t.Fatalf("Failed to create WatsonX LLM: %v", err)
	}

	llmtest.TestLLM(t, llm)
}
