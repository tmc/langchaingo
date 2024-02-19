package callbacks

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type filterFinalStringTestCase struct {
	inputStr string
}

func TestFilterFinalString(t *testing.T) {
	t.Parallel()

	keywork := "Final Answer:"

	// correct final string
	correctStr := "This is a correct final string."

	extraStrAbore := fmt.Sprintf(" some other text.\nFinal Answer: %s", correctStr)
	extraStrBefore := fmt.Sprintf(" another text. Final Answer: %s", correctStr)
	extraColonWithSpacesBefore := fmt.Sprintf(`   :    %s`, correctStr)

	testCases := []filterFinalStringTestCase{
		{correctStr},
		{extraStrAbore},
		{extraStrBefore},
		{extraColonWithSpacesBefore},
	}

	for _, tc := range testCases {
		filteredStr := filterFinalString(tc.inputStr, keywork)
		require.Equal(t, filteredStr, correctStr)
	}
}
