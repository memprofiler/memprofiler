package memprofiler

import (
	"crypto/md5"
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/vitalyisaev2/memprofiler/schema"
)

// updateMemoryUsage updates MemoryUsage fields with values obtained from runtime
func updateMemoryUsage(mu *schema.MemoryUsage, r *runtime.MemProfileRecord) {
	mu.InUseObjects += r.InUseObjects()
	mu.InUseBytes += r.InUseBytes()
	mu.AllocObjects += r.AllocObjects
	mu.AllocBytes += r.AllocBytes
}

// fillStack uses raw data to populate stack
func fillStack(cs *schema.CallStack, rawStack []uintptr, allFrames bool) {
	var (
		show   = allFrames
		frames = runtime.CallersFrames(rawStack)
	)
	for {
		frame, more := frames.Next()
		name := frame.Function
		if name == "" {
			show = true
		} else if name != "runtime.goexit" && (show || !strings.HasPrefix(name, "runtime.")) {
			show = true
			frame := &schema.StackFrame{Name: name, File: frame.File, Line: int32(frame.Line)}
			cs.Frames = append(cs.Frames, frame)
		}
		if !more {
			break
		}
	}

	if !show {
		fillStack(cs, rawStack, true)
	}
}

// hashStack computes a hash value for a stack (useful for stack comparison etc.)
func hashStack(cs *schema.CallStack) (string, error) {
	h := md5.New()

	for _, sf := range cs.Frames {
		if _, err := io.WriteString(h, dumpStackFrame(sf)); err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func dumpStackFrame(sf *schema.StackFrame) string {
	return fmt.Sprintf("%s:%s:%d", sf.GetName(), sf.GetFile(), sf.GetLine())
}
