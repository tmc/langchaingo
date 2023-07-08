package zapier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateDescription(t *testing.T) {
	t.Parallel()

	tool, err := New(ToolOptions{
		Name:     "Test Tool",
		ActionID: "test1234",
		Params: map[string]string{
			"Param1": "Param1 Description",
			"Param2": "Param2 Description",
		},
	})
	assert.NoError(t, err)

	desc := tool.Description()
	assert.Contains(t, desc, "Test Tool")
	assert.Contains(t, desc, "Param2")
	assert.Contains(t, desc, "Param1")
}
