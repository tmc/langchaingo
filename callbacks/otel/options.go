package otelcallback

import "go.opentelemetry.io/otel/trace"

// Option is a functional option for configuring the OpenTelemetry Handler.
type Option func(*config)

type config struct {
	tracerProvider trace.TracerProvider
	captureInput   bool
	captureOutput  bool
}

// WithTracerProvider sets a custom TracerProvider.
// If not called, the global OpenTelemetry TracerProvider is used.
func WithTracerProvider(tp trace.TracerProvider) Option {
	return func(c *config) {
		c.tracerProvider = tp
	}
}

// WithCaptureInput enables capturing prompt/input content as span attributes.
// WARNING: enabling this may log sensitive user data into your trace backend.
// Default: false.
func WithCaptureInput(capture bool) Option {
	return func(c *config) {
		c.captureInput = capture
	}
}

// WithCaptureOutput enables capturing completion/output content as span attributes.
// WARNING: enabling this may log sensitive model output into your trace backend.
// Default: false.
func WithCaptureOutput(capture bool) Option {
	return func(c *config) {
		c.captureOutput = capture
	}
}
