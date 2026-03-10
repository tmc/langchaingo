package streaming

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

func TestToolCall(t *testing.T) {
	t.Parallel()

	// Test creating a new tool call
	toolCall := NewToolCall("123", "weather", `{"location": "New York"}`)
	assert.Equal(t, "123", toolCall.ID)
	assert.Equal(t, "weather", toolCall.Name)
	assert.Equal(t, `{"location": "New York"}`, toolCall.Arguments)
	assert.Nil(t, toolCall.Reasoning)

	// Test String method
	expected := "ToolCall{ID: 123, Name: weather, Arguments: {\"location\": \"New York\"}}"
	assert.Equal(t, expected, toolCall.String())

	// Test Parse method
	args, err := toolCall.Parse()
	require.NoError(t, err)
	assert.Equal(t, "New York", args["location"])

	// Test Parse with empty arguments
	emptyToolCall := NewToolCall("123", "weather", "")
	args, err = emptyToolCall.Parse()
	require.NoError(t, err)
	assert.Empty(t, args)

	// Test Parse with invalid JSON
	invalidToolCall := NewToolCall("123", "weather", "{invalid json}")
	_, err = invalidToolCall.Parse()
	require.Error(t, err)
}

func TestNewToolCallWithReasoning(t *testing.T) {
	t.Parallel()

	reasoningContent := &reasoning.ContentReasoning{
		Content:   "Analyzing weather request for New York",
		Signature: []byte("sig123"),
	}

	// Test creating a new tool call with reasoning
	toolCall := NewToolCallWithReasoning("123", "weather", `{"location": "New York"}`, reasoningContent)
	assert.Equal(t, "123", toolCall.ID)
	assert.Equal(t, "weather", toolCall.Name)
	assert.Equal(t, `{"location": "New York"}`, toolCall.Arguments)
	assert.NotNil(t, toolCall.Reasoning)
	assert.Equal(t, "Analyzing weather request for New York", toolCall.Reasoning.Content)
	assert.Equal(t, []byte("sig123"), toolCall.Reasoning.Signature)

	// Test with nil reasoning
	toolCallNilReasoning := NewToolCallWithReasoning("456", "getTime", `{}`, nil)
	assert.Equal(t, "456", toolCallNilReasoning.ID)
	assert.Equal(t, "getTime", toolCallNilReasoning.Name)
	assert.Equal(t, `{}`, toolCallNilReasoning.Arguments)
	assert.Nil(t, toolCallNilReasoning.Reasoning)

	// Test Parse method with reasoning
	args, err := toolCall.Parse()
	require.NoError(t, err)
	assert.Equal(t, "New York", args["location"])
}

func TestChunk(t *testing.T) {
	t.Parallel()

	// Test text chunk
	textChunk := NewTextChunk("Hello, world!")
	assert.Equal(t, ChunkTypeText, textChunk.Type)
	assert.Equal(t, "Hello, world!", textChunk.Content)
	assert.Equal(t, "Text: Hello, world!", textChunk.String())

	// Test reasoning chunk
	reasoningContent := "Step 1: Analyze the problem."
	reasoningChunk := NewReasoningChunkWithContent(reasoningContent)
	assert.Equal(t, ChunkTypeReasoning, reasoningChunk.Type)
	assert.Equal(t, reasoningContent, reasoningChunk.Reasoning.Content)
	assert.Equal(t, "Reasoning: "+reasoningContent, reasoningChunk.String())

	// Extended reasoning chunk
	signature := []byte("signature")
	reasoningChunk = NewReasoningChunk(&reasoning.ContentReasoning{
		Content:   reasoningContent,
		Signature: signature,
	})
	assert.Equal(t, reasoningContent, reasoningChunk.Reasoning.Content)
	assert.Equal(t, string(signature), string(reasoningChunk.Reasoning.Signature))
	expectedReasoningChunkString := fmt.Sprintf("Reasoning: %s\nSignature: %s", reasoningContent, string(signature))
	assert.Equal(t, expectedReasoningChunkString, reasoningChunk.String())

	// Test tool call chunk
	toolCall := NewToolCall("123", "weather", `{"location": "New York"}`)
	toolCallChunk := NewToolCallChunk(toolCall)
	assert.Equal(t, ChunkTypeToolCall, toolCallChunk.Type)
	assert.Equal(t, toolCall, toolCallChunk.ToolCall)
	assert.Contains(t, toolCallChunk.String(), "ToolCall: ToolCall{ID: 123")

	// Test tool call chunk with reasoning
	toolCallReasoning := &reasoning.ContentReasoning{Content: "Analyzing weather request"}
	toolCallWithReasoning := NewToolCallWithReasoning("456", "getTime", `{"timezone": "UTC"}`, toolCallReasoning)
	toolCallChunkWithReasoning := NewToolCallChunk(toolCallWithReasoning)
	assert.Equal(t, ChunkTypeToolCall, toolCallChunkWithReasoning.Type)
	assert.Equal(t, toolCallWithReasoning, toolCallChunkWithReasoning.ToolCall)
	assert.NotNil(t, toolCallChunkWithReasoning.ToolCall.Reasoning)
	assert.Equal(t, "Analyzing weather request", toolCallChunkWithReasoning.ToolCall.Reasoning.Content)

	// Test done chunk
	doneChunk := NewDoneChunk()
	assert.Equal(t, ChunkTypeDone, doneChunk.Type)
	assert.Equal(t, "Done", doneChunk.String())

	// Test empty chunk type
	noneChunk := Chunk{Type: ChunkTypeNone}
	assert.Equal(t, "None", noneChunk.String())

	// Test Unknown chunk type string representation
	unknownChunk := Chunk{Type: "unknown"}
	assert.Equal(t, "unexpected chunk type: unknown", unknownChunk.String())
}

func TestCallWithText(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	var receivedText string

	callback := func(_ context.Context, chunk Chunk) error {
		assert.Equal(t, ChunkTypeText, chunk.Type)
		receivedText = chunk.Content
		return nil
	}

	// Test with valid text
	err := CallWithText(ctx, callback, "Hello, world!")
	require.NoError(t, err)
	assert.Equal(t, "Hello, world!", receivedText)

	// Test with empty text
	err = CallWithText(ctx, callback, "")
	require.NoError(t, err)
	assert.Equal(t, "Hello, world!", receivedText) // Should not change

	// Test with nil callback
	err = CallWithText(ctx, nil, "Hello, world!")
	require.NoError(t, err)

	// Test with error from callback
	expectedErr := errors.New("callback error")
	errorCallback := func(_ context.Context, _ Chunk) error {
		return expectedErr
	}
	err = CallWithText(ctx, errorCallback, "Hello, world!")
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

//nolint:funlen
func TestCallWithReasoning(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	var receivedReasoning *reasoning.ContentReasoning

	callback := func(_ context.Context, chunk Chunk) error {
		assert.Equal(t, ChunkTypeReasoning, chunk.Type)
		receivedReasoning = chunk.Reasoning
		return nil
	}

	// Test with valid reasoning
	reasoningContent := "Step 1: Analyze."
	err := CallWithReasoning(ctx, callback, &reasoning.ContentReasoning{Content: reasoningContent})
	require.NoError(t, err)
	assert.Equal(t, reasoningContent, receivedReasoning.Content)

	// Test with empty reasoning
	err = CallWithReasoning(ctx, callback, &reasoning.ContentReasoning{})
	require.NoError(t, err)
	assert.Equal(t, reasoningContent, receivedReasoning.Content) // Should not change

	// Test with nil callback
	err = CallWithReasoning(ctx, nil, &reasoning.ContentReasoning{Content: reasoningContent})
	require.NoError(t, err)

	// Test with multiline reasoning
	multilineReasoning := "Step 1: Define the problem.\nStep 2: Collect data.\nStep 3: Form a hypothesis."
	err = CallWithReasoning(ctx, callback, &reasoning.ContentReasoning{Content: multilineReasoning})
	require.NoError(t, err)
	assert.Equal(t, multilineReasoning, receivedReasoning.Content)

	// Test with special characters and Unicode
	specialCharsReasoning := "分析步骤: 检查数据 ¥€$£ß"
	err = CallWithReasoning(ctx, callback, &reasoning.ContentReasoning{Content: specialCharsReasoning})
	require.NoError(t, err)
	assert.Equal(t, specialCharsReasoning, receivedReasoning.Content)

	// Test with very long reasoning text
	longReasoning := "Step 1: " + strings.Repeat("Analysis and evaluation. ", 100)
	err = CallWithReasoning(ctx, callback, &reasoning.ContentReasoning{Content: longReasoning})
	require.NoError(t, err)
	assert.Equal(t, longReasoning, receivedReasoning.Content)

	// Test multiple sequential calls with accumulation
	accumulatedReasoning := ""
	accumulationCallback := func(_ context.Context, chunk Chunk) error {
		assert.Equal(t, ChunkTypeReasoning, chunk.Type)
		accumulatedReasoning += chunk.Reasoning.Content
		return nil
	}

	err = CallWithReasoning(ctx, accumulationCallback, &reasoning.ContentReasoning{Content: "First part. "})
	require.NoError(t, err)
	err = CallWithReasoning(ctx, accumulationCallback, &reasoning.ContentReasoning{Content: "Second part. "})
	require.NoError(t, err)
	err = CallWithReasoning(ctx, accumulationCallback, &reasoning.ContentReasoning{Content: "Final part."})
	require.NoError(t, err)

	assert.Equal(t, "First part. Second part. Final part.", accumulatedReasoning)

	// Test with signature
	signature := []byte("signature")
	err = CallWithReasoning(ctx, callback, &reasoning.ContentReasoning{
		Content:   reasoningContent,
		Signature: signature,
	})
	require.NoError(t, err)
	assert.Equal(t, reasoningContent, receivedReasoning.Content)
	assert.Equal(t, string(signature), string(receivedReasoning.Signature))

	// Test with error from callback
	expectedErr := errors.New("callback error")
	errorCallback := func(_ context.Context, _ Chunk) error {
		return expectedErr
	}
	err = CallWithReasoning(ctx, errorCallback, &reasoning.ContentReasoning{Content: reasoningContent})
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestCallWithReasoningContent(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	var receivedReasoning *reasoning.ContentReasoning

	callback := func(_ context.Context, chunk Chunk) error {
		assert.Equal(t, ChunkTypeReasoning, chunk.Type)
		receivedReasoning = chunk.Reasoning
		return nil
	}

	// Test with valid reasoning content
	reasoningContent := "Step 1: Analyze the problem."
	err := CallWithReasoningContent(ctx, callback, reasoningContent)
	require.NoError(t, err)
	assert.Equal(t, reasoningContent, receivedReasoning.Content)
	assert.Nil(t, receivedReasoning.Signature)

	// Test with empty content
	err = CallWithReasoningContent(ctx, callback, "")
	require.NoError(t, err)
	assert.Equal(t, reasoningContent, receivedReasoning.Content) // Should not change

	// Test with nil callback
	err = CallWithReasoningContent(ctx, nil, "Some reasoning")
	require.NoError(t, err)

	// Test with error from callback
	expectedErr := errors.New("callback error")
	errorCallback := func(_ context.Context, _ Chunk) error {
		return expectedErr
	}
	err = CallWithReasoningContent(ctx, errorCallback, "Some reasoning")
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestCallWithToolCall(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	var receivedToolCall ToolCall

	callback := func(_ context.Context, chunk Chunk) error {
		assert.Equal(t, ChunkTypeToolCall, chunk.Type)
		receivedToolCall = chunk.ToolCall
		return nil
	}

	toolCall := NewToolCall("123", "weather", `{"location": "New York"}`)

	// Test with valid tool call
	err := CallWithToolCall(ctx, callback, toolCall)
	require.NoError(t, err)
	assert.Equal(t, toolCall, receivedToolCall)

	// Test with tool call with reasoning
	reasoningContent := &reasoning.ContentReasoning{
		Content:   "Analyzing weather request",
		Signature: []byte("sig123"),
	}
	toolCallWithReasoning := NewToolCallWithReasoning("456", "getTime", `{"timezone": "UTC"}`, reasoningContent)
	err = CallWithToolCall(ctx, callback, toolCallWithReasoning)
	require.NoError(t, err)
	assert.Equal(t, toolCallWithReasoning, receivedToolCall)
	assert.NotNil(t, receivedToolCall.Reasoning)
	assert.Equal(t, "Analyzing weather request", receivedToolCall.Reasoning.Content)

	// Test with missing ID
	invalidToolCall := NewToolCall("", "weather", `{"location": "New York"}`)
	err = CallWithToolCall(ctx, callback, invalidToolCall)
	require.Error(t, err)
	assert.Equal(t, ErrToolCallIDRequired, err)

	// Test with missing Name
	invalidToolCall = NewToolCall("123", "", `{"location": "New York"}`)
	err = CallWithToolCall(ctx, callback, invalidToolCall)
	require.Error(t, err)
	assert.Equal(t, ErrToolCallNameRequired, err)

	// Test with nil callback
	err = CallWithToolCall(ctx, nil, toolCall)
	require.NoError(t, err)

	// Test with error from callback
	expectedErr := errors.New("callback error")
	errorCallback := func(_ context.Context, _ Chunk) error {
		return expectedErr
	}
	err = CallWithToolCall(ctx, errorCallback, toolCall)
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestAppendToolCall(t *testing.T) {
	t.Parallel()

	src := NewToolCall("123", "weather", `{"location": "New York"}`)
	dst := ToolCall{
		ID:        "",
		Name:      "",
		Arguments: "",
	}

	AppendToolCall(src, &dst)

	assert.Equal(t, "123", dst.ID)
	assert.Equal(t, "weather", dst.Name)
	assert.Equal(t, `{"location": "New York"}`, dst.Arguments)

	// Test appending to existing arguments
	src = NewToolCall("123", "weather", `, "unit": "celsius"}`)
	AppendToolCall(src, &dst)

	assert.Equal(t, "123", dst.ID)
	assert.Equal(t, "weather", dst.Name)
	assert.Equal(t, `{"location": "New York"}, "unit": "celsius"}`, dst.Arguments)

	// Test appending with reasoning - note that AppendToolCall does NOT append reasoning
	// It only sets ID, Name and appends Arguments
	reasoningContent := &reasoning.ContentReasoning{Content: "Analyzing request"}
	srcWithReasoning := NewToolCallWithReasoning("456", "getTime", `{"timezone": "UTC"}`, reasoningContent)
	dstWithReasoning := ToolCall{}

	AppendToolCall(srcWithReasoning, &dstWithReasoning)

	assert.Equal(t, "456", dstWithReasoning.ID)
	assert.Equal(t, "getTime", dstWithReasoning.Name)
	assert.Equal(t, `{"timezone": "UTC"}`, dstWithReasoning.Arguments)
	// AppendToolCall doesn't copy reasoning field
	assert.Nil(t, dstWithReasoning.Reasoning)
}

func TestCallWithDone(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	var receivedDone bool

	callback := func(_ context.Context, chunk Chunk) error {
		assert.Equal(t, ChunkTypeDone, chunk.Type)
		receivedDone = true
		return nil
	}

	// Test with valid callback
	err := CallWithDone(ctx, callback)
	require.NoError(t, err)
	assert.True(t, receivedDone)

	// Test with nil callback
	err = CallWithDone(ctx, nil)
	require.NoError(t, err)

	// Test with error from callback
	expectedErr := errors.New("callback error")
	errorCallback := func(_ context.Context, _ Chunk) error {
		return expectedErr
	}
	err = CallWithDone(ctx, errorCallback)
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

//nolint:funlen
func TestIntegration(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	// Track received chunks
	var textChunks []string
	var reasoningChunks []string
	var toolCalls []ToolCall
	var doneReceived bool

	callback := func(_ context.Context, chunk Chunk) error {
		switch chunk.Type {
		case ChunkTypeNone:
			// Just ensure we can handle this type
		case ChunkTypeText:
			textChunks = append(textChunks, chunk.Content)
		case ChunkTypeReasoning:
			reasoningChunks = append(reasoningChunks, chunk.Reasoning.Content)
		case ChunkTypeToolCall:
			toolCalls = append(toolCalls, chunk.ToolCall)
		case ChunkTypeDone:
			doneReceived = true
		}
		return nil
	}

	// Simulate streaming text chunks
	_ = CallWithText(ctx, callback, "Hello")
	_ = CallWithText(ctx, callback, ", ")
	_ = CallWithText(ctx, callback, "world!")

	// Simulate streaming reasoning chunks
	_ = CallWithReasoning(ctx, callback, &reasoning.ContentReasoning{Content: "Step 1: Start with a greeting."})
	_ = CallWithReasoningContent(ctx, callback, "Step 2: Add punctuation.")
	_ = CallWithReasoning(ctx, callback, &reasoning.ContentReasoning{Content: "Step 3: Complete the phrase."})

	// Simulate streaming tool calls
	weatherTool := NewToolCall("123", "weather", `{"location": `)
	_ = CallWithToolCall(ctx, callback, weatherTool)

	weatherTool2 := NewToolCall("123", "weather", `"New York"}`)
	_ = CallWithToolCall(ctx, callback, weatherTool2)

	reasoningContent := &reasoning.ContentReasoning{Content: "Checking current time"}
	timeTool := NewToolCallWithReasoning("456", "getTime", `{}`, reasoningContent)
	_ = CallWithToolCall(ctx, callback, timeTool)

	// Signal stream completion
	_ = CallWithDone(ctx, callback)

	// Verify results
	assert.Equal(t, []string{"Hello", ", ", "world!"}, textChunks)
	assert.Equal(t, []string{
		"Step 1: Start with a greeting.",
		"Step 2: Add punctuation.",
		"Step 3: Complete the phrase.",
	}, reasoningChunks)

	require.Len(t, toolCalls, 3)
	assert.Equal(t, "123", toolCalls[0].ID)
	assert.Equal(t, "weather", toolCalls[0].Name)
	assert.Equal(t, `{"location": `, toolCalls[0].Arguments)
	assert.Nil(t, toolCalls[0].Reasoning)

	assert.Equal(t, "123", toolCalls[1].ID)
	assert.Equal(t, "weather", toolCalls[1].Name)
	assert.Equal(t, `"New York"}`, toolCalls[1].Arguments)
	assert.Nil(t, toolCalls[1].Reasoning)

	assert.Equal(t, "456", toolCalls[2].ID)
	assert.Equal(t, "getTime", toolCalls[2].Name)
	assert.Equal(t, `{}`, toolCalls[2].Arguments)
	assert.NotNil(t, toolCalls[2].Reasoning)
	assert.Equal(t, "Checking current time", toolCalls[2].Reasoning.Content)

	// Verify done was received
	assert.True(t, doneReceived)
}

func TestToolCallMarshalUnmarshal(t *testing.T) {
	t.Parallel()

	toolCall := NewToolCall("123", "weather", `{"location": "New York", "units": "celsius"}`)

	// Test marshaling
	data, err := json.Marshal(toolCall)
	require.NoError(t, err)

	// Test unmarshaling
	var unmarshaled ToolCall
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, toolCall, unmarshaled)

	// Test marshaling tool call with reasoning
	reasoningContent := &reasoning.ContentReasoning{
		Content:   "Analyzing request",
		Signature: []byte("sig123"),
	}
	toolCallWithReasoning := NewToolCallWithReasoning("456", "getTime", `{"timezone": "UTC"}`, reasoningContent)

	data, err = json.Marshal(toolCallWithReasoning)
	require.NoError(t, err)

	var unmarshaledWithReasoning ToolCall
	err = json.Unmarshal(data, &unmarshaledWithReasoning)
	require.NoError(t, err)

	assert.Equal(t, toolCallWithReasoning.ID, unmarshaledWithReasoning.ID)
	assert.Equal(t, toolCallWithReasoning.Name, unmarshaledWithReasoning.Name)
	assert.Equal(t, toolCallWithReasoning.Arguments, unmarshaledWithReasoning.Arguments)
	assert.NotNil(t, unmarshaledWithReasoning.Reasoning)
	assert.Equal(t, "Analyzing request", unmarshaledWithReasoning.Reasoning.Content)
	assert.Equal(t, []byte("sig123"), unmarshaledWithReasoning.Reasoning.Signature)
}

func TestChunkMarshalUnmarshal(t *testing.T) {
	t.Parallel()

	// Test text chunk
	textChunk := NewTextChunk("Hello, world!")
	data, err := json.Marshal(textChunk)
	require.NoError(t, err)

	var unmarshaledTextChunk Chunk
	err = json.Unmarshal(data, &unmarshaledTextChunk)
	require.NoError(t, err)
	assert.Equal(t, textChunk, unmarshaledTextChunk)

	// Test reasoning chunk
	reasoningChunk := NewReasoningChunk(&reasoning.ContentReasoning{
		Content:   "Step 1: Analyze.",
		Signature: []byte("signature"),
	})
	data, err = json.Marshal(reasoningChunk)
	require.NoError(t, err)

	var unmarshaledReasoningChunk Chunk
	err = json.Unmarshal(data, &unmarshaledReasoningChunk)
	require.NoError(t, err)
	assert.Equal(t, reasoningChunk, unmarshaledReasoningChunk)

	// Test reasoning chunk with content
	reasoningContent := "Step 1: Analyze."
	reasoningChunk = NewReasoningChunkWithContent(reasoningContent)
	data, err = json.Marshal(reasoningChunk)
	require.NoError(t, err)

	var unmarshaledReasoningChunkWithContent Chunk
	err = json.Unmarshal(data, &unmarshaledReasoningChunkWithContent)
	require.NoError(t, err)
	assert.Equal(t, reasoningChunk, unmarshaledReasoningChunkWithContent)

	// Test tool call chunk
	toolCall := NewToolCall("123", "weather", `{"location": "New York"}`)
	toolCallChunk := NewToolCallChunk(toolCall)
	data, err = json.Marshal(toolCallChunk)
	require.NoError(t, err)

	var unmarshaledToolCallChunk Chunk
	err = json.Unmarshal(data, &unmarshaledToolCallChunk)
	require.NoError(t, err)
	assert.Equal(t, toolCallChunk, unmarshaledToolCallChunk)

	// Test done chunk
	doneChunk := NewDoneChunk()
	data, err = json.Marshal(doneChunk)
	require.NoError(t, err)

	var unmarshaledDoneChunk Chunk
	err = json.Unmarshal(data, &unmarshaledDoneChunk)
	require.NoError(t, err)
	assert.Equal(t, doneChunk, unmarshaledDoneChunk)
}
