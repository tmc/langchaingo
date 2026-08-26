package reasoning

import "testing"

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
