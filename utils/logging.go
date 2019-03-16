package utils

import (
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/grpclog"
)

var _ grpclog.LoggerV2 = (*logrusGRPCV2)(nil)

type logrusGRPCV2 struct {
	logrus.FieldLogger
}

func (l2 *logrusGRPCV2) V(l int) bool {
	// I have no idea what's going on here
	return false
}

// LogrusToGRPCLogger wraps logrus into GRPC logger
func LogrusToGRPCLogger(src logrus.FieldLogger) grpclog.LoggerV2 {
	return &logrusGRPCV2{FieldLogger: src}
}
