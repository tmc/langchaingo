package reasoning

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sort"
	"testing"
)

type listingEntry struct {
	EffortSupported   bool     `json:"effort_supported"`
	Efforts           []string `json:"efforts"`
	ThinkingTypes     []string `json:"thinking_types"`
	StructuredOutputs bool     `json:"structured_outputs"`
}

const listingSnapshot = "testdata/anthropic_models.json"

func readListing(t *testing.T) map[string]listingEntry {
	t.Helper()

	body, err := os.ReadFile(listingSnapshot)
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	var snapshot map[string]listingEntry
	if err := json.Unmarshal(body, &snapshot); err != nil {
		t.Fatalf("parse snapshot: %v", err)
	}
	if len(snapshot) == 0 {
		t.Fatal("snapshot is empty; regenerate it with go generate ./llms/reasoning")
	}
	return snapshot
}

func kindFromThinkingTypes(types []string) ClaudeReasoningKind {
	enabled := slices.Contains(types, "enabled")
	adaptive := slices.Contains(types, "adaptive")
	switch {
	case enabled && adaptive:
		return ClaudeReasoningAdaptiveAndBudget
	case adaptive:
		return ClaudeReasoningAdaptiveOnly
	case enabled:
		return ClaudeReasoningBudgetOnly
	default:
		return ClaudeReasoningUnknown
	}
}

func claudeListingDiff(snapshot map[string]listingEntry) []string {
	ids := make([]string, 0, len(snapshot))
	for id := range snapshot {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	var diffs []string
	for _, id := range ids {
		e := snapshot[id]

		ours := ClaudeEffortsFor(id, ProviderAnthropic)
		if len(ours) != len(e.Efforts) || (len(ours) > 0 && !slices.Equal(ours, e.Efforts)) {
			diffs = append(diffs, fmt.Sprintf("%s: we accept efforts %v, the vendor lists %v", id, ours, e.Efforts))
		}
		if want, got := kindFromThinkingTypes(e.ThinkingTypes), ClaudeReasoningKindFor(id); want != got {
			diffs = append(diffs, fmt.Sprintf("%s: we classify thinking as %d, the vendor reports %v (%d)",
				id, got, e.ThinkingTypes, want))
		}
		if got := ClaudeSupportsStructuredOutput(id); got != e.StructuredOutputs {
			diffs = append(diffs, fmt.Sprintf("%s: we say structured output %v, the vendor says %v",
				id, got, e.StructuredOutputs))
		}
		if e.EffortSupported != (len(e.Efforts) > 0) {
			diffs = append(diffs, fmt.Sprintf("%s: the vendor reports effort supported=%v with %d levels",
				id, e.EffortSupported, len(e.Efforts)))
		}
	}
	return diffs
}

func TestTheVendorListingAgreesWithOurTables(t *testing.T) {
	t.Parallel()

	for _, d := range claudeListingDiff(readListing(t)) {
		t.Errorf("drift: %s. Measure it against the vendor before touching either side", d)
	}
}

func TestTheListingSnapshotCoversEveryThinkingKind(t *testing.T) {
	t.Parallel()

	seen := map[ClaudeReasoningKind]int{}
	for _, e := range readListing(t) {
		seen[kindFromThinkingTypes(e.ThinkingTypes)]++
	}
	for _, kind := range []ClaudeReasoningKind{
		ClaudeReasoningAdaptiveOnly, ClaudeReasoningAdaptiveAndBudget, ClaudeReasoningBudgetOnly,
	} {
		if seen[kind] == 0 {
			t.Errorf("no model of kind %d in the snapshot: a listing that lost a whole generation "+
				"would still pass the drift test", kind)
		}
	}
}

func TestTheKeyThatRecordedTheSnapshotCannotSeeMythos(t *testing.T) {
	t.Parallel()

	snapshot := readListing(t)
	for _, id := range []string{"claude-mythos-5", "claude-mythos-preview"} {
		if _, ok := snapshot[id]; ok {
			t.Errorf("%s is in the snapshot now: our table for it was never vendor-checked, "+
				"so verify it and drop this guard", id)
		}
		if !ClaudeSupportsThinking(id) {
			t.Errorf("%s left our tables while still unverifiable against the listing", id)
		}
	}
}

func TestTheDriftCheckRedensOnASingleChangedLevel(t *testing.T) {
	t.Parallel()

	snapshot := readListing(t)
	const target = "claude-sonnet-4-6"
	entry, ok := snapshot[target]
	if !ok {
		t.Fatalf("%s left the snapshot; point this test at another dual-thinking model", target)
	}
	entry.Efforts = append(slices.Clone(entry.Efforts), "xhigh")
	snapshot[target] = entry

	diffs := claudeListingDiff(snapshot)
	if len(diffs) != 1 {
		t.Fatalf("want exactly one divergence from one changed level, got %d: %v", len(diffs), diffs)
	}
	if !slices.ContainsFunc(diffs, func(d string) bool { return len(d) > 0 && d[:len(target)] == target }) {
		t.Errorf("the divergence does not name %s: %v", target, diffs)
	}
}

func TestTheListingSaysNothingAboutSamplingOrDefaultThinking(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile(listingSnapshot)
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	var raw map[string]map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("parse snapshot: %v", err)
	}

	want := []string{"effort_supported", "efforts", "structured_outputs", "thinking_types"}
	for id, fields := range raw {
		keys := make([]string, 0, len(fields))
		for k := range fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		if !slices.Equal(keys, want) {
			t.Fatalf("%s carries %v: the snapshot grew a field, so decide whether the drift "+
				"test should compare it before it becomes silent data", id, keys)
		}
	}

	if !ClaudeThinkingDefaultsOn("claude-opus-5") {
		t.Error("claude-opus-5 thinks by default; the listing carries no field for it, so the table must")
	}
	if !ClaudeRejectsSampling("claude-opus-4-7") {
		t.Error("claude-opus-4-7 rejects sampling while thinking; the listing carries no field for it")
	}
	if !ClaudeThinkingAlwaysOn("claude-fable-5") {
		t.Error("claude-fable-5 cannot be switched off; the listing carries no field for it")
	}
}
