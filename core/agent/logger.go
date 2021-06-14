package agent

import log "github.com/sirupsen/logrus"

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

var logger Logger

func init() {
	logger = log.StandardLogger()
}

// Logger - set Logger for the agent
func SetLogger(l Logger) {
	logger = l
}
