package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/vitalyisaev2/memprofiler/schema"
)

// UpdateMemoryUsage updates MemoryUsage fields with values obtained from runtime
func UpdateMemoryUsage(mu *schema.MemoryUsage, r *runtime.MemProfileRecord) {
	mu.AllocObjects += r.AllocObjects
	mu.AllocBytes += r.AllocBytes
	mu.FreeObjects += r.FreeObjects
	mu.FreeBytes += r.FreeBytes
}

// FillCallStack uses raw data to populate stack
func FillCallStack(cs *schema.CallStack, rawStack []uintptr, allFrames bool) {
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
		FillCallStack(cs, rawStack, true)
	}
}

// HashCallStack computes a hash value for a stack (useful for stack comparison etc.)
func HashCallStack(cs *schema.CallStack) (string, error) {
	h := md5.New()

	for _, sf := range cs.Frames {
		if _, err := io.WriteString(h, DumpStackFrame(sf)); err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// DumpStackFrame makes some string representation of a stack frame
func DumpStackFrame(sf *schema.StackFrame) string {
	return fmt.Sprintf("%s:%s:%d", sf.GetName(), sf.GetFile(), sf.GetLine())
}
