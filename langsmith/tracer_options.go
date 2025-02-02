package langsmith

type LangChainTracerOption interface {
	apply(t *Tracer)
}

type langChainTracerOptionFunc func(t *Tracer)

func (f langChainTracerOptionFunc) apply(t *Tracer) {
	f(t)
}

// WithClient sets the client to use for the tracer.
func WithClient(client *Client) LangChainTracerOption {
	return langChainTracerOptionFunc(func(t *Tracer) {
		t.client = client
	})
}

// WithLogger sets the logger to use for the tracer.
func WithLogger(logger LeveledLogger) LangChainTracerOption {
	return langChainTracerOptionFunc(func(t *Tracer) {
		t.logger = logger
	})
}

// WithProjectName sets the project name to use for the tracer.
func WithName(name string) LangChainTracerOption {
	return langChainTracerOptionFunc(func(t *Tracer) {
		t.name = name
	})
}

// WithProjectName sets the project name to use for the tracer.
func WithProjectName(projectName string) LangChainTracerOption {
	return langChainTracerOptionFunc(func(t *Tracer) {
		t.projectName = projectName
	})
}

// WithRunID sets the run ID to use for the tracer.
func WithRunID(runID string) LangChainTracerOption {
	return langChainTracerOptionFunc(func(t *Tracer) {
		t.runID = runID
	})
}

// WithExtras sets the extras instrumented in tracer.
func WithExtras(extras KVMap) LangChainTracerOption {
	return langChainTracerOptionFunc(func(t *Tracer) {
		t.extras = extras
	})
}
