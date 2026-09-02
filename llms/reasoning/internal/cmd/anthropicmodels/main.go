// Command anthropicmodels refreshes the committed Anthropic listing snapshot
// read by the drift test.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/vxcontrol/langchaingo/llms/anthropic"
)

type entry struct {
	EffortSupported   bool     `json:"effort_supported"`
	Efforts           []string `json:"efforts"`
	ThinkingTypes     []string `json:"thinking_types"`
	StructuredOutputs bool     `json:"structured_outputs"`
}

func main() {
	out := flag.String("out", "testdata/anthropic_models.json", "where to write the snapshot")
	baseURL := flag.String("base-url", "", "override the vendor base URL, e.g. a gateway passthrough")
	flag.Parse()

	if err := run(*out, *baseURL); err != nil {
		fmt.Fprintln(os.Stderr, "anthropicmodels:", err)
		os.Exit(1)
	}
}

func run(out, baseURL string) error {
	opts := []anthropic.Option{}
	if baseURL != "" {
		opts = append(opts, anthropic.WithBaseURL(baseURL))
	}
	llm, err := anthropic.New(opts...)
	if err != nil {
		return err
	}

	resp, err := llm.ListModels(context.Background(), anthropic.ListModelsRequest{Limit: 1000})
	if err != nil {
		return err
	}
	if resp.HasMore {
		return fmt.Errorf("the vendor has more models after %q: teach this command to page "+
			"before writing a snapshot that silently covers less", resp.LastID)
	}
	if len(resp.Models) == 0 {
		return fmt.Errorf("the vendor returned no models")
	}

	snapshot := make(map[string]entry, len(resp.Models))
	for _, m := range resp.Models {
		if !m.CapabilitiesReported {
			return fmt.Errorf("model %q carries no capabilities object: writing it would "+
				"claim the vendor denied every feature", m.ID)
		}
		if m.EffortSupported != (len(m.EffortLevels) > 0) {
			return fmt.Errorf("model %q reports effort supported=%v with %d levels: the "+
				"vendor contradicts itself, look before writing", m.ID, m.EffortSupported, len(m.EffortLevels))
		}
		snapshot[m.ID] = entry{
			EffortSupported:   m.EffortSupported,
			Efforts:           m.EffortLevels,
			ThinkingTypes:     m.ThinkingTypes,
			StructuredOutputs: m.Features.StructuredOutputs,
		}
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
	fmt.Fprintf(os.Stderr, "anthropicmodels: %d models -> %s\n", len(snapshot), out)
	return nil
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
