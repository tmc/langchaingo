package agents

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/i18n"
	"github.com/tmc/langchaingo/schema"
)

func TestMRKLOutputParser(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input           string
		expectedActions []schema.AgentAction
		expectedFinish  *schema.AgentFinish
		expectedErr     error
	}{
		{
			input: "Action:  foo Action Input: bar",
			expectedActions: []schema.AgentAction{{
				Tool:      "foo",
				ToolInput: "bar",
				Log:       "Action:  foo Action Input: bar",
			}},
			expectedFinish: nil,
			expectedErr:    nil,
		},
		{
			input: "Action: foo\nAction Input:\nbar\nbaz",
			expectedActions: []schema.AgentAction{{
				Tool:      "foo",
				ToolInput: "bar\nbaz",
				Log:       "Action: foo\nAction Input:\nbar\nbaz",
			}},
			expectedFinish: nil,
			expectedErr:    nil,
		},
	}

	lang := i18n.EN
	a := OneShotZeroAgent{
		FinalAnswer: i18n.AgentsMustPhrase(lang, "mrkl final answer"),
		Lang:        lang,
	}
	for _, tc := range testCases {
		actions, finish, err := a.parseOutput(tc.input)
		require.ErrorIs(t, tc.expectedErr, err)
		require.Equal(t, tc.expectedActions, actions)
		require.Equal(t, tc.expectedFinish, finish)
	}
}
