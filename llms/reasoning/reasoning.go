package reasoning

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
	"sync"
)

const trimCharset = "\t\r\n "

var (
	reMatchStartReasoningContent = regexp.MustCompile(`(?s)^(.*?)<(think|thinking)>(.*?)$`)
	reMatchEndReasoningContent   = regexp.MustCompile(`(?s)^(.*?)</(think|thinking)>(.*?)$`)
	reMatchReasoningContent      = regexp.MustCompile(`(?s)^(.*?)<(think|thinking)>(.*?)</(?:think|thinking)>\s*(.*?)$`)
)

type ContentReasoning struct {
	// Content is the reasoning content of the assistant message before the final answer.
	Content string `json:"content,omitempty"`

	// Signature is the signature of the reasoning contents.
	Signature []byte `json:"signature,omitempty"`
}

func (r *ContentReasoning) IsEmpty() bool {
	return r == nil || (r.Content == "" && len(r.Signature) == 0)
}

func (r *ContentReasoning) String() string {
	if r.IsEmpty() {
		return "Reasoning: (empty)"
	}

	var buf bytes.Buffer
	buf.WriteString("Reasoning: ")
	buf.WriteString(r.Content)
	if len(r.Signature) > 0 {
		buf.WriteString("\nSignature: ")
		buf.Write(r.Signature)
	}

	return buf.String()
}

func (r *ContentReasoning) MarshalJSON() ([]byte, error) {
	type Alias ContentReasoning
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}

func (r *ContentReasoning) UnmarshalJSON(data []byte) error {
	type Alias ContentReasoning
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	return json.Unmarshal(data, aux)
}

type ChunkContentSplitterState int

const (
	ChunkContentSplitterStateText ChunkContentSplitterState = iota
	ChunkContentSplitterStateReasoning
)

type ChunkContentSplitter interface {
	Split(chunk string) (string, string)
	GetState() ChunkContentSplitterState
}

type chunkContentSplitter struct {
	mx    sync.Mutex
	state ChunkContentSplitterState
}

func NewChunkContentSplitter() ChunkContentSplitter {
	return &chunkContentSplitter{}
}

func (c *chunkContentSplitter) Split(chunk string) (string, string) {
	if c == nil { // splitter is not initialized and it does not work
		return "", chunk
	}

	c.mx.Lock()
	defer c.mx.Unlock()

	if c.state == ChunkContentSplitterStateReasoning {
		matches := reMatchEndReasoningContent.FindStringSubmatch(chunk)
		if len(matches) < 3 {
			return "", chunk
		}

		c.state = ChunkContentSplitterStateText
		suffix := matches[1]
		chunk = matches[3]
		return chunk, suffix
	}

	matches := reMatchStartReasoningContent.FindStringSubmatch(chunk)
	if len(matches) < 3 {
		return chunk, ""
	}

	c.state = ChunkContentSplitterStateReasoning
	prefix := matches[1]
	reasoning := matches[3]

	matches = reMatchEndReasoningContent.FindStringSubmatch(reasoning)
	if len(matches) < 3 {
		return prefix, reasoning
	}

	c.state = ChunkContentSplitterStateText
	suffix := matches[1]
	chunk = matches[3]
	return prefix + " " + chunk, suffix
}

func (c *chunkContentSplitter) GetState() ChunkContentSplitterState {
	if c == nil { // splitter is not initialized and returns default state
		return ChunkContentSplitterStateText
	}

	c.mx.Lock()
	defer c.mx.Unlock()

	return c.state
}

func SplitContent(content string) (string, string) {
	content = strings.Trim(content, trimCharset)

	matches := reMatchReasoningContent.FindStringSubmatch(content)
	if len(matches) < 5 {
		return "", content
	}

	prefix := strings.Trim(matches[1], trimCharset)
	reasoning := strings.Trim(matches[3], trimCharset)
	text := strings.Trim(matches[4], trimCharset)

	if prefix != "" {
		text = prefix + "\n" + text
	}

	return reasoning, text
}

func SplitContentWithReasoning(content string) (*ContentReasoning, string) {
	reasoning, text := SplitContent(content)
	return &ContentReasoning{
		Content:   reasoning,
		Signature: nil, // received in special field in the response
	}, text
}

// IsReasoningModel returns true if the model is a reasoning/thinking model.
// This includes OpenAI o1/o3/GPT-5 series, Anthropic Claude 3.7+, DeepSeek reasoner, etc.
// For runtime checking of LLM instances, use SupportsReasoningModel instead.
func IsReasoningModel(model string) bool {
	return DefaultIsReasoningModel(model)
}

// DefaultIsReasoningModel provides the default reasoning model detection logic.
// This can be used by LLM implementations that want to extend rather than replace
// the default detection logic.
func DefaultIsReasoningModel(model string) bool { //nolint:funlen // a flat catalog of model-family prefix checks; splitting hurts readability
	modelLower := strings.ToLower(model)

	// Remove provider prefix if present (e.g., "openai/", "anthropic/", "google/")
	if idx := strings.LastIndex(modelLower, "/"); idx != -1 {
		modelLower = modelLower[idx+1:]
	}
	modelLower = stripPlatformPrefix(modelLower)

	if strings.Contains(modelLower, "-chat-latest") {
		return false
	}

	// OpenAI reasoning models
	if strings.HasPrefix(modelLower, "gpt-5") ||
		strings.HasPrefix(modelLower, "gpt-oss-") ||
		strings.HasPrefix(modelLower, "o1") ||
		strings.HasPrefix(modelLower, "o3") ||
		strings.HasPrefix(modelLower, "o4-mini") {
		return true
	}

	// Anthropic extended thinking / adaptive models
	if strings.Contains(modelLower, "claude-3.7") ||
		strings.Contains(modelLower, "claude-3-7") ||
		strings.HasPrefix(modelLower, "claude-opus-4") ||
		strings.HasPrefix(modelLower, "claude-opus-5") ||
		strings.HasPrefix(modelLower, "claude-sonnet-4") ||
		strings.HasPrefix(modelLower, "claude-sonnet-5") ||
		strings.Contains(modelLower, "claude-fable-5") ||
		strings.Contains(modelLower, "claude-mythos-5") ||
		strings.Contains(modelLower, "claude-haiku-4.5") ||
		strings.Contains(modelLower, "claude-haiku-4-5") {
		return true
	}

	// DeepSeek reasoning models
	if strings.Contains(modelLower, "deepseek-r1") ||
		strings.Contains(modelLower, "deepseek-chat-v3") ||
		strings.Contains(modelLower, "deepseek-v3.1-terminus") ||
		strings.HasPrefix(modelLower, "deepseek-v3.2") ||
		strings.HasPrefix(modelLower, "deepseek-v4") {
		return true
	}

	// Google Gemini / Gemma reasoning models (Gemini 2.5, Gemini 3.x, Gemma 4).
	if GeminiSupportsThinking(modelLower) {
		return true
	}

	// X-AI Grok reasoning models
	if strings.HasPrefix(modelLower, "grok-3-mini") ||
		(strings.HasPrefix(modelLower, "grok-4") && !strings.HasSuffix(modelLower, "-non-reasoning")) ||
		(strings.HasPrefix(modelLower, "grok-5") && !strings.HasSuffix(modelLower, "-non-reasoning")) ||
		strings.HasPrefix(modelLower, "grok-build") ||
		strings.Contains(modelLower, "grok-code-fast") {
		return true
	}

	// Z-AI GLM reasoning models (Zhipu AI)
	if strings.HasPrefix(modelLower, "glm-4.5") ||
		strings.HasPrefix(modelLower, "glm-4.6") ||
		strings.HasPrefix(modelLower, "glm-4.7") ||
		strings.HasPrefix(modelLower, "glm-5") {
		return true
	}

	// Qwen reasoning models. Qwen3.x thinking is user-toggleable with configurable
	// sampling, so the bare qwen3 prefix must not force temperature pinning; only
	// an explicit "thinking" marker or the QwQ line count as always-on reasoning.
	if (strings.HasPrefix(modelLower, "qwen") && strings.Contains(modelLower, "thinking")) ||
		strings.Contains(modelLower, "qwq-") {
		return true
	}

	// Minimax reasoning models
	if strings.HasPrefix(modelLower, "minimax-m") {
		return true
	}

	// Moonshot AI Kimi reasoning models
	if strings.Contains(modelLower, "kimi-") &&
		(strings.Contains(modelLower, "k2-thinking") ||
			strings.Contains(modelLower, "2.5") ||
			strings.Contains(modelLower, "2.6") ||
			strings.Contains(modelLower, "2.7") ||
			strings.Contains(modelLower, "k3") ||
			strings.Contains(modelLower, "dev-72b")) {
		return true
	}

	// Perplexity reasoning models
	if strings.Contains(modelLower, "sonar-") &&
		(strings.Contains(modelLower, "deep-research") ||
			strings.Contains(modelLower, "reasoning") ||
			strings.Contains(modelLower, "pro-search")) {
		return true
	}

	// Other reasoning models
	if strings.Contains(modelLower, "aion-") ||
		(strings.Contains(modelLower, "olmo-") && strings.Contains(modelLower, "-think")) ||
		strings.Contains(modelLower, "nova-2-lite") ||
		strings.Contains(modelLower, "trinity-mini") ||
		strings.Contains(modelLower, "ernie-4.5") ||
		strings.Contains(modelLower, "seed-1.6") ||
		strings.Contains(modelLower, "cogito-v2") ||
		(strings.Contains(modelLower, "lfm-") && strings.Contains(modelLower, "-thinking")) ||
		strings.Contains(modelLower, "deephermes") ||
		strings.Contains(modelLower, "hermes-4-") ||
		strings.Contains(modelLower, "nemotron") ||
		strings.Contains(modelLower, "intellect-3") ||
		strings.Contains(modelLower, "step3") ||
		strings.Contains(modelLower, "hunyuan-a13b") ||
		strings.Contains(modelLower, "chimera") ||
		strings.Contains(modelLower, "mimo-v2") ||
		strings.Contains(modelLower, "tongyi-deepresearch") {
		return true
	}

	// Additional reasoning families from the current OpenRouter catalog. Matched by
	// the distinctive part of the model name so detection works both on OpenRouter
	// (with a provider prefix, stripped above) and on the original provider.
	if strings.HasPrefix(modelLower, "seed-2.0") ||
		strings.HasPrefix(modelLower, "step-3") ||
		strings.HasPrefix(modelLower, "north-mini") ||
		strings.HasPrefix(modelLower, "mercury-2") ||
		strings.HasPrefix(modelLower, "ring-2.6") ||
		strings.HasPrefix(modelLower, "muse-spark") ||
		strings.HasPrefix(modelLower, "mistral-medium-3") ||
		strings.HasPrefix(modelLower, "mistral-small-2603") ||
		strings.HasPrefix(modelLower, "nex-n2") ||
		strings.HasPrefix(modelLower, "perceptron-mk") ||
		strings.HasPrefix(modelLower, "laguna-") ||
		strings.HasPrefix(modelLower, "reka-flash-3") ||
		strings.HasPrefix(modelLower, "fugu-ultra") ||
		strings.HasPrefix(modelLower, "hy3") ||
		strings.HasPrefix(modelLower, "solar-pro-3") ||
		strings.Contains(modelLower, "trinity-large-thinking") {
		return true
	}

	return false
}

func stripPlatformPrefix(model string) string {
	for {
		idx := strings.Index(model, ".")
		if idx <= 0 {
			return model
		}
		if !isAlpha(model[:idx]) {
			return model
		}
		model = model[idx+1:]
	}
}

func isAlpha(s string) bool {
	for i := range len(s) {
		if s[i] < 'a' || s[i] > 'z' {
			return false
		}
	}
	return true
}
