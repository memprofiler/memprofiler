package utils

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"google.golang.org/grpc/grpclog"

	"github.com/memprofiler/memprofiler/server/config"
)

var _ grpclog.LoggerV2 = (*zeroLogGRPCV2)(nil)

type zeroLogGRPCV2 struct {
	zeroLog *zerolog.Logger
}

func (l2 *zeroLogGRPCV2) Info(args ...interface{}) {
	l2.zeroLog.Info().Msg(fmt.Sprint(args...))
}

func (l2 *zeroLogGRPCV2) Infoln(args ...interface{}) {
	l2.zeroLog.Info().Msg(fmt.Sprint(args...))
}

func (l2 *zeroLogGRPCV2) Infof(format string, args ...interface{}) {
	l2.zeroLog.Info().Msg(fmt.Sprintf(format, args...))
}

func (l2 *zeroLogGRPCV2) Warning(args ...interface{}) {
	l2.zeroLog.Warn().Msg(fmt.Sprint(args...))
}

func (l2 *zeroLogGRPCV2) Warningln(args ...interface{}) {
	l2.zeroLog.Warn().Msg(fmt.Sprint(args...))
}

func (l2 *zeroLogGRPCV2) Warningf(format string, args ...interface{}) {
	l2.zeroLog.Warn().Msg(fmt.Sprintf(format, args...))
}

func (l2 *zeroLogGRPCV2) Error(args ...interface{}) {
	l2.zeroLog.Error().Msg(fmt.Sprint(args...))
}

func (l2 *zeroLogGRPCV2) Errorln(args ...interface{}) {
	l2.zeroLog.Error().Msg(fmt.Sprint(args...))
}

func (l2 *zeroLogGRPCV2) Errorf(format string, args ...interface{}) {
	l2.zeroLog.Error().Msg(fmt.Sprintf(format, args...))
}

func (l2 *zeroLogGRPCV2) Fatal(args ...interface{}) {
	l2.zeroLog.Fatal().Msg(fmt.Sprint(args...))
}

func (l2 *zeroLogGRPCV2) Fatalln(args ...interface{}) {
	l2.zeroLog.Fatal().Msg(fmt.Sprint(args...))
}

func (l2 *zeroLogGRPCV2) Fatalf(format string, args ...interface{}) {
	l2.zeroLog.Fatal().Msg(fmt.Sprintf(format, args...))
}

func (l2 *zeroLogGRPCV2) V(l int) bool {
	// I have no idea what's going on here
	return false
}

// ZeroLogToGRPCLogger wraps zerolog into GRPC logger
func ZeroLogToGRPCLogger(src *zerolog.Logger) grpclog.LoggerV2 {
	return &zeroLogGRPCV2{zeroLog: src}
}

func NewLogger(cfg *config.LoggingConfig) *zerolog.Logger {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, NoColor: false}).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(cfg.Level)

	return &logger
}
