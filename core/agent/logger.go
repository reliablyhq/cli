package agent

// Logger - interface used to fascilitate logging
// within the agent
type Logger interface {
	Infof(format string, args ...interface{})
	Info(args ...interface{})

	Debugf(format string, args ...interface{})
	Debug(args ...interface{})

	Warnf(format string, args ...interface{})
	Warn(args ...interface{})

	Errorf(format string, args ...interface{})
	Error(args ...interface{})
}
