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
	"alibaba/qwen-turbo":                                    "не измерено",
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
	"mistral/zai-glm-5-2":                                   "не измерено: шлюз режет параметр, вывод невозможен в обе стороны",
}

type snapshotModel struct {
	Reasoning bool `json:"reasoning"`
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
