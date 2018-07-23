package utils

import (
	"runtime"
	"strings"
)

// Caller returns name of the calling function
func Caller(skip int) string {
	counter, _, _, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	f := runtime.FuncForPC(counter)
	return f.Name()[strings.LastIndexByte(f.Name(), '.')+1:]
}
