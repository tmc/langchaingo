package callbacks

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilterFinalString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		keyword  string
		inputStr string
		expected string
	}{
		{
			keyword:  "Final Answer:",
			inputStr: "This is a correct final string.",
			expected: "This is a correct final string.",
		},
		{
			keyword:  "Final Answer:",
			inputStr: " some other text above.\nFinal Answer: This is a correct final string.",
			expected: "This is a correct final string.",
		},
		{
			keyword:  "Final Answer:",
			inputStr: " another text before. Final Answer: This is a correct final string.",
			expected: "This is a correct final string.",
		},
		{
			keyword:  "Final Answer:",
			inputStr: `   :    This is a correct final string.`,
			expected: "This is a correct final string.",
		},
		{
			keyword:  "Customed KeyWord_2:",
			inputStr: " some other text above.\nSome Customed KeyWord_2: This is a correct final string.",
			expected: "This is a correct final string.",
		},
		{
			keyword:  "Customed KeyWord_$#@-123:",
			inputStr: " another text before keyword. Some Customed KeyWord_$#@-123: This is a correct final string.",
			expected: "This is a correct final string.",
		},
	}

	for _, tc := range cases {
		filteredStr := filterFinalString(tc.inputStr, tc.keyword)
		require.Equal(t, tc.expected, filteredStr)
	}
}
