package client

import (
	"github.com/sirupsen/logrus"
)

// Logger that is used within memprofiler client package
type Logger interface {
	Debug(msg string)
	Warning(msg string)
	Error(msg string)
}

type logrusWrapper struct {
	logger logrus.FieldLogger
}

func (w logrusWrapper) Debug(msg string)   { w.logger.Debug(msg) }
func (w logrusWrapper) Warning(msg string) { w.logger.Warning(msg) }
func (w logrusWrapper) Error(msg string)   { w.logger.Error(msg) }

// LoggerFromLogrus wraps logrus' logger instance
func LoggerFromLogrus(src logrus.FieldLogger) Logger {
	return &logrusWrapper{logger: src}
}
