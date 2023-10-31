package callbacks

import (
	"context"
	"strings"
)

//nolint:all
var DefaultKeywords = []string{"Final Answer:", "Final:", "AI:"}

// nolint:all
type AgentFinalStreamHandler struct {
	SimpleHandler
	ctx             context.Context
	egress          chan []byte
	Keywords        []string
	LastTokens      string
	KeywordDetected bool
	PrintOutput     bool
}

var _ Handler = &AgentFinalStreamHandler{}

func NewFinalStreamHandler(keywords ...string) *AgentFinalStreamHandler {
	if len(keywords) > 0 {
		DefaultKeywords = keywords
	}

	return &AgentFinalStreamHandler{
		egress:   make(chan []byte),
		Keywords: DefaultKeywords,
	}
}

func (handler *AgentFinalStreamHandler) GetEgress() chan []byte {
	return handler.egress
}

func (handler *AgentFinalStreamHandler) ReadFromEgress(callback func(ctx context.Context, chunk []byte)) {
	egressCtx, cancel := context.WithCancel(handler.ctx)
	defer cancel()

	go func() {
		defer close(handler.egress)
		for data := range handler.egress {
			callback(egressCtx, data)
		}
	}()
}

func (handler *AgentFinalStreamHandler) HandleStreamingFunc(ctx context.Context, chunk []byte) {
	handler.ctx = ctx
	chunkStr := string(chunk)
	handler.LastTokens += chunkStr

	// Buffer the last few chunks to match the longest keyword size
	longestSize := len(handler.Keywords[0])
	for _, k := range handler.Keywords {
		if len(k) > longestSize {
			longestSize = len(k)
		}
	}

	if len(handler.LastTokens) > longestSize {
		handler.LastTokens = handler.LastTokens[len(handler.LastTokens)-longestSize:]
	}

	// Check for keywords
	for _, k := range DefaultKeywords {
		if strings.Contains(handler.LastTokens, k) {
			handler.KeywordDetected = true
		}
	}

	// Check for colon and set print mode.
	if handler.KeywordDetected && chunkStr != ":" {
		handler.PrintOutput = true
	}

	// Print the final output after the detection of keyword.
	if handler.PrintOutput {
		handler.egress <- chunk
	}
}
