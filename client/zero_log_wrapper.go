package client

import "github.com/rs/zerolog"

type zeroLogWrapper struct {
	logger *zerolog.Logger
}

func (w zeroLogWrapper) Debug(msg string)   { w.logger.Debug().Msg(msg) }
func (w zeroLogWrapper) Warning(msg string) { w.logger.Warn().Msg(msg) }
func (w zeroLogWrapper) Error(msg string)   { w.logger.Error().Msg(msg) }

// LoggerFromZeroLog wraps zerologs' logger instance
func LoggerFromZeroLog(src *zerolog.Logger) Logger {
	subLogger := src.With().Fields(map[string]interface{}{
		"side": "profiler",
	}).Logger()
	return &zeroLogWrapper{logger: &subLogger}
}
