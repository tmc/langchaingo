package mistral

import (
	"encoding/json"
	"os"
	"testing"
)

const vendorListingSnapshot = "../reasoning/testdata/mistral_models.json"

func TestTheDefaultModelIsOneTheVendorStillServes(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile(vendorListingSnapshot)
	if err != nil {
		t.Fatalf("read %s: %v", vendorListingSnapshot, err)
	}
	var listing map[string]json.RawMessage
	if err := json.Unmarshal(body, &listing); err != nil {
		t.Fatalf("parse %s: %v", vendorListingSnapshot, err)
	}
	if len(listing) == 0 {
		t.Fatalf("%s is empty, so it would accept any default; refresh it with "+
			"go generate ./llms/reasoning", vendorListingSnapshot)
	}

	llm, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	got := llm.clientOptions.model
	if _, served := listing[got]; !served {
		t.Fatalf("default model %q is not in the vendor listing %s: a caller that never "+
			"passes WithModel gets a retired id and the vendor rejects the call",
			got, vendorListingSnapshot)
	}
}
