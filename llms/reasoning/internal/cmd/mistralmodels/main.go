// Command mistralmodels refreshes the committed Mistral listing snapshot read by
// the drift test.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"time"
)

const defaultBaseURL = "https://api.mistral.ai/v1"

type vendorModel struct {
	ID           string   `json:"id"`
	Aliases      []string `json:"aliases"`
	Capabilities *struct {
		Reasoning bool `json:"reasoning"`
	} `json:"capabilities"`
}

type entry struct {
	Reasoning bool     `json:"reasoning"`
	Aliases   []string `json:"aliases"`
}

func main() {
	out := flag.String("out", "testdata/mistral_models.json", "where to write the snapshot")
	baseURL := flag.String("base-url", defaultBaseURL, "override the vendor base URL, e.g. a gateway passthrough")
	flag.Parse()

	if err := run(*out, *baseURL); err != nil {
		fmt.Fprintln(os.Stderr, "mistralmodels:", err)
		os.Exit(1)
	}
}

func run(out, baseURL string) error {
	models, err := fetch(baseURL, os.Getenv("MISTRAL_API_KEY"))
	if err != nil {
		return err
	}
	if len(models) == 0 {
		return fmt.Errorf("the vendor returned no models")
	}

	snapshot := make(map[string]entry, len(models))
	for _, m := range models {
		if m.Capabilities == nil {
			return fmt.Errorf("model %q carries no capabilities object: writing it would claim "+
				"the vendor denied reasoning", m.ID)
		}
		aliases := append([]string(nil), m.Aliases...)
		sort.Strings(aliases)
		snapshot[m.ID] = entry{Reasoning: m.Capabilities.Reasoning, Aliases: aliases}
	}

	if err := keepsEveryCommittedModel(out, snapshot); err != nil {
		return err
	}

	body, err := json.MarshalIndent(snapshot, "", " ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(out, append(body, '\n'), 0o600); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "mistralmodels: %d models -> %s\n", len(snapshot), out)
	return nil
}

func fetch(baseURL, key string) ([]vendorModel, error) {
	req, err := http.NewRequest(http.MethodGet, baseURL+"/models", http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+key)

	client := &http.Client{Timeout: 2 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s/models: %s", baseURL, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var payload struct {
		Data []vendorModel `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	return payload.Data, nil
}

func keepsEveryCommittedModel(out string, snapshot map[string]entry) error {
	body, err := os.ReadFile(out)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var committed map[string]entry
	if err := json.Unmarshal(body, &committed); err != nil {
		return err
	}

	var gone []string
	for id := range committed {
		if _, ok := snapshot[id]; !ok {
			gone = append(gone, id)
		}
	}
	if len(gone) > 0 {
		sort.Strings(gone)
		return fmt.Errorf("the committed snapshot has %v and this key does not see them: "+
			"a narrower key would shrink coverage without a word", gone)
	}
	return nil
}
