package callbacks

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilterFinalString(t *testing.T) {
	t.Parallel()

	keywork := "Final Answer:"
	correctStr := "This is a correct final string."

	mockchunkStr1 := fmt.Sprintf(` some other text.
	Final Answer: %s`, correctStr)

	mockchunkStr2 := fmt.Sprintf(`   :    %s`, correctStr)

	filteredStr := filterFinalString(correctStr, keywork)

	filteredStr1 := filterFinalString(mockchunkStr1, keywork)
	filteredStr2 := filterFinalString(mockchunkStr2, keywork)

	require.Equal(t, filteredStr, correctStr)
	require.Equal(t, filteredStr1, correctStr)
	require.Equal(t, filteredStr2, correctStr)
}
