package openaiclient

import (
	"strings"
	"testing"
)

func TestStatusErrorSurfacesProviderText(t *testing.T) {
	cases := []struct {
		name, body, want string
	}{
		{"openai shape", `{"error":{"message":"bad request","type":"invalid"}}`, "bad request"},
		{"xai shape", `{"code":"invalid-argument","error":"Model grok-4.5 does not support parameter presencePenalty."}`, "presencePenalty"},
		{"plain text", "upstream exploded", "upstream exploded"},
		{"empty body", "", "400"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := statusError(400, strings.NewReader(tc.body))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want it to contain %q", err, tc.want)
			}
		})
	}
}

func TestStatusErrorCapsTheQuotedBody(t *testing.T) {
	err := statusError(500, strings.NewReader(strings.Repeat("x", maxErrorBodyBytes*4)))
	if len(err.Error()) > maxErrorBodyBytes+128 {
		t.Fatalf("error grew to %d bytes, want it capped near %d", len(err.Error()), maxErrorBodyBytes)
	}
}
