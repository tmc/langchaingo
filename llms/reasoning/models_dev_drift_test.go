package reasoning

import (
	"encoding/json"
	"os"
	"sort"
	"testing"
)

// The catalogue is a community source that has been measured wrong, so a
// disagreement is a hypothesis: add the entry with its reason, and change the
// tables only when a vendor probe says so.
var knownDrift = map[string]string{
	"google/gemini-2.5-flash-image":            "картиночный маршрут: эффорт вызывает отказ Gemini",
	"google/gemini-3-pro-image":                "картиночный маршрут",
	"google/gemini-3-pro-image-preview":        "картиночный маршрут",
	"google/gemini-3.1-flash-image":            "картиночный маршрут",
	"google/gemini-3.1-flash-image-preview":    "картиночный маршрут",
	"google/gemini-3.1-flash-lite-image":       "картиночный маршрут",
	"google-vertex/gemini-3-pro-image":         "картиночный маршрут",
	"google-vertex/gemini-3.1-flash-image":     "картиночный маршрут",
	"google/gemini-3.1-flash-tts-preview":      "речевой маршрут: 400 Thinking level is not supported",
	"google/gemini-omni-flash-preview":         "не чат: 400 от вендора",
	"google/deep-research-preview-04-2026":     "маршрут не обслуживается",
	"google/deep-research-max-preview-04-2026": "маршрут не обслуживается",
	"openai/gpt-5.2-chat-latest":               "чат-вариант: вендор отвечает 404",
	"openai/gpt-realtime-2.1":                  "не чат: This is not a chat model",
	"amazon-bedrock/qwen.qwen3-32b-v1:0":       "думает только в стриме по просьбе, а предикат отвечает на вопрос о голом вызове",

	"amazon-bedrock/deepseek.v3-v1:0": "имени нет на Bedrock: The provided model identifier is invalid; линейка V3 замерена как нерассуждающая",

	"alibaba/qvq-max":                                       "не измерено",
	"alibaba/qwen3-omni-flash":                              "не измерено",
	"alibaba/qwen3-vl-235b-a22b":                            "не измерено",
	"alibaba/qwen3-vl-30b-a3b":                              "не измерено",
	"alibaba/qwen3-max":                                     "записи расходятся между собой: контроль дал ноль, коммерческий гибрид дал 3095 символов по просьбе",
	"amazon-bedrock/qwen.qwen3-coder-next":                  "не измерено",
	"amazon-bedrock/writer.palmyra-x4-v1:0":                 "не измерено",
	"amazon-bedrock/writer.palmyra-x5-v1:0":                 "не измерено",
	"amazon-bedrock/nvidia.nemotron-nano-12b-v2":            "не измерено",
	"amazon-bedrock/nvidia.nemotron-nano-9b-v2":             "не измерено",
	"amazon-bedrock/openai.gpt-oss-safeguard-120b":          "не измерено: кандидат в не-чатовые поверхности",
	"amazon-bedrock/openai.gpt-oss-safeguard-20b":           "не измерено: кандидат в не-чатовые поверхности",
	"google-vertex/qwen/qwen3-235b-a22b-instruct-2507-maas": "не измерено",
}

type snapshotModel struct {
	Reasoning bool `json:"reasoning"`
	Options   []struct {
		Type string `json:"type"`
	} `json:"reasoning_options"`
}

func (m snapshotModel) claimsEffort() bool {
	for _, o := range m.Options {
		if o.Type == "effort" {
			return true
		}
	}
	return false
}

func (m snapshotModel) claimsBudget() bool {
	for _, o := range m.Options {
		if o.Type == "budget_tokens" {
			return true
		}
	}
	return false
}

// An entry records a divergence measured against the vendor, with the reason.
var knownEffortDrift = map[string]string{
	"xai/grok-4.20-multi-agent-0309": "вендор не обслуживает multi-agent на chat completions: " +
		"«Multi Agent requests are not allowed on chat completions», замер 02.09.2026",
}

var knownBudgetDrift = map[string]string{
	"qwen3-max": "снимок бюджета не объявляет, вендор его исполняет: голый вызов даёт 9166 " +
		"символов размышления против 285 при thinking_budget 100, замер 03.09.2026",
}

func TestModelsDevBudgetDrift(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("testdata/models_dev.json")
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	var snapshot map[string]map[string]snapshotModel
	if err := json.Unmarshal(body, &snapshot); err != nil {
		t.Fatalf("parse snapshot: %v", err)
	}

	diverging := map[string]bool{}
	claims := 0
	for id, entry := range snapshot["alibaba"] {
		if entry.claimsBudget() {
			claims++
		}
		if entry.claimsBudget() != QwenTakesThinkingBudget(id) {
			diverging[id] = true
		}
	}
	if claims == 0 {
		t.Fatal("snapshot carries no budget options for alibaba; regenerate it with go generate ./llms/reasoning")
	}

	for _, name := range sorted(diverging) {
		if _, known := knownBudgetDrift[name]; !known {
			t.Errorf("new budget drift on %s: the catalogue and QwenTakesThinkingBudget disagree on "+
				"whether DashScope caps its thinking by a token budget. Measure it against the vendor, "+
				"then either fix the predicate or add it to knownBudgetDrift with the reason", name)
		}
	}
	for _, name := range sortedKeys(knownBudgetDrift) {
		if !diverging[name] {
			t.Errorf("stale knownBudgetDrift entry %s: the catalogue and our predicate now agree, "+
				"drop the line", name)
		}
	}
}

func TestModelsDevEffortDrift(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("testdata/models_dev.json")
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	var snapshot map[string]map[string]snapshotModel
	if err := json.Unmarshal(body, &snapshot); err != nil {
		t.Fatalf("parse snapshot: %v", err)
	}

	diverging := map[string]bool{}
	claims := 0
	for provider, entries := range snapshot {
		for id, entry := range entries {
			if !entry.claimsEffort() {
				continue
			}
			claims++
			if !AcceptsEffortWire(id) {
				diverging[provider+"/"+id] = true
			}
		}
	}
	if claims == 0 {
		t.Fatal("snapshot carries no effort options; regenerate it with go generate ./llms/reasoning")
	}

	for _, name := range sorted(diverging) {
		if _, known := knownEffortDrift[name]; !known {
			t.Errorf("new effort drift on %s: the catalogue gives it an effort control, our "+
				"AcceptsEffortWire keeps the field off the wire. Measure it against the vendor, "+
				"then either fix the predicate or add it to knownEffortDrift with the reason", name)
		}
	}
	for _, name := range sortedKeys(knownEffortDrift) {
		if !diverging[name] {
			t.Errorf("stale knownEffortDrift entry %s: the catalogue and our predicate now agree, "+
				"drop the line", name)
		}
	}
}

func TestModelsDevSnapshotDrift(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("testdata/models_dev.json")
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	var snapshot map[string]map[string]snapshotModel
	if err := json.Unmarshal(body, &snapshot); err != nil {
		t.Fatalf("parse snapshot: %v", err)
	}

	diverging := map[string]bool{}
	models := 0
	for provider, entries := range snapshot {
		for id, entry := range entries {
			models++
			if entry.Reasoning != IsReasoningModel(id) {
				diverging[provider+"/"+id] = true
			}
		}
	}
	if models == 0 {
		t.Fatal("snapshot is empty; regenerate it with go generate ./llms/reasoning")
	}

	for _, name := range sorted(diverging) {
		if _, known := knownDrift[name]; !known {
			t.Errorf("new drift on %s: the catalogue and our tables disagree. Measure it against the "+
				"vendor, then add it to knownDrift with the reason", name)
		}
	}
	for _, name := range sortedKeys(knownDrift) {
		if !diverging[name] {
			t.Errorf("stale knownDrift entry %s: the catalogue and our tables now agree, drop the line", name)
		}
	}
}

func sorted(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
