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

// IsEmpty reports whether there is nothing to carry back into the next turn.
func (r *ContentReasoning) IsEmpty() bool {
	return r == nil || (r.Content == "" && len(r.Signature) == 0)
}

// HasContent reports whether the model actually reasoned.
func (r *ContentReasoning) HasContent() bool {
	return r != nil && r.Content != ""
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
func DefaultIsReasoningModel(model string) bool {
	for _, form := range modelSpellings(model) {
		if namesReasoningModel(form) {
			return true
		}
	}
	return false
}

func mistralReasons(model string) bool {
	if model == "mistral-medium" {
		return true
	}
	for _, prefix := range []string{
		"mistral-medium-3", "mistral-medium-latest", "mistral-medium-2604",
		"mistral-small-2603", "mistral-small-latest",
		"mistral-vibe-cli",
	} {
		if strings.HasPrefix(model, prefix) {
			return true
		}
	}
	return false
}

func nvidiaNonChatSurface(model string) bool {
	return strings.Contains(model, "-embed-") ||
		strings.Contains(model, "-rerank-") ||
		strings.Contains(model, "-asr-") ||
		strings.Contains(model, "content-safety")
}

func namesReasoningModel(modelLower string) bool { //nolint:funlen // a flat catalog of model-family prefix checks; splitting hurts readability
	if strings.Contains(modelLower, "-chat-latest") || openAIChatVariant(modelLower) {
		return false
	}

	// OpenAI reasoning models
	if strings.HasPrefix(modelLower, "gpt-5") ||
		strings.HasPrefix(modelLower, "gpt-oss") ||
		strings.HasPrefix(modelLower, "o1") ||
		strings.HasPrefix(modelLower, "o3") ||
		strings.HasPrefix(modelLower, "o4-mini") ||
		modelLower == "gpt-latest" {
		return true
	}

	// Anthropic extended thinking / adaptive models
	claudeName := canonicalClaude(modelLower)
	if strings.Contains(claudeName, "claude-3-7") ||
		strings.HasPrefix(claudeName, "claude-opus-4") ||
		strings.HasPrefix(claudeName, "claude-opus-5") ||
		strings.HasPrefix(claudeName, "claude-sonnet-4") ||
		strings.HasPrefix(claudeName, "claude-sonnet-5") ||
		strings.Contains(claudeName, "claude-fable-5") ||
		strings.Contains(claudeName, "claude-mythos-5") ||
		strings.Contains(claudeName, "claude-haiku-4-5") ||
		ClaudeSupportsThinking(claudeName) {
		return true
	}

	// DeepSeek reasoning models
	if strings.Contains(modelLower, "deepseek-r1") ||
		strings.Contains(modelLower, "deepseek-reasoner") ||
		strings.Contains(modelLower, "deepseek-chat-v3.") ||
		strings.Contains(modelLower, "deepseek-v3.1") ||
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
		(strings.HasPrefix(modelLower, "grok-4") && !strings.Contains(modelLower, "-non-reasoning")) ||
		(strings.HasPrefix(modelLower, "grok-5") && !strings.Contains(modelLower, "-non-reasoning")) ||
		strings.HasPrefix(modelLower, "grok-build") ||
		strings.Contains(modelLower, "grok-code-fast") ||
		modelLower == "grok-latest" {
		return true
	}

	// Z-AI GLM reasoning models (Zhipu AI)
	glm := strings.TrimPrefix(modelLower, "zai-")
	if strings.HasPrefix(glm, "glm-4.5") ||
		strings.HasPrefix(glm, "glm-4.6") ||
		strings.HasPrefix(glm, "glm-4.7") ||
		strings.HasPrefix(glm, "glm-5") ||
		glm == "glm-latest" || glm == "glm-flash-latest" {
		return true
	}

	if strings.HasPrefix(modelLower, "labs-leanstral") {
		return true
	}

	// Qwen reasoning models. Qwen3.x thinking is user-toggleable with configurable
	// sampling, so the bare qwen3 prefix must not force temperature pinning; only
	// an explicit "thinking" marker or the QwQ line count as always-on reasoning.
	if (strings.HasPrefix(modelLower, "qwen") && ThinkingMarkedInName(modelLower)) ||
		strings.Contains(modelLower, "qwq-") ||
		QwenThinkingRequiresStream(modelLower) ||
		QwenThinkingEnabledByFlag(modelLower) ||
		qwenReasonsUnasked(modelLower) {
		return true
	}

	// Minimax reasoning models
	if strings.HasPrefix(modelLower, "minimax-m") && !strings.Contains(modelLower, "-her") {
		return true
	}

	// Moonshot AI Kimi reasoning models
	if strings.Contains(modelLower, "kimi-") &&
		(strings.Contains(modelLower, "k2-thinking") ||
			strings.Contains(modelLower, "2.5") ||
			strings.Contains(modelLower, "2.6") ||
			strings.Contains(modelLower, "2.7") ||
			strings.Contains(modelLower, "k3") ||
			strings.Contains(modelLower, "dev-72b")) ||
		modelLower == "kimi-latest" {
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
	if (strings.Contains(modelLower, "aion-") && !strings.Contains(modelLower, "aion-rp")) ||
		(strings.Contains(modelLower, "olmo-") && strings.Contains(modelLower, "-think")) ||
		strings.Contains(modelLower, "nova-2-lite") ||
		strings.Contains(modelLower, "trinity-mini") ||
		strings.Contains(modelLower, "ernie-4.5") ||
		strings.Contains(modelLower, "seed-1.6") ||
		strings.Contains(modelLower, "cogito-v2") ||
		(strings.Contains(modelLower, "lfm-") && strings.Contains(modelLower, "-thinking")) ||
		strings.HasPrefix(modelLower, "lfm-2.5") ||
		strings.Contains(modelLower, "seed-2-1") ||
		strings.Contains(modelLower, "inkling") ||
		strings.Contains(modelLower, "deephermes") ||
		strings.Contains(modelLower, "hermes-4-") ||
		(strings.Contains(modelLower, "nemotron") && !nvidiaNonChatSurface(modelLower)) ||
		strings.Contains(modelLower, "intellect-3") ||
		strings.Contains(modelLower, "step3") ||
		strings.Contains(modelLower, "hunyuan-a13b") ||
		strings.Contains(modelLower, "chimera") ||
		(strings.Contains(modelLower, "mimo-v2") && !strings.Contains(modelLower, "-tts")) ||
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
		mistralReasons(modelLower) ||
		strings.HasPrefix(modelLower, "magistral") ||
		strings.HasPrefix(modelLower, "nex-n2") ||
		strings.HasPrefix(modelLower, "perceptron-mk") ||
		strings.HasPrefix(modelLower, "laguna-") ||
		strings.HasPrefix(modelLower, "reka-flash-3") ||
		strings.HasPrefix(modelLower, "fugu-ultra") ||
		strings.HasPrefix(modelLower, "hy3") ||
		strings.HasPrefix(modelLower, "hy4") ||
		strings.HasPrefix(modelLower, "solar-pro-3") ||
		strings.HasPrefix(modelLower, "solar-pro4") ||
		strings.HasPrefix(modelLower, "ling-3.0") ||
		strings.HasPrefix(modelLower, "longcat-2") ||
		strings.Contains(modelLower, "trinity-large-thinking") {
		return true
	}

	return false
}

// ThinkingOptIn reports whether the model reasons only when a request asks it to.
func ThinkingOptIn(model string) bool {
	if OpenAIThinkingOptIn(model) || QwenThinkingEnabledByFlag(model) {
		return true
	}
	for _, form := range modelSpellings(model) {
		if mistralReasons(form) || strings.HasPrefix(form, "magistral") ||
			strings.HasPrefix(form, "ernie-4.5") {
			return true
		}
		for _, prefix := range []string{"solar-pro-3", "solar-pro4", "deepseek-v3.1", "deepseek-chat-v3.1", "deepseek-v3.2"} {
			if strings.HasPrefix(form, prefix) {
				return true
			}
		}
	}
	return false
}

// ThinkingMarkedInName reports whether the model name selects a routing variant
// that already runs with thinking on, not merely a model capable of thinking.
func ThinkingMarkedInName(model string) bool {
	m := strings.ToLower(model)
	const marker = "thinking"
	for i := 0; i < len(m); {
		idx := strings.Index(m[i:], marker)
		if idx == -1 {
			return false
		}
		start, end := i+idx, i+idx+len(marker)
		if !letterAt(m, start-1) && !letterAt(m, end) {
			return true
		}
		i = end
	}
	return false
}

func letterAt(s string, i int) bool {
	return i >= 0 && i < len(s) && s[i] >= 'a' && s[i] <= 'z'
}

func isAlpha(s string) bool {
	for i := range len(s) {
		if s[i] < 'a' || s[i] > 'z' {
			return false
		}
	}
	return true
}

func qwenReasonsUnasked(model string) bool {
	for _, prefix := range []string{"qwen3.5", "qwen3.6", "qwen3.7", "qwen3.8"} {
		if strings.HasPrefix(model, prefix) {
			return true
		}
	}
	return false
}
