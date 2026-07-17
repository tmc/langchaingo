package bedrockclient

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

type ValidStruct struct {
	Field1 string `document:"field1"`
	Field2 int    `document:"field2"`
}

type InvalidStruct struct {
	Field1 string `document:"field1"`
	Field2 int    // missing document tag
}

type NestedValidStruct struct {
	Nested ValidStruct `document:"nested"`
	Value  string      `document:"value"`
}

type NestedInvalidStruct struct {
	Nested InvalidStruct `document:"nested"`
	Value  string        `document:"value"`
}

type StructWithSlice struct {
	Items []string `document:"items"`
	Name  string   `document:"name"`
}

type StructWithMap struct {
	Data map[string]int `document:"data"`
	Name string         `document:"name"`
}

func TestIsSmithyValidObject(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		// Simple types
		{"nil", nil, true},
		{"string", "test", true},
		{"int", 42, true},
		{"bool", true, true},
		{"float64", 3.14, true},

		// Valid structs
		{"valid struct", ValidStruct{Field1: "test", Field2: 42}, true},
		{"valid struct pointer", &ValidStruct{Field1: "test", Field2: 42}, true},
		{"nil pointer", (*ValidStruct)(nil), true},
		{"nested valid struct", NestedValidStruct{Nested: ValidStruct{}, Value: "test"}, true},

		// Invalid structs
		{"invalid struct", InvalidStruct{Field1: "test", Field2: 42}, false},
		{"invalid struct pointer", &InvalidStruct{Field1: "test", Field2: 42}, false},
		{"nested invalid struct", NestedInvalidStruct{Nested: InvalidStruct{}, Value: "test"}, false},

		// Valid slices
		{"empty slice", []string{}, true},
		{"slice of simple types", []string{"test1", "test2", "test3"}, true},
		{"slice of integers", []int{1, 2, 3}, true},
		{"slice of valid structs", []ValidStruct{{Field1: "a", Field2: 1}, {Field1: "b", Field2: 2}}, true},
		{"slice of valid struct pointers", []*ValidStruct{{Field1: "a", Field2: 1}, {Field1: "b", Field2: 2}}, true},
		{"slice with nil pointers", []*ValidStruct{nil, {Field1: "test", Field2: 42}}, true},

		// Invalid slices
		{"slice of invalid structs", []InvalidStruct{{Field1: "test", Field2: 42}}, false},
		{"slice of invalid struct pointers", []*InvalidStruct{{Field1: "test", Field2: 42}}, false},
		{"mixed slice with invalid struct", []any{ValidStruct{Field1: "test", Field2: 42}, InvalidStruct{Field1: "test", Field2: 42}}, false},

		// Valid maps
		{"empty map", map[string]int{}, true},
		{"map with simple types", map[string]int{"key1": 1, "key2": 2}, true},
		{"map with int keys", map[int]string{1: "one", 2: "two"}, true},
		{"map with valid struct values", map[string]ValidStruct{"a": {Field1: "test1", Field2: 1}, "b": {Field1: "test2", Field2: 2}}, true},
		{"map with valid struct pointer values", map[string]*ValidStruct{"a": {Field1: "test1", Field2: 1}, "b": nil}, true},

		// Invalid maps
		{"map with invalid struct values", map[string]InvalidStruct{"a": {Field1: "test", Field2: 42}}, false},
		{"map with complex values", map[string][]InvalidStruct{"key": {{Field1: "test", Field2: 42}}}, false},

		// Nested collections
		{"slice of slices", [][]string{{"a", "b"}, {"c", "d"}}, true},
		{"slice of maps", []map[string]int{{"key1": 1}, {"key2": 2}}, true},
		{"map of slices", map[string][]int{"key1": {1, 2}, "key2": {3, 4}}, true},
		{"map of maps", map[string]map[string]int{"outer": {"inner": 42}}, true},

		// Invalid nested collections
		{"map of invalid slices", map[string][]InvalidStruct{"key": {{Field1: "test", Field2: 42}}}, false},

		// Structs with collections
		{"struct with valid slice", StructWithSlice{Items: []string{"a", "b"}, Name: "test"}, true},
		{"struct with valid map", StructWithMap{Data: map[string]int{"key": 42}, Name: "test"}, true},

		// Non-simple, non-struct, non-collection types
		{"channel", make(chan int), false},
		{"function", func() {}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSmithyValidObject(tt.input)
			if result != tt.expected {
				t.Errorf("isSmithyValidObject(%v) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

const bedrockSOSchema = `{"type":"object","properties":{"answer":{"type":"string"}},"required":["answer"],"additionalProperties":false}`

func soConfig() *llms.StructuredOutputConfig {
	return &llms.StructuredOutputConfig{
		Name:        "answer_schema",
		Description: "an answer",
		Schema:      []byte(bedrockSOSchema),
	}
}

func textOutput(text string, reason types.StopReason) *bedrockruntime.ConverseOutput {
	return &bedrockruntime.ConverseOutput{
		StopReason: reason,
		Output: &types.ConverseOutputMemberMessage{
			Value: types.Message{
				Role:    types.ConversationRoleAssistant,
				Content: []types.ContentBlock{&types.ContentBlockMemberText{Value: text}},
			},
		},
	}
}

func TestConverse_StructuredOutput_TextFormat(t *testing.T) {
	mockClient := &MockBedrockRuntimeClient{}
	client := NewConverseClient(mockClient)

	var captured *bedrockruntime.ConverseInput
	mockClient.On("Converse", mock.Anything, mock.MatchedBy(func(in *bedrockruntime.ConverseInput) bool {
		captured = in
		return true
	}), mock.Anything).Return(textOutput(`{"answer":"ok"}`, types.StopReasonEndTurn), nil)

	input := &ConverseInput{
		ModelID:          "anthropic.claude-sonnet-4-5-v1:0",
		Messages:         []Message{{Role: llms.ChatMessageTypeHuman, Content: "hi", Type: "text"}},
		StructuredOutput: soConfig(),
	}

	resp, err := client.CreateCompletionConverse(t.Context(), input)
	require.NoError(t, err)
	require.Equal(t, `{"answer":"ok"}`, resp.Choices[0].Content)
	require.Equal(t, string(types.StopReasonEndTurn), resp.Choices[0].StopReason)

	require.NotNil(t, captured.OutputConfig)
	require.NotNil(t, captured.OutputConfig.TextFormat)
	assert.Equal(t, types.OutputFormatTypeJsonSchema, captured.OutputConfig.TextFormat.Type)
	js, ok := captured.OutputConfig.TextFormat.Structure.(*types.OutputFormatStructureMemberJsonSchema)
	require.True(t, ok)
	require.NotNil(t, js.Value.Schema)
	// Schema is a JSON string passed verbatim, no second encoding.
	assert.JSONEq(t, bedrockSOSchema, *js.Value.Schema)
	assert.Equal(t, "answer_schema", *js.Value.Name)
	assert.Equal(t, "an answer", *js.Value.Description)
}

func TestConverse_StructuredOutput_RejectsOpenObject(t *testing.T) {
	mockClient := &MockBedrockRuntimeClient{}
	client := NewConverseClient(mockClient)
	// No Converse call is expected: the open object schema is rejected before send.

	input := &ConverseInput{
		ModelID:  "anthropic.claude-sonnet-4-5-v1:0",
		Messages: []Message{{Role: llms.ChatMessageTypeHuman, Content: "hi", Type: "text"}},
		StructuredOutput: &llms.StructuredOutputConfig{
			Name:   "open",
			Schema: []byte(`{"type":"object","properties":{"answer":{"type":"string"}},"required":["answer"]}`),
		},
	}

	_, err := client.CreateCompletionConverse(t.Context(), input)
	if !errors.Is(err, llms.ErrStructuredOutputConfig) {
		t.Fatalf("open object schema must be rejected with ErrStructuredOutputConfig, got %v", err)
	}
	mockClient.AssertNotCalled(t, "Converse")
}

func TestConverse_StructuredOutput_CoexistsWithReasoning(t *testing.T) {
	mockClient := &MockBedrockRuntimeClient{}
	client := NewConverseClient(mockClient)

	var captured *bedrockruntime.ConverseInput
	mockClient.On("Converse", mock.Anything, mock.MatchedBy(func(in *bedrockruntime.ConverseInput) bool {
		captured = in
		return true
	}), mock.Anything).Return(textOutput(`{"answer":"ok"}`, types.StopReasonEndTurn), nil)

	input := &ConverseInput{
		ModelID:          "anthropic.claude-sonnet-5-v1:0",
		Messages:         []Message{{Role: llms.ChatMessageTypeHuman, Content: "hi", Type: "text"}},
		ReasoningConfig:  &llms.ReasoningConfig{Mode: llms.ReasoningOn, Adaptive: true, Effort: llms.ReasoningHigh},
		StructuredOutput: soConfig(),
	}

	_, err := client.CreateCompletionConverse(t.Context(), input)
	require.NoError(t, err)
	// Native OutputConfig carries the schema; reasoning stays in the provider fields.
	require.NotNil(t, captured.OutputConfig.TextFormat)
	require.NotNil(t, captured.AdditionalModelRequestFields, "reasoning must remain in additionalModelRequestFields")
}

func TestConverse_StructuredOutput_Validation(t *testing.T) {
	cases := []struct {
		name    string
		text    string
		reason  types.StopReason
		wantErr bool
	}{
		{"valid end_turn", `{"answer":"ok"}`, types.StopReasonEndTurn, false},
		{"valid stop_sequence", `{"answer":"ok"}`, types.StopReasonStopSequence, false},
		{"invalid stop_sequence", `{"answer":5}`, types.StopReasonStopSequence, true},
		{"invalid end_turn", `{"answer":5}`, types.StopReasonEndTurn, true},
		{"tool_use skipped", `not json`, types.StopReasonToolUse, false},
		{"max_tokens skipped", `{"answer"`, types.StopReasonMaxTokens, false},
		{"guardrail skipped", `blocked`, types.StopReasonGuardrailIntervened, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockBedrockRuntimeClient{}
			client := NewConverseClient(mockClient)
			mockClient.On("Converse", mock.Anything, mock.Anything, mock.Anything).
				Return(textOutput(tc.text, tc.reason), nil)

			input := &ConverseInput{
				ModelID:          "anthropic.claude-sonnet-4-5-v1:0",
				Messages:         []Message{{Role: llms.ChatMessageTypeHuman, Content: "hi", Type: "text"}},
				StructuredOutput: soConfig(),
			}
			_, err := client.CreateCompletionConverse(t.Context(), input)
			if tc.wantErr {
				var ve *llms.ErrStructuredOutputValidation
				require.True(t, errors.As(err, &ve), "want validation error, got %v", err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConverse_StructuredOutput_StreamCopiesOutputConfig(t *testing.T) {
	mockClient := &MockBedrockRuntimeClient{}
	client := NewConverseClient(mockClient)

	var captured *bedrockruntime.ConverseStreamInput
	mockClient.On("ConverseStream", mock.Anything, mock.MatchedBy(func(in *bedrockruntime.ConverseStreamInput) bool {
		captured = in
		return true
	}), mock.Anything).Return(&bedrockruntime.ConverseStreamOutput{}, errors.New("stop early"))

	input := &ConverseInput{
		ModelID:          "anthropic.claude-sonnet-4-5-v1:0",
		Messages:         []Message{{Role: llms.ChatMessageTypeHuman, Content: "hi", Type: "text"}},
		StructuredOutput: soConfig(),
		StreamingFunc:    func(_ context.Context, _ streaming.Chunk) error { return nil },
	}
	_, _ = client.CreateCompletionConverse(t.Context(), input)

	require.NotNil(t, captured, "ConverseStream must be called")
	require.NotNil(t, captured.OutputConfig, "stream input must copy OutputConfig")
	require.NotNil(t, captured.OutputConfig.TextFormat)
}

func TestLegacyAnthropic_StructuredOutput_FormatAndEffort(t *testing.T) {
	t.Parallel()
	input := anthropicTextGenerationInput{
		OutputConfig: &anthropicOutputConfig{Effort: "high"},
	}
	require.NoError(t, applyAnthropicStructuredOutput(&input, "anthropic.claude-sonnet-4-5-v1:0", soConfig()))

	got := marshalAnthropicInput(t, input)
	oc, _ := got["output_config"].(map[string]any)
	require.NotNil(t, oc)
	assert.Equal(t, "high", oc["effort"], "effort must survive alongside format")
	format, ok := oc["format"].(map[string]any)
	require.True(t, ok, "format must be present")
	assert.Equal(t, "json_schema", format["type"])
	// Schema must be an embedded JSON object, not a double-encoded string.
	_, isObject := format["schema"].(map[string]any)
	assert.True(t, isObject, "schema must not be double-encoded, got %T", format["schema"])
}

func TestLegacyAnthropic_StructuredOutput_UnsupportedModel(t *testing.T) {
	t.Parallel()
	input := anthropicTextGenerationInput{}
	err := applyAnthropicStructuredOutput(&input, "anthropic.claude-3-sonnet-20240229-v1:0", soConfig())
	var unsup *llms.ErrStructuredOutputUnsupported
	require.True(t, errors.As(err, &unsup), "legacy 3.x must be unsupported, got %v", err)
}
