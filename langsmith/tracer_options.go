package langsmith

type LangChainTracerOption interface {
	apply(t *Tracer)
}

type langChainTracerOptionFunc func(t *Tracer)

func (f langChainTracerOptionFunc) apply(t *Tracer) {
	f(t)
}

func WithClient(client *Client) LangChainTracerOption {
	return langChainTracerOptionFunc(func(t *Tracer) {
		t.client = client
	})
}

func WithLogger(logger LeveledLogger) LangChainTracerOption {
	return langChainTracerOptionFunc(func(t *Tracer) {
		t.logger = logger
	})
}

func WithName(name string) LangChainTracerOption {
	return langChainTracerOptionFunc(func(t *Tracer) {
		t.name = name
	})
}

func WithProjectName(projectName string) LangChainTracerOption {
	return langChainTracerOptionFunc(func(t *Tracer) {
		t.projectName = projectName
	})
}

func WithRunID(runID string) LangChainTracerOption {
	return langChainTracerOptionFunc(func(t *Tracer) {
		t.runID = runID
	})
}

func WithExtras(extras KVMap) LangChainTracerOption {
	return langChainTracerOptionFunc(func(t *Tracer) {
		t.extras = extras
	})
}
