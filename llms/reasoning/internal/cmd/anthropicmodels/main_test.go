package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func serve(t *testing.T, body string) string {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)
	t.Setenv("ANTHROPIC_API_KEY", "test")
	return srv.URL
}

func model(id, effort, thinking string) string {
	return `{"type":"model","id":"` + id + `","capabilities":{` +
		`"structured_outputs":{"supported":true},"effort":` + effort +
		`,"thinking":` + thinking + `}}`
}

const (
	fullEffort = `{"supported":true,"low":{"supported":true},"medium":{"supported":true},` +
		`"high":{"supported":true},"xhigh":{"supported":true},"max":{"supported":true}}`
	noEffort = `{"supported":false,"low":{"supported":false},"medium":{"supported":false},` +
		`"high":{"supported":false},"xhigh":{"supported":false},"max":{"supported":false}}`
	adaptive = `{"supported":true,"types":{"enabled":{"supported":false},"adaptive":{"supported":true}}}`
)

func TestASnapshotIsWrittenForAWellFormedListing(t *testing.T) {
	url := serve(t, `{"data":[`+model("claude-opus-5", fullEffort, adaptive)+`],"has_more":false}`)
	out := filepath.Join(t.TempDir(), "snap.json")

	if err := run(out, url); err != nil {
		t.Fatalf("run: %v", err)
	}

	var written map[string]entry
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, &written); err != nil {
		t.Fatal(err)
	}
	got := written["claude-opus-5"]
	if len(got.Efforts) != 5 || got.ThinkingTypes[0] != "adaptive" || !got.StructuredOutputs {
		t.Errorf("snapshot lost detail: %+v", got)
	}
}

func TestAPagedListingIsRefusedInsteadOfTruncated(t *testing.T) {
	url := serve(t, `{"data":[`+model("claude-opus-5", fullEffort, adaptive)+`],`+
		`"has_more":true,"last_id":"claude-opus-5"}`)
	out := filepath.Join(t.TempDir(), "snap.json")

	err := run(out, url)
	if err == nil {
		t.Fatal("a truncated listing was written as if it were the whole vendor")
	}
	if !strings.Contains(err.Error(), "more models") {
		t.Errorf("unexpected refusal: %v", err)
	}
	if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
		t.Error("the refusal still left a snapshot behind")
	}
}

func TestARecordWithoutCapabilitiesIsRefused(t *testing.T) {
	url := serve(t, `{"data":[{"type":"model","id":"claude-newborn-9"}],"has_more":false}`)
	out := filepath.Join(t.TempDir(), "snap.json")

	err := run(out, url)
	if err == nil {
		t.Fatal("a model the vendor said nothing about was written as denying every feature")
	}
	if !strings.Contains(err.Error(), "claude-newborn-9") {
		t.Errorf("the refusal does not name the model: %v", err)
	}
}

func TestAVendorContradictingItselfIsRefused(t *testing.T) {
	url := serve(t, `{"data":[`+model("claude-opus-5", noEffort, adaptive)+
		`,`+model("claude-odd-1", `{"supported":true,"low":{"supported":false},`+
		`"medium":{"supported":false},"high":{"supported":false},"xhigh":{"supported":false},`+
		`"max":{"supported":false}}`, adaptive)+`],"has_more":false}`)
	out := filepath.Join(t.TempDir(), "snap.json")

	err := run(out, url)
	if err == nil {
		t.Fatal("effort supported with no levels was written without a word")
	}
	if !strings.Contains(err.Error(), "claude-odd-1") {
		t.Errorf("the refusal does not name the model: %v", err)
	}
}

func TestANarrowerKeyCannotShrinkTheCommittedSnapshot(t *testing.T) {
	url := serve(t, `{"data":[`+model("claude-opus-5", fullEffort, adaptive)+`],"has_more":false}`)
	out := filepath.Join(t.TempDir(), "snap.json")
	committed := `{"claude-opus-5":{"effort_supported":true,"efforts":["low"],` +
		`"thinking_types":["adaptive"],"structured_outputs":true},` +
		`"claude-sonnet-5":{"effort_supported":true,"efforts":["low"],` +
		`"thinking_types":["adaptive"],"structured_outputs":true}}`
	if err := os.WriteFile(out, []byte(committed), 0o600); err != nil {
		t.Fatal(err)
	}

	err := run(out, url)
	if err == nil {
		t.Fatal("a key that sees fewer models silently rewrote the snapshot")
	}
	if !strings.Contains(err.Error(), "claude-sonnet-5") {
		t.Errorf("the refusal does not name what would be lost: %v", err)
	}

	after, readErr := os.ReadFile(out)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(after) != committed {
		t.Error("the refusal still overwrote the committed snapshot")
	}
}
