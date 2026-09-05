package bedrockclient

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

func reasoningTextOf(t *testing.T, blocks []types.ContentBlock) types.ReasoningTextBlock {
	t.Helper()

	for _, block := range blocks {
		wrapper, ok := block.(*types.ContentBlockMemberReasoningContent)
		if !ok {
			continue
		}
		text, ok := wrapper.Value.(*types.ReasoningContentBlockMemberReasoningText)
		require.True(t, ok, "the reasoning block must carry reasoningText")
		return text.Value
	}

	t.Fatal("no reasoning block was emitted")
	return types.ReasoningTextBlock{}
}

func TestSignatureOnlyReasoningKeepsTheTextField(t *testing.T) {
	t.Parallel()

	signatureOnly := &reasoning.ContentReasoning{Signature: []byte("sig-only")}

	t.Run("accumulated message", func(t *testing.T) {
		t.Parallel()

		acc := &aiMessageAccumulator{reasoning: signatureOnly}

		block := reasoningTextOf(t, acc.build().Content)
		require.NotNil(t, block.Text)
		assert.Empty(t, *block.Text)
		require.NotNil(t, block.Signature)
		assert.Equal(t, "sig-only", *block.Signature)
	})
}
