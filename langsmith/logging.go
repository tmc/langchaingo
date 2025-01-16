package langsmith

type LeveledLogger interface {
	Debugf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
}

var _ LeveledLogger = &NopLogger{}

type NopLogger struct{}

func (n *NopLogger) Debugf(_ string, _ ...interface{}) {
}

func (n *NopLogger) Errorf(_ string, _ ...interface{}) {
}

func (n *NopLogger) Infof(_ string, _ ...interface{}) {
}

func (n *NopLogger) Warnf(_ string, _ ...interface{}) {
}
