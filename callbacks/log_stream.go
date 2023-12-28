//nolint:forbidigo
package callbacks

import (
	"context"
	"fmt"
)

// StreamLogHandler is a callback handler that prints to the standard output streaming.
type StreamLogHandler struct {
	SimpleHandler
}

var _ Handler = StreamLogHandler{}

func (StreamLogHandler) HandleStreamingFunc(_ context.Context, chunk []byte) {
	fmt.Println(string(chunk))
}
