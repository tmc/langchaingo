package langsmith

type LeveledLoggerInterface interface {
	Debugf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
}

var _ LeveledLoggerInterface = &NopLogger{}

type NopLogger struct{}

func (n *NopLogger) Debugf(format string, v ...interface{}) {
	return
}

func (n *NopLogger) Errorf(format string, v ...interface{}) {
	return
}

func (n *NopLogger) Infof(format string, v ...interface{}) {
	return
}

func (n *NopLogger) Warnf(format string, v ...interface{}) {
	return
}
