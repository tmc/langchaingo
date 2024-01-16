package qwen_client

import "context"

type Parameters struct {
	ResultFormat      string  `json:"result_format,omitempty"`
	Seed              int     `json:"seed,omitempty"`
	MaxTokens         int     `json:"max_tokens,omitempty"`
	TopP              float64 `json:"top_p,omitempty"`
	TopK              int     `json:"top_k,omitempty"`
	Temperature       float64 `json:"temperature,omitempty"`
	EnableSearch      bool    `json:"enable_search,omitempty"`
	IncrementalOutput bool    `json:"incremental_output,omitempty"`
}

func NewParameters() *Parameters {
	return &Parameters{}
}

func DefaultParameters() *Parameters {
	q := Parameters{}
	q.
		SetResultFormat("message").
		SetTemperature(0.7)

	return &q
}

func (p *Parameters) SetResultFormat(value string) *Parameters {
	p.ResultFormat = value
	return p
}

func (p *Parameters) SetSeed(value int) *Parameters {
	p.Seed = value
	return p
}

func (p *Parameters) SetMaxTokens(value int) *Parameters {
	p.MaxTokens = value
	return p
}

func (p *Parameters) SetTopP(value float64) *Parameters {
	p.TopP = value
	return p
}

func (p *Parameters) SetTopK(value int) *Parameters {
	p.TopK = value
	return p
}

func (p *Parameters) SetTemperature(value float64) *Parameters {
	p.Temperature = value
	return p
}

func (p *Parameters) SetEnableSearch(value bool) *Parameters {
	p.EnableSearch = value
	return p
}

func (p *Parameters) SetIncrementalOutput(value bool) *Parameters {
	p.IncrementalOutput = value
	return p
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Input struct {
	Messages []Message `json:"messages"`
}

type QwenRequest struct {
	Model      string      `json:"model"`
	Input      Input       `json:"input"`
	Parameters *Parameters `json:"parameters,omitempty"`

	StreamingFunc func(ctx context.Context, chunk []byte) error `json:"-"`
}

func (q *QwenRequest) SetModel(value string) *QwenRequest {
	q.Model = value
	return q
}

func (q *QwenRequest) SetInput(value Input) *QwenRequest {
	q.Input = value
	return q
}

func (q *QwenRequest) SetParameters(value *Parameters) *QwenRequest {
	q.Parameters = value
	return q
}

func (q *QwenRequest) SetStreamingFunc(fn func(ctx context.Context, chunk []byte) error) *QwenRequest {
	q.StreamingFunc = fn
	return q
}

// old version response output
type QwenOutputText struct {
	Output struct {
		FinishReason string `json:"finish_reason"`
		Text         string `json:"text"`
	} `json:"output"`
	Usage struct {
		TotalTokens  int `json:"total_tokens"`
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	RequestID string `json:"request_id"`
}

type QwenResponse struct {
	ID         string            `json:"id"`
	Event      string            `json:"event"`
	HttpStatus int               `json:"http_status"`
	Output     QwenOutputMessage `json:"output"`
	Err        error             `json:"error"`
}

// new version response format
type QwenOutput struct {
	Choices []struct {
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
}

type QwenOutputMessage struct {
	Output QwenOutput `json:"output"`
	Usage  struct {
		TotalTokens  int `json:"total_tokens"`
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	RequestID string `json:"request_id"`
	// ErrMsg    string `json:"error_msg"`
}

