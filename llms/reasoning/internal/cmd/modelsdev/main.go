// Command modelsdev refreshes the committed models.dev snapshot read by the
// drift test.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const catalogue = "https://models.dev/api.json"

// Ollama is deliberately absent: the catalogue does not carry local runtimes, so
// listing it here would make every run exit 1.
var providers = []string{
	"anthropic", "openai", "xai", "google", "google-vertex", "amazon-bedrock",
	"deepseek", "mistral", "moonshotai", "zhipuai", "minimax", "alibaba",
}

type model struct {
	Reasoning        bool              `json:"reasoning"`
	ReasoningOptions []reasoningOption `json:"reasoning_options,omitempty"`
	Temperature      *bool             `json:"temperature,omitempty"`
	Limit            *limit            `json:"limit,omitempty"`
}

type reasoningOption struct {
	Type   string   `json:"type"`
	Values []string `json:"values,omitempty"`
	Min    *int     `json:"min,omitempty"`
	Max    *int     `json:"max,omitempty"`
}

type limit struct {
	Context int `json:"context,omitempty"`
	Output  int `json:"output,omitempty"`
}

func main() {
	out := flag.String("out", "testdata/models_dev.json", "where to write the snapshot")
	flag.Parse()

	catalog, err := fetch()
	if err != nil {
		fmt.Fprintln(os.Stderr, "modelsdev:", err)
		os.Exit(1)
	}

	snapshot := make(map[string]map[string]model, len(providers))
	kept := 0
	for _, p := range providers {
		models, ok := catalog[p]
		if !ok {
			fmt.Fprintf(os.Stderr, "modelsdev: provider %q is gone from the catalogue\n", p)
			os.Exit(1)
		}
		snapshot[p] = models.Models
		kept += len(models.Models)
	}

	body, err := json.MarshalIndent(snapshot, "", " ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "modelsdev:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(*out, append(body, '\n'), 0o600); err != nil {
		fmt.Fprintln(os.Stderr, "modelsdev:", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "modelsdev: %d models from %d providers -> %s\n", kept, len(providers), *out)
}

func fetch() (map[string]struct {
	Models map[string]model `json:"models"`
}, error) {
	client := &http.Client{Timeout: 2 * time.Minute}
	resp, err := client.Get(catalogue)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", catalogue, resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var catalog map[string]struct {
		Models map[string]model `json:"models"`
	}
	if err := json.Unmarshal(body, &catalog); err != nil {
		return nil, err
	}
	return catalog, nil
}
