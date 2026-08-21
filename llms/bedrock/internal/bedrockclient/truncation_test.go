package bedrockclient

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
)

func TestEveryFamilySpellsTruncationInAWayWeRecognise(t *testing.T) {
	t.Parallel()

	for _, reason := range []string{
		AmazonCompletionReasonMaxTokens,
		AnthropicCompletionReasonMaxTokens,
		Ai21CompletionReasonLength,
		CohereCompletionReasonMaxTokens,
		MetaCompletionReasonLength,
		NovaCompletionReasonMaxTokens,
		string(types.StopReasonMaxTokens),
	} {
		assert.True(t, llms.IsTruncated(reason), "%q means the output limit was reached", reason)
	}

	for _, reason := range []string{
		AmazonCompletionReasonFinish,
		AmazonCompletionReasonContentFiltered,
		AnthropicCompletionReasonEndTurn,
		AnthropicCompletionReasonStopSequence,
		Ai21CompletionReasonStop,
		Ai21CompletionReasonEndOfText,
		CohereCompletionReasonComplete,
		CohereCompletionReasonError,
		CohereCompletionReasonErrorToxic,
		MetaCompletionReasonStop,
		NovaCompletionReasonEndTurn,
		NovaCompletionReasonStopSequence,
		NovaCompletionReasonContentFiltered,
		string(types.StopReasonEndTurn),
		string(types.StopReasonToolUse),
		string(types.StopReasonStopSequence),
		string(types.StopReasonGuardrailIntervened),
	} {
		assert.False(t, llms.IsTruncated(reason), "%q does not mean truncation", reason)
	}
}

func TestConverseReportsTruncation(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name          string
		stopReason    types.StopReason
		wantTruncated bool
	}{
		{"out of budget", types.StopReasonMaxTokens, true},
		{"finished on its own", types.StopReasonEndTurn, false},
		{"asked for a tool", types.StopReasonToolUse, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockClient := &MockBedrockRuntimeClient{}
			mockClient.On("Converse", mock.Anything, mock.Anything, mock.Anything).Return(
				&bedrockruntime.ConverseOutput{
					StopReason: tc.stopReason,
					Output: &types.ConverseOutputMemberMessage{
						Value: types.Message{
							Role:    types.ConversationRoleAssistant,
							Content: []types.ContentBlock{&types.ContentBlockMemberText{Value: "half an ans"}},
						},
					},
				}, nil)

			resp, err := NewConverseClient(mockClient).CreateCompletionConverse(t.Context(), &ConverseInput{
				ModelID:  "anthropic.claude-3-sonnet-20240229-v1:0",
				Messages: []Message{{Role: llms.ChatMessageTypeHuman, Content: "hi", Type: "text"}},
			})
			require.NoError(t, err)
			assert.Equal(t, tc.wantTruncated, resp.Choices[0].Truncated)
			assert.Equal(t, string(tc.stopReason), resp.Choices[0].StopReason)
		})
	}
}
