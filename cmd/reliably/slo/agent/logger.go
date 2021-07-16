package agent

import (
	"fmt"
	"time"
)

// AgentUILogger - implements agent.Logger interface
type AgentUILogger struct {
	InfoChan  chan interface{}
	DebugChan chan interface{}
	WarnChan  chan interface{}
	ErrorChan chan interface{}
}

func (a *AgentUILogger) prefix() string {
	t := time.Now()
	return fmt.Sprintf("%s - ", t.Format("2006/01/02 15:04:05"))
}

func (a *AgentUILogger) Infof(format string, args ...interface{}) {
	a.InfoChan <- a.prefix() + fmt.Sprintf(format, args...)
}

func (a AgentUILogger) Info(args ...interface{}) {
	a.InfoChan <- a.prefix() + fmt.Sprintln(args...)
}

func (a *AgentUILogger) Debugf(format string, args ...interface{}) {
	a.DebugChan <- a.prefix() + fmt.Sprintf(format, args...)
}

func (a AgentUILogger) Debug(args ...interface{}) {
	a.DebugChan <- a.prefix() + fmt.Sprintln(args...)
}

func (a *AgentUILogger) Warnf(format string, args ...interface{}) {
	a.WarnChan <- a.prefix() + fmt.Sprintf(format, args...)
}

func (a AgentUILogger) Warn(args ...interface{}) {
	a.WarnChan <- a.prefix() + fmt.Sprintln(args...)
}

func (a *AgentUILogger) Errorf(format string, args ...interface{}) {
	a.ErrorChan <- a.prefix() + fmt.Sprintf(format, args...)
}

func (a AgentUILogger) Error(args ...interface{}) {
	a.ErrorChan <- a.prefix() + fmt.Sprintln(args...)
}

func NewAgentLogger() *AgentUILogger {
	return &AgentUILogger{
		WarnChan:  make(chan interface{}, 100),
		DebugChan: make(chan interface{}, 100),
		InfoChan:  make(chan interface{}, 100),
		ErrorChan: make(chan interface{}, 100),
	}
}
