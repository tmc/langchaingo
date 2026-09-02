package reasoning

import (
	"encoding/json"
	"os"
	"sort"
	"testing"
)

type mistralEntry struct {
	Reasoning bool     `json:"reasoning"`
	Aliases   []string `json:"aliases"`
}

const mistralSnapshot = "testdata/mistral_models.json"

var mistralKnownDrift = map[string]string{}

func readMistralListing(t *testing.T) map[string]mistralEntry {
	t.Helper()

	body, err := os.ReadFile(mistralSnapshot)
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	var snapshot map[string]mistralEntry
	if err := json.Unmarshal(body, &snapshot); err != nil {
		t.Fatalf("parse snapshot: %v", err)
	}
	if len(snapshot) == 0 {
		t.Fatal("snapshot is empty; regenerate it with go generate ./llms/reasoning")
	}
	return snapshot
}

func mistralListingDiff(snapshot map[string]mistralEntry) []string {
	ids := make([]string, 0, len(snapshot))
	for id := range snapshot {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	var diffs []string
	for _, id := range ids {
		if got := IsReasoningModel(id); got != snapshot[id].Reasoning {
			diffs = append(diffs, id)
		}
	}
	return diffs
}

func TestTheMistralListingAgreesWithOurTables(t *testing.T) {
	t.Parallel()

	snapshot := readMistralListing(t)
	reasoning, plain := 0, 0
	for _, e := range snapshot {
		if e.Reasoning {
			reasoning++
		} else {
			plain++
		}
	}
	if reasoning == 0 || plain == 0 {
		t.Fatalf("snapshot has %d reasoning and %d plain models: a listing that lost one side "+
			"would still pass every comparison below", reasoning, plain)
	}

	diverging := map[string]bool{}
	for _, id := range mistralListingDiff(snapshot) {
		diverging[id] = true
		if _, known := mistralKnownDrift[id]; !known {
			t.Errorf("new drift on %s: we say reasoning=%v, the vendor says %v. Measure it with "+
				"reasoning_effort before touching either side", id, IsReasoningModel(id), snapshot[id].Reasoning)
		}
	}
	for id := range mistralKnownDrift {
		if !diverging[id] {
			t.Errorf("stale entry %s: the listing and our tables now agree, drop the line", id)
		}
	}
}

func TestEveryLiveSpellingOfAMistralModelGetsOneAnswer(t *testing.T) {
	t.Parallel()

	snapshot := readMistralListing(t)
	checked := 0
	for id, e := range snapshot {
		if _, known := mistralKnownDrift[id]; known {
			continue
		}
		for _, alias := range e.Aliases {
			if _, known := mistralKnownDrift[alias]; known {
				continue
			}
			checked++
			if IsReasoningModel(id) != IsReasoningModel(alias) {
				t.Errorf("%s and %s are the same model by the vendor's own listing, but we answer "+
					"%v and %v", id, alias, IsReasoningModel(id), IsReasoningModel(alias))
			}
		}
	}
	if checked == 0 {
		t.Fatal("no alias pair was compared; the snapshot lost its aliases")
	}
}

func TestTheMistralDriftCheckRedensOnASingleFlippedFlag(t *testing.T) {
	t.Parallel()

	snapshot := readMistralListing(t)
	const target = "codestral-2508"
	entry, ok := snapshot[target]
	if !ok {
		t.Fatalf("%s left the snapshot; point this test at another non-reasoning model", target)
	}
	entry.Reasoning = !entry.Reasoning
	snapshot[target] = entry

	diffs := mistralListingDiff(snapshot)
	var fresh []string
	for _, id := range diffs {
		if _, known := mistralKnownDrift[id]; !known {
			fresh = append(fresh, id)
		}
	}
	if len(fresh) != 1 || fresh[0] != target {
		t.Fatalf("want exactly one new divergence naming %s, got %v", target, fresh)
	}
}
