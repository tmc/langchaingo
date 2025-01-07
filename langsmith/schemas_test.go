package langsmith

import (
	"encoding/json"
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

func TestRunCreate_AsJson(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		data *RunCreate
		want string
	}{
		{
			"base",
			&RunCreate{},
			`{}`,
		},
		{
			"start time",
			&RunCreate{
				BaseRun: BaseRun{
					StartTime: timeToMillisecondsPtr(time.UnixMilli(1234567890)),
				},
			},
			`{"start_time":1234567890}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, err := json.Marshal(tt.data)
			require.NoError(t, err)

			assert.Equal(t, tt.want, string(out))
		})
	}
}
