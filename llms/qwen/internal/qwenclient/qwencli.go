package qwen_client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"strconv"
	"strings"
	"sync"

	"log"
)

// request input

var ErrEmptyResponse = errors.New("empty response")

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

func NewParameters() Parameters {
	return Parameters{}
}

func DefaultParameters() Parameters {
	q := Parameters{}
	q.
		SetResultFormat("message").
		SetTemperature(0.7)
		// SetIncrementalOutput(true)

	return q
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
	Model      string     `json:"model"`
	Input      Input      `json:"input"`
	Parameters Parameters `json:"parameters"`

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

func (q *QwenRequest) SetParameters(value Parameters) *QwenRequest {
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

// =======================

type QwenClient struct {
	Model   Qwen_Model
	baseURL string
	token   string
}

func NewQwenClient(model string) *QwenClient {
	qwen_model := ChoseQwenModel(model)

	return &QwenClient{
		Model:   qwen_model,
		baseURL: QWEN_DASHSCOPE_URL,
		token:   os.Getenv("DASHSCOPE_API_KEY"),
	}
}

func (q *QwenClient) parseStreamingChatResponse(ctx context.Context, payload *QwenRequest) (*QwenOutputMessage, error) {
	responseChan := q.AsyncChatStreaming(ctx, payload)
	outputMessage := QwenOutputMessage{}
	for rspData := range responseChan {
		if rspData.Err != nil {
			return nil, fmt.Errorf("parseStreamingChatResponse err: %v", rspData.Err)
		}
		if len(rspData.Output.Output.Choices) == 0 {
			return nil, ErrEmptyResponse
			// continue
		}

		chunk := []byte(rspData.Output.Output.Choices[0].Message.Content)

		if payload.StreamingFunc != nil {
			err := payload.StreamingFunc(ctx, chunk)
			if err != nil {
				return nil, fmt.Errorf("parseStreamingChatResponse err: %v", err)
			}
		}

		// fmt.Printf("11111 %+v\n", rspData.Output.Output.Choices[0])
		outputMessage.RequestID = rspData.Output.RequestID
		outputMessage.Usage = rspData.Output.Usage
		if outputMessage.Output.Choices == nil {
			outputMessage.Output.Choices = rspData.Output.Output.Choices
		}
		outputMessage.Output.Choices[0].FinishReason = rspData.Output.Output.Choices[0].FinishReason
		outputMessage.Output.Choices[0].Message.Role = rspData.Output.Output.Choices[0].Message.Role
		outputMessage.Output.Choices[0].Message.Content += rspData.Output.Output.Choices[0].Message.Content
		fmt.Println("debug... outputMessage.Content: ", outputMessage.Output.Choices[0].Message.Content)
	}

	return &outputMessage, nil
}

func (q *QwenClient) CreateCompletion(ctx context.Context, payload *QwenRequest) (*QwenOutputMessage, error) {
	if payload.StreamingFunc != nil {
		payload.Parameters.SetIncrementalOutput(true)
		return q.parseStreamingChatResponse(ctx, payload)
	} else {
		// TODO: should also support disable SSE streaming
		return q.SyncCall(ctx, payload)
		// panic("unimplementd non-streaming completion")
	}
}

func (q *QwenClient) CreateEmbedding(ctx context.Context, r *EmbeddingRequest) ([][]float32, error) {
	if r.Model == "" {
		r.Model = defaultEmbeddingModel
	}
	if r.Params.TextType == "" {
		r.Params.TextType = "document"
	}
	resp, err := q.createEmbedding(ctx, r)
	if err != nil {
		return nil, err
	}
	if len(resp.Output.Embeddings) == 0 {
		return nil, ErrEmptyResponse

	}
	embeddings := make([][]float32, 0)
	for i := 0; i < len(resp.Output.Embeddings); i++ {
		embeddings = append(embeddings, resp.Output.Embeddings[i].Embedding)
	}
	return embeddings, nil
}

func (q *QwenClient) AsyncChatStreaming(ctx context.Context, r *QwenRequest) <-chan QwenResponse {
	_respChunkChannel := make(chan QwenResponse, 100)

	go func() {
		withHeader := map[string]string{
			"Authorization":   "Bearer " + q.token,
			"X-DashScope-SSE": "enable",
		}

		q._combineStreamingChunk(ctx, r, withHeader, _respChunkChannel)
	}()
	return _respChunkChannel
}

func (q *QwenClient) SyncCall(ctx context.Context, r *QwenRequest) (*QwenOutputMessage, error) {
	// fmt.Println("SyncCall: ", r)
	withHeader := map[string]string{
		"Authorization": "Bearer " + q.token,
	}
	resp, err := Post[QwenOutputMessage](ctx, q.baseURL, r, 0, withHeader)
	if err != nil {
		return nil, err
	}
	if len(resp.Output.Choices) == 0 {
		return nil, ErrEmptyResponse
	}
	return &resp, nil
}

/*
 * combine SSE lines to be a structed response data
 * id: xxxx
 * event: xxxxx
 * ......
 */
func (q *QwenClient) _combineStreamingChunk(ctx context.Context, reqBody *QwenRequest, withHeader map[string]string, _respChunkChannel chan QwenResponse) {
	// fmt.Println("go: combine streaming Chunk...: header: ", withHeader)
	defer func() {
		// fmt.Println("close channel")
		close(_respChunkChannel)
	}()
	_rawStreamOutChannel := make(chan string, 500)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(header map[string]string) {
		defer wg.Done()

		err := PostSSE(ctx, q.baseURL, reqBody, _rawStreamOutChannel, header)
		if err != nil {
			fmt.Println("go: PostSSE err: ", err)
			_respChunkChannel <- QwenResponse{Err: err}
			return
		}
	}(withHeader)

	var rsp QwenResponse = QwenResponse{}
	for v := range _rawStreamOutChannel {
		// fmt.Println("go: raw stream: ", v)
		if strings.TrimSpace(v) == "" {
			// 发送组合好的 rsp 给调用方
			_respChunkChannel <- rsp
			rsp = QwenResponse{}
			continue
		} else {
			// 循环给 rsp 中的每个字段赋值
			q.parseEvent(v, &rsp)
		}
	}
	wg.Wait()
	// fmt.Println("go: stream chat over")
}

func (q *QwenClient) parseEvent(line string, output *QwenResponse) error {
	if strings.TrimSpace(line) == "" {
		return nil
	}

	switch {
	case strings.HasPrefix(line, "id:"):
		output.ID = strings.TrimPrefix(line, "id:")
	case strings.HasPrefix(line, "event:"):
		output.Event = strings.TrimPrefix(line, "event:")
	case strings.HasPrefix(line, ":HTTP_STATUS/"):
		code, err := strconv.Atoi(strings.TrimPrefix(line, ":HTTP_STATUS/"))
		if err != nil {
			output.Err = fmt.Errorf("http_status err: strconv.Atoi  %v", err)
		}
		output.HttpStatus = code
	case strings.HasPrefix(line, "data:"):
		if output.Event == "error" {
			output.Err = errors.New(strings.TrimPrefix(line, "data:"))
			return nil
		}
		dataJson := strings.TrimPrefix(line, "data:")
		outputData := QwenOutputMessage{}
		err := json.Unmarshal([]byte(dataJson), &outputData)
		if err != nil {
			panic(err)
		}

		output.Output = outputData
	default:
		data := bytes.TrimSpace([]byte(line))
		log.Printf("unknown line: %s", data)
	}

	return nil
}
