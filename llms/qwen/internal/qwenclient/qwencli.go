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

	"log"
)

var ErrEmptyResponse = errors.New("empty response")

type QwenClient struct {
	Model   Qwen_Model
	baseURL string
	token   string
	httpCli IHttpClient
}

func NewQwenClient(model string, httpCli IHttpClient) *QwenClient {
	qwen_model := ChoseQwenModel(model)

	return &QwenClient{
		Model:   qwen_model,
		baseURL: QWEN_DASHSCOPE_URL,
		token:   os.Getenv("DASHSCOPE_API_KEY"),
		httpCli: httpCli,
	}
}

func (q *QwenClient) parseStreamingChatResponse(ctx context.Context, payload *QwenRequest) (*QwenOutputMessage, error) {
	if payload.Model == "" {
		payload.Model = string(q.Model)
	}
	responseChan := q.asyncChatStreaming(ctx, payload)
	outputMessage := QwenOutputMessage{}
	for rspData := range responseChan {
		if rspData.Err != nil {
			return nil, fmt.Errorf("parseStreamingChatResponse err: %v", rspData.Err)
		}
		if len(rspData.Output.Output.Choices) == 0 {
			return nil, ErrEmptyResponse
		}

		chunk := []byte(rspData.Output.Output.Choices[0].Message.Content)

		if payload.StreamingFunc != nil {
			err := payload.StreamingFunc(ctx, chunk)
			if err != nil {
				return nil, fmt.Errorf("parseStreamingChatResponse err: %v", err)
			}
		}

		outputMessage.RequestID = rspData.Output.RequestID
		outputMessage.Usage = rspData.Output.Usage
		if outputMessage.Output.Choices == nil {
			outputMessage.Output.Choices = rspData.Output.Output.Choices
		} else {
			outputMessage.Output.Choices[0].Message.Role = rspData.Output.Output.Choices[0].Message.Role
			outputMessage.Output.Choices[0].Message.Content += rspData.Output.Output.Choices[0].Message.Content
			outputMessage.Output.Choices[0].FinishReason = rspData.Output.Output.Choices[0].FinishReason
		}
	}

	return &outputMessage, nil
}

func (q *QwenClient) CreateCompletion(ctx context.Context, payload *QwenRequest) (*QwenOutputMessage, error) {
	if payload.Parameters == nil {
		payload.Parameters = DefaultParameters()
	}
	if payload.StreamingFunc != nil {
		payload.Parameters.SetIncrementalOutput(true)
		return q.parseStreamingChatResponse(ctx, payload)
	} else {
		return q.SyncCall(ctx, payload)
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

func (q *QwenClient) asyncChatStreaming(ctx context.Context, r *QwenRequest) <-chan QwenResponse {
	_respChunkChannel := make(chan QwenResponse, 100)

	go func() {
		withHeader := map[string]string{
			"Authorization": "Bearer " + q.token,
			// "X-DashScope-SSE": "enable",
			"Accept": "text/event-stream",
		}

		q._combineStreamingChunk(ctx, r, withHeader, _respChunkChannel)
	}()
	return _respChunkChannel
}

func (q *QwenClient) SyncCall(ctx context.Context, req *QwenRequest) (*QwenOutputMessage, error) {
	withHeader := map[string]string{
		"Authorization": "Bearer " + q.token,
	}

	headerOpt := WithHeader(withHeader)

	resp := QwenOutputMessage{}
	err := q.httpCli.Post(ctx, q.baseURL, req, &resp, headerOpt)
	if err != nil {
		return nil, err
	}
	if len(resp.Output.Choices) == 0 {
		return nil, ErrEmptyResponse
	}
	return &resp, nil
}

/*
 * combine SSE streaming lines to be a structed response data
 * id: xxxx
 * event: xxxxx
 * ......
 */
func (q *QwenClient) _combineStreamingChunk(ctx context.Context, reqBody *QwenRequest, withHeader map[string]string, _respChunkChannel chan QwenResponse) {
	defer close(_respChunkChannel)
	var _rawStreamOutChannel chan string

	var err error
	headerOpt := WithHeader(withHeader)


	_rawStreamOutChannel, err = q.httpCli.PostSSE(ctx, q.baseURL, reqBody, headerOpt)
	if err != nil {
		_respChunkChannel <- QwenResponse{Err: err}
		return
	}

	var rsp QwenResponse = QwenResponse{}

	for v := range _rawStreamOutChannel {
		if strings.TrimSpace(v) == "" {
			// streaming out combined response
			_respChunkChannel <- rsp
			rsp = QwenResponse{}
			continue
		} else {
			q.fillInRespData(v, &rsp)
		}
	}
}

// filled in response data line by line
func (q *QwenClient) fillInRespData(line string, output *QwenResponse) error {
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
