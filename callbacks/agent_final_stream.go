package callbacks

import (
	"context"
	"strings"
)

// DefaultKeywords is map of the agents final out prefix keywords.
//
//nolint:all
var DefaultKeywords = []string{"Final Answer:", "Final:", "AI:"}

type AgentFinalStreamHandler struct {
	SimpleHandler
	egress          chan []byte
	Keywords        []string
	LastTokens      string
	KeywordDetected bool
	PrintOutput     bool
}

var _ Handler = &AgentFinalStreamHandler{}

// NewFinalStreamHandler creates a new instance of the AgentFinalStreamHandler struct.
//
// It accepts a variadic number of strings as keywords. If any keywords are provided,
// the DefaultKeywords variable is updated with the provided keywords.
//
// DefaultKeywords is map of the agents final out prefix keywords.
//
// The function returns a pointer to the created AgentFinalStreamHandler struct.
func NewFinalStreamHandler(keywords ...string) *AgentFinalStreamHandler {
	if len(keywords) > 0 {
		DefaultKeywords = keywords
	}

	return &AgentFinalStreamHandler{
		egress:   make(chan []byte),
		Keywords: DefaultKeywords,
	}
}

// GetEgress returns the egress channel of the AgentFinalStreamHandler.
//
// It does not take any parameters.
// It returns a channel of type []byte.
func (handler *AgentFinalStreamHandler) GetEgress() chan []byte {
	return handler.egress
}

// ReadFromEgress reads data from the egress channel and invokes the provided
// callback function with each chunk of data.
//
// The callback function receives two parameters:
// - ctx: the context.Context object for the egress operation.
// - chunk: a byte slice representing a chunk of data from the egress channel.
func (handler *AgentFinalStreamHandler) ReadFromEgress(
	ctx context.Context,
	callback func(ctx context.Context, chunk []byte),
) {
	go func() {
		defer close(handler.egress)
		for data := range handler.egress {
			callback(ctx, data)
		}
	}()
}

// HandleStreamingFunc implements the callback interface that handles the streaming
// of data in the AgentFinalStreamHandler. The handler reads the incoming data and checks for the
// agents final output keywords, ie, `Final Answer:`, `Final:`, `AI:`. Upon detection of
// the keyword, it starst to stream the agents final output to the egress channel.
//
// It takes in the context and a chunk of bytes as parameters.
// There is no return type for this function.
func (handler *AgentFinalStreamHandler) HandleStreamingFunc(_ context.Context, chunk []byte) {
	chunkStr := string(chunk)
	handler.LastTokens += chunkStr
	var detectedKeyword string

	// Buffer the last few chunks to match the longest keyword size
	var longestSize int
	for _, k := range handler.Keywords {
		if len(k) > longestSize {
			longestSize = len(k)
		}
	}

	// Check for keywords
	for _, k := range DefaultKeywords {
		if strings.Contains(handler.LastTokens, k) {
			handler.KeywordDetected = true
			detectedKeyword = k
		}
	}

	if len(handler.LastTokens) > longestSize {
		handler.LastTokens = handler.LastTokens[len(handler.LastTokens)-longestSize:]
	}

	// Check for colon and set print mode.
	if handler.KeywordDetected && !handler.PrintOutput {
		// remove any other strings before the final answer
		chunk = []byte(filterFinalString(chunkStr, detectedKeyword))
		handler.PrintOutput = true
	}

	// Print the final output after the detection of keyword.
	if handler.PrintOutput {
		handler.egress <- chunk
	}
}

func filterFinalString(chunkStr, keyword string) string {
	chunkStr = strings.TrimLeft(chunkStr, " ")

	index := strings.Index(chunkStr, keyword)
	switch {
	case index != -1:
		chunkStr = chunkStr[index+len(keyword):]
	case strings.HasPrefix(chunkStr, ":"):
		chunkStr = strings.TrimPrefix(chunkStr, ":")
	}

	return strings.TrimLeft(chunkStr, " ")
}
