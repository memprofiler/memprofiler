package tsdb

import (
	"fmt"

	goKitLog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/rs/zerolog"
)

var _ goKitLog.Logger = (*goKitLogWrapper)(nil)

type goKitLogWrapper struct {
	zeroLog *zerolog.Logger
}

// Log write log
func (l *goKitLogWrapper) Log(keyValues ...interface{}) error {
	// goKitLog.Logger interface contains the Log method,
	// which has a list of keys and values as arguments,
	// therefore it is necessary to check the number of arguments,
	// if the number is even then we take the values in pairs (key, value),
	// if not even add the key UNKNOWN
	var (
		currentLevel = ""

		zeroFields = make(map[string]interface{}, len(keyValues))

		namedCount = len(keyValues) / 2 * 2
		hasUnknown = len(keyValues)%2 > 0
	)

	for i := 0; i < namedCount; i += 2 {
		key := fmt.Sprintf("%v", keyValues[i])
		val := keyValues[i+1]

		switch key {
		case "level":
			currentLevel = val.(level.Value).String()
			continue
		case "ts":
			continue
		}

		zeroFields[key] = val
	}

	if hasUnknown {
		zeroFields["UNKNOWN"] = keyValues[len(keyValues)-1]
	}

	switch currentLevel {
	case level.WarnValue().String():
		l.zeroLog.Warn().Fields(zeroFields).Send()
	case level.ErrorValue().String():
		l.zeroLog.Error().Fields(zeroFields).Send()
	case level.DebugValue().String():
		l.zeroLog.Debug().Fields(zeroFields).Send()
	case level.InfoValue().String():
		l.zeroLog.Info().Fields(zeroFields).Send()
	default:
		l.zeroLog.Error().Fields(zeroFields).Msg("Found unexpected fields")
	}

	return nil
}

// NewGoKitLogWrapper creates zeroLog logger with goKitLog.Logger interface
func NewGoKitLogWrapper(logger *zerolog.Logger) (goKitLog.Logger, error) {
	subLogger := logger.With().Fields(map[string]interface{}{
		"side": "prometheus",
	}).Logger()

	return &goKitLogWrapper{
		zeroLog: &subLogger,
	}, nil
}
