package reasoning

import (
	"strings"
	"testing"
)

func TestHasContentSeparatesReasoningFromSignature(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		reasoning  *ContentReasoning
		hasContent bool
		isEmpty    bool
	}{
		{"nil", nil, false, true},
		{"zero value", &ContentReasoning{}, false, true},
		{"signature only", &ContentReasoning{Signature: []byte("sig")}, false, false},
		{"content only", &ContentReasoning{Content: "step 1"}, true, false},
		{"content and signature", &ContentReasoning{Content: "step 1", Signature: []byte("sig")}, true, false},
	} {
		if got := tc.reasoning.HasContent(); got != tc.hasContent {
			t.Errorf("%s: HasContent() = %v, want %v", tc.name, got, tc.hasContent)
		}
		if got := tc.reasoning.IsEmpty(); got != tc.isEmpty {
			t.Errorf("%s: IsEmpty() = %v, want %v", tc.name, got, tc.isEmpty)
		}
	}
}

func TestRedactedReasoningIsCarriedButIsNotText(t *testing.T) {
	t.Parallel()

	redacted := &ContentReasoning{Redacted: []byte{0x01, 0x02, 0x03}}
	if redacted.IsEmpty() {
		t.Error("an encrypted block is something to carry back, so it is not empty")
	}
	if redacted.HasContent() {
		t.Error("an encrypted block is not readable thinking")
	}
	if got := redacted.String(); !strings.Contains(got, "3 bytes") {
		t.Errorf("String() = %q, want the encrypted size named", got)
	}
	if !(&ContentReasoning{}).IsEmpty() {
		t.Error("a value with nothing in it stays empty")
	}
}
