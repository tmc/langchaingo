package bedrockclient

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSignatureOnlyThinkingBlockKeepsItsField(t *testing.T) {
	t.Parallel()

	body, err := json.Marshal(anthropicTextGenerationInputContent{
		Type:      "thinking",
		Signature: "sig-abc",
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(body), `"thinking":""`) {
		t.Errorf("a signature-only thinking block must still carry the field, got %s", body)
	}
	if !strings.Contains(string(body), `"signature":"sig-abc"`) {
		t.Errorf("the signature must survive, got %s", body)
	}
}

func TestOtherBlocksDoNotGainAThinkingField(t *testing.T) {
	t.Parallel()

	for _, blockType := range []string{"text", "image", "tool_use", "tool_result"} {
		t.Run(blockType, func(t *testing.T) {
			t.Parallel()
			body, err := json.Marshal(anthropicTextGenerationInputContent{
				Type: blockType,
				Text: "hi",
			})
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if strings.Contains(string(body), `"thinking"`) {
				t.Errorf("a %s block must not carry a thinking field, got %s", blockType, body)
			}
		})
	}
}
