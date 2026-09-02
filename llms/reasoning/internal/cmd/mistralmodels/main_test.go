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
	t.Setenv("MISTRAL_API_KEY", "test")
	return srv.URL
}

const oneReasoner = `{"object":"list","data":[{"id":"magistral-medium-latest",` +
	`"aliases":["mistral-medium-latest"],"capabilities":{"reasoning":true}}]}`

func TestASnapshotKeepsTheFlagAndTheAliases(t *testing.T) {
	url := serve(t, oneReasoner)
	out := filepath.Join(t.TempDir(), "snap.json")

	if err := run(out, url); err != nil {
		t.Fatalf("run: %v", err)
	}

	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var written map[string]entry
	if err := json.Unmarshal(body, &written); err != nil {
		t.Fatal(err)
	}
	got := written["magistral-medium-latest"]
	if !got.Reasoning || len(got.Aliases) != 1 || got.Aliases[0] != "mistral-medium-latest" {
		t.Errorf("snapshot lost detail: %+v", got)
	}
}

func TestARecordWithoutCapabilitiesIsRefused(t *testing.T) {
	url := serve(t, `{"object":"list","data":[{"id":"mistral-newborn-9"}]}`)
	out := filepath.Join(t.TempDir(), "snap.json")

	err := run(out, url)
	if err == nil {
		t.Fatal("a model the vendor said nothing about was written as denying reasoning")
	}
	if !strings.Contains(err.Error(), "mistral-newborn-9") {
		t.Errorf("the refusal does not name the model: %v", err)
	}
	if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
		t.Error("the refusal still left a snapshot behind")
	}
}

func TestAnEmptyListingIsRefused(t *testing.T) {
	url := serve(t, `{"object":"list","data":[]}`)
	out := filepath.Join(t.TempDir(), "snap.json")

	if err := run(out, url); err == nil {
		t.Fatal("an empty listing was written, and every comparison against it would pass")
	}
}

func TestANarrowerKeyCannotShrinkTheCommittedSnapshot(t *testing.T) {
	url := serve(t, oneReasoner)
	out := filepath.Join(t.TempDir(), "snap.json")
	committed := `{"magistral-medium-latest":{"reasoning":true,"aliases":[]},` +
		`"codestral-2508":{"reasoning":false,"aliases":[]}}`
	if err := os.WriteFile(out, []byte(committed), 0o600); err != nil {
		t.Fatal(err)
	}

	err := run(out, url)
	if err == nil {
		t.Fatal("a key that sees fewer models silently rewrote the snapshot")
	}
	if !strings.Contains(err.Error(), "codestral-2508") {
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
