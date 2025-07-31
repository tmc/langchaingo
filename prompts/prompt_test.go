package prompts

import (
	"testing"
)

func TestStringPromptValueString(t *testing.T) {
	t.Parallel()

	spv := StringPromptValue("")
	str := spv.String()
	if str != "" {
		t.Errorf("expected empty string, got %q", str)
	}

	spv = StringPromptValue("test")
	str = spv.String()
	if str != "test" {
		t.Errorf("expected %q, got %q", "test", str)
	}
}

func TestStringPromptValueMessages(t *testing.T) {
	t.Parallel()

	spv := StringPromptValue("")
	msgs := spv.Messages()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	spv = StringPromptValue("test")
	msgs = spv.Messages()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].GetContent() != "test" {
		t.Errorf("expected %q, got %q", "test", msgs[0].GetContent())
	}
}
