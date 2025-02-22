package agents

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/tools"
)

func TestMrklPromptEdit(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		promptPrefix                     string
		formatInstructions               string
		promptSuffix                     string
		promptPrefixInputVariables       []string
		formatInstructionsInputVariables []string
		promptSuffixInputVariables       []string
		expectPromptTemplate             prompts.PromptTemplate
	}{
		{
			promptPrefix:                     "this {{.top}}",
			formatInstructions:               "this {{.instruction}}",
			promptSuffix:                     "this {{.content}} and {{.end}}",
			promptPrefixInputVariables:       []string{"top"},
			formatInstructionsInputVariables: []string{"instruction"},
			promptSuffixInputVariables:       []string{"content", "end"},
			expectPromptTemplate: prompts.NewPromptTemplate(strings.Join([]string{
				"this {{.top}}",
				"this {{.instruction}}",
				"this {{.content}} and {{.end}}",
			}, "\n\n"),
				[]string{"top", "instruction", "content", "end"},
			),
		},
		{
			promptPrefix:                     "",
			formatInstructions:               "",
			promptSuffix:                     "",
			promptPrefixInputVariables:       []string{},
			formatInstructionsInputVariables: []string{},
			promptSuffixInputVariables:       []string{},
			expectPromptTemplate: prompts.NewPromptTemplate(strings.Join([]string{
				_defaultMrklPrefix,
				_defaultMrklFormatInstructions,
				_defaultMrklSuffix,
			}, "\n\n"),
				[]string{"today", "agent_scratchpad", "input"},
			),
		},
	}
	for k, v := range testCases {
		expectPromptTemp := v.expectPromptTemplate
		expectPromptTemp.PartialVariables = map[string]any{
			"tool_names":        "",
			"tool_descriptions": "",
		}
		testCases[k].expectPromptTemplate = expectPromptTemp
	}
	for _, tc := range testCases {
		opt := mrklDefaultOptions()
		if tc.promptPrefix != "" {
			WithPromptPrefix(tc.promptPrefix)(&opt)
		}
		if tc.promptSuffix != "" {
			WithPromptSuffix(tc.promptSuffix)(&opt)
		}
		if tc.formatInstructions != "" {
			WithPromptFormatInstructions(tc.formatInstructions)(&opt)
		}
		if len(tc.promptPrefixInputVariables) != 0 {
			WithPromptPrefixInputVariables(tc.promptPrefixInputVariables)(&opt)
		}
		if len(tc.promptSuffixInputVariables) != 0 {
			WithPromptSuffixInputVariables(tc.promptSuffixInputVariables)(&opt)
		}
		if len(tc.formatInstructionsInputVariables) != 0 {
			WithPromptInstructionsInputVariables(tc.formatInstructionsInputVariables)(&opt)
		}

		temp := opt.getMrklPrompt([]tools.Tool{})
		require.Equal(t, tc.expectPromptTemplate, temp)
	}
}

func TestConversationPromptEdit(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		promptPrefix                     string
		formatInstructions               string
		promptSuffix                     string
		promptPrefixInputVariables       []string
		formatInstructionsInputVariables []string
		promptSuffixInputVariables       []string
		expectPromptTemplate             prompts.PromptTemplate
	}{
		{
			promptPrefix:                     "this {{.top}}",
			formatInstructions:               "this {{.instruction}}",
			promptSuffix:                     "this {{.history}} and {{.end}}",
			promptPrefixInputVariables:       []string{"top"},
			formatInstructionsInputVariables: []string{"instruction"},
			promptSuffixInputVariables:       []string{"content", "end"},
			expectPromptTemplate: prompts.NewPromptTemplate(strings.Join([]string{
				"this {{.top}}",
				"this {{.instruction}}",
				"this {{.history}} and {{.end}}",
			}, "\n\n"),
				[]string{"top", "instruction", "content", "end"},
			),
		},
		{
			promptPrefix:                     "",
			formatInstructions:               "",
			promptSuffix:                     "",
			promptPrefixInputVariables:       []string{},
			formatInstructionsInputVariables: []string{},
			promptSuffixInputVariables:       []string{},
			expectPromptTemplate: prompts.NewPromptTemplate(strings.Join([]string{
				_defaultConversationalPrefix,
				_defaultConversationalFormatInstructions,
				_defaultConversationalSuffix,
			}, "\n\n"),
				[]string{"agent_scratchpad", "input"},
			),
		},
	}
	for k, v := range testCases {
		expectPromptTemp := v.expectPromptTemplate
		expectPromptTemp.PartialVariables = map[string]any{
			"tool_names":        "",
			"tool_descriptions": "",
			"history":           "",
		}
		testCases[k].expectPromptTemplate = expectPromptTemp
	}
	for _, tc := range testCases {
		opt := conversationalDefaultOptions()
		if tc.promptPrefix != "" {
			WithPromptPrefix(tc.promptPrefix)(&opt)
		}
		if tc.promptSuffix != "" {
			WithPromptSuffix(tc.promptSuffix)(&opt)
		}
		if tc.formatInstructions != "" {
			WithPromptFormatInstructions(tc.formatInstructions)(&opt)
		}
		if len(tc.promptPrefixInputVariables) != 0 {
			WithPromptPrefixInputVariables(tc.promptPrefixInputVariables)(&opt)
		}
		if len(tc.promptSuffixInputVariables) != 0 {
			WithPromptSuffixInputVariables(tc.promptSuffixInputVariables)(&opt)
		}
		if len(tc.formatInstructionsInputVariables) != 0 {
			WithPromptInstructionsInputVariables(tc.formatInstructionsInputVariables)(&opt)
		}

		temp := opt.getConversationalPrompt([]tools.Tool{})
		require.Equal(t, tc.expectPromptTemplate, temp)
	}
}
