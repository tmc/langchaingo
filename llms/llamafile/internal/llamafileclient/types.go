package llamafileclient

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type StatusError struct {
	Status       string `json:"status,omitempty"`
	ErrorMessage string `json:"error"`
	StatusCode   int    `json:"code,omitempty"`
}

func (e StatusError) Error() string {
	switch {
	case e.Status != "" && e.ErrorMessage != "":
		return fmt.Sprintf("%s: %s", e.Status, e.ErrorMessage)
	case e.Status != "":
		return e.Status
	case e.ErrorMessage != "":
		return e.ErrorMessage
	default:
		// this should not happen
		return "something went wrong, please see the ollama server logs for details"
	}
}

type GenerateRequest struct {
	Prompt   string `json:"prompt"`
	System   string `json:"system"`
	Template string `json:"template"`
	Context  []int  `json:"context,omitempty"`
	Stream   *bool  `json:"stream"`

	GenerationSettings
}

type ImageData []byte

type Message struct {
	Role    string      `json:"role"` // one of ["system", "user", "assistant"]
	Content string      `json:"content"`
	Images  []ImageData `json:"images,omitempty"`
}

type ChatRequest struct {
	Messages []*Message `json:"-"`
	Prompt   *string    `json:"prompt,omitempty"`
	Stream   *bool      `json:"stream,omitempty"`
	GenerationSettings
}

type GenerationSettings struct {
	FrequencyPenalty       float64       `json:"frequency_penalty,omitempty"`
	Grammar                string        `json:"grammar,omitempty"`
	IgnoreEOS              bool          `json:"ignore_eos,omitempty"`
	LogitBias              []interface{} `json:"logit_bias,omitempty"` // Assuming array of unknown structure, adjust as needed
	MinP                   float64       `json:"min_p,omitempty"`
	Mirostat               int           `json:"mirostat,omitempty"`
	MirostatEta            float64       `json:"mirostat_eta,omitempty"`
	MirostatTau            float64       `json:"mirostat_tau,omitempty"`
	Model                  string        `json:"model,omitempty"`
	NCtx                   int           `json:"n_ctx,omitempty"`
	NKeep                  int           `json:"n_keep,omitempty"`
	NPredict               int           `json:"n_predict,omitempty"`
	NProbs                 int           `json:"n_probs,omitempty"`
	PenalizeNL             bool          `json:"penalize_nl,omitempty"`
	PenaltyPromptTokens    []interface{} `json:"penalty_prompt_tokens,omitempty"` // Assuming array of unknown structure, adjust as needed
	PresencePenalty        float64       `json:"presence_penalty,omitempty"`
	RepeatLastN            int           `json:"repeat_last_n,omitempty"`
	RepeatPenalty          float64       `json:"repeat_penalty,omitempty"`
	Seed                   uint32        `json:"seed,omitempty"` // uint32 due to the value 4294967295
	Stop                   []string      `json:"stop,omitempty"`
	Stream                 bool          `json:"stream,omitempty"`
	Temperature            float64       `json:"temperature,omitempty"`
	TfsZ                   float64       `json:"tfs_z,omitempty"`
	TopK                   int           `json:"top_k,omitempty"`
	TopP                   float64       `json:"top_p,omitempty"`
	TypicalP               float64       `json:"typical_p,omitempty"`
	UsePenaltyPromptTokens bool          `json:"use_penalty_prompt_tokens,omitempty"`
	LlamafileServerURL     *url.URL
	HTTPClient             *http.Client
}

type Timings struct {
	PredictedMS         float64 `json:"predicted_ms"`
	PredictedN          int     `json:"predicted_n"`
	PredictedPerSecond  float64 `json:"predicted_per_second"`
	PredictedPerTokenMS float64 `json:"predicted_per_token_ms"`
	PromptMS            float64 `json:"prompt_ms"`
	PromptN             int     `json:"prompt_n"`
	PromptPerSecond     float64 `json:"prompt_per_second"`
	PromptPerTokenMS    float64 `json:"prompt_per_token_ms"`
}

type GenerateResponse struct {
	CreatedAt          time.Time     `json:"created_at"`
	Response           string        `json:"response"`
	Context            []int         `json:"context,omitempty"`
	TotalDuration      time.Duration `json:"total_duration,omitempty"`
	LoadDuration       time.Duration `json:"load_duration,omitempty"`
	PromptEvalCount    int           `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration time.Duration `json:"prompt_eval_duration,omitempty"`
	EvalCount          int           `json:"eval_count,omitempty"`
	EvalDuration       time.Duration `json:"eval_duration,omitempty"`
	Done               bool          `json:"done"`
}

type ChatResponse struct {
	Content            string             `json:"content"`
	GenerationSettings GenerationSettings `json:"generation_settings"`
	Model              string             `json:"model"`
	Prompt             string             `json:"prompt"`
	SlotID             int                `json:"slot_id"`
	Stop               bool               `json:"stop"`
	StoppedEOS         bool               `json:"stopped_eos"`
	StoppedLimit       bool               `json:"stopped_limit"`
	StoppedWord        bool               `json:"stopped_word"`
	StoppingWord       string             `json:"stopping_word"`
	Timings            Timings            `json:"timings"`
	TokensCached       int                `json:"tokens_cached"`
	TokensEvaluated    int                `json:"tokens_evaluated"`
	TokensPredicted    int                `json:"tokens_predicted"`
	Truncated          bool               `json:"truncated"`
}
